package auth

import (
	"errors"
	"reflect"
	"testing"

	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
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

func TestCreateNewTempSession_HandlesDBError_WhenCreatingSessions(t *testing.T) {
	expectedSessionID := int64(12345)
	expectedAccessToken := "tmp-something"
	monkey.Patch(GenerateKey, func() (string, error) { return expectedAccessToken, nil })
	monkey.PatchInstanceMethod(reflect.TypeOf(&database.DataStore{}), "NewID",
		func(_ *database.DataStore) int64 { return expectedSessionID })
	defer monkey.UnpatchAll()

	db, mock := database.NewDBMock()
	defer func() { _ = db.Close() }()

	expectedUserID := int64(12345)
	expectedError := errors.New("some error")
	dataStore := database.NewDataStore(db)

	patch := monkey.PatchInstanceMethod(reflect.TypeOf(dataStore), "InsertMap",
		func(_ *database.DataStore, row map[string]interface{}) error {
			require.Equal(t, map[string]interface{}{
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
	expectedAccessToken := "tmp-04abcdefghijklmnopqrstuvwxyz"
	monkey.Patch(GenerateKey, func() (string, error) { return expectedAccessToken, nil })
	monkey.PatchInstanceMethod(reflect.TypeOf(&database.DataStore{}), "NewID",
		func(_ *database.DataStore) int64 { return expectedSessionID })
	defer monkey.UnpatchAll()

	db, mock := database.NewDBMock()
	defer func() { _ = db.Close() }()

	expectedUserID := int64(12345)
	expectedError := errors.New("some error")
	dataStore := database.NewDataStore(db)

	monkey.PatchInstanceMethod(reflect.TypeOf(dataStore), "InsertMap",
		func(_ *database.DataStore, row map[string]interface{}) error {
			require.Equal(t, map[string]interface{}{
				"session_id": expectedSessionID,
				"user_id":    expectedUserID,
			}, row)
			return nil
		})
	monkey.PatchInstanceMethod(reflect.TypeOf(&database.AccessTokenStore{}), "InsertNewToken",
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

func Test_mustNotBeError(t *testing.T) {
	mustNotBeError(nil)
	expectedError := errors.New("some error")
	assert.PanicsWithValue(t, expectedError, func() {
		mustNotBeError(expectedError)
	})
}
