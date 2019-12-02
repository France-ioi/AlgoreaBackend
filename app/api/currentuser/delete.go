package currentuser

import (
	"errors"
	"net/http"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/app/logging"
	"github.com/France-ioi/AlgoreaBackend/app/loginmodule"
	"github.com/France-ioi/AlgoreaBackend/app/service"
)

// swagger:operation DELETE /current-user users currentUserDeletion
// ---
// summary: Delete the current user
// description: >
//                Deletes all information stored in the platform related to the current user
//                and calls the “UnlinkClient” service of the login module.
//
//                The data to be deleted:
//
//                1. [`users_threads`, `users_answers`, `users_items`, `filters`, `sessions`, `refresh_tokens`]
//                   having `user_id` = `users.group_id`;
//                2. [`permissions_granted`, `permissions_generated`, `groups_attempts`, `groups_login_prefixes`]
//                   having `group_id` = `users.group_id`;
//
//                3. `groups_groups` having `parent_group_id` or `child_group_id` equal to `users.group_id`;
//                4. `group_pending_requests`/`group_membership_changes` having `group_id` or `member_id` equal
//                   to `users.group_id`;
//                5. `groups_ancestors` having `ancestor_group_id` or `child_group_id` equal
//                   to `users.group_id`;
//                6. [`groups_propagate`, `groups`] having `id` equal to `users.group_id`;
//                7. `users` having `group_id` = `users.group_id`.
//
//
//                The deletion is rejected if the user is a member of at least one group with
//                `now() < lock_user_deletion_until`.
// responses:
//   "200":
//     "$ref": "#/responses/deletedResponse"
//   "401":
//     "$ref": "#/responses/unauthorizedResponse"
//   "403":
//     "$ref": "#/responses/forbiddenResponse"
//   "500":
//     "$ref": "#/responses/internalErrorResponse"
func (srv *Service) delete(w http.ResponseWriter, r *http.Request) service.APIError {
	user := srv.GetUser(r)

	doNotDelete, err := srv.Store.GroupGroups().WhereUserIsMember(user).
		Joins("JOIN `groups` ON `groups`.id = groups_groups.parent_group_id").
		Where("NOW() < `groups`.lock_user_deletion_until").HasRows()
	service.MustNotBeError(err)

	if doNotDelete {
		logging.GetLogEntry(r).
			Infof("A user with group_id = %d tried to delete himself, but he is a member of a group with lock_user_deletion_until >= NOW()",
				user.GroupID)
		return service.ErrForbidden(errors.New("you cannot delete yourself right now"))
	}

	var loginID int64
	if !user.IsTempUser {
		service.MustNotBeError(srv.Store.Users().ByID(user.GroupID).
			PluckFirst("login_id", &loginID).Error())
	}
	service.MustNotBeError(srv.Store.Users().DeleteWithTraps(user))
	if !user.IsTempUser {
		service.MustNotBeError(loginmodule.NewClient(srv.Config.Auth.LoginModuleURL).
			UnlinkClient(r.Context(), srv.Config.Auth.ClientID, srv.Config.Auth.ClientSecret, loginID))
	}

	render.Respond(w, r, service.DeletionSuccess(nil))

	return service.NoError
}
