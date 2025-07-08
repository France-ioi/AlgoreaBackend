package auth

import (
	"context"
	"errors"
	"net/http"
	"sync"

	"golang.org/x/oauth2"

	"github.com/France-ioi/AlgoreaBackend/v2/app/auth"
	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/app/logging"
	"github.com/France-ioi/AlgoreaBackend/v2/app/service"
)

type sessionIDsInProgressMap sync.Map

func (m *sessionIDsInProgressMap) WithLock(sessionID int64, httpRequest *http.Request, funcToRun func() error) error {
	sessionMutex := make(chan bool)
	defer close(sessionMutex)
	sessionMutexInterface, loaded := (*sync.Map)(m).LoadOrStore(sessionID, sessionMutex)
	// retry storing our mutex into the map
	for ; loaded; sessionMutexInterface, loaded = (*sync.Map)(m).LoadOrStore(sessionID, sessionMutex) {
		select { // like mutex.Lock(), but with cancel/deadline
		case <-sessionMutexInterface.(chan bool): // it is much better than <-time.After(...)
		case <-httpRequest.Context().Done():
			logging.GetLogEntry(httpRequest).Warnf("The request is canceled: %s", httpRequest.Context().Err())
			return httpRequest.Context().Err()
		}
	}
	defer (*sync.Map)(m).Delete(sessionID)

	return funcToRun()
}

var sessionIDsInProgress sessionIDsInProgressMap

func (srv *Service) refreshAccessToken(responseWriter http.ResponseWriter, httpRequest *http.Request) error {
	requestData := httpRequest.Context().Value(parsedRequestData).(map[string]interface{})
	cookieAttributes, _ := srv.resolveCookieAttributes(httpRequest, requestData) // the error has been checked in createAccessToken()

	user := srv.GetUser(httpRequest)
	store := srv.GetStore(httpRequest)
	sessionID := srv.GetSessionID(httpRequest)
	oldAccessToken := auth.BearerTokenFromContext(httpRequest.Context())

	var newToken string
	var expiresIn int32
	var err error

	sessionMostRecentToken, err := store.
		AccessTokens().
		GetMostRecentValidTokenForSession(sessionID)
	service.MustNotBeError(err)

	if sessionMostRecentToken.Token != oldAccessToken || sessionMostRecentToken.TooNewToRefresh {
		// We return the most recent token if the input token is not the most recent one or if it is too new to refresh.
		// Note: we know that the token is valid because we checked it in the middleware.
		newToken = sessionMostRecentToken.Token
		expiresIn = sessionMostRecentToken.SecondsUntilExpiry
	} else {
		if user.IsTempUser {
			newToken, expiresIn, err = auth.RefreshTempUserSession(store, user.GroupID, sessionID)
			service.MustNotBeError(err)
		} else {
			// We should not allow concurrency in this part because the login module generates not only
			// a new access token, but also a new refresh token and revokes the old one. We want to prevent
			// usage of the old refresh token for that reason.
			service.MustNotBeError(sessionIDsInProgress.WithLock(sessionID, httpRequest, func() error {
				newToken, expiresIn, err = srv.refreshTokens(httpRequest.Context(), store, user, sessionID)
				return err
			}))
		}

		service.MustNotBeError(store.AccessTokens().DeleteExpiredTokensOfUser(user.GroupID))
	}

	srv.respondWithNewAccessToken(
		responseWriter, httpRequest, service.CreationSuccess[map[string]interface{}], newToken, expiresIn, cookieAttributes)
	return nil
}

func (srv *Service) refreshTokens(
	ctx context.Context,
	store *database.DataStore,
	user *database.User,
	sessionID int64,
) (newToken string, expiresIn int32, err error) {
	var refreshToken string
	err = store.Sessions().Where("session_id = ?", sessionID).
		PluckFirst("refresh_token", &refreshToken).Error()
	if refreshToken == "" {
		logging.SharedLogger.WithContext(ctx).
			Warnf("No refresh token found in the DB for user %d", user.GroupID)
		return "", 0, service.ErrNotFound(errors.New("no refresh token found in the DB for the authenticated user"))
	}
	service.MustNotBeError(err)
	// oldToken is invalid since its AccessToken is empty, so the lib will refresh it
	oldToken := &oauth2.Token{RefreshToken: refreshToken}
	oauthConfig := auth.GetOAuthConfig(srv.AuthConfig)
	token, err := oauthConfig.TokenSource(ctx, oldToken).Token()

	deleteSessionFunc := func() {
		service.MustNotBeError(store.Sessions().Delete("session_id = ? AND refresh_token = ?", sessionID, refreshToken).Error())
	}

	var retrieveError *oauth2.RetrieveError
	if errors.As(err, &retrieveError) &&
		retrieveError.Response.StatusCode == http.StatusUnauthorized {
		// The refresh token is invalid
		logging.SharedLogger.WithContext(ctx).
			Warnf("The refresh token is invalid for user %d", user.GroupID)

		deleteSessionFunc()
		return "", 0, service.ErrNotFound(errors.New("the refresh token is invalid"))
	}

	service.MustNotBeError(err)

	expiresIn, err = validateAndGetExpiresInFromOAuth2Token(token)
	if err != nil {
		deleteSessionFunc()
		return "", 0, err
	}

	service.MustNotBeError(store.InTransaction(func(store *database.DataStore) error {
		// insert the new access token
		service.MustNotBeError(store.AccessTokens().InsertNewToken(sessionID, token.AccessToken, expiresIn))
		if refreshToken != token.RefreshToken {
			service.MustNotBeError(store.Sessions().
				Where("session_id = ?", sessionID).
				UpdateColumn("refresh_token", token.RefreshToken).
				Error(),
			)
		}

		newToken = token.AccessToken

		return nil
	}))
	return newToken, expiresIn, nil
}
