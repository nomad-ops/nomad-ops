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
	pauseFunc  func(context.Context, PauseOptions) error
	pauseCh    chan PauseOptions
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
}

type RepoWatcherConfig struct {
	Interval time.Duration
	AppName  string
}

type SourceStatusPatcher interface {
	SetSourceStatus(srcID string, s *domain.SourceStatus) error
}

func CreateRepoWatcher(ctx context.Context,
	logger log.Logger,
	cfg RepoWatcherConfig,
	sourceStatusPatcher SourceStatusPatcher,
	dsw DesiredStateWatcher,
	notifier Notifier) (*RepoWatcher, error) {
	t := &RepoWatcher{
		ctx:                 ctx,
		logger:              logger,
		cfg:                 cfg,
		sourceStatusPatcher: sourceStatusPatcher,
		dsw:                 dsw,
		watchList:           map[string]*WatchInfo{},
		notifier:            notifier,
	}

	return t, nil
}

type PauseOptions struct {
	Pause bool
}

type SyncSourceOptions struct {
	ForceRestart bool
}

func (w *RepoWatcher) SyncSourceByID(ctx context.Context, id string, opts SyncSourceOptions) error {
	w.lock.Lock()
	wi, ok := w.watchList[id]
	if !ok {
		return errors.ErrNotFound
	}
	w.lock.Unlock()
	w.logger.LogInfo(ctx, "Syncing repo %s on branch %s", wi.Source.URL, wi.Source.Branch)
	err := wi.syncFunc(ctx, opts)
	if err != nil {
		return err
	}
	return nil
}

func (w *RepoWatcher) PauseSourceByID(ctx context.Context, id string, opts PauseOptions) error {
	w.lock.Lock()
	wi, ok := w.watchList[id]
	if !ok {
		return errors.ErrNotFound
	}
	w.lock.Unlock()
	w.logger.LogInfo(ctx, "Pausing repo %s on branch %s: %v", wi.Source.URL, wi.Source.Branch, opts.Pause)
	err := wi.pauseFunc(ctx, opts)
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
			}
		},
		pauseCh: make(chan PauseOptions),
		pauseFunc: func(ctx context.Context, opts PauseOptions) error {
			select {
			case wi.pauseCh <- opts:
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		},
		updateCh: make(chan *domain.Source),
		updateFunc: func(ctx context.Context, src *domain.Source) error {
			select {
			case wi.updateCh <- src:
				return nil
			case <-ctx.Done():
				return ctx.Err()
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
		w.logger.LogInfo(wi.ctx, "Starting watch on %s %s - %s", wi.Source.Name, wi.Source.URL, wi.Source.Path)
		defer func() {
			w.logger.LogInfo(wi.ctx, "Stopped watch on %s %s - %s", wi.Source.Name, wi.Source.URL, wi.Source.Path)
		}()
		defer func() {
			if r := recover(); r != nil {
				w.logger.LogError(wi.ctx, "Recovered worker %s - %s:%v - %s", wi.Source.URL, wi.Source.Path, r, string(debug.Stack()))
			}
		}()

		hasError := false
		firstRun := true

		metrics.GetOrCreateCounter("nomad_ops_watched_repos_gauge" +
			fmt.Sprintf(`{app="%s"}`,
				w.cfg.AppName)).Inc()

		defer func() {
			metrics.GetOrCreateCounter("nomad_ops_watched_repos_gauge" +
				fmt.Sprintf(`{app="%s"}`,
					w.cfg.AppName)).Dec()
		}()

		paused := false

		waitTime := w.cfg.Interval

		for {
			if !firstRun {
				metrics.GetOrCreateCounter("nomad_ops_reconciliations_counter" +
					fmt.Sprintf(`{app="%s",repo_url="%s",repo_branch="%s",nomad_namespace="%s",nomad_dc="%s",key_id="%s",repo_path="%s",has_error="%v"}`,
						w.cfg.AppName, wi.Source.URL, wi.Source.Branch,
						wi.Source.Namespace, wi.Source.DataCenter,
						wi.Source.DeployKeyID, wi.Source.Path, hasError)).Inc()
			}
			firstRun = false
			restart := false
			select {
			case <-time.After(waitTime):
			case opts := <-wi.syncCh:
				restart = opts.ForceRestart
			case opts := <-wi.pauseCh:
				paused = opts.Pause
			case src := <-wi.updateCh:
				w.logger.LogInfo(wi.ctx, "Updating watch on %s %s - %s", wi.Source.Name, wi.Source.URL, wi.Source.Path)
				wi.Source = src
			case <-wi.ctx.Done():
				return
			}

			desiredState, err := w.dsw.FetchDesiredState(wi.ctx, wi.Source)
			if err != nil {
				w.logger.LogError(wi.ctx, "Could not FetchDesiredState: %v - %v - %v", err, wi.Source.URL, wi.Source.Path)
				if !hasError {
					err = w.sourceStatusPatcher.SetSourceStatus(wi.Source.ID, &domain.SourceStatus{
						Status:        domain.SourceStatusStatusError,
						Message:       err.Error(),
						LastCheckTime: toTimePtr(time.Now()),
					})
					if err != nil {
						w.logger.LogError(ctx, "Could not SetSourceStatus on %s:%v", wi.Source.ID, err)
					}
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
				hasError = true
				continue
			}

			err = w.applyOverrides(wi.ctx, wi.Source, desiredState)
			if err != nil {
				w.logger.LogError(wi.ctx, "Could not apply overrides: %v - %v - %v", err, wi.Source.URL, wi.Source.Path)
				if !hasError {
					err = w.sourceStatusPatcher.SetSourceStatus(wi.Source.ID, &domain.SourceStatus{
						Status:        domain.SourceStatusStatusError,
						Message:       err.Error(),
						LastCheckTime: toTimePtr(time.Now()),
					})
					if err != nil {
						w.logger.LogError(ctx, "Could not SetSourceStatus on %s:%v", wi.Source.ID, err)
					}
					err = w.notifier.Notify(ctx, NotifyOptions{
						Source:  wi.Source,
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
				hasError = true
				continue
			}

			status, changeInfo, err := wi.Reconciler(wi.ctx, wi.Source, desiredState, restart, paused)
			if err != nil {
				w.logger.LogError(wi.ctx, "Could not Reconcile: %v - %v - %v", err, wi.Source.URL, wi.Source.Path)
				if !hasError {
					err = w.sourceStatusPatcher.SetSourceStatus(wi.Source.ID, &domain.SourceStatus{
						Status:        domain.SourceStatusStatusError,
						Message:       err.Error(),
						LastCheckTime: toTimePtr(time.Now()),
					})
					if err != nil {
						w.logger.LogError(ctx, "Could not SetSourceStatus on %s:%v", wi.Source.ID, err)
					}
					err = w.notifier.Notify(ctx, NotifyOptions{
						Source:  wi.Source,
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
				hasError = true
				continue
			}
			if hasError {
				hasError = false
				err = w.notifier.Notify(ctx, NotifyOptions{
					Source:  wi.Source,
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

			if paused {
				status.Status = domain.SourceStatusStatusPaused
				msg := "Still in sync"
				if len(changeInfo.Create) > 0 || len(changeInfo.Update) > 0 || len(changeInfo.Delete) > 0 {
					msg = fmt.Sprintf("Out of sync: %d to create, %d to update, %d to delete",
						len(changeInfo.Create), len(changeInfo.Update), len(changeInfo.Delete))
				}
				status.Message = msg
			}

			status.DetermineSyncStatus()

			err = w.sourceStatusPatcher.SetSourceStatus(wi.Source.ID, status)
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
