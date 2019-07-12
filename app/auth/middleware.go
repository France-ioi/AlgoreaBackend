package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/jinzhu/gorm"

	"github.com/France-ioi/AlgoreaBackend/app/appenv"
	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/logging"
)

type ctxKey int

const (
	ctxUserID ctxKey = iota
)

// UserIDMiddleware is a middleware retrieving user ID from the request content.
// It takes the access token from the 'Authorization' header and loads the user ID from the 'sessions' table
func UserIDMiddleware(sessionStore *database.SessionStore) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			var accessToken string
			var userID int64
			var authorized bool
			for _, authValue := range r.Header["Authorization"] {
				parsedAuthValue := strings.SplitN(authValue, " ", 3)
				// credentials = "Bearer" 1*SP b64token (see https://tools.ietf.org/html/rfc6750#section-2.1)
				if len(parsedAuthValue) == 2 && parsedAuthValue[0] == "Bearer" {
					accessToken = parsedAuthValue[1]
					break
				}
			}

			if accessToken == "" {
				http.Error(w, "No access token provided", http.StatusUnauthorized)
				return
			}

			if len(accessToken) < 255 {
				err := sessionStore.Where("sAccessToken = ?", accessToken).
					Where("sExpirationDate > NOW()").PluckFirst("idUser", &userID).
					Error()
				authorized = err == nil
				if err != nil && !gorm.IsRecordNotFoundError(err) {
					logging.Errorf("Can't validate an access token: %s", err)
					http.Error(w, "Can't validate the access token", http.StatusBadGateway)
					return
				}
			}

			if !authorized {
				http.Error(w, "The access token has expired", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), ctxUserID, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// SetAuthorizationHeaderFromQueryMiddleware is a middleware that copies an access token
// given as a query parameter into the "Authorization" header.
// This middleware is for development purposes only. Don not use it in prod!
func SetAuthorizationHeaderFromQueryMiddleware() func(next http.Handler) http.Handler {
	if appenv.IsEnvProd() {
		panic("SetAuthorizationHeaderFromQueryMiddleware should not be used in production")
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if accessToken := r.URL.Query().Get("access_token"); accessToken != "" {
				r.Header.Set("Authorization", "Bearer "+accessToken)
			}
			next.ServeHTTP(w, r)
		})
	}
}
