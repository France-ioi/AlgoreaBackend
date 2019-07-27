package auth

import (
	"context"

	"github.com/France-ioi/AlgoreaBackend/app/database"
)

// UserFromContext retrieves a user from a context set by the middleware
func UserFromContext(ctx context.Context) *database.User {
	userFromContext := ctx.Value(ctxUser).(*database.User)
	return userFromContext.Clone()
}

// BearerTokenFromContext retrieves a bearer token from a context set by the middleware
func BearerTokenFromContext(ctx context.Context) string {
	return ctx.Value(ctxBearer).(string)
}
