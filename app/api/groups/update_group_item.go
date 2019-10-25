package groups

import (
	"net/http"
	"time"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/formdata"
	"github.com/France-ioi/AlgoreaBackend/app/service"
)

// Access rights to be set
// swagger:model
type updateGroupItemInput struct {
	// Nullable
	PartialAccessSince *time.Time `json:"partial_access_since"`
	// Nullable
	FullAccessSince *time.Time `json:"full_access_since"`
	// Nullable
	SolutionsAccessSince *time.Time `json:"solutions_access_since"`
	// Nullable
	// maxLength: 200
	AccessReason *string `json:"access_reason" validate:"max=200"`
}

// swagger:operation PUT /groups/{group_id}/items/{item_id} groups items groupItemEdit
// ---
// summary: Update access rights of a group on an item
// description: Let an admin of a group give/withdraw access rights on an item (identified by `item_id`)
//   to a user (identified by `group_id` of his self group).
//
//   * The user giving the access must be an owner of one of the ancestors of the group.
//
//   * The user giving the access must be an owner of any of the item’s ancestors or the item itself (`owner_access`)
//   or be a manager of the item (`groups_items.cached_manager_access`).
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
//     "$ref": "#/definitions/updateGroupItemInput"
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
func (srv *Service) updateGroupItem(w http.ResponseWriter, r *http.Request) service.APIError {
	groupID, err := service.ResolveURLQueryPathInt64Field(r, "group_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	itemID, err := service.ResolveURLQueryPathInt64Field(r, "item_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	user := srv.GetUser(r)

	data := formdata.NewFormData(&updateGroupItemInput{})
	err = data.ParseJSONRequestData(r)
	if err != nil {
		return service.ErrInvalidRequest(err)
	}
	apiErr := service.NoError

	err = srv.Store.InTransaction(func(s *database.DataStore) error {
		var found bool

		// the item or one of its ancestors should be owned/managed by the authorized user
		found, err = s.GroupItems().
			MatchingUserAncestors(user).
			Select("groups_items.item_id").
			Where("groups_items.cached_manager_access OR groups_items.owner_access").
			Where("groups_items.item_id = ? OR groups_items.item_id IN (SELECT ancestor_item_id FROM items_ancestors WHERE child_item_id = ?)",
				itemID, itemID).HasRows()

		service.MustNotBeError(err)
		if !found {
			apiErr = service.InsufficientAccessRightsError
			return apiErr.Error // rollback
		}

		// the authorized user should own the group
		found, err = s.Groups().OwnedBy(user).Where("groups.id = ?", groupID).HasRows()
		service.MustNotBeError(err)
		if !found {
			apiErr = service.InsufficientAccessRightsError
			return apiErr.Error // rollback
		}

		// at least one of the item's parents should be visible to the group
		itemsVisibleToGroupSubQuery := s.GroupItems().AccessRightsForItemsVisibleToGroup(&groupID).SubQuery()

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

		saveGroupItemDataIntoDB(groupID, itemID, user.GroupID, data, s)
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

func saveGroupItemDataIntoDB(groupID, itemID, creatorUserGroupID int64, data *formdata.FormData, s *database.DataStore) {
	dbMap := data.ConstructMapForDB()
	groupItemScope := s.GroupItems().
		Where("group_id = ?", groupID).
		Where("item_id = ?", itemID)
	found, err := groupItemScope.WithWriteLock().HasRows()
	service.MustNotBeError(err)
	if found {
		service.MustNotBeError(groupItemScope.UpdateColumn(dbMap).Error())
	} else {
		dbMap["group_id"] = groupID
		dbMap["item_id"] = itemID
		service.MustNotBeError(s.RetryOnDuplicatePrimaryKeyError(func(retryStore *database.DataStore) error {
			dbMap["id"] = retryStore.NewID()
			dbMap["creator_user_group_id"] = creatorUserGroupID
			return s.GroupItems().InsertMap(dbMap)
		}))
	}
	service.MustNotBeError(s.GroupItems().After())
}
