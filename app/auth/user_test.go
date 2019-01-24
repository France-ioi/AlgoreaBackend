package auth

import (
	"context"
	"testing"

	assert_lib "github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	sqlmock "gopkg.in/DATA-DOG/go-sqlmock.v1"

	"github.com/France-ioi/AlgoreaBackend/app/database"
)

type MockUserStore struct {
	mock.Mock
}

func (m *MockUserStore) ByID(userID int64) database.DB {
	args := m.Called(userID)
	return args.Get(0).(database.DB)
}

func TestUserFromContext(t *testing.T) {
	assert := assert_lib.New(t)

	var userID int64 = 8
	ctx := context.WithValue(context.Background(), ctxUserID, userID)
	store := &MockUserStore{}
	user := UserFromContext(ctx, store)

	assert.EqualValues(userID, user.UserID)
	assert.NotNil(user.store)
	assert.Nil(user.data)
	store.AssertNotCalled(t, "ByID")
}

func TestSelfGroupID(t *testing.T) {
	assert := assert_lib.New(t)

	db, dbMock := database.NewDBMock()
	userStore := &database.UserStore{DataStore: &database.DataStore{DB: db}}
	dbMock.ExpectQuery("^SELECT").WithArgs(42).WillReturnRows(
		sqlmock.
			NewRows([]string{"idGroupSelf"}).
			FromCSVString("43"),
	)
	user := User{42, userStore, nil}

	assert.EqualValues(43, user.SelfGroupID())
	assert.NotNil(user.data)
}
