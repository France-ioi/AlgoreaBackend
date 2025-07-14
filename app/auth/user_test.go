package auth

import (
	"context"
	"testing"

	assertlib "github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
)

func TestUserFromContext(t *testing.T) {
	assert := assertlib.New(t)

	myUser := &database.User{GroupID: 8}
	ctx := context.WithValue(context.Background(), ctxUser, myUser)
	user := UserFromContext(ctx)

	assert.NotSame(myUser, user)
	assert.EqualValues(myUser, user)
}

func TestBearerTokenFromContext(t *testing.T) {
	assert := assertlib.New(t)

	expectedToken := "sometoken"
	ctx := context.WithValue(context.Background(), ctxBearer, expectedToken)
	token := BearerTokenFromContext(ctx)

	assert.Equal(expectedToken, token)
}

func TestSessionCookieAttributesFromContext(t *testing.T) {
	assert := assertlib.New(t)

	expectedCookieAttributes := &SessionCookieAttributes{
		UseCookie: true,
		Secure:    true,
		SameSite:  true,
		Domain:    "somedomain.org",
		Path:      "/api/",
	}
	ctx := context.WithValue(context.Background(), ctxSessionCookieAttributes, expectedCookieAttributes)
	cookieAttributes := SessionCookieAttributesFromContext(ctx)

	assert.NotSame(expectedCookieAttributes, cookieAttributes)
	assert.EqualValues(expectedCookieAttributes, cookieAttributes)

	ctx = context.WithValue(context.Background(), ctxSessionCookieAttributes, nil)
	assert.Nil(SessionCookieAttributesFromContext(ctx))
}
