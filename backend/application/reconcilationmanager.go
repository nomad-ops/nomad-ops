package application

import (
	"context"

	"github.com/nomad-ops/nomad-ops/backend/domain"
	"github.com/nomad-ops/nomad-ops/backend/utils/log"
)

type NotifyAdditionalInfos struct {
	Header string
	Text   string
	Large  bool
}

type NotificationType string

var (
	NotificationSuccess NotificationType = "success"
	NotificationError   NotificationType = "error"
)

type NotifyOptions struct {
	Source  *domain.Source
	GitInfo GitInfo
	Type    NotificationType
	Message string
	Infos   []NotifyAdditionalInfos
}

type Notifier interface {
	Notify(ctx context.Context, opts NotifyOptions) error
}

type ListSourcesOptions struct {
}

type SourceRepo interface {
	ListSources(ctx context.Context, opts ListSourcesOptions) ([]*domain.Source, error)
}

type ListKeysOptions struct {
}

type KeyRepo interface {
	GetKey(ctx context.Context, id string) (*domain.DeployKey, error)
}

type EventRepo interface {
	SaveEvent(ctx context.Context, ev *domain.Event) error
}

type SourceWatcher interface {
	WatchSource(ctx context.Context, src *domain.Source, cb ReconcilerFunc) error
	StopSourceWatch(ctx context.Context, id string) error
}

type ReconciliationManager struct {
	ctx           context.Context
	logger        log.Logger
	cfg           ReconciliationManagerConfig
	repo          SourceRepo
	watcher       SourceWatcher
	clusterAccess ClusterAPI
	evRepo        EventRepo
	notifier      Notifier
}

type ReconciliationManagerConfig struct {
}

func CreateReconciliationManager(ctx context.Context,
	logger log.Logger,
	cfg ReconciliationManagerConfig,
	repo SourceRepo,
	watcher SourceWatcher,
	clusterAccess ClusterAPI,
	evRepo EventRepo,
	notifier Notifier) (*ReconciliationManager, error) {
	t := &ReconciliationManager{
		ctx:           ctx,
		logger:        logger,
		cfg:           cfg,
		repo:          repo,
		watcher:       watcher,
		clusterAccess: clusterAccess,
		evRepo:        evRepo,
		notifier:      notifier,
	}

	// Get all sources from repo on startup
	srcs, err := repo.ListSources(ctx, ListSourcesOptions{})
	if err != nil {
		return nil, err
	}

	for _, src := range srcs {
		cpy := src
		err = watcher.WatchSource(ctx, cpy, t.OnReconcile)
		if err != nil {
			return nil, err
		}
	}

	return t, nil
}

func (m *ReconciliationManager) OnAddedSource(ctx context.Context, src *domain.Source) error {
	err := m.watcher.WatchSource(ctx, src, m.OnReconcile)
	if err != nil {
		return err
	}
	return nil
}

func (m *ReconciliationManager) ListSources(ctx context.Context, opts ListSourcesOptions) ([]*domain.Source, error) {
	return m.repo.ListSources(ctx, opts)
}

func (m *ReconciliationManager) OnDeletedSource(ctx context.Context, id string) error {
	err := m.watcher.StopSourceWatch(ctx, id)
	if err != nil {
		return err
	}
	return nil
}
