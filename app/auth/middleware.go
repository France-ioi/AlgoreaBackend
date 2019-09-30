package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/jinzhu/gorm"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/logging"
)

type ctxKey int

const (
	ctxUser ctxKey = iota
	ctxBearer
)

// UserMiddleware is a middleware retrieving a user from the request content.
// It takes the access token from the 'Authorization' header and loads the user info from the DB
func UserMiddleware(sessionStore *database.SessionStore) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			var accessToken string
			var user database.User
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
				w.Header().Set("Content-Type", "application/json; charset=utf-8")
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = w.Write([]byte(`{"success":false,"message":"Unauthorized","error_text":"No access token provided"}` + "\n"))
				return
			}

			if len(accessToken) <= 2000 {
				err := sessionStore.
					Select(`
						users.id, users.login, users.is_admin, users.self_group_id, users.owned_group_id, users.access_group_id,
						users.temp_user, users.allow_subgroups, users.notifications_read_at,
						users.default_language, l.id as default_language_id`).
					Joins("JOIN users ON users.id = sessions.user_id").
					Joins("LEFT JOIN languages l ON users.default_language = l.code").
					Where("access_token = ?", accessToken).
					Where("expires_at > NOW()").Take(&user).
					Error()
				authorized = err == nil
				if err != nil && !gorm.IsRecordNotFoundError(err) {
					logging.Errorf("Can't validate an access token: %s", err)
					w.Header().Set("Content-Type", "application/json; charset=utf-8")
					w.WriteHeader(http.StatusInternalServerError)
					_, _ = w.Write([]byte(`{"success":false,"message":"Internal server error","error_text":"Can't validate the access token"}` + "\n"))
					return
				}
			}

			if !authorized {
				w.Header().Set("Content-Type", "application/json; charset=utf-8")
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = w.Write([]byte(`{"success":false,"message":"Unauthorized","error_text":"Invalid access token"}` + "\n"))
				return
			}

			ctx := context.WithValue(r.Context(), ctxBearer, accessToken)
			ctx = context.WithValue(ctx, ctxUser, &user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
