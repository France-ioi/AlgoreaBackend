package auth

import (
	"context"
	"errors"
	"net/http"
	"regexp"
	"testing"
	"time"

	"bou.ke/monkey"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-chi/chi/middleware"
	"github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/app/config"
	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/logging"
	"github.com/France-ioi/AlgoreaBackend/app/loggingtest"
)

func TestCreateLoginState(t *testing.T) {
	counter := -1
	randomStrings := []string{"randomstate", "randomcookie"}
	monkey.Patch(GenerateKey, func() (string, error) {
		counter++
		return randomStrings[counter], nil
	})

	currentTime := time.Now()
	monkey.Patch(time.Now, func() time.Time { return currentTime })
	defer monkey.UnpatchAll()

	db, mock := database.NewDBMock()
	mock.ExpectExec("^"+regexp.QuoteMeta(
		"INSERT INTO `login_states` (`cookie`, `expires_at`, `state`) VALUES (?, NOW() + INTERVAL ? SECOND, ?)",
	)+"$").WithArgs("randomcookie", 7200, "randomstate").WillReturnResult(sqlmock.NewResult(12, 1))

	conf := config.Server{
		Domain:   "backend.algorea.org",
		RootPath: "/in/subdirectory/",
	}
	cookie, state, err := CreateLoginState(database.NewDataStore(db).LoginStates(), &conf)
	assert.NoError(t, err)
	assert.Equal(t, "randomstate", state)

	assert.Equal(t, "login_csrf=randomcookie; Path=/in/subdirectory/; Domain=backend.algorea.org; Expires="+
		currentTime.Add(2*time.Hour).Truncate(time.Second).UTC().Format("Mon, 02 Jan 2006 15:04:05 GMT")+
		"; Max-Age=7200; HttpOnly",
		cookie.String())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateLoginState_HandlesGenerateKeyError(t *testing.T) {
	expectedError := errors.New("some error")
	monkey.Patch(GenerateKey, func() (string, error) {
		return "", expectedError
	})
	defer monkey.UnpatchAll()

	db, mock := database.NewDBMock()

	conf := config.Server{}
	cookie, state, err := CreateLoginState(database.NewDataStore(db).LoginStates(), &conf)
	assert.Equal(t, expectedError, err)
	assert.Equal(t, "", state)
	assert.Nil(t, cookie)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateLoginState_HandlesGenerateKeyErrorForCookie(t *testing.T) {
	expectedError := errors.New("some error")
	counter := -1
	monkey.Patch(GenerateKey, func() (string, error) {
		counter++
		if counter == 0 {
			return "randomstate", nil
		}
		return "", expectedError
	})
	defer monkey.UnpatchAll()

	db, mock := database.NewDBMock()

	conf := config.Server{}
	cookie, state, err := CreateLoginState(database.NewDataStore(db).LoginStates(), &conf)
	assert.Equal(t, expectedError, err)
	assert.Equal(t, "", state)
	assert.Nil(t, cookie)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateLoginState_RetriesOnCollision(t *testing.T) {
	counter := -1
	randomStrings := []string{"randomstate", "randomcookie", "newrandomcookie"}
	monkey.Patch(GenerateKey, func() (string, error) {
		counter++
		return randomStrings[counter], nil
	})

	currentTime := time.Now()
	monkey.Patch(time.Now, func() time.Time { return currentTime })
	defer monkey.UnpatchAll()

	db, mock := database.NewDBMock()
	mock.ExpectExec("^"+regexp.QuoteMeta(
		"INSERT INTO `login_states` (`cookie`, `expires_at`, `state`) VALUES (?, NOW() + INTERVAL ? SECOND, ?)",
	)+"$").WithArgs("randomcookie", 7200, "randomstate").
		WillReturnError(&mysql.MySQLError{Number: 1062, Message: "Duplicate entry 'randomcookie' for key 'PRIMARY'"})
	mock.ExpectExec("^"+regexp.QuoteMeta(
		"INSERT INTO `login_states` (`cookie`, `expires_at`, `state`) VALUES (?, NOW() + INTERVAL ? SECOND, ?)",
	)+"$").WithArgs("newrandomcookie", 7200, "randomstate").WillReturnResult(sqlmock.NewResult(12, 1))

	conf := config.Server{
		Domain:   "backend.algorea.org",
		RootPath: "/in/subdirectory/",
	}
	cookie, state, err := CreateLoginState(database.NewDataStore(db).LoginStates(), &conf)
	assert.NoError(t, err)
	assert.Equal(t, "randomstate", state)

	assert.Equal(t, "login_csrf=newrandomcookie; Path=/in/subdirectory/; Domain=backend.algorea.org; Expires="+
		currentTime.Add(2*time.Hour).Truncate(time.Second).UTC().Format("Mon, 02 Jan 2006 15:04:05 GMT")+
		"; Max-Age=7200; HttpOnly",
		cookie.String())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestLoadLoginState(t *testing.T) {
	expectedError := errors.New("some error")
	tests := []struct {
		name           string
		stateInDB      string
		stateToCompare string
		dbError        error
		expectedResult *LoginState
		expectedError  error
		expectedLogs   string
	}{
		{name: "success", stateInDB: "somestate", stateToCompare: "somestate", expectedResult: &LoginState{ok: true, cookie: "somecookie"}},
		{name: "expired", stateToCompare: "somestate", expectedResult: &LoginState{ok: false}},
		{name: "wrong state", stateInDB: "somestate", stateToCompare: "wrongstate", expectedResult: &LoginState{ok: false},
			expectedLogs: `level=warning msg="Wrong login state"`},
		{name: "db error", stateInDB: "somestate", stateToCompare: "somestate", dbError: expectedError,
			expectedResult: &LoginState{ok: false}, expectedError: expectedError},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			db, mock := database.NewDBMock()
			expectation := mock.ExpectQuery("^" + regexp.QuoteMeta(
				"SELECT state FROM `login_states` WHERE (cookie = ?) AND (expires_at > NOW()) LIMIT 1",
			) + "$").WithArgs("somecookie")
			if test.dbError != nil {
				expectation.WillReturnError(test.dbError)
			} else {
				rowsToReturn := mock.NewRows([]string{"state"})
				if test.stateInDB != "" {
					rowsToReturn = rowsToReturn.AddRow(test.stateInDB)
				}
				expectation.WillReturnRows(rowsToReturn)
			}

			logHook, restoreFunc := logging.MockSharedLoggerHook()
			defer restoreFunc()

			request, err := http.NewRequest("GET", "/", nil)
			assert.NoError(t, err)
			request = request.WithContext(context.WithValue(request.Context(),
				middleware.LogEntryCtxKey, (&logging.StructuredLogger{Logger: logging.SharedLogger.Logger}).NewLogEntry(request)))
			assert.NoError(t, err)
			request.AddCookie(&http.Cookie{Name: "login_csrf", Value: "somecookie"})
			loginState, err := LoadLoginState(database.NewDataStore(db).LoginStates(), request, test.stateToCompare)

			assert.Equal(t, test.expectedError, err)
			assert.Equal(t, test.expectedResult, loginState)
			logs := (&loggingtest.Hook{Hook: logHook}).GetAllStructuredLogs()
			assert.Contains(t, logs, test.expectedLogs)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestLoadLoginState_NoCookie(t *testing.T) {
	db, mock := database.NewDBMock()

	request, err := http.NewRequest("GET", "/", nil)
	assert.NoError(t, err)
	loginState, err := LoadLoginState(database.NewDataStore(db).LoginStates(), request, "somestate")

	assert.NoError(t, err)
	assert.Equal(t, &LoginState{ok: false}, loginState)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestLoginState_IsOK(t *testing.T) {
	assert.False(t, (&LoginState{ok: false}).IsOK())
	assert.True(t, (&LoginState{ok: true}).IsOK())
}

func TestLoginState_Delete(t *testing.T) {
	currentTime := time.Now()
	monkey.Patch(time.Now, func() time.Time { return currentTime })
	defer monkey.UnpatchAll()

	cookieValue := "anothercookie"

	db, mock := database.NewDBMock()
	mock.ExpectExec("^" + regexp.QuoteMeta(
		"DELETE FROM `login_states` WHERE (cookie = ?)",
	) + "$").WithArgs(cookieValue).WillReturnResult(sqlmock.NewResult(-1, 1))

	conf := config.Server{
		Domain:   "backend.algorea.org",
		RootPath: "/in/subdirectory/",
	}

	cookie, err := (&LoginState{ok: true, cookie: cookieValue}).Delete(database.NewDataStore(db).LoginStates(), &conf)
	assert.NoError(t, err)

	assert.Equal(t, "login_csrf=; Path=/in/subdirectory/; Domain=backend.algorea.org; Expires="+
		currentTime.Add(-24*365*time.Hour).Truncate(time.Second).UTC().Format("Mon, 02 Jan 2006 15:04:05 GMT")+
		"; Max-Age=0; HttpOnly", cookie.String())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestLoginState_Delete_InvalidState(t *testing.T) {
	currentTime := time.Now()
	monkey.Patch(time.Now, func() time.Time { return currentTime })
	defer monkey.UnpatchAll()

	db, mock := database.NewDBMock()

	conf := config.Server{}

	cookie, err := (&LoginState{ok: false}).Delete(database.NewDataStore(db).LoginStates(), &conf)
	assert.NoError(t, err)

	assert.Equal(t, "login_csrf=; Expires="+
		currentTime.Add(-24*365*time.Hour).Truncate(time.Second).UTC().Format("Mon, 02 Jan 2006 15:04:05 GMT")+
		"; Max-Age=0; HttpOnly", cookie.String())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestLoginState_Delete_HandlesDBError(t *testing.T) {
	currentTime := time.Now()
	monkey.Patch(time.Now, func() time.Time { return currentTime })
	defer monkey.UnpatchAll()

	cookieValue := "somecookie"
	expectedError := errors.New("some error")

	db, mock := database.NewDBMock()
	mock.ExpectExec("^" + regexp.QuoteMeta(
		"DELETE FROM `login_states` WHERE (cookie = ?)",
	) + "$").WithArgs(cookieValue).WillReturnError(expectedError)

	conf := config.Server{}
	cookie, err := (&LoginState{ok: true, cookie: cookieValue}).Delete(database.NewDataStore(db).LoginStates(), &conf)

	assert.Nil(t, cookie)
	assert.Equal(t, expectedError, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
