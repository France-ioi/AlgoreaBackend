package auth

import (
	"context"
	"errors"
	"regexp"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
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

var expectedLazyLoadDataQueryRegexp = "^" + regexp.QuoteMeta(
	"SELECT users.*, l.ID as idDefaultLanguage FROM `users` LEFT JOIN languages l ON (users.sDefaultLanguage = l.sCode) WHERE (users.ID = ?)",
) + "$"

func TestSelfGroupID(t *testing.T) {
	assert := assertlib.New(t)

	db, dbMock := database.NewDBMock()
	userStore := database.NewDataStore(db).Users()
	dbMock.ExpectQuery(expectedLazyLoadDataQueryRegexp).WithArgs(42).WillReturnRows(
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
	dbMock.ExpectQuery(expectedLazyLoadDataQueryRegexp).WithArgs(42).WillReturnError(errors.New("db error"))
	user := User{42, userStore, nil}

	assert.EqualValues(0, user.SelfGroupID())
	assert.Nil(user.data)
}

func TestUser_AllowSubgroups(t *testing.T) {
	assert := assertlib.New(t)

	db, dbMock := database.NewDBMock()
	userStore := database.NewDataStore(db).Users()
	dbMock.ExpectQuery(expectedLazyLoadDataQueryRegexp).WithArgs(42).
		WillReturnRows(sqlmock.NewRows([]string{"allowSubgroups"}).AddRow(int64(1)))
	user := User{42, userStore, nil}

	assert.True(user.AllowSubgroups())
	assert.NotNil(user.data)
}

func TestUser_AllowSubgroups_Fail(t *testing.T) {
	assert := assertlib.New(t)

	db, dbMock := database.NewDBMock()
	userStore := database.NewDataStore(db).Users()
	dbMock.ExpectQuery(expectedLazyLoadDataQueryRegexp).WithArgs(42).WillReturnError(errors.New("db error"))
	user := User{42, userStore, nil}

	assert.False(user.AllowSubgroups())
	assert.Nil(user.data)
}

func TestUser_AllowSubgroups_False(t *testing.T) {
	assert := assertlib.New(t)

	db, dbMock := database.NewDBMock()
	userStore := database.NewDataStore(db).Users()
	dbMock.ExpectQuery(expectedLazyLoadDataQueryRegexp).WithArgs(42).
		WillReturnRows(sqlmock.NewRows([]string{"allowSubgroups"}).AddRow(int64(0)))
	user := User{42, userStore, nil}

	assert.False(user.AllowSubgroups())
	assert.NotNil(user.data)
}
