package auth

import (
	"net/http"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/app/auth"
	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/service"
)

// swagger:operation POST /auth/logout auth authLogout
// ---
// summary: User logout
// description: Removes the current user’s session (all access and refresh tokens)
// responses:
//   "200":
//     "$ref": "#/responses/successResponse"
//   "401":
//     "$ref": "#/responses/unauthorizedResponse"
//   "500":
//     "$ref": "#/responses/internalErrorResponse"
func (srv *Service) logout(w http.ResponseWriter, r *http.Request) service.APIError {
	user := srv.GetUser(r)

	service.MustNotBeError(srv.Store.InTransaction(func(store *database.DataStore) error {
		service.MustNotBeError(store.Sessions().Delete("user_id = ?", user.GroupID).Error())
		service.MustNotBeError(store.RefreshTokens().Delete("user_id = ?", user.GroupID).Error())
		return nil
	}))

	cookieAttributes := auth.SessionCookieAttributesFromContext(r.Context())
	if _, cookieErr := r.Cookie("access_token"); cookieErr == nil && cookieAttributes.UseCookie {
		http.SetCookie(w, cookieAttributes.SessionCookie("", -1000))
	}

	render.Respond(w, r, &service.Response{Success: true, Message: "success"})
	return service.NoError
}
