package auth

import (
	"context"
	"net/http"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
)

// MockCtxSessionID is a fixed session ID to be used by the mock user middleware.
const MockCtxSessionID = int64(1)

// MockUserMiddleware is a middleware to be used to mock a fixed user in the context.
func MockUserMiddleware(user *database.User) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), ctxBearer, "accesstoken")
			ctx = context.WithValue(ctx, ctxSessionCookieAttributes, &SessionCookieAttributes{})
			ctx = context.WithValue(ctx, ctxUser, user)
			ctx = context.WithValue(ctx, ctxSessionID, MockCtxSessionID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
