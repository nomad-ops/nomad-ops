package teamstore

import (
	"context"
	"database/sql"

	"github.com/pocketbase/pocketbase/core"

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

func (s *PocketBaseStore) IsTeamMember(ctx context.Context, teamID string, userID string) (bool, error) {

	teamRecord, err := s.cfg.App.Dao().FindRecordById("teams", teamID)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}

	members := teamRecord.GetStringSlice("members")
	for _, mem := range members {
		if mem == userID {
			return true, nil
		}
	}

	return false, nil
}
