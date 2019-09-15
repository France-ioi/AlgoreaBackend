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
//                1. [`users_threads`, `history_users_threads`, `users_answers`, `users_items`, `history_users_items`,
//                    `filters`, `history_filters`, `sessions`, `refresh_tokens`]
//                   having `idUser` = `users.ID`;
//                2. [`groups_items`, `history_groups_items`, `groups_attempts`, `history_groups_attempts`,
//                    `groups_login_prefixes`, `history_groups_login_prefixes`]
//                   having `idGroup` = `users.idGroupSelf` or `idGroup` = `users.idGroupOwned`;
//                3. `groups_items_propagate` having the same `ID`s as the rows removed from `groups_items`;
//
//                4. [`groups_groups`, `history_groups_groups`] having `idGroupParent` or `idGroupChild` equal
//                   to one of `users.idGroupSelf`/`users.idGroupOwned`;
//                5. [`groups_ancestors`, `history_groups_ancestors`] having `idGroupAncestor` or `idGroupChild` equal
//                   to one of `users.idGroupSelf`/`users.idGroupOwned`;
//                6. [`groups_propagate`, `groups`, `history_groups`] having `ID` equal to one of
//                   `users.idGroupSelf`/`users.idGroupOwned`;
//                7. `users`, `history_users` having `ID` = `users.ID`.
//
//
//                The deletion is rejected if the user is a member of at least one group with
//                `now() < lockUserDeletionDate`.
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
		Joins("JOIN `groups` ON `groups`.ID = groups_groups.idGroupParent").
		Where("NOW() < `groups`.lockUserDeletionDate").HasRows()
	service.MustNotBeError(err)

	if doNotDelete {
		logging.GetLogEntry(r).
			Infof("A user with ID = %d tried to delete himself, but he is a member of a group with lockUserDeletionDate >= NOW()",
				user.ID)
		return service.ErrForbidden(errors.New("you cannot delete yourself right now"))
	}

	var loginID int64
	if !user.IsTempUser {
		service.MustNotBeError(srv.Store.Users().ByID(user.ID).PluckFirst("loginID", &loginID).Error())
	}
	service.MustNotBeError(srv.Store.Users().DeleteWithTraps(user))
	if !user.IsTempUser {
		service.MustNotBeError(loginmodule.NewClient(srv.Config.Auth.LoginModuleURL).
			UnlinkClient(r.Context(), srv.Config.Auth.ClientID, srv.Config.Auth.ClientSecret, loginID))
	}

	render.Respond(w, r, service.DeletionSuccess(nil))

	return service.NoError
}
