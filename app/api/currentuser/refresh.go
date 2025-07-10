package currentuser

import (
	"net/http"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/v2/app/auth"
	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/app/loginmodule"
	"github.com/France-ioi/AlgoreaBackend/v2/app/service"
)

// swagger:operation PUT /current-user/refresh users userDataRefresh
//
//	---
//	summary: Refresh the local user info cache
//	description: Gets the user info from the login module, updates the local user info cache stored in the `users` table
//						 and badges.
//	responses:
//		"200":
//			"$ref": "#/responses/updatedResponse"
//		"401":
//			"$ref": "#/responses/unauthorizedResponse"
//		"408":
//			"$ref": "#/responses/requestTimeoutResponse"
//		"500":
//			"$ref": "#/responses/internalErrorResponse"
func (srv *Service) refresh(responseWriter http.ResponseWriter, httpRequest *http.Request) error {
	user := srv.GetUser(httpRequest)
	accessToken := auth.BearerTokenFromContext(httpRequest.Context())

	userProfile, err := loginmodule.NewClient(srv.AuthConfig.GetString("loginModuleURL")).GetUserProfile(httpRequest.Context(), accessToken)
	service.MustNotBeError(err)
	badges := userProfile["badges"].([]database.Badge)

	service.MustNotBeError(srv.GetStore(httpRequest).InTransaction(func(store *database.DataStore) error {
		service.MustNotBeError(store.Groups().StoreBadges(badges, user.GroupID, false))
		userProfile["latest_activity_at"] = database.Now()
		userProfile["latest_profile_sync_at"] = database.Now()
		delete(userProfile, "default_language")
		delete(userProfile, "badges")
		service.MustNotBeError(store.Users().ByID(user.GroupID).UpdateColumn(userProfile).Error())
		return nil
	}))

	response := service.UpdateSuccess[*struct{}](nil)
	render.Respond(responseWriter, httpRequest, &response)

	return nil
}
