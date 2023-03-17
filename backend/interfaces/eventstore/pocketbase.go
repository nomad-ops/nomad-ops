package eventstore

import (
	"context"

	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/forms"
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
func (s *PocketBaseStore) SaveEvent(ctx context.Context, ev *domain.Event) error {
	collection, err := s.cfg.App.Dao().FindCollectionByNameOrId("events")
	if err != nil {
		return err
	}

	record := models.NewRecord(collection)

	form := forms.NewRecordUpsert(s.cfg.App, record)

	err = form.LoadData(map[string]any{
		"message":   ev.Message,
		"type":      string(ev.Type),
		"timestamp": ev.Timestamp,
		"source":    ev.Source.ID,
	})
	if err != nil {
		return err
	}

	// validate and submit (internally it calls app.Dao().SaveRecord(record) in a transaction)
	if err := form.Submit(); err != nil {
		return err
	}
	return nil
}
