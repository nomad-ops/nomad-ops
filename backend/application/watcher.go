package application

import (
	"context"
	"fmt"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/VictoriaMetrics/metrics"
	giturls "github.com/whilp/git-urls"

	"github.com/nomad-ops/nomad-ops/backend/domain"
	"github.com/nomad-ops/nomad-ops/backend/utils/errors"
	"github.com/nomad-ops/nomad-ops/backend/utils/log"
)

type DesiredStateWatcher interface {
	FetchDesiredState(ctx context.Context, src *domain.Source) (*DesiredState, error)
}

type WatchInfo struct {
	ctx        context.Context
	cancel     context.CancelFunc
	Source     *domain.Source
	Reconciler ReconcilerFunc
	syncFunc   func(context.Context, SyncSourceOptions) error
	syncCh     chan SyncSourceOptions
	updateFunc func(context.Context, *domain.Source) error
	updateCh   chan *domain.Source
}

type RepoWatcher struct {
	ctx                 context.Context
	logger              log.Logger
	cfg                 RepoWatcherConfig
	sourceStatusPatcher SourceStatusPatcher
	dsw                 DesiredStateWatcher
	lock                sync.Mutex
	watchList           map[string]*WatchInfo
	notifier            Notifier
	vaultRepo           VaultTokenRepo
}

type RepoWatcherConfig struct {
	Interval        time.Duration
	ErrorRetryCount int
	AppName         string
}

type SourceStatusPatcher interface {
	SetSourceStatus(srcID string, s *domain.SourceStatus) error
}

func CreateRepoWatcher(ctx context.Context,
	logger log.Logger,
	cfg RepoWatcherConfig,
	sourceStatusPatcher SourceStatusPatcher,
	dsw DesiredStateWatcher,
	notifier Notifier,
	vaultRepo VaultTokenRepo) (*RepoWatcher, error) {
	t := &RepoWatcher{
		ctx:                 ctx,
		logger:              logger,
		cfg:                 cfg,
		sourceStatusPatcher: sourceStatusPatcher,
		dsw:                 dsw,
		watchList:           map[string]*WatchInfo{},
		notifier:            notifier,
		vaultRepo:           vaultRepo,
	}

	return t, nil
}

type SyncSourceOptions struct {
	ForceRestart bool
}

func (w *RepoWatcher) SyncSourceByID(ctx context.Context, id string, opts SyncSourceOptions) error {
	w.lock.Lock()
	wi, ok := w.watchList[id]
	w.lock.Unlock()
	if !ok {
		return errors.ErrNotFound
	}
	w.logger.LogInfo(ctx, "Syncing repo %s on branch %s", wi.Source.URL, wi.Source.Branch)
	err := wi.syncFunc(ctx, opts)
	if err != nil {
		return err
	}
	return nil
}

func (w *RepoWatcher) SyncSource(ctx context.Context, repo, branch string, opts SyncSourceOptions) error {
	w.logger.LogInfo(ctx, "Syncing repo %s on branch %s", repo, branch)
	w.lock.Lock()
	var wis []*WatchInfo
	for _, iwi := range w.watchList {
		if iwi.Source.Branch != branch {
			continue
		}

		w.logger.LogTrace(ctx, "Parsing URL %s", iwi.Source.URL)
		u, err := giturls.Parse(iwi.Source.URL)
		if err != nil {
			w.logger.LogError(ctx, "Could not parse url:%v - %v", iwi.Source.URL, err)
			continue
		}
		w.logger.LogTrace(ctx, "Raw Path: %s", u.RawPath)
		w.logger.LogTrace(ctx, "Path: %s", u.Path)
		if u.Path != fmt.Sprintf("%s.git", repo) {
			continue
		}
		cpy := iwi
		wis = append(wis, cpy)
	}
	w.lock.Unlock()
	if len(wis) == 0 {
		return errors.ErrNotFound
	}
	for _, wi := range wis {
		err := wi.syncFunc(ctx, opts)
		if err != nil {
			return err
		}
	}
	return nil
}

