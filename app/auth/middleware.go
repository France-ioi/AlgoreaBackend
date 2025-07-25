package auth

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/jinzhu/gorm"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/app/logging"
)

type ctxKey int

const (
	ctxUser ctxKey = iota
	ctxBearer
	ctxSessionCookieAttributes
	ctxSessionID
)

// GetStorer is an interface allowing to get a data store bound to the context of the given request.
type GetStorer interface {
	GetStore(r *http.Request) *database.DataStore
}

// UserMiddleware is a middleware retrieving a user from the request content.
// It takes the access token from the 'Authorization' header and loads the user info from the DB.
func UserMiddleware(service GetStorer) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(responseWriter http.ResponseWriter, httpRequest *http.Request) {
			requestContext, isSuccess, reason, err := ValidatesUserAuthentication(service, responseWriter, httpRequest)
			if err != nil || !isSuccess {
				responseWriter.Header().Set("Content-Type", "application/json; charset=utf-8")
				if err != nil {
					responseWriter.WriteHeader(http.StatusInternalServerError)
					_, _ = fmt.Fprintf(responseWriter, `{"success":false,"message":"Internal server error"}`+"\n")
					return
				}
				responseWriter.WriteHeader(http.StatusUnauthorized)
				_, _ = fmt.Fprintf(responseWriter, `{"success":false,"message":"Unauthorized","error_text":"%s"}`+"\n", reason)
				return
			}

			next.ServeHTTP(responseWriter, httpRequest.WithContext(requestContext))
		})
	}
}

// ValidatesUserAuthentication checks the authentication in the Authorization header and in the "access_token" cookie.
// It returns:
//   - A request context with the user authenticated on success
//   - Whether the authentication was a success
//   - The reason why the user couldn't be authenticated
func ValidatesUserAuthentication(service GetStorer, responseWriter http.ResponseWriter, httpRequest *http.Request) (
	ctx context.Context, authorized bool, reason string, err error,
) {
	var user database.User
	var sessionID int64

	accessToken, cookieAttributes := ParseSessionCookie(httpRequest)

	for _, authValue := range httpRequest.Header["Authorization"] {
		//nolint:mnd // credentials = "Bearer" 1*SP b64token (see https://tools.ietf.org/html/rfc6750#section-2.1)
		parsedAuthValue := strings.SplitN(authValue, " ", 3)
		if len(parsedAuthValue) == 2 && parsedAuthValue[0] == "Bearer" {
			accessToken = parsedAuthValue[1]
			// Delete the cookie since the Authorization header is given
			deleteSessionCookie(responseWriter, &cookieAttributes)
			break
		}
	}

	if accessToken == "" {
		return httpRequest.Context(), false, "No access token provided", nil
	}

	if len(accessToken) > database.AccessTokenMaxLength {
		authorized = false
	} else {
		user, sessionID, err = service.GetStore(httpRequest).Sessions().GetUserAndSessionIDByValidAccessToken(accessToken)
		authorized = err == nil
		if err != nil && !gorm.IsRecordNotFoundError(err) {
			logging.SharedLogger.WithContext(httpRequest.Context()).Errorf("Can't validate an access token: %s", err)

			return httpRequest.Context(), false, "", err
		}
	}

	if !authorized {
		return httpRequest.Context(), false, "Invalid access token", nil
	}

	ctx = context.WithValue(httpRequest.Context(), ctxBearer, accessToken)
	ctx = context.WithValue(ctx, ctxSessionCookieAttributes, &cookieAttributes)
	ctx = context.WithValue(ctx, ctxUser, &user)
	ctx = context.WithValue(ctx, ctxSessionID, sessionID)

	logging.LogEntrySetField(httpRequest, "user_id", user.GroupID)

	return ctx, true, "", nil
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
