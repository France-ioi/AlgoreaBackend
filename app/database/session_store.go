package database

import (
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
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
func (s *SessionStore) InsertNewOAuth(userID int64, token string, secondsUntilExpiry int32, issuer string,
	cookieAttributes *SessionCookieAttributes) error {
	return s.InsertMap(map[string]interface{}{
		"access_token":     token,
		"expires_at":       gorm.Expr("?  + INTERVAL ? SECOND", Now(), secondsUntilExpiry),
		"user_id":          userID,
		"issuer":           issuer,
		"issued_at":        Now(),
		"use_cookie":       cookieAttributes.UseCookie,
		"cookie_secure":    cookieAttributes.Secure,
		"cookie_same_site": cookieAttributes.SameSite,
		"cookie_domain":    stringOrNil(cookieAttributes.Domain),
		"cookie_path":      stringOrNil(cookieAttributes.Path),
	})
}

func stringOrNil(s string) interface{} {
	if len(s) > 0 {
		return s
	}
	return nil
}
