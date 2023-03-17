package keystore

import (
	"context"
	"database/sql"

	"github.com/pocketbase/pocketbase/core"

	"github.com/nomad-ops/nomad-ops/backend/domain"
	"github.com/nomad-ops/nomad-ops/backend/utils/errors"
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

func (s *PocketBaseStore) GetKey(ctx context.Context, id string) (*domain.DeployKey, error) {
	record, err := s.cfg.App.Dao().FindRecordById("keys", id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.ErrNotFound
		}
		return nil, err
	}
	return domain.DeployKeyFromRecord(record), nil
}
