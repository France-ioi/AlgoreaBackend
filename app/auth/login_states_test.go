package auth

import (
	"errors"
	"regexp"
	"testing"
	"time"

	"bou.ke/monkey"
	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/app/config"
	"github.com/France-ioi/AlgoreaBackend/app/database"
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
		"INSERT INTO `login_states` (sCookie, sExpirationDate, sState) VALUES (?, NOW() + INTERVAL ? SECOND, ?)",
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
		"INSERT INTO `login_states` (sCookie, sExpirationDate, sState) VALUES (?, NOW() + INTERVAL ? SECOND, ?)",
	)+"$").WithArgs("randomcookie", 7200, "randomstate").
		WillReturnError(&mysql.MySQLError{Number: 1062, Message: "Duplicate entry 'randomcookie' for key 'PRIMARY'"})
	mock.ExpectExec("^"+regexp.QuoteMeta(
		"INSERT INTO `login_states` (sCookie, sExpirationDate, sState) VALUES (?, NOW() + INTERVAL ? SECOND, ?)",
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
