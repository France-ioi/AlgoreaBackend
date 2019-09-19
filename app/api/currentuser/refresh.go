package currentuser

import (
	"net/http"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/app/auth"
	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/loginmodule"
	"github.com/France-ioi/AlgoreaBackend/app/service"
)

// swagger:operation PUT /current-user/refresh users userDataUpdate
// ---
// summary: Update the local user info cache
// description: Gets the user info from the login module and updates the local user info cache stored in the `users` table
// responses:
//   "200":
//     "$ref": "#/responses/updatedResponse"
//   "401":
//     "$ref": "#/responses/unauthorizedResponse"
//   "500":
//     "$ref": "#/responses/internalErrorResponse"
func (srv *Service) refresh(w http.ResponseWriter, r *http.Request) service.APIError {
	user := srv.GetUser(r)
	accessToken := auth.BearerTokenFromContext(r.Context())

	userProfile, err := loginmodule.NewClient(srv.Config.Auth.LoginModuleURL).GetUserProfile(r.Context(), accessToken)
	service.MustNotBeError(err)

	userProfile["last_activity_date"] = database.Now()
	if defaultLanguage, ok := userProfile["default_language"]; ok && defaultLanguage == nil {
		userProfile["default_language"] = database.Default()
	}
	service.MustNotBeError(srv.Store.Users().ByID(user.ID).UpdateColumn(userProfile).Error())

	response := service.UpdateSuccess(nil)
	render.Respond(w, r, &response)

	return service.NoError
}
