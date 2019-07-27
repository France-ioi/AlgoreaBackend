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
		"sAccessToken":    token.AccessToken,
		"sExpirationDate": token.Expiry.UTC(),
		"idUser":          userID,
		"sIssuer":         "login-module",
		"sIssuedAtDate":   Now(),
	})
}
