package domain

import (
	"context"

	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/forms"
	"github.com/pocketbase/pocketbase/models/schema"

	"github.com/nomad-ops/nomad-ops/backend/utils/log"
)

func InitModels(ctx context.Context, logger log.Logger, app core.App) error {
	keyCollection, err := initKeyCollection(app)
	if err != nil {
		logger.LogError(ctx, "Could not initKeyCollection:%v - %T", err, err)
		return err
	}

	srcCollection, err := initSourceCollection(app, keyCollection)
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
		f.Unique = field.Unique
		f.Options = field.Options
	} else {
		form.Schema.AddField(field)
	}
}