func (w *RepoWatcher) UpdateSource(ctx context.Context, src *domain.Source) error {
	w.logger.LogInfo(ctx, "Updating source %s", src.Name)
	w.lock.Lock()
	wi, ok := w.watchList[src.ID]
	w.lock.Unlock()
	if !ok {
		return errors.ErrNotFound
	}
	err := wi.updateFunc(ctx, src)
	if err != nil {
		return err
	}
	return nil
}

func (w *RepoWatcher) applyOverrides(ctx context.Context, src *domain.Source, desiredState *DesiredState) error {

	for _, v := range desiredState.Jobs {
		if src.DataCenter != "" {
			dcs := strings.Split(src.DataCenter, ",")
			v.Datacenters = dcs
		}
		if src.Namespace != "" {
			v.Namespace = &src.Namespace
		}
	}

	return nil
}

func (w *RepoWatcher) WatchSource(ctx context.Context, origSrc *domain.Source, cb ReconcilerFunc) error {
	w.lock.Lock()
	defer w.lock.Unlock()
	wi, ok := w.watchList[origSrc.ID]
	if ok {
		// already watching
		return nil
	}
	workerCtx, cancel := context.WithCancel(w.ctx)
	wi = &WatchInfo{
		ctx:        workerCtx,
		cancel:     cancel,
		Reconciler: cb,
		Source:     origSrc,
		syncCh:     make(chan SyncSourceOptions),
		syncFunc: func(ctx context.Context, opts SyncSourceOptions) error {
			select {
			case wi.syncCh <- opts:
				return nil
			case <-ctx.Done():
				return ctx.Err()
			case <-workerCtx.Done():
				return workerCtx.Err()
			}
		},
		updateCh: make(chan *domain.Source),
		updateFunc: func(ctx context.Context, src *domain.Source) error {
			select {
			case wi.updateCh <- src:
				return nil
			case <-ctx.Done():
				return ctx.Err()
			case <-workerCtx.Done():
				return workerCtx.Err()
			}
		},
	}

	err := w.sourceStatusPatcher.SetSourceStatus(wi.Source.ID, &domain.SourceStatus{
		Message: "Waiting on first sync",
		Status:  domain.SourceStatusStatusInit,
	})
	if err != nil {
		w.logger.LogError(ctx, "Could not SetSourceStatus on %s:%v", wi.Source.ID, err)
	}

	w.watchList[origSrc.ID] = wi

	go func(wi *WatchInfo) {
		defer wi.cancel()
		w.logger.LogInfo(wi.ctx, "Starting watch on %s %s - %s", wi.Source.Name, wi.Source.URL, wi.Source.Path)
		defer func() {
			w.logger.LogInfo(wi.ctx, "Stopped watch on %s %s - %s", wi.Source.Name, wi.Source.URL, wi.Source.Path)
		}()
		defer func() {
			if r := recover(); r != nil {
				w.logger.LogError(wi.ctx, "Recovered worker %s - %s:%v - %s", wi.Source.URL, wi.Source.Path, r, string(debug.Stack()))
				err := w.StopSourceWatch(ctx, wi.Source.ID)
				if err != nil {
					// we panic to force a restart
					panic(err)
				}
				err = w.WatchSource(ctx, origSrc, cb)
				if err != nil {
					// we panic to force a restart
					panic(err)
				}
			}
		}()

		//hasError := false
		errorCount := 0
		firstRun := true

		metrics.GetOrCreateCounter("nomad_ops_watched_repos_gauge" +
			fmt.Sprintf(`{app="%s"}`,
				w.cfg.AppName)).Inc()

		defer func() {
			metrics.GetOrCreateCounter("nomad_ops_watched_repos_gauge" +
				fmt.Sprintf(`{app="%s"}`,
					w.cfg.AppName)).Dec()
		}()

		waitTime := w.cfg.Interval

		for {
			select {
			case <-wi.ctx.Done():
				return
			default:
			}
			if !firstRun {
				metrics.GetOrCreateCounter("nomad_ops_reconciliations_counter" +
					fmt.Sprintf(`{app="%s",repo_url="%s",repo_branch="%s",nomad_namespace="%s",nomad_dc="%s",key_id="%s",repo_path="%s",has_error="%v"}`,
						w.cfg.AppName, wi.Source.URL, wi.Source.Branch,
						wi.Source.Namespace, wi.Source.DataCenter,
						wi.Source.DeployKeyID, wi.Source.Path, errorCount != 0)).Inc()
			}
			firstRun = false
			restart := false
			select {
			case <-time.After(waitTime):
			case opts := <-wi.syncCh:
				restart = opts.ForceRestart
			case src := <-wi.updateCh:
				w.logger.LogInfo(wi.ctx, "Updating watch on %s %s - %s", wi.Source.Name, wi.Source.URL, wi.Source.Path)
				wi.Source = src
			case <-wi.ctx.Done():
				return
			}
			wi.Source.Status.Status = domain.SourceStatusStatusSyncing
			wi.Source.Status.Message = "Syncing"

			err = w.sourceStatusPatcher.SetSourceStatus(wi.Source.ID, wi.Source.Status)
			if err != nil {
				w.logger.LogError(ctx, "Could not SetSourceStatus on %s:%v", wi.Source.ID, err)
			}

			desiredState, err := w.dsw.FetchDesiredState(wi.ctx, wi.Source)
			if err != nil {
				w.logger.LogError(wi.ctx, "Could not FetchDesiredState: %v - %v - %v", err, wi.Source.URL, wi.Source.Path)
				err = w.sourceStatusPatcher.SetSourceStatus(wi.Source.ID, &domain.SourceStatus{
					Status:        domain.SourceStatusStatusError,
					Message:       err.Error(),
					LastCheckTime: toTimePtr(time.Now()),
				})
				if err != nil {
					w.logger.LogError(ctx, "Could not SetSourceStatus on %s:%v", wi.Source.ID, err)
				}
				if errorCount == w.cfg.ErrorRetryCount {
					err = w.notifier.Notify(ctx, NotifyOptions{
						Source:  wi.Source,
						Type:    NotificationError,
						Message: "Could not fetch desired state",
						Infos: []NotifyAdditionalInfos{
							{
								Header: "Git-Url",
								Text:   wi.Source.URL,
							},
							{
								Header: "Git-Rev",
								Text:   wi.Source.Branch,
							},
							{
								Header: "Git-Repo-Path",
								Text:   wi.Source.Path,
							},
							{
								Header: "Nomad-Namespace",
								Text:   wi.Source.Namespace,
							},
							{
								Header: "Nomad-Region",
								Text:   wi.Source.Region,
							},
							{
								Header: "Nomad-DataCenter",
								Text:   wi.Source.DataCenter,
							},
							{
								Header: "Force Restart",
								Text:   fmt.Sprintf("%v", restart),
							},
							{
								Header: "Error",
								Text:   fmt.Sprintf("Could not fetch desired state:%v", err),
								Large:  true,
							},
						},
					})
					if err != nil {
						w.logger.LogError(ctx, "Could not notify:%v", err)
					}
				}
				errorCount++
				continue
			}

			if wi.Source.VaultTokenID != "" {
				t, err := w.vaultRepo.GetVaultToken(ctx, wi.Source.VaultTokenID)
				if err != nil {
					w.logger.LogError(wi.ctx, "Could not GetVaultToken: %v - %v - %v", err, wi.Source.URL, wi.Source.Path)
					err = w.sourceStatusPatcher.SetSourceStatus(wi.Source.ID, &domain.SourceStatus{
						Status:        domain.SourceStatusStatusError,
						Message:       err.Error(),
						LastCheckTime: toTimePtr(time.Now()),
					})
					if err != nil {
						w.logger.LogError(ctx, "Could not SetSourceStatus on %s:%v", wi.Source.ID, err)
					}
					if errorCount == w.cfg.ErrorRetryCount {
						err = w.notifier.Notify(ctx, NotifyOptions{
							Source:  wi.Source,
							GitInfo: desiredState.GitInfo,
							Type:    NotificationError,
							Message: "Could not GetVaultToken",
							Infos: []NotifyAdditionalInfos{
								{
									Header: "Git-Url",
									Text:   wi.Source.URL,
								},
								{
									Header: "Git-Rev",
									Text:   wi.Source.Branch,
								},
								{
									Header: "Git-Repo-Path",
									Text:   wi.Source.Path,
								},
								{
									Header: "Nomad-Namespace",
									Text:   wi.Source.Namespace,
								},
								{
									Header: "Nomad-Region",
									Text:   wi.Source.Region,
								},
								{
									Header: "Force Restart",
									Text:   fmt.Sprintf("%v", restart),
								},
								{
									Header: "Error",
									Text:   fmt.Sprintf("Could not apply overrides:%v", err),
									Large:  true,
								},
							},
						})
						if err != nil {
							w.logger.LogError(ctx, "Could not notify:%v", err)
						}
					}
					errorCount++
					continue
				}
				// using vault token
				for k := range desiredState.Jobs {
					desiredState.Jobs[k].VaultToken = &t.Value
				}
			}

			err = w.applyOverrides(wi.ctx, wi.Source, desiredState)
			if err != nil {
				w.logger.LogError(wi.ctx, "Could not apply overrides: %v - %v - %v", err, wi.Source.URL, wi.Source.Path)
				err = w.sourceStatusPatcher.SetSourceStatus(wi.Source.ID, &domain.SourceStatus{
					Status:        domain.SourceStatusStatusError,
					Message:       err.Error(),
					LastCheckTime: toTimePtr(time.Now()),
				})
				if err != nil {
					w.logger.LogError(ctx, "Could not SetSourceStatus on %s:%v", wi.Source.ID, err)
				}
				if errorCount == w.cfg.ErrorRetryCount {
					err = w.notifier.Notify(ctx, NotifyOptions{
						Source:  wi.Source,
						GitInfo: desiredState.GitInfo,
						Type:    NotificationError,
						Message: "Could not apply overrides",
						Infos: []NotifyAdditionalInfos{
							{
								Header: "Git-Url",
								Text:   wi.Source.URL,
							},
							{
								Header: "Git-Rev",
								Text:   wi.Source.Branch,
							},
							{
								Header: "Git-Repo-Path",
								Text:   wi.Source.Path,
							},
							{
								Header: "Nomad-Namespace",
								Text:   wi.Source.Namespace,
							},
							{
								Header: "Nomad-Region",
								Text:   wi.Source.Region,
							},
							{
								Header: "Force Restart",
								Text:   fmt.Sprintf("%v", restart),
							},
							{
								Header: "Error",
								Text:   fmt.Sprintf("Could not apply overrides:%v", err),
								Large:  true,
							},
						},
					})
					if err != nil {
						w.logger.LogError(ctx, "Could not notify:%v", err)
					}
				}
				errorCount++
				continue
			}

			changeInfo, err := wi.Reconciler(wi.ctx, wi.Source, desiredState, restart)
			if err != nil {
				w.logger.LogError(wi.ctx, "Could not Reconcile: %v - %v - %v", err, wi.Source.URL, wi.Source.Path)
				err = w.sourceStatusPatcher.SetSourceStatus(wi.Source.ID, &domain.SourceStatus{
					Status:        domain.SourceStatusStatusError,
					Message:       err.Error(),
					LastCheckTime: toTimePtr(time.Now()),
				})
				if err != nil {
					w.logger.LogError(ctx, "Could not SetSourceStatus on %s:%v", wi.Source.ID, err)
				}
				if errorCount == w.cfg.ErrorRetryCount {
					err = w.notifier.Notify(ctx, NotifyOptions{
						Source:  wi.Source,
						GitInfo: desiredState.GitInfo,
						Type:    NotificationError,
						Message: "Could not Reconcile",
						Infos: []NotifyAdditionalInfos{
							{
								Header: "Git-Url",
								Text:   wi.Source.URL,
							},
							{
								Header: "Git-Rev",
								Text:   wi.Source.Branch,
							},
							{
								Header: "Git-Repo-Path",
								Text:   wi.Source.Path,
							},
							{
								Header: "Nomad-Namespace",
								Text:   wi.Source.Namespace,
							},
							{
								Header: "Nomad-Region",
								Text:   wi.Source.Region,
							},
							{
								Header: "Nomad-DataCenter",
								Text:   wi.Source.DataCenter,
							},
							{
								Header: "Force Restart",
								Text:   fmt.Sprintf("%v", restart),
							},
							{
								Header: "Error",
								Text:   fmt.Sprintf("Could not Reconcile:%v", err),
								Large:  true,
							},
						},
					})
					if err != nil {
						w.logger.LogError(ctx, "Could not notify:%v", err)
					}
				}
				errorCount++
				continue
			}
			if errorCount > 0 {
				// only notify if we broke the retry threshold
				notify := errorCount >= w.cfg.ErrorRetryCount
				errorCount = 0
				if notify {
					err = w.notifier.Notify(ctx, NotifyOptions{
						Source:  wi.Source,
						GitInfo: desiredState.GitInfo,
						Type:    NotificationSuccess,
						Message: "Synced successfully",
						Infos: []NotifyAdditionalInfos{
							{
								Header: "Git-Url",
								Text:   wi.Source.URL,
							},
							{
								Header: "Git-Rev",
								Text:   wi.Source.Branch,
							},
							{
								Header: "Git-Repo-Path",
								Text:   wi.Source.Path,
							},
							{
								Header: "Nomad-Namespace",
								Text:   wi.Source.Namespace,
							},
							{
								Header: "Nomad-Region",
								Text:   wi.Source.Region,
							},
							{
								Header: "Nomad-DataCenter",
								Text:   wi.Source.DataCenter,
							},
							{
								Header: "Force Restart",
								Text:   fmt.Sprintf("%v", restart),
							},
						},
					})
					if err != nil {
						w.logger.LogError(ctx, "Could not notify:%v", err)
					}
				}
			}

			if wi.Source.Paused {
				wi.Source.Status.Status = domain.SourceStatusStatusSynced
				msg := "Still in sync"
				if len(changeInfo.Create) > 0 || len(changeInfo.Update) > 0 || len(changeInfo.Delete) > 0 {
					msg = fmt.Sprintf("Out of sync: %d to create, %d to update, %d to delete",
						len(changeInfo.Create), len(changeInfo.Update), len(changeInfo.Delete))
					wi.Source.Status.Status = domain.SourceStatusStatusOutOfSync
				}
				wi.Source.Status.Message = msg
			}

			wi.Source.Status.DetermineSyncStatus()

			err = w.sourceStatusPatcher.SetSourceStatus(wi.Source.ID, wi.Source.Status)
			if err != nil {
				w.logger.LogError(ctx, "Could not SetSourceStatus on %s:%v", wi.Source.ID, err)
			}
		}
	}(wi)

	return nil
}

func (w *RepoWatcher) StopSourceWatch(ctx context.Context, id string) error {
	w.lock.Lock()
	defer w.lock.Unlock()
	wi, ok := w.watchList[id]
	if !ok {
		// already watching
		return nil
	}
	wi.cancel()
	delete(w.watchList, id)

	return nil
}
