package groups

import (
	"net/http"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/app/service"
)

// swagger:model permissionsHasPathResponse
type permissionsHasPathResponse struct {
	// required: true
	HasPath bool `json:"has_path"`
}

// swagger:operation GET /groups/{group_id}/permissions/{item_id}/has-path groups permissionsHasPath
//
//	---
//	summary: Check if a group has a path to an item
//	description: >
//		Returns whether the given group can reach the given item from its content tree.
//
//		`has_path` is true when at least one of the following holds:
//
//		* at least one of the item's parents is visible to the group (`can_view` >= `info`),
//		* the item itself is visible to the group (`can_view` >= `info`),
//		* the item is a root activity/skill for an ancestor of the group.
//
//		The current user must be a manager (with `can_grant_group_access` permission) of a non-user
//		ancestor of `{group_id}`, i.e. be allowed to call `updatePermissions` for `{group_id}` on
//		`{item_id}` with some source group.
//	parameters:
//		- name: group_id
//			in: path
//			required: true
//			type: integer
//			format: int64
//		- name: item_id
//			in: path
//			required: true
//			type: integer
//			format: int64
//	responses:
//		"200":
//			description: OK. Whether the group has a path to the item.
//			schema:
//				"$ref": "#/definitions/permissionsHasPathResponse"
//		"400":
//			"$ref": "#/responses/badRequestResponse"
//		"401":
//			"$ref": "#/responses/unauthorizedResponse"
//		"403":
//			"$ref": "#/responses/forbiddenResponse"
//		"408":
//			"$ref": "#/responses/requestTimeoutResponse"
//		"500":
//			"$ref": "#/responses/internalErrorResponse"
func (srv *Service) getPermissionsHasPath(responseWriter http.ResponseWriter, httpRequest *http.Request) error {
	groupID, err := service.ResolveURLQueryPathInt64Field(httpRequest, "group_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	itemID, err := service.ResolveURLQueryPathInt64Field(httpRequest, "item_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	user := srv.GetUser(httpRequest)
	store := srv.GetStore(httpRequest)

	err = checkIfUserCanGrantPermissionsToGroup(store, user, groupID)
	service.MustNotBeError(err)

	hasPath, err := groupHasPathToItemOrItemIsRoot(store, groupID, itemID)
	service.MustNotBeError(err)

	render.Respond(responseWriter, httpRequest, permissionsHasPathResponse{HasPath: hasPath})
	return nil
}

func checkIfUserCanGrantPermissionsToGroup(store *database.DataStore, user *database.User, groupID int64) error {
	found, err := managerCanGrantGroupAccessScope(store, user, groupID).HasRows()
	service.MustNotBeError(err)
	if !found {
		return service.ErrAPIInsufficientAccessRights
	}
	return nil
}

func groupHasPathToItemOrItemIsRoot(store *database.DataStore, groupID, itemID int64) (bool, error) {
	// at least one of the item's parents should be visible to the group
	found, err := store.Permissions().MatchingGroupAncestors(groupID).
		WherePermissionIsAtLeast("view", info).
		Joins("JOIN items_items ON items_items.parent_item_id = permissions.item_id").
		Where("items_items.child_item_id = ?", itemID).
		HasRows()
	if err != nil {
		return false, err
	}
	if found {
		return true, nil
	}

	// if not, the item itself should be visible to the group
	found, err = store.Permissions().MatchingGroupAncestors(groupID).WherePermissionIsAtLeast("view", info).
		Where("item_id = ?", itemID).HasRows()
	if err != nil {
		return false, err
	}
	if found {
		return true, nil
	}

	// if not, the item should be a root item for one of the group's ancestors
	found, err = store.Groups().
		Joins("JOIN groups_ancestors_active ON groups_ancestors_active.ancestor_group_id = groups.id").
		Where("groups_ancestors_active.child_group_id = ?", groupID).
		Where("root_activity_id = ? OR root_skill_id = ?", itemID, itemID).HasRows()
	if err != nil {
		return false, err
	}

	return found, nil
}
