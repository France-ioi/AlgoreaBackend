package auth

import (
	"net/http"
	"sync"
	"time"

	"golang.org/x/oauth2"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/app/auth"
	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/service"
)

var userIDsInProgress sync.Map

// swagger:operation POST /auth/token auth authTokenCreate
// ---
// summary: Request a new access token
// description: Creates a new access token (locally for temporary users or via the login module for normal users) and
//              saves it in the DB keeping only the input token (from authorization headers) and the new token.
//              Since the login module responds with both access and refresh tokens, the service updates the user's
//              refresh token in this case as well.
// responses:
//   "201":
//     description: "Created. Success response with the new access token"
//     in: body
//     schema:
//       "$ref": "#/definitions/userCreateTmpResponse"
//   "401":
//     "$ref": "#/responses/unauthorizedResponse"
//   "500":
//     "$ref": "#/responses/internalErrorResponse"
func (srv *Service) createToken(w http.ResponseWriter, r *http.Request) service.APIError {
	user := srv.GetUser(r)
	oldAccessToken := auth.BearerTokenFromContext(r.Context())

	var newToken string
	var expiresIn int32

	if user.IsTempUser {
		service.MustNotBeError(srv.Store.InTransaction(func(store *database.DataStore) error {
			sessionStore := store.Sessions()
			// delete all the user's access tokens keeping the input token only
			service.MustNotBeError(sessionStore.Delete("idUser = ? AND sAccessToken != ?",
				user.ID, oldAccessToken).Error())
			var err error
			newToken, expiresIn, err = auth.CreateNewTempSession(sessionStore, user.ID)
			return err
		}))
	} else {
		// We should not allow concurrency in this part because the login module generates not only
		// a new access token, but also a new refresh token and revokes the old one. We want to prevent
		// usage of the old refresh token for that reason.
		userMutex := make(chan bool)
		defer close(userMutex)
		userMutexInterface, loaded := userIDsInProgress.LoadOrStore(user.ID, userMutex)
		// retry storing our mutex into the map
		for ; loaded; userMutexInterface, loaded = userIDsInProgress.LoadOrStore(user.ID, userMutex) {
			select { // like mutex.Lock(), but with cancel/deadline
			case <-userMutexInterface.(chan bool): // it is much better than <-time.After(...)
			case <-r.Context().Done():
				return service.ErrUnexpected(r.Context().Err())
			}
		}
		defer userIDsInProgress.Delete(user.ID)

		var refreshToken string
		service.MustNotBeError(
			srv.Store.RefreshTokens().Where("idUser = ?", user.ID).
				PluckFirst("sRefreshToken", &refreshToken).Error())

		// oldToken is invalid since its AccessToken is empty, so the lib will refresh it
		oldToken := &oauth2.Token{RefreshToken: refreshToken}
		oauthConfig := getOAuthConfig(&srv.Config.Auth)
		token, err := oauthConfig.TokenSource(r.Context(), oldToken).Token()
		service.MustNotBeError(err)

		service.MustNotBeError(srv.Store.InTransaction(func(store *database.DataStore) error {
			sessionStore := store.Sessions()
			// delete all the user's access tokens keeping the input token only
			service.MustNotBeError(sessionStore.Delete("idUser = ? AND sAccessToken != ?",
				user.ID, oldAccessToken).Error())
			// insert the new access token
			service.MustNotBeError(sessionStore.InsertNewOAuth(user.ID, token))
			if refreshToken != token.RefreshToken {
				service.MustNotBeError(store.RefreshTokens().Where("idUser = ?", user.ID).
					UpdateColumn("sRefreshToken", token.RefreshToken).Error())
			}
			newToken = token.AccessToken
			expiresIn = int32(time.Until(token.Expiry).Round(time.Second) / time.Second)
			return nil
		}))
	}

	service.MustNotBeError(render.Render(w, r, service.CreationSuccess(map[string]interface{}{
		"access_token": newToken,
		"expires_in":   expiresIn,
	})))

	return service.NoError
}
