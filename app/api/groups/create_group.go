package groups

import (
	"errors"
	"net/http"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/formdata"
	"github.com/France-ioi/AlgoreaBackend/app/service"
)

// swagger:model createGroupRequest
type createGroupRequest struct {
	// required: true
	// minLength: 1
	Name *string `json:"name" validate:"set,min=1"`
	// required: true
	// enum: Class,Team,Club,Friends,Other
	Type *string `json:"type" validate:"set,oneof=Class Team Club Friends Other"`
	// only if `type` = "Team"
	// type: string
	// format: int64
	// minimum: 1
	ItemID *int64 `json:"item_id,string" validate:"min=1"`
}

// swagger:operation POST /groups groups groupCreate
// ---
// summary: Create a group
// description: >
//
//   Creates a group with the input `name`, `type`, `created_at` = now(), and default values in other columns.
//   If `item_id` is given:
//
//     * If `type` != "Team", returns the "badRequest" response
//
//     * Otherwise, checks that the authenticated user
//
//       * has at least `info` access on the item (otherwise returns the "forbidden" response)
//
//       * sets this `item_id` as `team_item_id` of the new group.
//
//   Also, the service sets the authenticated user as a manager of the group with the highest level of permissions.
//   After everything, it propagates group ancestors.
//
//
//   The user should have `owned_group_id` set and should not be temporary,
//   otherwise the "forbidden" response is returned.
// parameters:
// - in: body
//   name: data
//   required: true
//   description: The group to create
//   schema:
//     "$ref": "#/definitions/createGroupRequest"
// responses:
//   "201":
//     "$ref": "#/responses/createdWithIDResponse"
//   "400":
//     "$ref": "#/responses/badRequestResponse"
//   "401":
//     "$ref": "#/responses/unauthorizedResponse"
//   "403":
//     "$ref": "#/responses/forbiddenResponse"
//   "500":
//     "$ref": "#/responses/internalErrorResponse"
func (srv *Service) createGroup(w http.ResponseWriter, r *http.Request) service.APIError {
	var err error
	user := srv.GetUser(r)

	input := createGroupRequest{}
	formData := formdata.NewFormData(&input)
	err = formData.ParseJSONRequestData(r)
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	if input.ItemID != nil && *input.Type != "Team" {
		return service.ErrInvalidRequest(errors.New("only teams can be created with item_id set"))
	}

	// owned_group_id should be set (normal users have it)
	if user.IsTempUser || user.OwnedGroupID == nil {
		return service.InsufficientAccessRightsError
	}

	apiError := service.NoError
	var groupID int64
	err = srv.Store.InTransaction(func(store *database.DataStore) error {
		if input.ItemID != nil {
			hasRows, itemErr := store.Raw("SELECT 1 FROM ? AS access_rights",
				store.Permissions().VisibleToUser(user).Where("item_id = ?", *input.ItemID).
					WithWriteLock().SubQuery()).HasRows()
			service.MustNotBeError(itemErr)
			if !hasRows {
				apiError = service.InsufficientAccessRightsError
				return apiError.Error // rollback
			}
		}
		groupID, err = store.Groups().CreateNew(input.Name, input.Type, input.ItemID)
		service.MustNotBeError(err)
		return store.GroupManagers().InsertMap(map[string]interface{}{
			"group_id":               groupID,
			"manager_id":             user.GroupID,
			"can_manage":             "memberships_and_group",
			"can_grant_group_access": 1,
			"can_watch_members":      1,
		})
	})

	if apiError != service.NoError {
		return apiError
	}
	service.MustNotBeError(err)

	// response
	response := struct {
		GroupID int64 `json:"id,string"`
	}{GroupID: groupID}
	service.MustNotBeError(render.Render(w, r, service.CreationSuccess(&response)))
	return service.NoError
}
