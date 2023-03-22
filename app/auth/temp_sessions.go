package auth

import (
	"time"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/logging"
)

// TemporaryUserSessionLifetimeInSeconds specifies the lifetime of the access token for a temporary user.
const TemporaryUserSessionLifetimeInSeconds = int32(2 * time.Hour / time.Second) // 2 hours (7200 seconds)

// CreateNewTempSession creates a new session for a temporary user.
func CreateNewTempSession(s *database.SessionStore, userID int64) (
	accessToken string, expiresIn int32, err error) {
	expiresIn = TemporaryUserSessionLifetimeInSeconds

	err = s.RetryOnDuplicatePrimaryKeyError(func(retryStore *database.DataStore) error {
		accessToken, err = GenerateKey()
		if err != nil {
			return err
		}
		return retryStore.Sessions().InsertNewOAuth(userID, accessToken, expiresIn, "backend")
	})
	if err != nil {
		accessToken = ""
		return
	}

	logging.Infof("Generated a session token expiring in %d seconds for a temporary user with group_id = %d",
		expiresIn, userID)

	return
}
