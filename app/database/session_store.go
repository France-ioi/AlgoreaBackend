package database

import (
	"net/http"
	"time"

	"golang.org/x/oauth2"
)

// SessionStore implements database operations on `sessions`
type SessionStore struct {
	*DataStore
}

// SessionCookieAttributes represents attributes of the session cookie
type SessionCookieAttributes struct {
	UseCookie bool
	Secure    bool   `sql:"column:cookie_secure"`
	SameSite  bool   `sql:"column:cookie_same_site"`
	Domain    string `sql:"column:cookie_domain"`
	Path      string `sql:"column:cookie_path"`
}

// SessionCookie constructs a session cookie with the given attributes
func (attributes *SessionCookieAttributes) SessionCookie(token string, secondsUntilExpiry int32) *http.Cookie {
	cookieSameSiteMode := http.SameSiteNoneMode
	if attributes.SameSite {
		cookieSameSiteMode = http.SameSiteStrictMode
	}
	return &http.Cookie{
		Name:     "access_token",
		Value:    token,
		Expires:  time.Now().Add(time.Duration(secondsUntilExpiry) * time.Second),
		MaxAge:   int(secondsUntilExpiry),
		HttpOnly: true,
		Secure:   attributes.Secure,
		SameSite: cookieSameSiteMode,
		Domain:   attributes.Domain,
		Path:     attributes.Path,
	}
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
