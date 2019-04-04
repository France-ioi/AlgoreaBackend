package auth

import (
	"context"
	"testing"

	assertlib "github.com/stretchr/testify/assert"
)

func TestUserFromContext(t *testing.T) {
	assert := assertlib.New(t)

	var userID int64 = 8
	ctx := context.WithValue(context.Background(), ctxUserID, userID)
	user := UserFromContext(ctx, nil)

	assert.EqualValues(userID, user.UserID)
}
