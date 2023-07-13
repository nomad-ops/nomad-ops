package teamstore

import (
	"context"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/daos"
	"github.com/pocketbase/pocketbase/models"

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

func (s *PocketBaseStore) UpsertTeam(ctx context.Context, t *domain.Team) error {
	err := s.cfg.App.Dao().RunInTransaction(func(txDao *daos.Dao) error {
		records, err := txDao.FindRecordsByExpr("teams",
			dbx.HashExp{
				"name": t.Name,
			},
		)
		if err != nil {
			return err
		}
		if len(records) != 0 {
			teamRecord := records[0]
			t.ID = teamRecord.Id

			members := teamRecord.GetStringSlice("members")
			newTeam := domain.Team{
				ID:        t.ID,
				Name:      t.Name,
				MemberIDs: members,
			}
			if !newTeam.MergeMembers(ctx, t.MemberIDs) {
				return nil
			}
			// changed => update team
			teamRecord.Set("members", newTeam.MemberIDs)
			t.MemberIDs = newTeam.MemberIDs

			s.logger.LogInfo(ctx, "Upserting team with new member...%s", t.Name)
			if err := txDao.SaveRecord(teamRecord); err != nil {
				s.logger.LogError(ctx, "Could not upsert team %s:%v", t.Name, err)
				return err
			}
			return nil
		}
		coll, err := txDao.FindCollectionByNameOrId("teams")
		if err != nil {
			return err
		}
		record := models.NewRecord(coll)
		record.Set("name", t.Name)
		record.Set("members", t.MemberIDs)

		// validate and submit (internally it calls app.Dao().SaveRecord(record) in a transaction)
		s.logger.LogInfo(ctx, "Upserting team...%s", t.Name)
		if err := txDao.SaveRecord(record); err != nil {
			s.logger.LogError(ctx, "Could not upsert team %s:%v", t.Name, err)
			return err
		}
		t.ID = record.Id
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}
