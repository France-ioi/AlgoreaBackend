package groups

import (
	"net/http"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/app/formdata"
	"github.com/France-ioi/AlgoreaBackend/v2/app/service"
)

// swagger:operation PUT /groups/{group_id}/managers/{manager_id} groups groupManagerEdit
//
//	---
//	summary: Change permissions of a group manager
//	description: >
//
//		Modifies permissions of a group manager.
//
//
//		The authenticated user should have 'can_manage:memberships_and_group' permission on the group
//		and the `{group_id}`-`{manager_id}` pair should exist in `group_managers,
//		otherwise the "forbidden" error is returned.
//	parameters:
//		- name: group_id
//			in: path
//			required: true
//			type: integer
//			format: int64
//		- name: manager_id
//			in: path
//			required: true
//			type: integer
//			format: int64
//		- in: body
//			name: data
//			required: true
//			description: New permissions of the manager
//			schema:
//				"$ref": "#/definitions/createGroupManagerRequest"
//	responses:
//		"200":
//			"$ref": "#/responses/updatedResponse"
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
func (srv *Service) updateGroupManager(responseWriter http.ResponseWriter, httpRequest *http.Request) error {
	var err error
	user := srv.GetUser(httpRequest)

	groupID, err := service.ResolveURLQueryPathInt64Field(httpRequest, "group_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}
	managerID, err := service.ResolveURLQueryPathInt64Field(httpRequest, "manager_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	input := createGroupManagerRequest{}
	formData := formdata.NewFormData(&input)
	err = formData.ParseJSONRequestData(httpRequest)
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	err = srv.GetStore(httpRequest).InTransaction(func(store *database.DataStore) error {
		var found bool
		// 1) the authenticated user should have can_manage:memberships_and_group permission on the groupID
		// 2) there should be a row in group_managers for the given groupID-managerID pair
		found, err = store.Groups().ManagedBy(user).WithExclusiveWriteLock().
			Where("groups.id = ?", groupID).
			Joins(`
				JOIN group_managers AS this_manager
					ON this_manager.group_id = groups.id AND this_manager.manager_id = ?`, managerID).
			Where("group_managers.can_manage = 'memberships_and_group'").HasRows()
		service.MustNotBeError(err)
		if !found {
			return service.ErrAPIInsufficientAccessRights // rollback
		}

		values := formData.ConstructMapForDB()
		return store.GroupManagers().
			Where("group_id = ?", groupID).
			Where("manager_id = ?", managerID).
			UpdateColumn(values).Error()
	})

	service.MustNotBeError(err)

	service.MustNotBeError(render.Render(responseWriter, httpRequest, service.UpdateSuccess[*struct{}](nil)))
	return nil
}
