package auth

import (
	"context"
	"errors"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/sirupsen/logrus/hooks/test"
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
	assert.NotNil(user.store)
	assert.Nil(user.data)
	store.AssertNotCalled(t, "ByID")
}

func TestSelfGroupID(t *testing.T) {
	assert := assertlib.New(t)

	db, dbMock := database.NewDBMock()
	userStore := database.NewDataStore(db).Users()
	dbMock.ExpectQuery("^SELECT").WithArgs(42).WillReturnRows(
		sqlmock.
			NewRows([]string{"idGroupSelf"}).
			FromCSVString("43"),
	)
	user := User{42, userStore, nil}

	assert.EqualValues(43, user.SelfGroupID())
	assert.NotNil(user.data)
}

func TestSelfGroupIDFail(t *testing.T) {
	assert := assertlib.New(t)

	db, dbMock := database.NewDBMock()
	userStore := database.NewDataStore(db).Users()
	dbMock.ExpectQuery("^SELECT").WithArgs(42).WillReturnError(errors.New("db error"))
	user := User{42, userStore, nil}
	hook := test.NewGlobal()

	assert.EqualValues(0, user.SelfGroupID())
	assert.Nil(user.data)
	assert.NotNil(hook.LastEntry())
	assert.Contains(hook.LastEntry().Message, "Unable to load user data")
}
