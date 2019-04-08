package auth

import (
	"context"
	"net/http"
)

// MockUserIDMiddleware is a middleware to be used to mock a fixed user id in the context
func MockUserIDMiddleware(userID int64) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), ctxUserID, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
