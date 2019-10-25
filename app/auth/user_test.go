package auth

import (
	"context"
	"testing"

	assertlib "github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/app/database"
)

func TestUserFromContext(t *testing.T) {
	assert := assertlib.New(t)

	myUser := &database.User{GroupID: 8}
	ctx := context.WithValue(context.Background(), ctxUser, myUser)
	user := UserFromContext(ctx)

	assert.False(myUser == user)
	assert.EqualValues(myUser, user)
}

func TestBearerTokenFromContext(t *testing.T) {
	assert := assertlib.New(t)

	expectedToken := "sometoken"
	ctx := context.WithValue(context.Background(), ctxBearer, expectedToken)
	token := BearerTokenFromContext(ctx)

	assert.Equal(expectedToken, token)
}
