package main

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/hashicorp/nomad/api"
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
	"github.com/nomad-ops/nomad-ops/backend/interfaces/vaulttokenstore"
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
		set.Meta.AppName = env.GetStringEnv(ctx, logger, "POCKETBASE_APP_NAME", "Nomad-Ops")
		set.Meta.AppUrl = env.GetStringEnv(ctx, logger, "POCKETBASE_APP_URL", "http://localhost:8090")
		set.Meta.SenderName = env.GetStringEnv(ctx, logger, "POCKETBASE_SENDER_NAME", "Support")
		set.Meta.SenderAddress = env.GetStringEnv(ctx, logger, "POCKETBASE_SENDER_ADDRESS", "support@localhost.com")
		set.Meta.HideControls = env.GetStringEnv(ctx, logger, "POCKETBASE_HIDE_CONTROLS", "TRUE") == "TRUE"

		set.Smtp.Enabled = env.GetStringEnv(ctx, logger, "POCKETBASE_ENABLE_SMTP", "FALSE") == "TRUE"
		set.Smtp.Host = env.GetStringEnv(ctx, logger, "POCKETBASE_SMTP_HOST", "localhost")
		set.Smtp.Port = env.GetIntEnv(ctx, logger, "POCKETBASE_SMTP_PORT", 25)
		set.Smtp.Username = env.GetStringEnv(ctx, logger, "POCKETBASE_SMTP_USERNAME", "")
		set.Smtp.Password = env.GetStringEnv(ctx, logger, "POCKETBASE_SMTP_PASSWORD", "")
		set.Smtp.AuthMethod = env.GetStringEnv(ctx, logger, "POCKETBASE_SMTP_AUTH_METHOD", "PLAIN")
		set.Smtp.Tls = env.GetStringEnv(ctx, logger, "POCKETBASE_SMTP_TLS", "FALSE") == "TRUE"

		set.MicrosoftAuth = settings.AuthProviderConfig{
			Enabled:      env.GetStringEnv(ctx, logger, "POCKETBASE_AUTH_MICROSOFT_ENABLED", "FALSE") == "TRUE",
			ClientId:     env.GetStringEnv(ctx, logger, "POCKETBASE_AUTH_MICROSOFT_CLIENT_ID", ""),
			ClientSecret: env.GetStringEnv(ctx, logger, "POCKETBASE_AUTH_MICROSOFT_CLIENT_SECRET", ""),
			AuthUrl:      env.GetStringEnv(ctx, logger, "POCKETBASE_AUTH_MICROSOFT_AUTH_URL", ""),
			TokenUrl:     env.GetStringEnv(ctx, logger, "POCKETBASE_AUTH_MICROSOFT_TOKEN_URL", ""),
		}

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
		vaultTokenStore, err := vaulttokenstore.CreatePocketBaseStore(ctx,
			log.NewSimpleLogger(trace, "VaultTokenStore-PocketBase"),
			vaulttokenstore.PocketBaseStoreConfig{
				App: e.App,
			})

		if err != nil {
			logger.LogError(ctx, "Could not CreatePocketBaseStore for vaultTokens:%v", err)
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
				NomadToken: nomadToken,
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

		getNotifiers := func() map[string]application.Notifier {
			res := map[string]application.Notifier{}

			if slackWebhookURL := env.GetStringEnv(ctx, logger, "SLACK_WEBHOOK_URL", ""); slackWebhookURL != "" {
				slackNotifier, err := notifier.CreateSlack(ctx,
					log.NewSimpleLogger(trace, "Slack-Notifier"),
					notifier.SlackConfig{
						WebhookURL:  slackWebhookURL,
						BaseURL:     env.GetStringEnv(ctx, logger, "SLACK_BASE_URL", "localhost:3000/ui/sources/"),
						IconSuccess: env.GetStringEnv(ctx, logger, "SLACK_ICON_SUCCESS", ":check:"),
						IconError:   env.GetStringEnv(ctx, logger, "SLACK_ICON_ERROR", ":check-no:"),
						EnvInfoText: env.GetStringEnv(ctx, logger, "SLACK_ENV_INFO_TEXT", "Sent by nomad-ops (dev)"),
					})
				if err != nil {
					logger.LogError(ctx, "Could not CreateSlack:%v", err)
					os.Exit(-2)
				}
				res["slack"] = slackNotifier
			}

			if webhookURL := env.GetStringEnv(ctx, logger, "WEBHOOK_URL", ""); webhookURL != "" {
				webhookNotifier, err := notifier.CreateWebhook(ctx,
					log.NewSimpleLogger(trace, "Webhook-Notifier"),
					notifier.WebhookConfig{
						WebhookURL:          webhookURL,
						Timeout:             env.GetDurationEnv(ctx, logger, "WEBHOOK_TIMEOUT", 10*time.Second),
						Method:              env.GetStringEnv(ctx, logger, "WEBHOOK_METHOD", ""),
						Insecure:            env.GetStringEnv(ctx, logger, "WEBHOOK_INSECURE", "FALSE") == "TRUE",
						LogTemplateResults:  env.GetStringEnv(ctx, logger, "WEBHOOK_LOG_TEMPLATE_RESULTS", "FALSE") == "TRUE",
						FireOn:              strings.Split(env.GetStringEnv(ctx, logger, "WEBHOOK_FIRE_ON", "success"), ","),
						AuthHeaderName:      env.GetStringEnv(ctx, logger, "WEBHOOK_AUTH_HEADER_NAME", ""),
						AuthHeaderValue:     ReadFromFile(ctx, logger, "WEBHOOK_AUTH_HEADER_VALUE_FILE", ""),
						BodyTemplate:        ReadFromFile(ctx, logger, "WEBHOOK_BODY_TEMPLATE_FILE", ""),
						QueryParamsTemplate: ReadFromFile(ctx, logger, "WEBHOOK_QUERY_TEMPLATE_FILE", ""),
					})
				if err != nil {
					logger.LogError(ctx, "Could not CreateWebhook:%v", err)
					os.Exit(-2)
				}
				res["webhook"] = webhookNotifier
			}

			return res
		}

		notificationComposer, err := notifier.CreateComposer(ctx,
			log.NewSimpleLogger(trace, "Notification-Composer"),
			notifier.ComposerConfig{
				Notifiers: getNotifiers(),
			})
		if err != nil {
			logger.LogError(ctx, "Could not CreateComposer:%v", err)
			os.Exit(-2)
		}

		watcher, err := application.CreateRepoWatcher(ctx,
			log.NewSimpleLogger(trace, "RepoWatcher"),
			application.RepoWatcherConfig{
				Interval:        env.GetDurationEnv(ctx, logger, "NOMAD_OPS_POLLING_INTERVAL", 60*time.Second),
				ErrorRetryCount: env.GetIntEnv(ctx, logger, "NOMAD_OPS_ERROR_RETRY_COUNT", 2),
				AppName:         env.GetStringEnv(ctx, logger, "APP_NAME", "nomad-ops"),
			},
			srcStore,
			dsw,
			notificationComposer,
			vaultTokenStore)
		if err != nil {
			logger.LogError(ctx, "Could not CreateRepoWatcher:%v", err)
			os.Exit(-2)
		}

		err = nomadAPI.SubscribeJobChanges(ctx, func(jobName string) {
			err := watcher.SyncSourceByID(ctx, jobName, application.SyncSourceOptions{})
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
			notificationComposer)
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
			if e.Collection.Name == "sources" {
				// Update watch
				err := watcher.UpdateSource(e.HttpContext.Request().Context(), domain.SourceFromRecord(e.Record, true))
				if err != nil {
					logger.LogError(ctx, "Could not UpdateSource:%v", err)
					return err
				}
				logger.LogInfo(ctx, "updated source")
			}

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
				err := watcher.SyncSourceByID(c.Request().Context(), id, application.SyncSourceOptions{
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
				func(next echo.HandlerFunc) echo.HandlerFunc {
					return func(c echo.Context) error {
						authRecord, _ := c.Get(apis.ContextAuthRecordKey).(*models.Record)
						if authRecord == nil {
							return apis.NewForbiddenError("Only auth records can access this endpoint", nil)
						}
						id := c.QueryParam("id")
						if id == "" {
							return c.JSON(http.StatusBadRequest, domain.Error{
								Message: log.ToStrPtr("Expected a valid 'id' parameter"),
							})
						}

						return next(c)
					}
				},
				apis.RequireAdminOrRecordAuth("users"),
				apis.ActivityLogger(e.App),
				middleware.CORSWithConfig(middleware.CORSConfig{}),
				middleware.BodyLimitWithConfig(middleware.BodyLimitConfig{}),
				middleware.Recover(),
				middleware.LoggerWithConfig(middleware.LoggerConfig{}),
			},
		})

		e.Router.AddRoute(echo.Route{
			Method: http.MethodGet, // Read only, but still a user might see too much
			Path:   "/api/nomad/proxy/*",
			Handler: func(c echo.Context) error {

				var params map[string]string

				for k, v := range c.QueryParams() {
					if len(v) == 0 {
						continue
					}
					if params == nil {
						params = map[string]string{}
					}
					params[k] = v[0]
				}

				resp, err := nomadAPI.ProxyHandler(c.Request().Context(),
					strings.TrimPrefix(c.Request().URL.EscapedPath(), "/api/nomad/proxy"),
					api.QueryOptions{
						Params: params,
					})

				if err != nil {
					logger.LogError(c.Request().Context(), "Could not handle Nomad Proxy Request:%v", err)
					return c.JSON(http.StatusInternalServerError, domain.Error{
						Message: log.ToStrPtr("Unexpected error"),
					})
				}
				defer resp.Close()

				return c.Stream(http.StatusOK, "application/json", resp)
			},
			Middlewares: []echo.MiddlewareFunc{
				apis.RequireAdminOrRecordAuth("users"),
				middleware.CORSWithConfig(middleware.CORSConfig{}),
				middleware.BodyLimitWithConfig(middleware.BodyLimitConfig{}),
				middleware.Recover(),
				middleware.LoggerWithConfig(middleware.LoggerConfig{}),
			},
		})

		e.Router.AddRoute(echo.Route{
			Method: http.MethodGet,
			Path:   "/api/nomad/urls",
			Handler: func(c echo.Context) error {

				u, err := nomadAPI.GetURL(c.Request().Context())
				if err != nil {
					return c.JSONPretty(http.StatusInternalServerError, domain.Error{
						Message: log.ToStrPtr("Unexpected error"),
					}, "    ")
				}

				return c.JSONPretty(http.StatusOK, map[string]string{
					"ui": u,
				}, "    ")
			},
			Middlewares: []echo.MiddlewareFunc{
				apis.RequireAdminOrRecordAuth("users"),
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

func ReadFromFile(ctx context.Context, logger log.Logger, key string, def string) string {
	fp := env.GetStringEnv(ctx, logger, key, "")
	if fp == "" {
		return def
	}
	b, err := os.ReadFile(fp)
	if err != nil {
		logger.LogError(ctx, "Could not read file %s for key %s", fp, key)
		return def
	}
	return string(b)
}
