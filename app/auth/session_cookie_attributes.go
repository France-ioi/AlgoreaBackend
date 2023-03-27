package auth

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

// SessionCookieAttributes represents attributes of the session cookie.
type SessionCookieAttributes struct {
	UseCookie bool
	Secure    bool
	SameSite  bool
	Domain    string
	Path      string
}

// SessionCookie constructs a session cookie with the given attributes.
func (attributes *SessionCookieAttributes) SessionCookie(token string, secondsUntilExpiry int32) *http.Cookie {
	cookieSameSiteMode := http.SameSiteNoneMode
	if attributes.SameSite {
		cookieSameSiteMode = http.SameSiteStrictMode
	}
	return &http.Cookie{
		Name:     "access_token",
		Value:    attributes.marshalCookieValue(token),
		Expires:  time.Now().Add(time.Duration(secondsUntilExpiry) * time.Second),
		MaxAge:   int(secondsUntilExpiry),
		HttpOnly: true,
		Secure:   attributes.Secure,
		SameSite: cookieSameSiteMode,
		Domain:   attributes.Domain,
		Path:     attributes.Path,
	}
}

func (attributes *SessionCookieAttributes) marshalCookieValue(token string) string {
	if token == "" {
		return ""
	}
	// | secure | same_site | code |
	// | 0      | 1         | 1    |
	// | 1      | 0         | 2    |
	// | 1      | 1         | 3    |
	securityKind := '0'
	if attributes.Secure {
		securityKind += 2
	}
	if attributes.SameSite {
		securityKind++
	}
	return fmt.Sprintf("%c!%s!%s!%s", securityKind, token, attributes.Domain, attributes.Path)
}

func unmarshalSessionCookieValue(cookieValue string) (token string, attributes SessionCookieAttributes) {
	parts := strings.Split(cookieValue, "!")
	// should be "kind!value!domain!path"
	if len(parts) != 4 || len(parts[0]) != 1 {
		return
	}
	kind := []rune(parts[0])[0] - '0'
	attributes.UseCookie = true
	// | secure | same_site | code |
	// | 0      | 1         | 1    |
	// | 1      | 0         | 2    |
	// | 1      | 1         | 3    |
	attributes.Secure = (kind & 2) == 2
	attributes.SameSite = (kind & 1) == 1
	token = parts[1]
	attributes.Domain = parts[2]
	attributes.Path = parts[3]
	return token, attributes
}
