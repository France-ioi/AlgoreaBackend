package groups

import (
	"errors"
	"net/http"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/app/domain"
	"github.com/France-ioi/AlgoreaBackend/v2/app/service"
)

// swagger:operation DELETE /groups/{group_id} groups groupDelete
//
//	---
//	summary: Delete a group
//	description: >
//		Removes a group, its parent-child relations and other objects linked to it.
//
//
//		The service deletes `groups_groups`, `attempts`, `results`,
//		`group_membership_changes`, `group_pending_requests`,
//		`permissions_granted`, `permissions_generated`, and `filters` linked to the group.
//		Access rights are updated accordingly too.
//
//
//		If the group who is referenced in a `threads.helper_group_id`, the value of `threads.helper_group_id` will
//		be updated to the group `AllUsers`.
//
//
//		Restrictions (otherwise the 'forbidden' error is returned):
//			* the authenticated user should be a manager with `can_manage` = 'memberships_and_group' on the `{group_id}`,
//			* the group should not be of type "User".
//
//		Also, the group must be empty (no active subgroups of any type), otherwise the 'not found' error is returned.
//	parameters:
//		- name: group_id
//			in: path
//			type: integer
//			format: int64
//			required: true
//	responses:
//		"201":
//			"$ref": "#/responses/deletedResponse"
//		"400":
//			"$ref": "#/responses/badRequestResponse"
//		"401":
//			"$ref": "#/responses/unauthorizedResponse"
//		"403":
//			"$ref": "#/responses/forbiddenResponse"
//		"404":
//			"$ref": "#/responses/notFoundResponse"
//		"500":
//			"$ref": "#/responses/internalErrorResponse"
func (srv *Service) deleteGroup(w http.ResponseWriter, r *http.Request) service.APIError {
	groupID, err := service.ResolveURLQueryPathInt64Field(r, "group_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	user := srv.GetUser(r)
	apiErr := service.NoError

	err = srv.GetStore(r).InTransaction(func(s *database.DataStore) error {
		var found bool
		found, err = s.Groups().ManagedBy(user).
			WithWriteLock().
			Where("groups.id = ?", groupID).
			Where("group_managers.can_manage = 'memberships_and_group'").
			Where("groups.type != 'User'").HasRows()
		service.MustNotBeError(err)
		if !found {
			apiErr = service.InsufficientAccessRightsError
			return apiErr.Error // rollback
		}
		found, err = s.ActiveGroupGroups().Where("parent_group_id = ?", groupID).WithWriteLock().HasRows()
		service.MustNotBeError(err)
		if found {
			apiErr = service.ErrNotFound(errors.New("the group must be empty"))
			return apiErr.Error // rollback
		}

		// Updates all threads where helper_group_id was the deleted groupID to the AllUsers group.
		allUsersGroupID := domain.ConfigFromContext(r.Context()).AllUsersGroupID
		s.Threads().UpdateHelperGroupID(groupID, allUsersGroupID)

		return s.Groups().DeleteGroup(groupID)
	})

	if apiErr != service.NoError {
		return apiErr
	}

	service.MustNotBeError(err)
	service.MustNotBeError(render.Render(w, r, service.DeletionSuccess(nil)))
	return service.NoError
}
