package userstore

import (
	"context"

	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/forms"
	"github.com/pocketbase/pocketbase/models"

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
func (s *PocketBaseStore) CreateUser(ctx context.Context,
	username, email, pw string) error {
	collection, err := s.cfg.App.Dao().FindCollectionByNameOrId("users")
	if err != nil {
		return err
	}

	record := models.NewRecord(collection)

	form := forms.NewRecordUpsert(s.cfg.App, record)

	form.Username = username
	form.Email = email

	form.Password = pw
	form.PasswordConfirm = pw

	// validate and submit (internally it calls app.Dao().SaveRecord(record) in a transaction)
	if err := form.Submit(); err != nil {
		return err
	}
	return nil
}
