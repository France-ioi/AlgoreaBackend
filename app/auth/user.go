package auth

import (
	"context"

	"github.com/France-ioi/AlgoreaBackend/app/database"
)

// UserFromContext creates a User context from a context set by the middleware
func UserFromContext(ctx context.Context, store *database.UserStore) *database.User {
	userID := UserIDFromContext(ctx)
	return database.NewUser(userID, store, nil)
}

// UserIDFromContext returns userID from a context set by the middleware
func UserIDFromContext(ctx context.Context) int64 {
	return ctx.Value(ctxUserID).(int64)
}
