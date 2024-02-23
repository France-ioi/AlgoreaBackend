package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/jinzhu/gorm"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/logging"
	"github.com/France-ioi/AlgoreaBackend/app/utils"
)

type ctxKey int

const (
	ctxUser ctxKey = iota
	ctxBearer
	ctxSessionCookieAttributes
	ctxSessionID
)

var errCannotValidateAccessToken = errors.New("can't validate the access token")

// GetStorer is an interface allowing to get a data store bound to the context of the given request.
type GetStorer interface {
	GetStore(r *http.Request) *database.DataStore
}

// UserMiddleware is a middleware retrieving a user from the request content.
// It takes the access token from the 'Authorization' header and loads the user info from the DB.
func UserMiddleware(service GetStorer) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestContext, isSuccess, err := ValidatesUserAuthentication(service, w, r)
			if err != nil {
				w.Header().Set("Content-Type", "application/json; charset=utf-8")

				errorCode := GetAuthErrorCodeFromError(err)

				message := "Unauthorized"
				if errorCode == http.StatusInternalServerError {
					message = "Internal server error"
				}

				w.WriteHeader(errorCode)
				_, _ = fmt.Fprintf(w, `{"success":false,"message":"%s","error_text":"%s"}`+"\n", message, utils.Capitalize(err.Error()))
			}

			if !isSuccess {
				return
			}

			next.ServeHTTP(w, r.WithContext(requestContext))
		})
	}
}

// GetAuthErrorCodeFromError gets the HTTP error code.
func GetAuthErrorCodeFromError(err error) int {
	if errors.Is(err, errCannotValidateAccessToken) {
		return http.StatusInternalServerError
	}

	return http.StatusUnauthorized
}

// ValidatesUserAuthentication checks the authentication in the Authorization header and in the "access_token" cookie.
// It returns:
//   - A request context with the user authenticated on success
//   - Whether the authentication was a success
//   - The reason why the user couldn't be authenticated
func ValidatesUserAuthentication(service GetStorer, w http.ResponseWriter, r *http.Request) (context.Context, bool, error) {
	var user database.User
	var sessionID int64
	var authorized bool
	var err error

	accessToken, cookieAttributes := ParseSessionCookie(r)

	for _, authValue := range r.Header["Authorization"] {
		parsedAuthValue := strings.SplitN(authValue, " ", 3)
		// credentials = "Bearer" 1*SP b64token (see https://tools.ietf.org/html/rfc6750#section-2.1)
		if len(parsedAuthValue) == 2 && parsedAuthValue[0] == "Bearer" {
			accessToken = parsedAuthValue[1]
			// Delete the cookie since the Authorization header is given
			deleteSessionCookie(w, &cookieAttributes)
			break
		}
	}

	if accessToken == "" {
		return r.Context(), false, errors.New("no access token provided")
	}

	if len(accessToken) <= 2000 {
		user, sessionID, err = service.GetStore(r).Sessions().GetUserAndSessionIDByValidAccessToken(accessToken)
		authorized = err == nil
		if err != nil && !gorm.IsRecordNotFoundError(err) {
			logging.Errorf("Can't validate an access token: %s", err)

			return r.Context(), false, errCannotValidateAccessToken
		}
	}

	if !authorized {
		return r.Context(), false, errors.New("invalid access token")
	}

	ctx := context.WithValue(r.Context(), ctxBearer, accessToken)
	ctx = context.WithValue(ctx, ctxSessionCookieAttributes, &cookieAttributes)
	ctx = context.WithValue(ctx, ctxUser, &user)
	ctx = context.WithValue(ctx, ctxSessionID, sessionID)

	return ctx, true, nil
}

// ParseSessionCookie parses the 'access_token' cookie (if given) and returns the access token among with cookie attributes.
func ParseSessionCookie(r *http.Request) (accessToken string, cookieAttributes SessionCookieAttributes) {
	if cookie, cookieErr := r.Cookie("access_token"); cookieErr == nil {
		accessToken, cookieAttributes = unmarshalSessionCookieValue(cookie.Value)
	}
	return accessToken, cookieAttributes
}

func deleteSessionCookie(w http.ResponseWriter, cookieAttributes *SessionCookieAttributes) {
	if cookieAttributes.UseCookie {
		http.SetCookie(w, cookieAttributes.SessionCookie("", -1000))
		*cookieAttributes = SessionCookieAttributes{}
	}
}
