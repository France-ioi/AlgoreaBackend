package auth

import (
	"context"

	"github.com/France-ioi/AlgoreaBackend/app/database"
)

// UserFromContext creates a User context from a context set by the middleware
func UserFromContext(context context.Context, store *database.UserStore) *database.User {
	userID := UserIDFromContext(context)
	return database.NewUser(userID, store, nil)
}

// UserIDFromContext returns userID from a context set by the middleware
func UserIDFromContext(context context.Context) int64 {
	return context.Value(ctxUserID).(int64)
}
