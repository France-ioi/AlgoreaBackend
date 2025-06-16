package auth

import (
	"net/http"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/v2/app/auth"
	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/app/service"
)

// swagger:operation POST /auth/logout auth authLogout
//
//	---
//	summary: User logout
//	description: Removes the current user’s session (all access and refresh tokens)
//	responses:
//		"200":
//			"$ref": "#/responses/successResponse"
//		"401":
//			"$ref": "#/responses/unauthorizedResponse"
//		"408":
//			"$ref": "#/responses/requestTimeoutResponse"
//		"500":
//			"$ref": "#/responses/internalErrorResponse"
func (srv *Service) logout(w http.ResponseWriter, r *http.Request) error {
	sessionID := srv.GetSessionID(r)

	service.MustNotBeError(srv.GetStore(r).InTransaction(func(store *database.DataStore) error {
		service.MustNotBeError(store.Sessions().Delete("session_id = ?", sessionID).Error())
		service.MustNotBeError(store.AccessTokens().Delete("session_id = ?", sessionID).Error())
		return nil
	}))

	cookieAttributes := auth.SessionCookieAttributesFromContext(r.Context())
	if cookieAttributes.UseCookie {
		http.SetCookie(w, cookieAttributes.SessionCookie("", -1000))
	}

	render.Respond(w, r, &service.Response[*struct{}]{Success: true, Message: "success"})
	return nil
}
