package domain

import (
	"context"

	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/forms"
	"github.com/pocketbase/pocketbase/models/schema"
	"github.com/pocketbase/pocketbase/tools/types"

	"github.com/nomad-ops/nomad-ops/backend/utils/log"
)

func InitModels(ctx context.Context, logger log.Logger, app core.App) error {

	usersCollection, err := app.Dao().FindCollectionByNameOrId("users")
	if err != nil {
		logger.LogError(ctx, "Could not FindCollectionByNameOrId('users'):%v - %T", err, err)
		return err
	}

	// allow everyone authenticated to list users
	form := forms.NewCollectionUpsert(app, usersCollection)
	form.ListRule = types.Pointer("@request.auth.id != ''")
	if err := form.Submit(); err != nil {
		return err
	}

	teamCollection, err := initTeamCollection(app, usersCollection)
	if err != nil {
		logger.LogError(ctx, "Could not initTeamCollection:%v - %T", err, err)
		return err
	}

	keyCollection, err := initKeyCollection(app, teamCollection)
	if err != nil {
		logger.LogError(ctx, "Could not initKeyCollection:%v - %T", err, err)
		return err
	}
	vaultTokenCollection, err := initVaultTokenCollection(app, teamCollection)
	if err != nil {
		logger.LogError(ctx, "Could not initVaultTokenCollection:%v - %T", err, err)
		return err
	}

	srcCollection, err := initSourceCollection(app, keyCollection, teamCollection, vaultTokenCollection)
	if err != nil {
		logger.LogError(ctx, "Could not initSourceCollection:%v - %T", err, err)
		return err
	}

	_, err = initEventCollection(app, srcCollection)
	if err != nil {
		logger.LogError(ctx, "Could not initEventCollection:%v", err)
		return err
	}
	return nil
}

func addOrUpdateField(form *forms.CollectionUpsert, field *schema.SchemaField) {

	if f := form.Schema.GetFieldByName(field.Name); f != nil {
		f.Name = field.Name
		f.Type = field.Type
		f.Required = field.Required
		f.Options = field.Options
	} else {
		form.Schema.AddField(field)
	}
}
