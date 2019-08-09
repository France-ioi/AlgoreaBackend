package currentuser

import (
	"errors"
	"net/http"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/app/loginmodule"
	"github.com/France-ioi/AlgoreaBackend/app/service"
)

// swagger:operation DELETE /current-user users currentUserDeletion
// ---
// summary: Delete the current user
// description: Deletes all information stored in the platform related to the current user
//              and calls the “UnlinkClient” service of the login module
//
//              * The deletion is rejected if the user is a member of at least one group with
//              `now() < lockUserDeletionDate`
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
		Joins("JOIN groups ON groups.ID = groups_groups.idGroupParent").
		Where("NOW() < groups.lockUserDeletionDate").HasRows()
	service.MustNotBeError(err)

	if doNotDelete {
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
