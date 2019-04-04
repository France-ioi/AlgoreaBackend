package database

import (
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

// NewMockUser creates a mock user, to be used for testing
//func NewMockUser(id, selfGroupID, ownedGroupID, defaultLanguageID int64) *User {
func NewMockUser(id int64, userData *UserData) *User {
	userData.ID = id
	return NewUser(id, nil, userData)
}

func TestMockUser(t *testing.T) {
	user := NewMockUser(1, &UserData{SelfGroupID: 2, OwnedGroupID: 3, DefaultLanguageID: 4})
	assert.Equal(t, int64(1), user.UserID)
	assert.Equal(t, int64(1), user.data.ID)
	assert.Equal(t, int64(2), user.data.SelfGroupID)
	assert.Equal(t, int64(3), user.data.OwnedGroupID)
	assert.Equal(t, int64(4), user.data.DefaultLanguageID)
}

var expectedLazyLoadDataQueryRegexp = "^" + regexp.QuoteMeta(
	"SELECT users.*, l.ID as idDefaultLanguage FROM `users` LEFT JOIN languages l ON (users.sDefaultLanguage = l.sCode) WHERE (users.ID = ?)",
) + "$"

func TestUser_SelfGroupID(t *testing.T) {
	db, dbMock := NewDBMock()
	defer func() { _ = db.Close() }()

	userStore := NewDataStore(db).Users()
	dbMock.ExpectQuery(expectedLazyLoadDataQueryRegexp).WithArgs(42).WillReturnRows(
		sqlmock.
			NewRows([]string{"idGroupSelf"}).AddRow(int64(43)),
	)
	user := User{42, userStore, nil}

	got, err := user.SelfGroupID()
	assert.EqualValues(t, 43, got)
	assert.NoError(t, err)
	assert.NotNil(t, user.data)
}

func TestUser_SelfGroupID_Fail(t *testing.T) {
	db, dbMock := NewDBMock()
	defer func() { _ = db.Close() }()

	expectedError := errors.New("db error")
	userStore := NewDataStore(db).Users()
	dbMock.ExpectQuery(expectedLazyLoadDataQueryRegexp).WithArgs(42).WillReturnError(expectedError)
	user := User{42, userStore, nil}

	got, err := user.SelfGroupID()
	assert.EqualValues(t, 0, got)
	assert.Equal(t, expectedError, err)
	assert.Nil(t, user.data)
}

func TestUser_DefaultLanguageID(t *testing.T) {
	db, dbMock := NewDBMock()
	defer func() { _ = db.Close() }()

	userStore := NewDataStore(db).Users()
	dbMock.ExpectQuery(expectedLazyLoadDataQueryRegexp).WithArgs(42).WillReturnRows(
		sqlmock.
			NewRows([]string{"idDefaultLanguage"}).AddRow(int64(2)),
	)
	user := User{42, userStore, nil}

	got, err := user.DefaultLanguageID()
	assert.EqualValues(t, 2, got)
	assert.NoError(t, err)
	assert.NotNil(t, user.data)
}

func TestUser_DefaultLanguageID_Fail(t *testing.T) {
	db, dbMock := NewDBMock()
	defer func() { _ = db.Close() }()

	expectedError := errors.New("db error")
	userStore := NewDataStore(db).Users()
	dbMock.ExpectQuery(expectedLazyLoadDataQueryRegexp).WithArgs(42).WillReturnError(expectedError)
	user := User{42, userStore, nil}

	got, err := user.DefaultLanguageID()
	assert.Zero(t, got)
	assert.Equal(t, expectedError, err)
	assert.Nil(t, user.data)
}

func TestUser_OwnedGroupID(t *testing.T) {
	db, dbMock := NewDBMock()
	defer func() { _ = db.Close() }()

	userStore := NewDataStore(db).Users()
	dbMock.ExpectQuery(expectedLazyLoadDataQueryRegexp).WithArgs(42).WillReturnRows(
		sqlmock.
			NewRows([]string{"idGroupOwned"}).AddRow(int64(43)),
	)
	user := User{42, userStore, nil}

	got, err := user.OwnedGroupID()
	assert.EqualValues(t, 43, got)
	assert.NoError(t, err)
	assert.NotNil(t, user.data)
}

func TestUser_OwnedGroupID_Fail(t *testing.T) {
	db, dbMock := NewDBMock()
	defer func() { _ = db.Close() }()

	expectedError := errors.New("db error")
	userStore := NewDataStore(db).Users()
	dbMock.ExpectQuery(expectedLazyLoadDataQueryRegexp).WithArgs(42).WillReturnError(expectedError)
	user := User{42, userStore, nil}

	got, err := user.OwnedGroupID()
	assert.EqualValues(t, 0, got)
	assert.Equal(t, expectedError, err)
	assert.Nil(t, user.data)
}

func TestUser_AllowSubgroups(t *testing.T) {
	db, dbMock := NewDBMock()
	defer func() { _ = db.Close() }()

	userStore := NewDataStore(db).Users()
	dbMock.ExpectQuery(expectedLazyLoadDataQueryRegexp).WithArgs(42).
		WillReturnRows(sqlmock.NewRows([]string{"allowSubgroups"}).AddRow(int64(1)))
	user := User{42, userStore, nil}

	got, err := user.AllowSubgroups()
	assert.True(t, got)
	assert.NoError(t, err)
	assert.NotNil(t, user.data)
}

func TestUser_AllowSubgroups_Fail(t *testing.T) {
	db, dbMock := NewDBMock()

	userStore := NewDataStore(db).Users()
	expectedError := errors.New("db error")
	dbMock.ExpectQuery(expectedLazyLoadDataQueryRegexp).WithArgs(42).WillReturnError(expectedError)
	user := User{42, userStore, nil}

	got, err := user.AllowSubgroups()
	assert.False(t, got)
	assert.Equal(t, expectedError, err)
	assert.Nil(t, user.data)
}

func TestUser_AllowSubgroups_False(t *testing.T) {
	db, dbMock := NewDBMock()
	userStore := NewDataStore(db).Users()
	dbMock.ExpectQuery(expectedLazyLoadDataQueryRegexp).WithArgs(42).
		WillReturnRows(sqlmock.NewRows([]string{"allowSubgroups"}).AddRow(int64(0)))
	user := User{42, userStore, nil}

	got, err := user.AllowSubgroups()
	assert.False(t, got)
	assert.NoError(t, err)
	assert.NotNil(t, user.data)
}

func TestUser_AllowSubgroups_UserNotFound(t *testing.T) {
	db, dbMock := NewDBMock()
	userStore := NewDataStore(db).Users()
	dbMock.ExpectQuery(expectedLazyLoadDataQueryRegexp).WithArgs(42).
		WillReturnRows(sqlmock.NewRows([]string{"allowSubgroups"}))
	user := User{42, userStore, nil}

	got, err := user.AllowSubgroups()
	assert.False(t, got)
	assert.Equal(t, ErrUserNotFound, err)
	assert.Nil(t, user.data)
}
