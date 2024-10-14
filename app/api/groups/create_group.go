package groups

import (
	"net/http"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/app/formdata"
	"github.com/France-ioi/AlgoreaBackend/v2/app/service"
)

// swagger:model createGroupRequest
type createGroupRequest struct {
	// required: true
	// minLength: 1
	Name string `json:"name" validate:"set,min=1"`
	// required: true
	// enum: Class,Team,Club,Friends,Other,Session
	Type string `json:"type" validate:"set,oneof=Class Team Club Friends Other Session"`
}

// swagger:operation POST /groups groups groupCreate
//
//	---
//	summary: Create a group
//	description: >
//
//		Creates a group with the input `name`, `type`, `created_at` = now(), and default values in other columns.
//
//
//		Also, the service sets the authenticated user as a manager of the group with the highest level of permissions.
//		After everything, it propagates group ancestors.
//
//
//		The user should not be temporary, otherwise the "forbidden" response is returned.
//	parameters:
//		- in: body
//			name: data
//			required: true
//			description: The group to create
//			schema:
//				"$ref": "#/definitions/createGroupRequest"
//	responses:
//		"201":
//			"$ref": "#/responses/createdWithIDResponse"
//		"400":
//			"$ref": "#/responses/badRequestResponse"
//		"401":
//			"$ref": "#/responses/unauthorizedResponse"
//		"403":
//			"$ref": "#/responses/forbiddenResponse"
//		"500":
//			"$ref": "#/responses/internalErrorResponse"
func (srv *Service) createGroup(w http.ResponseWriter, r *http.Request) service.APIError {
	var err error
	user := srv.GetUser(r)

	input := createGroupRequest{}
	formData := formdata.NewFormData(&input)
	err = formData.ParseJSONRequestData(r)
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	if user.IsTempUser {
		return service.InsufficientAccessRightsError
	}

	var groupID int64
	err = srv.GetStore(r).InTransaction(func(store *database.DataStore) error {
		groupID, err = store.Groups().CreateNew(input.Name, input.Type)
		service.MustNotBeError(err)
		return store.GroupManagers().InsertMap(map[string]interface{}{
			"group_id":               groupID,
			"manager_id":             user.GroupID,
			"can_manage":             "memberships_and_group",
			"can_grant_group_access": 1,
			"can_watch_members":      1,
		})
	})
	service.MustNotBeError(err)

	// response
	response := struct {
		GroupID int64 `json:"id,string"`
	}{GroupID: groupID}
	service.MustNotBeError(render.Render(w, r, service.CreationSuccess(&response)))
	return service.NoError
}
