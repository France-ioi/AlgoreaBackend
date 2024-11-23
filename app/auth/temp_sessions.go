package auth

import (
	"time"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/app/logging"
	"github.com/France-ioi/AlgoreaBackend/v2/app/rand"
)

// TemporaryUserSessionLifetimeInSeconds specifies the lifetime of the access token for a temporary user.
const TemporaryUserSessionLifetimeInSeconds = int32(2 * time.Hour / time.Second) // 2 hours (7200 seconds)

// CreateNewTempSession creates a new session for a temporary user.
func CreateNewTempSession(s *database.DataStore, userID int64) (
	accessToken string, expiresIn int32, err error,
) {
	expiresIn = TemporaryUserSessionLifetimeInSeconds

	accessToken, err = GenerateKey()
	mustNotBeError(err)

	sessionID := rand.Int63()
	mustNotBeError(s.
		// No refresh tokens specified for temporary users.
		Exec("INSERT INTO sessions (session_id, user_id) VALUES (?, ?)", sessionID, userID).
		Error(),
	)

	mustNotBeError(s.AccessTokens().InsertNewToken(sessionID, accessToken, expiresIn))

	logging.SharedLogger.WithContext(s.GetContext()).
		Infof("Generated a session token expiring in %d seconds for a temporary user with group_id = %d",
			expiresIn, userID)

	return
}

// RefreshTempUserSession refreshes the session of a temporary user.
func RefreshTempUserSession(s *database.DataStore, userID, sessionID int64) (accessToken string, expiresIn int32, err error) {
	expiresIn = TemporaryUserSessionLifetimeInSeconds

	accessToken, err = GenerateKey()
	mustNotBeError(err)

	mustNotBeError(s.AccessTokens().InsertNewToken(sessionID, accessToken, expiresIn))

	logging.SharedLogger.WithContext(s.GetContext()).
		Infof("Refreshed a session token expiring in %d seconds for a temporary user with group_id = %d",
			expiresIn, userID)

	return
}

// mustNotBeError panics if the error is not nil.
func mustNotBeError(err error) {
	if err != nil {
		panic(err)
	}
}
