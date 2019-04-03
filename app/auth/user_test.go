package auth

import (
	"context"
	"testing"

	assertlib "github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/France-ioi/AlgoreaBackend/app/database"
)

type MockUserStore struct {
	mock.Mock
}

func (m *MockUserStore) ByID(userID int64) *database.DB {
	args := m.Called(userID)
	return args.Get(0).(*database.DB)
}

func TestUserFromContext(t *testing.T) {
	assert := assertlib.New(t)

	var userID int64 = 8
	ctx := context.WithValue(context.Background(), ctxUserID, userID)
	store := &MockUserStore{}
	user := UserFromContext(ctx, store)

	assert.EqualValues(userID, user.UserID)
	store.AssertNotCalled(t, "ByID")
}
