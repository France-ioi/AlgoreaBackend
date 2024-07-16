// Package auth provides user authentication.
package auth

import (
	"context"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
)

// UserFromContext retrieves a user from a context set by the middleware.
func UserFromContext(ctx context.Context) *database.User {
	userFromContext := ctx.Value(ctxUser).(*database.User)
	return userFromContext.Clone()
}

// SessionIDFromContext retrieves the session id from a context set by the middleware.
func SessionIDFromContext(ctx context.Context) int64 {
	return ctx.Value(ctxSessionID).(int64)
}

// BearerTokenFromContext retrieves a bearer token from a context set by the middleware.
func BearerTokenFromContext(ctx context.Context) string {
	return ctx.Value(ctxBearer).(string)
}

// SessionCookieAttributesFromContext retrieves session cookie attributes from a context set by the middleware.
// The nil result means the middleware hasn't been called.
func SessionCookieAttributesFromContext(ctx context.Context) *SessionCookieAttributes {
	cookieAttributesFromContext := ctx.Value(ctxSessionCookieAttributes)
	if cookieAttributesFromContext == nil {
		return nil
	}
	attributesCopy := *(cookieAttributesFromContext.(*SessionCookieAttributes))
	return &attributesCopy
}
