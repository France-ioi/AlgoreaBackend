package auth

import (
	"time"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/app/logging"
)

// TemporaryUserSessionLifetimeInSeconds specifies the lifetime of the access token for a temporary user.
const TemporaryUserSessionLifetimeInSeconds = int32(2 * time.Hour / time.Second) // 2 hours (7200 seconds)

// CreateNewTempSession creates a new session for a temporary user.
func CreateNewTempSession(store *database.DataStore, userID int64) (
	accessToken string, expiresIn int32, err error,
) {
	expiresIn = TemporaryUserSessionLifetimeInSeconds

	accessToken, err = GenerateKey()
	mustNotBeError(err)

	// the transaction is needed to retry in case of lock wait timeout or deadlock
	err = store.EnsureTransaction(func(store *database.DataStore) error {
		var sessionID int64
		err = store.RetryOnDuplicateKeyError("sessions", "PRIMARY", "session_id", func(s *database.DataStore) error {
			sessionID = s.NewID()
			// No refresh tokens specified for temporary users.
			return s.Sessions().InsertMap(map[string]interface{}{
				"session_id": sessionID,
				"user_id":    userID,
			})
		})
		if err == nil {
			err = store.AccessTokens().InsertNewToken(sessionID, accessToken, expiresIn)
		}
		return err
	})

	if err == nil {
		logging.EntryFromContext(store.GetContext()).
			Infof("Generated a session token expiring in %d seconds for a temporary user with group_id = %d",
				expiresIn, userID)
		return accessToken, expiresIn, nil
	}

	return "", 0, err
}

// RefreshTempUserSession refreshes the session of a temporary user.
func RefreshTempUserSession(store *database.DataStore, userID, sessionID int64) (accessToken string, expiresIn int32, err error) {
	expiresIn = TemporaryUserSessionLifetimeInSeconds

	accessToken, err = GenerateKey()
	mustNotBeError(err)

	// the transaction is needed to retry the insert in case of lock wait timeout or deadlock
	err = store.EnsureTransaction(func(store *database.DataStore) error {
		return store.AccessTokens().InsertNewToken(sessionID, accessToken, expiresIn)
	})

	if err == nil {
		logging.EntryFromContext(store.GetContext()).
			Infof("Refreshed a session token expiring in %d seconds for a temporary user with group_id = %d",
				expiresIn, userID)
		return accessToken, expiresIn, nil
	}

	return "", 0, err
}

// mustNotBeError panics if the error is not nil.
func mustNotBeError(err error) {
	if err != nil {
		panic(err)
	}
}
