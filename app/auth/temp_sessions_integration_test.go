package auth_test

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/France-ioi/AlgoreaBackend/v2/app/auth"
	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/app/logging"
	"github.com/France-ioi/AlgoreaBackend/v2/app/loggingtest"
	"github.com/France-ioi/AlgoreaBackend/v2/testhelpers"
)

const expectedAccessToken = "tmp-01abcdefghijklmnopqrstuvwxyz"

func TestCreateNewTempSession(t *testing.T) {
	expectedUserID := int64(67890)
	expectedSessionID := int64(12345)

	monkey.Patch(auth.GenerateKey, func() (string, error) { return expectedAccessToken, nil })
	monkey.PatchInstanceMethod(reflect.TypeOf(&database.DataStore{}), "NewID",
		func(_ *database.DataStore) int64 { return expectedSessionID })
	defer monkey.UnpatchAll()

	logger, logHook := logging.NewMockLogger()
	db := testhelpers.SetupDBWithFixtureString(testhelpers.CreateTestContextWithLogger(logger), "users: [{group_id: 67890}]")
	defer func() { _ = db.Close() }()

	expectedTime := time.Now().UTC()
	timePatch := testhelpers.MockDBTime(expectedTime.Format(time.DateTime))
	defer testhelpers.RestoreDBTime(timePatch)

	accessTokenStore := database.NewDataStore(db).AccessTokens()

	accessToken, expireIn, err := auth.CreateNewTempSession(accessTokenStore.DataStore, expectedUserID)
	require.NoError(t, err)
	assert.Equal(t, expectedAccessToken, accessToken)
	assert.Equal(t, int32(2*60*60), expireIn) // 2 hours

	var sessions []map[string]interface{}
	require.NoError(t, accessTokenStore.Sessions().Select("session_id, user_id").
		ScanIntoSliceOfMaps(&sessions).Error())
	assert.Equal(t, []map[string]interface{}{{
		"session_id": expectedSessionID,
		"user_id":    expectedUserID,
	}}, sessions)

	var accessTokens []map[string]interface{}
	require.NoError(t, accessTokenStore.Select("session_id, token, expires_at, issued_at").
		ScanIntoSliceOfMaps(&accessTokens).Error())
	assert.Equal(t, []map[string]interface{}{{
		"session_id": expectedSessionID,
		"token":      expectedAccessToken,
		"expires_at": expectedTime.Add(time.Duration(expireIn) * time.Second).Format(time.DateTime),
		"issued_at":  expectedTime.Format(time.DateTime),
	}}, accessTokens)

	logs := (&loggingtest.Hook{Hook: logHook}).GetAllStructuredLogs()
	assert.Contains(t, logs,
		fmt.Sprintf("Generated a session token expiring in %d seconds for a temporary user with group_id = %d",
			int32(2*60*60), expectedUserID))
}

func TestCreateNewTempSession_Retries(t *testing.T) {
	expectedSessionIDs := []int64{12345, 23456}
	sessionIDsIndex := -1
	expectedUserID := int64(67890)
	monkey.Patch(auth.GenerateKey, func() (string, error) { return expectedAccessToken, nil })
	monkey.PatchInstanceMethod(reflect.TypeOf(&database.DataStore{}), "NewID",
		func(_ *database.DataStore) int64 { sessionIDsIndex++; return expectedSessionIDs[sessionIDsIndex] })
	defer monkey.UnpatchAll()

	logger, logHook := logging.NewMockLogger()
	db := testhelpers.SetupDBWithFixtureString(testhelpers.CreateTestContextWithLogger(logger), `
		users: [{group_id: 67890, login: john}, {group_id: 67891, login: jane}]
		sessions: [{session_id: 12345, user_id: 67891}]`)
	defer func() { _ = db.Close() }()

	dataStore := database.NewDataStore(db)
	accessToken, expireIn, err := auth.CreateNewTempSession(dataStore, expectedUserID)
	require.NoError(t, err)
	assert.Equal(t, expectedAccessToken, accessToken)
	assert.Equal(t, int32(2*60*60), expireIn) // 2 hours
	assert.Equal(t, 1, sessionIDsIndex)

	var sessions []map[string]interface{}
	require.NoError(t, dataStore.Sessions().Select("session_id, user_id").Order("session_id").
		ScanIntoSliceOfMaps(&sessions).Error())
	assert.Equal(t, []map[string]interface{}{
		{
			"session_id": expectedSessionIDs[0],
			"user_id":    int64(67891),
		},
		{
			"session_id": expectedSessionIDs[1],
			"user_id":    expectedUserID,
		},
	}, sessions)

	logs := (&loggingtest.Hook{Hook: logHook}).GetAllStructuredLogs()
	assert.Contains(t, logs,
		fmt.Sprintf("Generated a session token expiring in %d seconds for a temporary user with group_id = %d",
			int32(2*60*60), expectedUserID))
	assert.Equal(t, 1, strings.Count(logs,
		fmt.Sprintf("Generated a session token expiring in %d seconds for a temporary user with group_id = %d",
			int32(2*60*60), expectedUserID)))
}
