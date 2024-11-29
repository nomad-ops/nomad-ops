package domain

import (
	"database/sql"
	"fmt"

	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/forms"
	"github.com/pocketbase/pocketbase/models"
	"github.com/pocketbase/pocketbase/models/schema"
	"github.com/pocketbase/pocketbase/tools/types"
)

// Source A source to watch
//
// swagger:model Source
type Source struct {

	// id
	// Read Only: true
	ID string `json:"id,omitempty"`

	// name
	// Required: true
	Name string `json:"name"`

	// branch
	// Required: true
	Branch string `json:"branch"`

	// if true the namespace will be created if it does not exist
	CreateNamespace bool `json:"createNamespace,omitempty"`

	// if set, will override whatever is written in the job file. Use comma to provide multiple.
	DataCenter string `json:"dataCenter,omitempty"`

	// deployKeyID to use
	DeployKeyID string `json:"deployKeyID,omitempty"`

	// vaultTokenID to use
	VaultTokenID string `json:"vaultTokenID,omitempty"`

	// if true every commit forces an job update
	Force bool `json:"force,omitempty"`

	// if true no syncing is paused
	Paused bool `json:"paused,omitempty"`

	// if set, will override whatever is written in the job file
	Namespace string `json:"namespace,omitempty"`

	// path in the repo
	// Required: true
	Path string `json:"path"`

	// region
	Region string `json:"region,omitempty"`

	// status
	// Read Only: true
	Status *SourceStatus `json:"status,omitempty"`

	// url to clone from
	// Required: true
	URL string `json:"url"`
}

func initSourceCollection(app core.App,
	keysCollection *models.Collection,
	teamsCollection *models.Collection,
	vaultTokenCollection *models.Collection) (*models.Collection, error) {

	collection, err := app.Dao().FindCollectionByNameOrId("sources")

	if err == sql.ErrNoRows {
		collection = &models.Collection{}
	}
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	form := forms.NewCollectionUpsert(app, collection)
	form.Name = "sources"
	form.Type = models.CollectionTypeBase
	form.ListRule = types.Pointer("@request.auth.id != ''")
	form.ViewRule = types.Pointer("@request.auth.id != ''")
	form.CreateRule = types.Pointer("@request.auth.id != ''")
	form.UpdateRule = types.Pointer("@request.auth.id != '' && (teams:length = 0 || teams.members.id ?= @request.auth.id)")
	form.DeleteRule = types.Pointer("@request.auth.id != '' && (teams:length = 0 || teams.members.id ?= @request.auth.id)")

	addOrUpdateField(form, &schema.SchemaField{
		Name:     "name",
		Type:     schema.FieldTypeText,
		Required: true,
		Options: &schema.TextOptions{
			Max: types.Pointer(200),
		},
	})
	addOrUpdateField(form, &schema.SchemaField{
		Name:     "url",
		Type:     schema.FieldTypeText,
		Required: true,
		Options: &schema.TextOptions{
			Max: types.Pointer(200),
		},
	})
	addOrUpdateField(form, &schema.SchemaField{
		Name:     "branch",
		Type:     schema.FieldTypeText,
		Required: true,
		Options: &schema.TextOptions{
			Max: types.Pointer(100),
		},
	})
	addOrUpdateField(form, &schema.SchemaField{
		Name:     "path",
		Type:     schema.FieldTypeText,
		Required: true,
		Options: &schema.TextOptions{
			Max: types.Pointer(200),
		},
	})
	addOrUpdateField(form, &schema.SchemaField{
		Name:     "dataCenter",
		Type:     schema.FieldTypeText,
		Required: false,
		Options: &schema.TextOptions{
			Max: types.Pointer(100),
		},
	})
	addOrUpdateField(form, &schema.SchemaField{
		Name:     "region",
		Type:     schema.FieldTypeText,
		Required: false,
		Options: &schema.TextOptions{
			Max: types.Pointer(100),
		},
	})
	addOrUpdateField(form, &schema.SchemaField{
		Name:     "namespace",
		Type:     schema.FieldTypeText,
		Required: false,
		Options: &schema.TextOptions{
			Max: types.Pointer(100),
		},
	})
	max := 1
	addOrUpdateField(form, &schema.SchemaField{
		Name:     "deployKey",
		Type:     schema.FieldTypeRelation,
		Required: false,
		Options: &schema.RelationOptions{
			CollectionId: keysCollection.Id,
			MaxSelect:    &max,
		},
	})
	addOrUpdateField(form, &schema.SchemaField{
		Name:     "force",
		Type:     schema.FieldTypeBool,
		Required: false,
	})
	addOrUpdateField(form, &schema.SchemaField{
		Name:     "paused",
		Type:     schema.FieldTypeBool,
		Required: false,
	})
	addOrUpdateField(form, &schema.SchemaField{
		Name:     "status",
		Type:     schema.FieldTypeJson,
		Required: false,
		Options: &schema.JsonOptions{
			// 1 MB
			MaxSize: 1048576,
		},
	})
	addOrUpdateField(form, &schema.SchemaField{
		Name:     "teams",
		Type:     schema.FieldTypeRelation,
		Required: false,
		Options: &schema.RelationOptions{
			CollectionId: teamsCollection.Id,
		},
	})
	addOrUpdateField(form, &schema.SchemaField{
		Name:     "vaultToken",
		Type:     schema.FieldTypeRelation,
		Required: false,
		Options: &schema.RelationOptions{
			CollectionId: vaultTokenCollection.Id,
			MaxSelect:    &max,
		},
	})

	// validate and submit (internally it calls app.Dao().SaveCollection(collection) in a transaction)
	if err := form.Submit(); err != nil {
		return nil, err
	}
	return collection, nil
}

func SourceFromRecord(record *models.Record, withStatus bool) *Source {

	status := &SourceStatus{}
	if withStatus {
		err := record.UnmarshalJSONField("status", &status)
		if err != nil {
			fmt.Printf("Could not unmarshal status field:%v", err)
			status = nil
		}
	} else {
		status = nil
	}
	src := &Source{
		ID:              record.Id,
		Name:            record.GetString("name"),
		URL:             record.GetString("url"),
		Branch:          record.GetString("branch"),
		Path:            record.GetString("path"),
		DataCenter:      record.GetString("dataCenter"),
		Region:          record.GetString("region"),
		Namespace:       record.GetString("namespace"),
		DeployKeyID:     record.GetString("deployKey"),
		VaultTokenID:    record.GetString("vaultToken"),
		CreateNamespace: record.GetBool("createNamespace"),
		Force:           record.GetBool("force"),
		Paused:          record.GetBool("paused"),
		Status:          status,
	}

	return src
}
