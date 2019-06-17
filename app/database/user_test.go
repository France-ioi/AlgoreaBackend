package database

import (
	"errors"
	"reflect"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

// NewMockUser creates a mock user, to be used for testing
func NewMockUser(id int64, userData *UserData) *User {
	if userData != nil {
		userData.ID = id
	}
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

func TestUser_Load(t *testing.T) {
	db, dbMock := NewDBMock()
	defer func() { _ = db.Close() }()

	userStore := NewDataStore(db).Users()
	dbMock.ExpectQuery(expectedLazyLoadDataQueryRegexp).WithArgs(42).WillReturnRows(
		sqlmock.
			NewRows([]string{"ID"}).AddRow(int64(43)),
	)
	user := User{42, userStore, nil}

	err := user.Load()
	assert.NotNil(t, user.data)
	assert.EqualValues(t, 43, user.data.ID)
	assert.NoError(t, err)
}

func TestUser_Load_Fail(t *testing.T) {
	db, dbMock := NewDBMock()
	defer func() { _ = db.Close() }()

	expectedError := errors.New("db error")
	userStore := NewDataStore(db).Users()
	dbMock.ExpectQuery(expectedLazyLoadDataQueryRegexp).WithArgs(42).WillReturnError(expectedError)
	user := User{42, userStore, nil}

	err := user.Load()
	assert.Equal(t, expectedError, err)
	assert.Nil(t, user.data)
}

func TestUser_Method(t *testing.T) {
	currentTime := time.Now().UTC()

	tests := []struct {
		name          string
		methodToCall  string
		dbColumn      string
		dbValue       interface{}
		expectedValue interface{}
	}{
		{methodToCall: "SelfGroupID", dbColumn: "idGroupSelf", dbValue: int64(43), expectedValue: int64(43)},
		{methodToCall: "DefaultLanguageID", dbColumn: "idDefaultLanguage", dbValue: int64(2), expectedValue: int64(2)},
		{methodToCall: "OwnedGroupID", dbColumn: "idGroupOwned", dbValue: int64(44), expectedValue: int64(44)},
		{name: "true", methodToCall: "AllowSubgroups", dbColumn: "allowSubgroups", dbValue: int64(1), expectedValue: true},
		{name: "false", methodToCall: "AllowSubgroups", dbColumn: "allowSubgroups", dbValue: int64(0), expectedValue: false},
		{name: "time", methodToCall: "NotificationReadDate", dbColumn: "sNotificationReadDate",
			dbValue: &currentTime, expectedValue: &currentTime},
		{name: "nil", methodToCall: "NotificationReadDate", dbColumn: "sNotificationReadDate",
			dbValue: nil, expectedValue: (*time.Time)(nil)},
	}
	for _, testCase := range tests {
		testCase := testCase
		t.Run(strings.Join([]string{testCase.methodToCall, testCase.name}, " "), func(t *testing.T) {
			db, dbMock := NewDBMock()
			defer func() { _ = db.Close() }()

			userStore := NewDataStore(db).Users()
			dbMock.ExpectQuery(expectedLazyLoadDataQueryRegexp).WithArgs(42).WillReturnRows(
				sqlmock.
					NewRows([]string{testCase.dbColumn}).AddRow(testCase.dbValue),
			)
			user := &User{42, userStore, nil}

			results := reflect.ValueOf(user).MethodByName(testCase.methodToCall).Call(nil)
			got, gotErr := results[0].Interface(), results[1].Interface()
			var err error
			if gotErr == nil {
				err = error(nil)
			} else {
				err = gotErr.(error)
			}
			assert.EqualValues(t, testCase.expectedValue, got)
			assert.NoError(t, err)
			assert.NotNil(t, user.data)
		})
	}
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

func TestUser_AllowSubgroups_Fail(t *testing.T) {
	db, dbMock := NewDBMock()
	defer func() { _ = db.Close() }()

	userStore := NewDataStore(db).Users()
	expectedError := errors.New("db error")
	dbMock.ExpectQuery(expectedLazyLoadDataQueryRegexp).WithArgs(42).WillReturnError(expectedError)
	user := User{42, userStore, nil}

	got, err := user.AllowSubgroups()
	assert.False(t, got)
	assert.Equal(t, expectedError, err)
	assert.Nil(t, user.data)
}

func TestUser_AllowSubgroups_UserNotFound(t *testing.T) {
	testMethodHandlesUserNotFoundError(t, func(db *DB, user *User) []interface{} {
		got, err := user.AllowSubgroups()
		return []interface{}{got, err}
	}, []interface{}{false, ErrUserNotFound})
}

func TestUser_NotificationReadDate_Fail(t *testing.T) {
	db, dbMock := NewDBMock()
	defer func() { _ = db.Close() }()

	userStore := NewDataStore(db).Users()
	expectedError := errors.New("db error")
	dbMock.ExpectQuery(expectedLazyLoadDataQueryRegexp).WithArgs(42).WillReturnError(expectedError)
	user := User{42, userStore, nil}

	got, err := user.NotificationReadDate()
	assert.Nil(t, got)
	assert.Equal(t, expectedError, err)
	assert.Nil(t, user.data)
}

func TestUser_NotificationReadDate_UserNotFound(t *testing.T) {
	testMethodHandlesUserNotFoundError(t, func(db *DB, user *User) []interface{} {
		got, err := user.NotificationReadDate()
		return []interface{}{got, err}
	}, []interface{}{(*time.Time)(nil), ErrUserNotFound})
}

func testMethodHandlesUserNotFoundError(t *testing.T,
	funcToCall func(*DB, *User) []interface{}, expectedResults []interface{}) {
	db, dbMock := NewDBMock()
	defer func() { _ = db.Close() }()

	userStore := NewDataStore(db).Users()
	user := User{42, userStore, nil}

	dbMock.ExpectQuery(expectedLazyLoadDataQueryRegexp).WithArgs(42).
		WillReturnRows(sqlmock.NewRows([]string{"ID"}))

	results := funcToCall(db, &user)
	assert.Equal(t, expectedResults, results)
}
