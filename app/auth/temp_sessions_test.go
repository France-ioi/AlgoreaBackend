package auth

import (
	"errors"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/testhelpers/monkey"
)

func TestCreateNewTempSession_PanicsGeneratorError(t *testing.T) {
	expectedError := errors.New("some error")
	monkey.Patch(GenerateKey, func() (string, error) { return "", expectedError })
	defer monkey.UnpatchAll()

	db, mock := database.NewDBMock()
	defer func() { _ = db.Close() }()

	expectedUserID := int64(12345)

	require.PanicsWithValue(t, expectedError, func() {
		_, _, _ = CreateNewTempSession(database.NewDataStore(db), expectedUserID)
	})

	assert.NoError(t, mock.ExpectationsWereMet())
}

const expectedAccessToken = "tmp-04abcdefghijklmnopqrstuvwxyz"

func TestCreateNewTempSession_HandlesDBError_WhenCreatingSessions(t *testing.T) {
	expectedSessionID := int64(12345)
	monkey.Patch(GenerateKey, func() (string, error) { return expectedAccessToken, nil })
	defer monkey.UnpatchAll()
	restoreNewID := database.SetNewIDFuncForTests(func() int64 { return expectedSessionID })
	defer restoreNewID()

	db, mock := database.NewDBMock()
	defer func() { _ = db.Close() }()

	expectedUserID := int64(12345)
	expectedError := errors.New("some error")
	dataStore := database.NewDataStore(db)

	patch := monkey.PatchInstanceMethod(reflect.TypeFor[*database.DataStore](), "InsertMap",
		func(_ *database.DataStore, row map[string]any) error {
			require.Equal(t, map[string]any{
				"session_id": expectedSessionID,
				"user_id":    expectedUserID,
			}, row)
			return expectedError
		})
	defer patch.Unpatch()

	mock.ExpectBegin()
	mock.ExpectRollback()

	accessToken, expireIn, err := CreateNewTempSession(dataStore, expectedUserID)
	assert.Equal(t, expectedError, err)
	assert.Empty(t, accessToken)
	assert.Equal(t, int32(0), expireIn) // 2 hours

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateNewTempSession_HandlesDBError_WhenCreatingAccessTokens(t *testing.T) {
	expectedSessionID := int64(12345)
	monkey.Patch(GenerateKey, func() (string, error) { return expectedAccessToken, nil })
	defer monkey.UnpatchAll()
	restoreNewID := database.SetNewIDFuncForTests(func() int64 { return expectedSessionID })
	defer restoreNewID()

	db, mock := database.NewDBMock()
	defer func() { _ = db.Close() }()

	expectedUserID := int64(12345)
	expectedError := errors.New("some error")
	dataStore := database.NewDataStore(db)

	monkey.PatchInstanceMethod(reflect.TypeFor[*database.DataStore](), "InsertMap",
		func(_ *database.DataStore, row map[string]any) error {
			require.Equal(t, map[string]any{
				"session_id": expectedSessionID,
				"user_id":    expectedUserID,
			}, row)
			return nil
		})
	monkey.PatchInstanceMethod(reflect.TypeFor[*database.AccessTokenStore](), "InsertNewToken",
		func(_ *database.AccessTokenStore, sessionID int64, token string, secondsUntilExpiry int32) error {
			require.Equal(t, expectedSessionID, sessionID)
			require.Equal(t, expectedAccessToken, token)
			require.Equal(t, int32(2*60*60), secondsUntilExpiry)
			return expectedError
		})
	defer monkey.UnpatchAll()

	mock.ExpectBegin()
	mock.ExpectRollback()

	accessToken, expireIn, err := CreateNewTempSession(dataStore, expectedUserID)
	assert.Equal(t, expectedError, err)
	assert.Empty(t, accessToken)
	assert.Equal(t, int32(0), expireIn) // 2 hours

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRefreshTempUserSession_HandlesDBError_WhenCreatingAccessTokens(t *testing.T) {
	expectedSessionID := int64(12345)
	monkey.Patch(GenerateKey, func() (string, error) { return expectedAccessToken, nil })
	defer monkey.UnpatchAll()

	db, mock := database.NewDBMock()
	defer func() { _ = db.Close() }()

	expectedUserID := int64(12345)
	expectedError := errors.New("some error")
	dataStore := database.NewDataStore(db)

	monkey.PatchInstanceMethod(reflect.TypeFor[*database.AccessTokenStore](), "InsertNewToken",
		func(_ *database.AccessTokenStore, sessionID int64, token string, secondsUntilExpiry int32) error {
			require.Equal(t, expectedSessionID, sessionID)
			require.Equal(t, expectedAccessToken, token)
			require.Equal(t, int32(2*60*60), secondsUntilExpiry)
			return expectedError
		})
	defer monkey.UnpatchAll()

	accessToken, expireIn, err := RefreshTempUserSession(dataStore, expectedUserID, expectedSessionID)
	assert.Equal(t, expectedError, err)
	assert.Empty(t, accessToken)
	assert.Equal(t, int32(0), expireIn) // 2 hours

	assert.NoError(t, mock.ExpectationsWereMet())
}

func Test_mustNotBeError(t *testing.T) {
	mustNotBeError(nil)
	expectedError := errors.New("some error")
	assert.PanicsWithValue(t, expectedError, func() {
		mustNotBeError(expectedError)
	})
}
