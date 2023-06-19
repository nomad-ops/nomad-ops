package vaulttokenstore

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

func (s *PocketBaseStore) GetVaultToken(ctx context.Context, id string) (*domain.VaultToken, error) {
	record, err := s.cfg.App.Dao().FindRecordById("vault_tokens", id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.ErrNotFound
		}
		return nil, err
	}
	return domain.VaultTokenFromRecord(record), nil
}
