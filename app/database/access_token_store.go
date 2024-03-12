package database

import "github.com/jinzhu/gorm"

const AccessTokenMaxLength = 2000

// AccessTokenStore implements database operations on `access_tokens`.
type AccessTokenStore struct {
	*DataStore
}

// InsertNewToken inserts a new OAuth token for the given sessionID into the DB.
func (s *AccessTokenStore) InsertNewToken(sessionID int64, token string, secondsUntilExpiry int32) error {
	return s.InsertMap(map[string]interface{}{
		"session_id": sessionID,
		"token":      token,
		"expires_at": gorm.Expr("?  + INTERVAL ? SECOND", Now(), secondsUntilExpiry),
		"issued_at":  Now(),
	})
}
