package groups

import (
	"net/http"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/formdata"
	"github.com/France-ioi/AlgoreaBackend/app/service"
)

// Access rights to be set
// swagger:model
type updatePermissionsInput struct {
	CanView string `json:"can_view" validate:"oneof=none info content content_with_descendants solution"`
}

// swagger:operation PUT /groups/{group_id}/items/{item_id} groups items updatePermissions
// ---
// summary: Update access rights of a group on an item
// description: Let an admin of a group give/withdraw access rights on an item (identified by `item_id`)
//   to a user (identified by `group_id` of his self group).
//
//   * The user giving the access must be a manager of one of the ancestors of the group.
//
//   * The user giving the access must have `permissions_generated.can_grant_view` >= given `can_view`
//     for the item.
//
//   * The group must already have access to one of the parents of the item or the item itself.
// parameters:
// - name: group_id
//   in: path
//   required: true
//   type: integer
// - name: item_id
//   in: path
//   required: true
//   type: integer
// - name: access rights information
//   in: body
//   required: true
//   schema:
//     "$ref": "#/definitions/updatePermissionsInput"
// responses:
//   "200":
//     "$ref": "#/responses/updatedResponse"
//   "400":
//     "$ref": "#/responses/badRequestResponse"
//   "401":
//     "$ref": "#/responses/unauthorizedResponse"
//   "403":
//     "$ref": "#/responses/forbiddenResponse"
//   "500":
//     "$ref": "#/responses/internalErrorResponse"
func (srv *Service) updatePermissions(w http.ResponseWriter, r *http.Request) service.APIError {
	groupID, err := service.ResolveURLQueryPathInt64Field(r, "group_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	itemID, err := service.ResolveURLQueryPathInt64Field(r, "item_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	user := srv.GetUser(r)

	data := formdata.NewFormData(&updatePermissionsInput{})
	err = data.ParseJSONRequestData(r)
	if err != nil {
		return service.ErrInvalidRequest(err)
	}
	apiErr := service.NoError

	err = srv.Store.InTransaction(func(s *database.DataStore) error {
		var found bool

		dataMap := data.ConstructMapForDB()

		if dataMap["can_view"] != nil {
			if !checkUserHasRightsToSetCanView(dataMap["can_view"].(string), s, user, itemID) {
				apiErr = service.InsufficientAccessRightsError
				return apiErr.Error // rollback
			}
		}

		// the authorized user should be a manager of the group
		found, err = s.Groups().ManagedBy(user).Where("groups.id = ?", groupID).HasRows()
		service.MustNotBeError(err)
		if !found {
			apiErr = service.InsufficientAccessRightsError
			return apiErr.Error // rollback
		}

		// at least one of the item's parents should be visible to the group
		itemsVisibleToGroupSubQuery := s.Permissions().VisibleToGroup(groupID).SubQuery()

		found, err = s.ItemItems().
			Joins("JOIN ? AS visible ON visible.item_id = items_items.parent_item_id", itemsVisibleToGroupSubQuery).
			Where("items_items.child_item_id = ?", itemID).
			HasRows()
		service.MustNotBeError(err)
		if !found {
			found, err = s.Items().ByID(itemID).
				Joins("JOIN ? AS visible ON visible.item_id = items.id", itemsVisibleToGroupSubQuery).HasRows()
			service.MustNotBeError(err)
			if !found {
				apiErr = service.InsufficientAccessRightsError
				return apiErr.Error // rollback
			}
		}

		savePermissionsIntoDB(groupID, itemID, user.GroupID, dataMap, s)
		return nil
	})

	if apiErr != service.NoError {
		return apiErr
	}

	service.MustNotBeError(err)

	response := service.Response{Success: true, Message: "updated"}
	render.Respond(w, r, &response)

	return service.NoError
}

func checkUserHasRightsToSetCanView(viewPermissionToSet string, s *database.DataStore, user *database.User, itemID int64) bool {
	requiredGrantViewPermission := viewPermissionToSet
	if requiredGrantViewPermission == "info" { // no "info" in can_grant_view
		requiredGrantViewPermission = "content"
	}
	// permissions_generated.can_grant_view_generated should be >= data["can_view"]
	found, err := s.Permissions().
		MatchingUserAncestors(user).
		Select("permissions.item_id").
		WherePermissionIsAtLeast("grant_view", requiredGrantViewPermission).
		Where("permissions.item_id = ?", itemID).HasRows()
	service.MustNotBeError(err)
	return found
}

func savePermissionsIntoDB(groupID, itemID, giverGroupID int64, dbMap map[string]interface{}, s *database.DataStore) {
	dbMap["latest_update_on"] = database.Now()

	columnsToUpdate := make([]string, 0, len(dbMap))
	for key := range dbMap {
		columnsToUpdate = append(columnsToUpdate, key)
	}

	dbMap["group_id"] = groupID
	dbMap["item_id"] = itemID
	dbMap["giver_group_id"] = giverGroupID

	permissionGrantedStore := s.PermissionsGranted()
	service.MustNotBeError(permissionGrantedStore.InsertOrUpdateMap(dbMap, columnsToUpdate))
	service.MustNotBeError(permissionGrantedStore.After())
}
