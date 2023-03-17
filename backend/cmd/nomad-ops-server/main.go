package main

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/models"
	"github.com/pocketbase/pocketbase/models/settings"

	"github.com/nomad-ops/nomad-ops/backend/application"
	"github.com/nomad-ops/nomad-ops/backend/domain"
	"github.com/nomad-ops/nomad-ops/backend/interfaces/eventstore"
	"github.com/nomad-ops/nomad-ops/backend/interfaces/github"
	"github.com/nomad-ops/nomad-ops/backend/interfaces/keystore"
	"github.com/nomad-ops/nomad-ops/backend/interfaces/nomadcluster"
	"github.com/nomad-ops/nomad-ops/backend/interfaces/notifier"
	"github.com/nomad-ops/nomad-ops/backend/interfaces/sourcestore"
	"github.com/nomad-ops/nomad-ops/backend/utils/env"
	"github.com/nomad-ops/nomad-ops/backend/utils/errors"
	"github.com/nomad-ops/nomad-ops/backend/utils/log"
)

//go:embed wwwroot/**
var public embed.FS

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	trace := os.Getenv("TRACE") == "TRUE"

	logger := log.NewSimpleLogger(trace, "Main")

	app := pocketbase.New()
	logger.LogInfo(ctx, "Start")

	app.OnRecordBeforeCreateRequest().Add(func(e *core.RecordCreateEvent) error {

		if e.Collection.Name == "sources" {
			e.Record.Set("status", &domain.SourceStatus{
				Status:  domain.SourceStatusStatusInit,
				Message: "Pending...",
			})
		}
		return nil
	})

	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {

		e.Router.Use( /* TODO */ )

		set := settings.New()
		set.Meta.AppName = "Nomad-Ops"

		err := e.App.Dao().SaveSettings(set)
		if err != nil {
			return err
		}

		adminCount, err := e.App.Dao().TotalAdmins()
		if err != nil {
			return err
		}

		if adminCount == 0 {

			defaultAdminUserEmail := env.GetStringEnv(ctx, logger, "DEFAULT_ADMIN_EMAIL", "admin@nomad-ops.org")
			defaultAdminUserPassword := env.GetStringEnv(ctx, logger, "DEFAULT_ADMIN_PASSWORD", "simple-nomad-ops")

			if defaultAdminUserEmail == "" || defaultAdminUserPassword == "" {
				return fmt.Errorf("need a DEFAULT_ADMIN_EMAIL and DEFAULT_ADMIN_PASSWORD to initialize the application")
			}

			logger.LogInfo(ctx, "Creating initial admin user...")
			admin := &models.Admin{
				Email: defaultAdminUserEmail,
			}
			err = admin.SetPassword(defaultAdminUserPassword)
			if err != nil {
				return err
			}
			err = e.App.Dao().SaveAdmin(admin)
			if err != nil {
				return err
			}
		}

		logger.LogInfo(ctx, "Initializing models...")

		err = domain.InitModels(ctx, logger, e.App)

		if err != nil {
			logger.LogError(ctx, "Could not init models:%v", err)
			return err
		}
		logger.LogInfo(ctx, "Initializing models...Done")

		// Start application and get all sources and start watch
		evStore, err := eventstore.CreatePocketBaseStore(ctx,
			log.NewSimpleLogger(trace, "EventStore-PocketBase"),
			eventstore.PocketBaseStoreConfig{
				App: e.App,
			})

		if err != nil {
			logger.LogError(ctx, "Could not CreatePocketBaseStore for events:%v", err)
			return err
		}
		srcStore, err := sourcestore.CreatePocketBaseStore(ctx,
			log.NewSimpleLogger(trace, "SourceStore-PocketBase"),
			sourcestore.PocketBaseStoreConfig{
				App: e.App,
			})

		if err != nil {
			logger.LogError(ctx, "Could not CreatePocketBaseStore for sources:%v", err)
			return err
		}
		keyStore, err := keystore.CreatePocketBaseStore(ctx,
			log.NewSimpleLogger(trace, "KeyStore-PocketBase"),
			keystore.PocketBaseStoreConfig{
				App: e.App,
			})

		if err != nil {
			logger.LogError(ctx, "Could not CreatePocketBaseStore for keys:%v", err)
			return err
		}

		nomadToken := ""
		if tokenPath := env.GetStringEnv(ctx, logger, "NOMAD_TOKEN_FILE", ""); tokenPath != "" {
			logger.LogInfo(ctx, "Using NOMAD_TOKEN_FILE...")
			b, err := os.ReadFile(tokenPath)
			if err != nil {
				logger.LogError(ctx, "Could not read NOMAD_TOKEN_FILE:%v", err)
				os.Exit(-2)
			}
			nomadToken = string(b)
		}

		nomadAPI, err := nomadcluster.CreateClient(ctx,
			log.NewSimpleLogger(trace, "NomadClient"),
			nomadcluster.ClientConfig{
				NomadToken:       nomadToken,
				DefaultNamespace: env.GetStringEnv(ctx, logger, "NOMAD_DEFAULT_NAMESPACE", "nomad-ops"),
				DefaultRegion:    env.GetStringEnv(ctx, logger, "NOMAD_DEFAULT_REGION", ""),
			})
		if err != nil {
			logger.LogError(ctx, "Could not CreateNomadClient:%v", err)
			os.Exit(-2)
		}

		dsw, err := github.CreateGitProvider(ctx,
			log.NewSimpleLogger(trace, "GitProvider"),
			github.GitProviderConfig{
				ReposDir: env.GetStringEnv(ctx, logger, "NOMAD_OPS_LOCAL_REPO_DIR", "repos"),
			},
			nomadAPI,
			keyStore)
		if err != nil {
			logger.LogError(ctx, "Could not CreateGitProvider:%v", err)
			os.Exit(-2)
		}

		slackNotifier, err := notifier.CreateSlack(ctx,
			log.NewSimpleLogger(trace, "Slack-Notifier"),
			notifier.SlackConfig{
				WebhookURL:  env.GetStringEnv(ctx, logger, "SLACK_WEBHOOK_URL", ""),
				BaseURL:     env.GetStringEnv(ctx, logger, "SLACK_BASE_URL", "localhost:3000/ui/sources/"),
				IconSuccess: env.GetStringEnv(ctx, logger, "SLACK_ICON_SUCCESS", ":check:"),
				IconError:   env.GetStringEnv(ctx, logger, "SLACK_ICON_ERROR", ":check-no:"),
				EnvInfoText: env.GetStringEnv(ctx, logger, "SLACK_ENV_INFO_TEXT", "Sent by nomad-ops (dev)"),
			})
		if err != nil {
			logger.LogError(ctx, "Could not CreateSlack:%v", err)
			os.Exit(-2)
		}

		watcher, err := application.CreateRepoWatcher(ctx,
			log.NewSimpleLogger(trace, "RepoWatcher"),
			application.RepoWatcherConfig{
				Interval: env.GetDurationEnv(ctx, logger, "NOMAD_OPS_POLLING_INTERVAL", 60*time.Second),
			},
			srcStore,
			dsw,
			slackNotifier)
		if err != nil {
			logger.LogError(ctx, "Could not CreateRepoWatcher:%v", err)
			os.Exit(-2)
		}

		err = nomadAPI.SubscribeJobChanges(ctx, func(jobName string) {
			err := watcher.UpdateSourceByID(ctx, jobName, application.UpdateSourceOptions{})
			if err == errors.ErrNotFound {
				// not handled by us --- ignore
				return
			}
			if err != nil {
				logger.LogError(ctx, "Could not UpdateSourceByID on Nomad Event:%v", err)
			}
		})
		if err != nil {
			logger.LogError(ctx, "Could not SubscribeJobChanges:%v", err)
			os.Exit(-2)
		}

		manager, err := application.CreateReconciliationManager(ctx,
			log.NewSimpleLogger(trace, "ReconciliationManager"),
			application.ReconciliationManagerConfig{},
			srcStore,
			watcher,
			nomadAPI,
			evStore,
			slackNotifier)
		if err != nil {
			logger.LogError(ctx, "Could not CreateReconciliationManager:%v", err)
			os.Exit(-2)
		}

		app.OnRecordAfterCreateRequest().Add(func(e *core.RecordCreateEvent) error {
			if e.Collection.Name == "sources" {
				logger.LogInfo(ctx, "Adding new source to watch...")
				err := manager.OnAddedSource(e.HttpContext.Request().Context(), domain.SourceFromRecord(e.Record, false))
				if err != nil {
					logger.LogError(ctx, "Could not handle added source:%v", err)
					return err
				}
			}

			return nil
		})

		app.OnModelAfterUpdate().Add(func(e *core.ModelEvent) error {
			if logger.IsTraceEnabled(ctx) {
				logger.LogTrace(ctx, "After Model Update...%s", log.ToJSONString(e.Model))
			}
			return nil
		})

		app.OnRecordAfterUpdateRequest().Add(func(e *core.RecordUpdateEvent) error {
			// if e.Collection.Name == "sources" {
			// Update watch
			// }

			return nil
		})

		app.OnRecordAfterDeleteRequest().Add(func(e *core.RecordDeleteEvent) error {
			if e.Collection.Name == "sources" {
				logger.LogInfo(ctx, "Removing source from watch...")
				err := manager.OnDeletedSource(e.HttpContext.Request().Context(), e.Record.Id)
				if err != nil {
					logger.LogError(ctx, "Could not handle deleted source:%v", err)
					return err
				}
			}

			return nil
		})

		app.OnRecordsListRequest().Add(func(e *core.RecordsListEvent) error {

			return nil
		})

		wwwroot, err := fs.Sub(public, "wwwroot")
		if err != nil {
			return err
		}

		e.Router.Add("GET", "/*", apis.StaticDirectoryHandler(wwwroot, true))

		// add new "POST /api/actions/sources/sync" route
		e.Router.AddRoute(echo.Route{
			Method: http.MethodPost,
			Path:   "/api/actions/sources/sync",
			Handler: func(c echo.Context) error {
				id := c.QueryParam("id")
				if id == "" {
					return c.JSON(http.StatusBadRequest, domain.Error{
						Message: log.ToStrPtr("Expected a valid 'id' parameter"),
					})
				}

				logger.LogInfo(c.Request().Context(), "Syncing source %s...", id)
				err := watcher.UpdateSourceByID(c.Request().Context(), id, application.UpdateSourceOptions{
					ForceRestart: false,
				})

				if err == errors.ErrNotFound {
					return c.JSON(http.StatusNotFound, domain.Error{
						Message: log.ToStrPtr("Source was not found"),
					})
				}

				if err != nil {
					logger.LogError(c.Request().Context(), "Could not UpdateSourceByID:%v", err)
					return c.JSON(http.StatusInternalServerError, domain.Error{
						Message: log.ToStrPtr("Unexpected error"),
					})
				}

				return c.JSON(http.StatusOK, map[string]string{}) // empty 200 OK response
			},
			Middlewares: []echo.MiddlewareFunc{
				apis.RequireAdminOrRecordAuth("users"),
				apis.ActivityLogger(e.App),
				middleware.CORSWithConfig(middleware.CORSConfig{}),
				middleware.BodyLimitWithConfig(middleware.BodyLimitConfig{}),
				middleware.Recover(),
				middleware.LoggerWithConfig(middleware.LoggerConfig{}),
			},
		})

		// add new "POST /api/actions/sources/pause" route
		e.Router.AddRoute(echo.Route{
			Method: http.MethodPost,
			Path:   "/api/actions/sources/pause",
			Handler: func(c echo.Context) error {
				id := c.QueryParam("id")
				if id == "" {
					return c.JSON(http.StatusBadRequest, domain.Error{
						Message: log.ToStrPtr("Expected a valid 'id' parameter"),
					})
				}

				pause := c.QueryParam("pause") == "true"

				if pause {
					logger.LogInfo(c.Request().Context(), "Pausing watch on source %s...", id)
				} else {
					logger.LogInfo(c.Request().Context(), "Resuming watch on source %s...", id)
				}
				err := watcher.PauseSourceByID(c.Request().Context(), id, application.PauseOptions{
					Pause: pause,
				})

				if err == errors.ErrNotFound {
					return c.JSON(http.StatusNotFound, domain.Error{
						Message: log.ToStrPtr("Source was not found"),
					})
				}

				if err != nil {
					logger.LogError(c.Request().Context(), "Could not PauseSourceByID:%v", err)
					return c.JSON(http.StatusInternalServerError, domain.Error{
						Message: log.ToStrPtr("Unexpected error"),
					})
				}

				return c.JSON(http.StatusOK, map[string]string{}) // empty 200 OK response
			},
			Middlewares: []echo.MiddlewareFunc{
				apis.RequireAdminOrRecordAuth("users"),
				apis.ActivityLogger(e.App),
				middleware.CORSWithConfig(middleware.CORSConfig{}),
				middleware.BodyLimitWithConfig(middleware.BodyLimitConfig{}),
				middleware.Recover(),
				middleware.LoggerWithConfig(middleware.LoggerConfig{}),
			},
		})

		logger.LogInfo(ctx, "Initialization done")

		return nil
	})

	if err := app.Start(); err != nil {
		logger.LogError(ctx, "Could not start app:%v", err)
	}
}
