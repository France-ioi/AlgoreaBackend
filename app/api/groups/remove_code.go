package groups

import (
	"net/http"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/v2/app/service"
)

// swagger:operation DELETE /groups/{group_id}/code groups groupCodeRemove
//
//	---
//	summary: Remove a group code
//	description: >
//
//		Removes the code of the given group (which prevents joining by code)
//
//
//		The authenticated user should be a manager of `group_id` with `can_manage` >= 'memberships',
//		otherwise the 'forbidden' error is returned. If the group is a user, the 'forbidden' error is returned as well.
//	parameters:
//		- name: group_id
//			in: path
//			type: integer
//			required: true
//	responses:
//		"200":
//			"$ref": "#/responses/deletedResponse"
//		"400":
//			"$ref": "#/responses/badRequestResponse"
//		"401":
//			"$ref": "#/responses/unauthorizedResponse"
//		"403":
//			"$ref": "#/responses/forbiddenResponse"
//		"500":
//			"$ref": "#/responses/internalErrorResponse"
func (srv *Service) removeCode(w http.ResponseWriter, r *http.Request) service.APIError {
	var err error
	user := srv.GetUser(r)

	groupID, err := service.ResolveURLQueryPathInt64Field(r, "group_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	store := srv.GetStore(r)
	if apiError := checkThatUserCanManageTheGroupMemberships(store, user, groupID); apiError != service.NoError {
		return apiError
	}

	service.MustNotBeError(
		store.Groups().Where("id = ?", groupID).
			UpdateColumn("code", nil).Error())

	service.MustNotBeError(render.Render(w, r, service.DeletionSuccess(nil)))
	return service.NoError
}
