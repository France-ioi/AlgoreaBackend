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

// swagger:operation PUT /groups/{source_group_id}/permissions/{group_id}/{item_id} groups updatePermissions
// ---
// summary: Update permissions
// description: Let an admin of a group give/withdraw access rights on an item (identified by `item_id`)
//   to a user (identified by `group_id` of his self group).
//
//   * The user giving the access must be a manager (with `can_grant_group_access` permission)
//     of `{source_group_id}` which should be an ancestor of the `{group_id}`.
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
// - name: source_group_id
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
	sourceGroupID, err := service.ResolveURLQueryPathInt64Field(r, "source_group_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

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
		dataMap := data.ConstructMapForDB()

		if dataMap["can_view"] != nil && !checkUserHasAppropriateCanGrantViewPermissionForItem(dataMap["can_view"].(string), s, user, itemID) {
			apiErr = service.InsufficientAccessRightsError
			return apiErr.Error // rollback
		}

		apiErr = checkUserIsManagerAllowedToGrantPermissionsAndItemIsVisibleToGroup(s, user, sourceGroupID, groupID, itemID)
		if apiErr != service.NoError {
			return apiErr.Error
		}

		savePermissionsIntoDB(groupID, itemID, sourceGroupID, dataMap, s)
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

func checkUserIsManagerAllowedToGrantPermissionsAndItemIsVisibleToGroup(s *database.DataStore, user *database.User,
	sourceGroupID, groupID, itemID int64) service.APIError {
	// the authorized user should be a manager of the sourceGroupID with `can_grant_group_access' permission and
	// the 'sourceGroupID' should be an ancestor of 'groupID'
	found, err := s.Groups().ManagedBy(user).Where("groups.id = ?", sourceGroupID).
		Joins(`
				JOIN groups_ancestors_active AS descendants
					ON descendants.ancestor_group_id = groups.id AND descendants.child_group_id = ?`, groupID).
		Where("group_managers.can_grant_group_access").
		HasRows()
	service.MustNotBeError(err)
	if !found {
		return service.InsufficientAccessRightsError
	}

	// at least one of the item's parents should be visible to the group
	found, err = s.Permissions().MatchingGroupAncestors(groupID).
		WherePermissionIsAtLeast("view", "info").
		Joins("JOIN items_items ON items_items.parent_item_id = permissions.item_id").
		Where("items_items.child_item_id = ?", itemID).
		HasRows()
	service.MustNotBeError(err)
	if !found {
		found, err = s.Permissions().MatchingGroupAncestors(groupID).WherePermissionIsAtLeast("view", "info").
			Where("item_id = ?", itemID).HasRows()
		service.MustNotBeError(err)
		if !found {
			return service.InsufficientAccessRightsError
		}
	}
	return service.NoError
}

func checkUserHasAppropriateCanGrantViewPermissionForItem(viewPermissionToSet string, s *database.DataStore,
	user *database.User, itemID int64) bool {
	requiredGrantViewPermission := viewPermissionToSet
	if requiredGrantViewPermission == "info" { // no "info" in can_grant_view
		requiredGrantViewPermission = "enter"
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

func savePermissionsIntoDB(groupID, itemID, sourceGroupID int64, dbMap map[string]interface{}, s *database.DataStore) {
	dbMap["latest_update_at"] = database.Now()

	columnsToUpdate := make([]string, 0, len(dbMap))
	for key := range dbMap {
		columnsToUpdate = append(columnsToUpdate, key)
	}

	dbMap["group_id"] = groupID
	dbMap["item_id"] = itemID
	dbMap["source_group_id"] = sourceGroupID
	dbMap["origin"] = "group_membership"

	permissionGrantedStore := s.PermissionsGranted()
	service.MustNotBeError(permissionGrantedStore.InsertOrUpdateMap(dbMap, columnsToUpdate))
	service.MustNotBeError(permissionGrantedStore.After())
	if dbMap["can_view"] != nil && dbMap["can_view"] != "none" {
		// permissionGrantedStore.After() implicitly (via triggers) marks some attempts as to_be_propagated
		// when a item becomes visible, so we should propagate attempts here
		service.MustNotBeError(s.Results().Propagate())
	}
}
