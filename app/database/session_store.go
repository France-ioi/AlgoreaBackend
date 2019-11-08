package database

import (
	"golang.org/x/oauth2"
)

// SessionStore implements database operations on `sessions`
type SessionStore struct {
	*DataStore
}

// InsertNewOAuth inserts a new OAuth token for the given user into the DB
func (s *SessionStore) InsertNewOAuth(userID int64, token *oauth2.Token) error {
	return s.InsertMap(map[string]interface{}{
		"access_token": token.AccessToken,
		"expires_at":   token.Expiry.UTC(),
		"user_id":      userID,
		"issuer":       "login-module",
		"issued_at":    Now(),
	})
}
