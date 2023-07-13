package domain

import (
	"context"
	"database/sql"

	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/forms"
	"github.com/pocketbase/pocketbase/models"
	"github.com/pocketbase/pocketbase/models/schema"
	"github.com/pocketbase/pocketbase/tools/types"
)

type Team struct {

	// id
	// Read Only: true
	ID string `json:"id,omitempty"`

	// name
	// Required: true
	Name string `json:"name"`

	MemberIDs []string `json:"members"`
}

func (t *Team) UpsertUser(ctx context.Context, userID string) {
	for _, id := range t.MemberIDs {
		if userID == id {
			return
		}
	}
	t.MemberIDs = append(t.MemberIDs, userID)
}

func (t *Team) MergeMembers(ctx context.Context, userIDs []string) bool {
	unique := map[string]bool{}
	changed := false
	var res []string
	for _, id := range t.MemberIDs {
		unique[id] = true
	}
	for _, newID := range userIDs {
		if ok := unique[newID]; !ok {
			changed = true
		}
		unique[newID] = true
	}
	for id := range unique {
		res = append(res, id)
	}
	t.MemberIDs = res
	return changed
}

func initTeamCollection(app core.App, usersCollection *models.Collection) (*models.Collection, error) {

	collection, err := app.Dao().FindCollectionByNameOrId("teams")

	if err == sql.ErrNoRows {
		collection = &models.Collection{}
	}
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	form := forms.NewCollectionUpsert(app, collection)
	form.Name = "teams"
	form.Type = models.CollectionTypeBase
	form.ListRule = types.Pointer("@request.auth.id != ''")
	form.ViewRule = types.Pointer("@request.auth.id != ''")
	form.CreateRule = types.Pointer("@request.auth.id != ''")
	form.UpdateRule = types.Pointer("@request.auth.id != '' && (members:length = 0 || members.id = @request.auth.id)")
	form.DeleteRule = types.Pointer("@request.auth.id != '' && (members:length = 0 || members.id = @request.auth.id)")

	addOrUpdateField(form, &schema.SchemaField{
		Name:     "name",
		Type:     schema.FieldTypeText,
		Required: true,
		Options: &schema.TextOptions{
			Max: types.Pointer(200),
		},
	})
	addOrUpdateField(form, &schema.SchemaField{
		Name:     "members",
		Type:     schema.FieldTypeRelation,
		Required: false,
		Options: &schema.RelationOptions{
			CollectionId: usersCollection.Id,
		},
	})

	// validate and submit (internally it calls app.Dao().SaveCollection(collection) in a transaction)
	if err := form.Submit(); err != nil {
		return nil, err
	}
	return collection, nil
}
