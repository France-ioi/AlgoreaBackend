package auth

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"bou.ke/monkey"
	"github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/logging"
	"github.com/France-ioi/AlgoreaBackend/app/loggingtest"
)

func TestCreateNewTempSession(t *testing.T) {
	expectedAccessToken := "tmp-01abcdefghijklmnopqrstuvwxyz"
	monkey.Patch(GenerateKey, func() (string, error) { return expectedAccessToken, nil })
	defer monkey.UnpatchAll()
	logHook, restoreFunc := logging.MockSharedLoggerHook()
	defer restoreFunc()

	db, mock := database.NewDBMock()
	defer func() { _ = db.Close() }()

	expectedUserID := int64(12345)
	expectedCookieAttributes := &database.SessionCookieAttributes{}
	sessionStore := database.NewDataStore(db).Sessions()
	patch := monkey.PatchInstanceMethod(reflect.TypeOf(sessionStore), "InsertNewOAuth",
		func(sessionStore *database.SessionStore, userID int64, token string, expiresIn int32, issuer string,
			cookieAttributes *database.SessionCookieAttributes) error {
			assert.Equal(t, expectedUserID, userID)
			assert.Equal(t, token, expectedAccessToken)
			assert.Equal(t, int32(2*60*60), expiresIn) // 2 hours
			assert.Equal(t, "backend", issuer)
			assert.Equal(t, expectedCookieAttributes, cookieAttributes)
			return nil
		})
	defer patch.Restore()

	accessToken, expireIn, err := CreateNewTempSession(sessionStore, expectedUserID, expectedCookieAttributes)
	assert.NoError(t, err)
	assert.Equal(t, expectedAccessToken, accessToken)
	assert.Equal(t, int32(2*60*60), expireIn) // 2 hours

	logs := (&loggingtest.Hook{Hook: logHook}).GetAllStructuredLogs()
	assert.Contains(t, logs, fmt.Sprintf("level=info msg=%q",
		fmt.Sprintf("Generated a session token expiring in %d seconds for a temporary user with group_id = %d",
			int32(2*60*60), expectedUserID)))

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateNewTempSession_Retries(t *testing.T) {
	expectedAccessTokens := []string{"tmp-02abcdefghijklmnopqrstuvwxyz", "tmp-03abcdefghijklmnopqrstuvwxyz"}
	accessTokensIndex := -1
	monkey.Patch(GenerateKey, func() (string, error) { accessTokensIndex++; return expectedAccessTokens[accessTokensIndex], nil })
	defer monkey.UnpatchAll()
	logHook, restoreFunc := logging.MockSharedLoggerHook()
	defer restoreFunc()

	db, mock := database.NewDBMock()
	defer func() { _ = db.Close() }()

	expectedUserID := int64(12345)
	expectedCookieAttributes := &database.SessionCookieAttributes{}
	sessionStore := database.NewDataStore(db).Sessions()

	var counter int
	patch := monkey.PatchInstanceMethod(reflect.TypeOf(sessionStore), "InsertNewOAuth",
		func(sessionStore *database.SessionStore, userID int64, token string, expiresIn int32, issuer string,
			cookieAttributes *database.SessionCookieAttributes) error {
			assert.Equal(t, expectedUserID, userID)
			assert.Equal(t, token, expectedAccessTokens[counter])
			assert.Equal(t, int32(2*60*60), expiresIn) // 2 hours
			assert.Equal(t, "backend", issuer)
			assert.Equal(t, expectedCookieAttributes, cookieAttributes)
			counter++
			if counter == 1 {
				return &mysql.MySQLError{
					Number:  1062,
					Message: fmt.Sprintf("ERROR 1062 (23000): Duplicate entry '%s' for key 'PRIMARY'", expectedAccessTokens[0]),
				}
			}
			return nil
		})
	defer patch.Restore()

	accessToken, expireIn, err := CreateNewTempSession(
		database.NewDataStore(db).Sessions(), expectedUserID, &database.SessionCookieAttributes{})
	assert.NoError(t, err)
	assert.Equal(t, expectedAccessTokens[1], accessToken)
	assert.Equal(t, int32(2*60*60), expireIn) // 2 hours
	assert.Equal(t, 2, counter)

	logs := (&loggingtest.Hook{Hook: logHook}).GetAllStructuredLogs()
	assert.Contains(t, logs, fmt.Sprintf("level=info msg=%q",
		fmt.Sprintf("Generated a session token expiring in %d seconds for a temporary user with group_id = %d",
			int32(2*60*60), expectedUserID)))
	assert.Equal(t, 1, strings.Count(logs, fmt.Sprintf("level=info msg=%q",
		fmt.Sprintf("Generated a session token expiring in %d seconds for a temporary user with group_id = %d",
			int32(2*60*60), expectedUserID))))

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateNewTempSession_HandlesGeneratorError(t *testing.T) {
	expectedError := errors.New("some error")
	monkey.Patch(GenerateKey, func() (string, error) { return "", expectedError })
	defer monkey.UnpatchAll()

	db, mock := database.NewDBMock()
	defer func() { _ = db.Close() }()

	expectedUserID := int64(12345)

	accessToken, expireIn, err := CreateNewTempSession(
		database.NewDataStore(db).Sessions(), expectedUserID, &database.SessionCookieAttributes{})
	assert.Equal(t, expectedError, err)
	assert.Equal(t, "", accessToken)
	assert.Equal(t, int32(2*60*60), expireIn) // 2 hours

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateNewTempSession_HandlesDBError(t *testing.T) {
	expectedAccessToken := "tmp-04abcdefghijklmnopqrstuvwxyz"
	monkey.Patch(GenerateKey, func() (string, error) { return expectedAccessToken, nil })
	defer monkey.UnpatchAll()

	db, mock := database.NewDBMock()
	defer func() { _ = db.Close() }()

	expectedUserID := int64(12345)
	expectedError := errors.New("some error")
	expectedCookieAttributes := &database.SessionCookieAttributes{}
	sessionStore := database.NewDataStore(db).Sessions()

	patch := monkey.PatchInstanceMethod(reflect.TypeOf(sessionStore), "InsertNewOAuth",
		func(sessionStore *database.SessionStore, userID int64, token string, expiresIn int32, issuer string,
			cookieAttributes *database.SessionCookieAttributes) error {
			assert.Equal(t, expectedUserID, userID)
			assert.Equal(t, token, expectedAccessToken)
			assert.Equal(t, int32(2*60*60), expiresIn) // 2 hours
			assert.Equal(t, "backend", issuer)
			assert.Equal(t, expectedCookieAttributes, cookieAttributes)
			return expectedError
		})
	defer patch.Restore()

	accessToken, expireIn, err := CreateNewTempSession(
		database.NewDataStore(db).Sessions(), expectedUserID, &database.SessionCookieAttributes{})
	assert.Equal(t, expectedError, err)
	assert.Equal(t, "", accessToken)
	assert.Equal(t, int32(2*60*60), expireIn) // 2 hours

	assert.NoError(t, mock.ExpectationsWereMet())
}
