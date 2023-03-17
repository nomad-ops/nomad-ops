package sourcestore

import (
	"context"

	"github.com/pocketbase/pocketbase/core"

	"github.com/nomad-ops/nomad-ops/backend/application"
	"github.com/nomad-ops/nomad-ops/backend/domain"
	"github.com/nomad-ops/nomad-ops/backend/utils/log"
)

type PocketBaseStore struct {
	ctx    context.Context
	logger log.Logger
	cfg    PocketBaseStoreConfig
}

type PocketBaseStoreConfig struct {
	App core.App
}

func CreatePocketBaseStore(ctx context.Context,
	logger log.Logger,
	cfg PocketBaseStoreConfig) (*PocketBaseStore, error) {
	t := &PocketBaseStore{
		ctx:    ctx,
		logger: logger,
		cfg:    cfg,
	}

	return t, nil
}

func (s *PocketBaseStore) ListSources(ctx context.Context, opts application.ListSourcesOptions) ([]*domain.Source, error) {
	records, err := s.cfg.App.Dao().FindRecordsByExpr("sources")
	if err != nil {
		return nil, err
	}

	var res []*domain.Source

	for _, record := range records {
		res = append(res, domain.SourceFromRecord(record, true))
	}
	return res, nil
}

func (s *PocketBaseStore) SetSourceStatus(srcID string, status *domain.SourceStatus) error {
	record, err := s.cfg.App.Dao().FindRecordById("sources", srcID)
	if err != nil {
		return err
	}

	record.Set("status", status)

	if err := s.cfg.App.Dao().SaveRecord(record); err != nil {
		return err
	}

	return nil
}
