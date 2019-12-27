package groups

import (
	"net/http"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/formdata"
	"github.com/France-ioi/AlgoreaBackend/app/service"
)

// swagger:model createGroupManagerRequest
type createGroupManagerRequest struct {
	// enum: none,memberships,memberships_and_group
	CanManage           string `json:"can_manage" validate:"oneof=none memberships memberships_and_group"`
	CanGrantGroupAccess bool   `json:"can_grant_group_access"`
	CanWatchMembers     bool   `json:"can_watch_members"`
}

// swagger:operation POST /groups/{group_id}/managers/{manager_id} groups groupManagerCreate
// ---
// summary: Make user a group manager
// description: >
//
//   Makes a user a group manager with given permissions.
//
//
//   The authenticated user should have 'can_manage:memberships_and_group' permission on the group
//   and `{manager_id}` should exist, otherwise the "forbidden" error is returned.
// parameters:
// - name: group_id
//   in: path
//   required: true
//   type: integer
// - name: manager_id
//   in: path
//   required: true
//   type: integer
// - in: body
//   name: data
//   required: true
//   description: Permissions of the new manager
//   schema:
//     "$ref": "#/definitions/createGroupManagerRequest"
// responses:
//   "201":
//     description: Created. The request has successfully added a user as a manager.
//     schema:
//       "$ref": "#/definitions/createdResponse"
//   "400":
//     "$ref": "#/responses/badRequestResponse"
//   "401":
//     "$ref": "#/responses/unauthorizedResponse"
//   "403":
//     "$ref": "#/responses/forbiddenResponse"
//   "500":
//     "$ref": "#/responses/internalErrorResponse"
func (srv *Service) createGroupManager(w http.ResponseWriter, r *http.Request) service.APIError {
	var err error
	user := srv.GetUser(r)

	groupID, err := service.ResolveURLQueryPathInt64Field(r, "group_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}
	managerID, err := service.ResolveURLQueryPathInt64Field(r, "manager_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	input := createGroupManagerRequest{}
	formData := formdata.NewFormData(&input)
	err = formData.ParseJSONRequestData(r)
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	apiError := service.NoError
	err = srv.Store.InTransaction(func(store *database.DataStore) error {
		var found bool
		// managerID should exist and the authenticated user should have
		// can_manage:memberships_and_group permission on the groupID
		found, err = store.Groups().ManagedBy(user).WithWriteLock().
			Where("groups.id = ?", groupID).
			Joins("JOIN users ON users.group_id = ?", managerID).
			Where("can_manage = 'memberships_and_group'").HasRows()
		service.MustNotBeError(err)
		if !found {
			apiError = service.InsufficientAccessRightsError
			return apiError.Error // rollback
		}

		values := formData.ConstructMapForDB()
		values["group_id"] = groupID
		values["manager_id"] = managerID
		return store.GroupManagers().InsertMap(values)
	})

	if apiError != service.NoError {
		return apiError
	}
	service.MustNotBeError(err)

	service.MustNotBeError(render.Render(w, r, service.CreationSuccess(nil)))
	return service.NoError
}
