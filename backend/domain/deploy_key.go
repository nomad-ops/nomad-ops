package domain

import (
	"database/sql"
	"time"

	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/forms"
	"github.com/pocketbase/pocketbase/models"
	"github.com/pocketbase/pocketbase/models/schema"
	"github.com/pocketbase/pocketbase/tools/types"
)

type DeployKey struct {

	// name
	// Required: true
	Name string `json:"name"`

	// created
	// Read Only: true
	Created time.Time `json:"timestamp,omitempty"`

	// value
	// Required: true
	Value string `json:"value"`
}

func initKeyCollection(app core.App) (*models.Collection, error) {

	collection, err := app.Dao().FindCollectionByNameOrId("keys")

	if err == sql.ErrNoRows {
		collection = &models.Collection{}
	}
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	form := forms.NewCollectionUpsert(app, collection)
	form.Name = "keys"
	form.Type = models.CollectionTypeBase
	form.ListRule = types.Pointer("@request.auth.id != ''")
	form.ViewRule = types.Pointer("@request.auth.id != ''")
	form.CreateRule = types.Pointer("@request.auth.id != ''")
	form.UpdateRule = types.Pointer("@request.auth.id != ''")
	form.DeleteRule = types.Pointer("@request.auth.id != ''")

	addOrUpdateField(form, &schema.SchemaField{
		Name:     "name",
		Type:     schema.FieldTypeText,
		Required: true,
		Unique:   true,
		Options: &schema.TextOptions{
			Max: types.Pointer(100),
		},
	})
	addOrUpdateField(form, &schema.SchemaField{
		Name:     "value",
		Type:     schema.FieldTypeText,
		Required: true,
		Options: &schema.TextOptions{
			Max: types.Pointer(1000),
		},
	})

	// validate and submit (internally it calls app.Dao().SaveCollection(collection) in a transaction)
	if err := form.Submit(); err != nil {
		return nil, err
	}
	return collection, nil
}

func DeployKeyFromRecord(record *models.Record) *DeployKey {
	return &DeployKey{
		Name:    record.GetString("name"),
		Created: record.Created.Time(),
		Value:   record.GetString("value"),
	}
}
