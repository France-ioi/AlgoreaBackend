package auth

import (
	"context"
	"net/http"

	"github.com/France-ioi/AlgoreaBackend/app/database"
)

// MockUserMiddleware is a middleware to be used to mock a fixed user in the context
func MockUserMiddleware(user *database.User) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), ctxBearer, "accesstoken")
			ctx = context.WithValue(ctx, ctxSessionCookieAttributes, &SessionCookieAttributes{})
			ctx = context.WithValue(ctx, ctxUser, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
