package currentuser

import (
	"errors"
	"net/http"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/v2/app/logging"
	"github.com/France-ioi/AlgoreaBackend/v2/app/loginmodule"
	"github.com/France-ioi/AlgoreaBackend/v2/app/service"
)

// swagger:operation DELETE /current-user users currentUserDeletion
//
//	---
//	summary: Delete the current user
//	description: >
//							 Deletes all information stored in the platform related to the current user
//							 and calls the “UnlinkClient” service of the login module.
//
//							 The data to be deleted:
//
//							 1. [`filters`, `sessions`, `access_tokens`]
//									having `user_id` = `users.group_id`;
//							 2. `answers` having `author_id` = `users.group_id`;
//
//							 3. [`permissions_granted`, `permissions_generated`, `attempts`]
//									having `group_id` = `users.group_id`;
//
//							 4. `permissions_granted` having `source_group_id` = `users.group_id`;
//							 5. `groups_groups` having `parent_group_id` or `child_group_id` equal to `users.group_id`;
//							 6. `group_pending_requests`/`group_membership_changes` having `group_id` or `member_id` equal
//									to `users.group_id`;
//							 7. `groups_ancestors` having `ancestor_group_id` or `child_group_id` equal
//									to `users.group_id`;
//							 8. [`groups_propagate`, `groups`] having `id` equal to `users.group_id`;
//							 9. `users` having `group_id` = `users.group_id`.
//
//
//							 The deletion is rejected if the user is a member of at least one group with
//							 `now() < require_lock_membership_approval_until` and `groups_groups.lock_membership_approved` set.
//	responses:
//		"200":
//			"$ref": "#/responses/deletedResponse"
//		"401":
//			"$ref": "#/responses/unauthorizedResponse"
//		"403":
//			"$ref": "#/responses/forbiddenResponse"
//		"408":
//			"$ref": "#/responses/requestTimeoutResponse"
//		"500":
//			"$ref": "#/responses/internalErrorResponse"
func (srv *Service) delete(responseWriter http.ResponseWriter, httpRequest *http.Request) error {
	user := srv.GetUser(httpRequest)
	store := srv.GetStore(httpRequest)

	doNotDelete, err := store.ActiveGroupGroups().WhereUserIsMember(user).
		Where("groups_groups_active.lock_membership_approved").
		Joins("JOIN `groups` ON `groups`.id = groups_groups_active.parent_group_id").
		Where("NOW() < `groups`.require_lock_membership_approval_until").HasRows()
	service.MustNotBeError(err)

	if doNotDelete {
		logging.GetLogEntry(httpRequest).
			Infof("A user with group_id = %d tried to delete himself, but he is a member of a group with lock_user_deletion_until >= NOW()",
				user.GroupID)
		return service.ErrForbidden(errors.New("you cannot delete yourself right now"))
	}

	var loginID int64
	if !user.IsTempUser {
		service.MustNotBeError(store.Users().ByID(user.GroupID).
			PluckFirst("login_id", &loginID).Error())
	}
	service.MustNotBeError(store.Users().DeleteWithTraps(user, user.IsTempUser))

	if !user.IsTempUser {
		var result bool
		result, err = loginmodule.NewClient(srv.AuthConfig.GetString("loginModuleURL")).
			UnlinkClient(httpRequest.Context(), srv.AuthConfig.GetString("clientID"), srv.AuthConfig.GetString("clientSecret"), loginID)
		service.MustNotBeError(err)
		if !result {
			return service.ErrUnexpected(errors.New("login module failed"))
		}
	}

	render.Respond(responseWriter, httpRequest, service.DeletionSuccess[*struct{}](nil))

	return nil
}
