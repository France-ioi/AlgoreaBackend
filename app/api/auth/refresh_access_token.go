package auth

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/jinzhu/gorm"
	"golang.org/x/oauth2"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/app/auth"
	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/logging"
	"github.com/France-ioi/AlgoreaBackend/app/service"
)

type userIDsInProgressMap sync.Map

func (m *userIDsInProgressMap) withLock(userID int64, r *http.Request, f func() error) error {
	userMutex := make(chan bool)
	defer close(userMutex)
	userMutexInterface, loaded := (*sync.Map)(m).LoadOrStore(userID, userMutex)
	// retry storing our mutex into the map
	for ; loaded; userMutexInterface, loaded = (*sync.Map)(m).LoadOrStore(userID, userMutex) {
		select { // like mutex.Lock(), but with cancel/deadline
		case <-userMutexInterface.(chan bool): // it is much better than <-time.After(...)
		case <-r.Context().Done():
			logging.GetLogEntry(r).Warnf("The request is canceled: %s", r.Context().Err())
			return r.Context().Err()
		}
	}
	defer (*sync.Map)(m).Delete(userID)

	return f()
}

var userIDsInProgress userIDsInProgressMap

func (srv *Service) refreshAccessToken(w http.ResponseWriter, r *http.Request) service.APIError {
	user := srv.GetUser(r)
	oldAccessToken := auth.BearerTokenFromContext(r.Context())

	var newToken string
	var expiresIn int32
	apiError := service.NoError

	if user.IsTempUser {
		service.MustNotBeError(srv.Store.InTransaction(func(store *database.DataStore) error {
			sessionStore := store.Sessions()
			// delete all the user's access tokens keeping the input token only
			service.MustNotBeError(sessionStore.Delete("user_id = ? AND access_token != ?",
				user.GroupID, oldAccessToken).Error())
			var err error
			newToken, expiresIn, err = auth.CreateNewTempSession(sessionStore, user.GroupID)
			return err
		}))
	} else {
		// We should not allow concurrency in this part because the login module generates not only
		// a new access token, but also a new refresh token and revokes the old one. We want to prevent
		// usage of the old refresh token for that reason.
		service.MustNotBeError(userIDsInProgress.withLock(user.GroupID, r, func() error {
			newToken, expiresIn, apiError = srv.refreshTokens(r.Context(), user, oldAccessToken)
			return nil
		}))
	}

	if apiError != service.NoError {
		return apiError
	}

	service.MustNotBeError(render.Render(w, r, service.CreationSuccess(map[string]interface{}{
		"access_token": newToken,
		"expires_in":   expiresIn,
	})))

	return service.NoError
}

func (srv *Service) refreshTokens(ctx context.Context, user *database.User, oldAccessToken string) (
	newToken string, expiresIn int32, apiError service.APIError) {
	var refreshToken string
	err := srv.Store.RefreshTokens().Where("user_id = ?", user.GroupID).
		PluckFirst("refresh_token", &refreshToken).Error()
	if gorm.IsRecordNotFoundError(err) {
		logging.Warnf("No refresh token found in the DB for user %d", user.GroupID)
		return "", 0, service.ErrNotFound(errors.New("no refresh token found in the DB for the authenticated user"))
	}
	service.MustNotBeError(err)
	// oldToken is invalid since its AccessToken is empty, so the lib will refresh it
	oldToken := &oauth2.Token{RefreshToken: refreshToken}
	oauthConfig := auth.GetOAuthConfig(srv.AuthConfig)
	token, err := oauthConfig.TokenSource(ctx, oldToken).Token()
	service.MustNotBeError(err)
	service.MustNotBeError(srv.Store.InTransaction(func(store *database.DataStore) error {
		sessionStore := store.Sessions()
		// delete all the user's access tokens keeping the input token only
		service.MustNotBeError(sessionStore.Delete("user_id = ? AND access_token != ?",
			user.GroupID, oldAccessToken).Error())
		// insert the new access token
		service.MustNotBeError(sessionStore.InsertNewOAuth(user.GroupID, token))
		if refreshToken != token.RefreshToken {
			service.MustNotBeError(store.RefreshTokens().Where("user_id = ?", user.GroupID).
				UpdateColumn("refresh_token", token.RefreshToken).Error())
		}
		newToken = token.AccessToken
		expiresIn = int32(time.Until(token.Expiry).Round(time.Second) / time.Second)
		return nil
	}))
	return newToken, expiresIn, service.NoError
}
