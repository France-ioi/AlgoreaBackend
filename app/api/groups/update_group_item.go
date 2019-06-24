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
	PartialAccessDate *time.Time `json:"partial_access_date" sql:"column:sPartialAccessDate"`
	// Nullable
	FullAccessDate *time.Time `json:"full_access_date" sql:"column:sFullAccessDate"`
	// Nullable
	AccessSolutionsDate *time.Time `json:"access_solutions_date" sql:"column:sAccessSolutionsDate"`
	// Nullable
	// maxLength: 200
	AccessReason *string `json:"access_reason" sql:"column:sAccessReason" validate:"max=200"`
}

// swagger:operation PUT /groups/{group_id}/items/{item_id} groups items groupItemEdit
// ---
// summary: Edit the group access rights on the item
// description: Let an admin of a group give/withdraw access rights on an item to a user (represented by group_id).
//
//  * The user giving the access must be an owner of one of the ancestors of the group.
//
//  * The user giving the access must be an owner of any of the item’s ancestors or the item itself (bOwnerAccess)
//  or be a manager of the item (groups_items.bCachedManagerAccess=1).
//
//  * The group must already have access to one of the parents of the item or the item itself.
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
	if user.Load() == database.ErrUserNotFound {
		return service.InsufficientAccessRightsError
	}
	service.MustNotBeError(err)

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
			Select("groups_items.idItem").
			Where("groups_items.bCachedManagerAccess OR groups_items.bOwnerAccess").
			Where("groups_items.idItem = ? OR groups_items.idItem IN (SELECT idItemAncestor FROM items_ancestors WHERE idItemChild = ?)",
				itemID, itemID).HasRows()

		service.MustNotBeError(err)
		if !found {
			apiErr = service.InsufficientAccessRightsError
			return apiErr.Error // rollback
		}

		// the authorized user should own the group
		found, err = s.Groups().OwnedBy(user).Where("groups.ID = ?", groupID).HasRows()
		service.MustNotBeError(err)
		if !found {
			apiErr = service.InsufficientAccessRightsError
			return apiErr.Error // rollback
		}

		// at least one of the item's parents should be visible to the group
		itemsVisibleToGroupSubquery := s.GroupItems().
			Select(
				"idItem, MIN(sCachedFullAccessDate) <= NOW() AS fullAccess, "+
					"MIN(sCachedPartialAccessDate) <= NOW() AS partialAccess, "+
					"MIN(sCachedGrayedAccessDate) <= NOW() AS grayedAccess").
			Joins(`
				JOIN (SELECT * FROM groups_ancestors WHERE (groups_ancestors.idGroupChild = ?)) AS ancestors
				ON ancestors.idGroupAncestor = groups_items.idGroup`, groupID).
			Group("groups_items.idItem").
			Having("fullAccess > 0 OR partialAccess > 0 OR grayedAccess > 0").SubQuery()

		found, err = s.ItemItems().
			Joins("JOIN ? AS visible ON visible.idItem = items_items.idItemParent", itemsVisibleToGroupSubquery).
			Where("items_items.idItemChild = ?", itemID).
			HasRows()
		service.MustNotBeError(err)
		if !found {
			found, err = s.Items().ByID(itemID).
				Joins("JOIN ? AS visible ON visible.idItem = items.ID", itemsVisibleToGroupSubquery).HasRows()
			service.MustNotBeError(err)
			if !found {
				apiErr = service.InsufficientAccessRightsError
				return apiErr.Error // rollback
			}
		}

		saveGroupItemDataIntoDB(groupID, itemID, data, s)
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

func saveGroupItemDataIntoDB(groupID, itemID int64, data *formdata.FormData, s *database.DataStore) {
	dbMap := data.ConstructMapForDB()
	groupItemScope := s.GroupItems().
		Where("idGroup = ?", groupID).
		Where("idItem = ?", itemID)
	found, err := groupItemScope.WithWriteLock().HasRows()
	service.MustNotBeError(err)
	if found {
		service.MustNotBeError(groupItemScope.UpdateColumn(dbMap).Error())
	} else {
		dbMap["idGroup"] = groupID
		dbMap["idItem"] = itemID
		service.MustNotBeError(s.RetryOnDuplicatePrimaryKeyError(func(retryStore *database.DataStore) error {
			dbMap["ID"] = retryStore.NewID()
			return s.GroupItems().InsertMap(dbMap)
		}))
	}
	service.MustNotBeError(s.GroupItems().After())
}
