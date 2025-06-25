package groups

import (
	"net/http"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/app/service"
	"github.com/France-ioi/AlgoreaBackend/v2/golang"
)

// swagger:operation DELETE /groups/{group_id}/managers/{manager_id} groups groupManagerDelete
//
//	---
//	summary: Remove a group manager
//	description: >
//
//		Removes a group manager.
//
//
//		The authenticated user should be the manager represented by `{manager_id}` or
//		have 'can_manage:memberships_and_group' permission on the group
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
//	responses:
//		"200":
//			"$ref": "#/responses/deletedResponse"
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
func (srv *Service) removeGroupManager(w http.ResponseWriter, r *http.Request) error {
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

	err = srv.GetStore(r).InTransaction(func(store *database.DataStore) error {
		var found bool
		// 1) the authenticated user should be the manager represented by managerID
		//    or have can_manage:memberships_and_group permission on the groupID
		// 2) there should be a row in group_managers for the given groupID-managerID pair
		if managerID == user.GroupID {
			found, err = store.GroupManagers().WithExclusiveWriteLock().
				Where("group_id = ?", groupID).
				Where("manager_id = ?", managerID).HasRows()
		} else {
			found, err = store.Groups().ManagedBy(user).
				WithCustomWriteLocks(
					golang.NewSet("groups", "groups_ancestors_active", "group_managers", "user_ancestors"),
					golang.NewSet("this_manager")).
				Where("groups.id = ?", groupID).
				Joins(`
				JOIN group_managers AS this_manager
					ON this_manager.group_id = groups.id AND this_manager.manager_id = ?`, managerID).
				Where("group_managers.can_manage = 'memberships_and_group'").HasRows()
		}
		service.MustNotBeError(err)
		if !found {
			return service.ErrAPIInsufficientAccessRights // rollback
		}

		return store.GroupManagers().
			Where("group_id = ?", groupID).
			Where("manager_id = ?", managerID).
			Delete().Error()
	})

	service.MustNotBeError(err)

	service.MustNotBeError(render.Render(w, r, service.DeletionSuccess[*struct{}](nil)))
	return nil
}
