//go:build !unit

package database_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/testhelpers"
	"github.com/France-ioi/AlgoreaBackend/v2/testhelpers/testoutput"
)

type tokenData struct {
	SessionID int64
	Token     string
	ExpiresAt database.Time
	IssuedAt  database.Time
}

func TestAccessTokenStore_InsertNewToken(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	timeNow, _ := time.Parse(time.DateTime, "2021-01-02 03:04:05")
	testhelpers.MockDBTime(timeNow.Format(time.DateTime))
	defer testhelpers.RestoreDBTime()

	db := testhelpers.SetupDBWithFixtureString(testhelpers.CreateTestContext(), `
		groups: [{id: 123}]
		sessions: [{session_id: 456, user_id: 123, refresh_token: 'refresh_token'}]`)
	defer func() { _ = db.Close() }()

	accessTokenStore := database.NewDataStore(db).AccessTokens()
	require.NoError(t, accessTokenStore.InsertNewToken(456, "token", 789))

	var token tokenData
	assert.NoError(t, accessTokenStore.Scan(&token).Error())
	assert.Equal(t, tokenData{
		SessionID: 456,
		Token:     "token",
		ExpiresAt: database.Time(timeNow.Add(789 * time.Second)),
		IssuedAt:  database.Time(timeNow),
	}, token)
}

func TestAccessTokenStore_GetMostRecentValidTokenForSession(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	timeNow, _ := time.Parse(time.DateTime, "2021-01-02 03:04:05")
	testhelpers.MockDBTime(timeNow.Format(time.DateTime))
	defer testhelpers.RestoreDBTime()

	db := testhelpers.SetupDBWithFixtureString(testhelpers.CreateTestContext(), `
		groups: [{id: 123}]
		sessions:
			- {session_id: 456, user_id: 123, refresh_token: 'refresh_token1'}
			- {session_id: 457, user_id: 123, refresh_token: 'refresh_token2'}
		access_tokens:
			- {session_id: 456, token: 'token1', expires_at: '2021-01-02 03:09:06', issued_at: '2021-01-02 02:04:05'}
			- {session_id: 456, token: 'token2', expires_at: '2021-01-02 03:09:07', issued_at: '2021-01-02 02:04:05'}
			- {session_id: 457, token: 'token3', expires_at: '2021-01-02 03:09:08', issued_at: '2021-01-02 02:59:06'}`)
	defer func() { _ = db.Close() }()

	accessTokenStore := database.NewDataStore(db).AccessTokens()
	token, err := accessTokenStore.GetMostRecentValidTokenForSession(456)
	require.NoError(t, err)
	assert.Equal(t, database.MostRecentToken{
		Token:              "token2",
		SecondsUntilExpiry: 302,
		TooNewToRefresh:    false,
	}, token)

	token, err = accessTokenStore.GetMostRecentValidTokenForSession(457)
	require.NoError(t, err)
	assert.Equal(t, database.MostRecentToken{
		Token:              "token3",
		SecondsUntilExpiry: 303,
		TooNewToRefresh:    true,
	}, token)
}

func TestAccessTokenStore_DeleteExpiredTokensOfUser(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	timeNow, _ := time.Parse(time.DateTime, "2021-01-02 03:09:08")
	testhelpers.MockDBTime(timeNow.Format(time.DateTime))
	defer testhelpers.RestoreDBTime()

	db := testhelpers.SetupDBWithFixtureString(testhelpers.CreateTestContext(), `
		groups: [{id: 123}, {id: 124}]
		sessions:
			- {session_id: 456, user_id: 123, refresh_token: 'refresh_token1'}
			- {session_id: 457, user_id: 123, refresh_token: 'refresh_token2'}
			- {session_id: 458, user_id: 124, refresh_token: 'refresh_token3'}
		access_tokens:
			- {session_id: 456, token: 'token1', expires_at: '2021-01-02 03:09:06', issued_at: '2021-01-02 02:04:05'}
			- {session_id: 456, token: 'token2', expires_at: '2021-01-02 03:09:07', issued_at: '2021-01-02 02:04:05'}
			- {session_id: 457, token: 'token3', expires_at: '2021-01-02 03:09:08', issued_at: '2021-01-02 02:59:06'}
			- {session_id: 457, token: 'token4', expires_at: '2021-01-02 03:09:09', issued_at: '2021-01-02 02:59:06'}
			- {session_id: 458, token: 'token5', expires_at: '2021-01-02 03:09:05', issued_at: '2021-01-02 02:59:06'}`)
	defer func() { _ = db.Close() }()

	accessTokenStore := database.NewDataStore(db).AccessTokens()
	assert.NoError(t, accessTokenStore.DeleteExpiredTokensOfUser(123))

	var count int
	assert.NoError(t, accessTokenStore.Count(&count).Error())
	assert.Equal(t, 2, count)

	t.Run("456", func(t *testing.T) {
		testoutput.SuppressIfPasses(t)
		_, err := accessTokenStore.GetMostRecentValidTokenForSession(456)
		require.EqualError(t, err, "record not found")
	})
	t.Run("457", func(t *testing.T) {
		testoutput.SuppressIfPasses(t)
		token, err := accessTokenStore.GetMostRecentValidTokenForSession(457)
		require.NoError(t, err)
		assert.Equal(t, database.MostRecentToken{
			Token:              "token4",
			SecondsUntilExpiry: 1,
			TooNewToRefresh:    false,
		}, token)
	})
	t.Run("458", func(t *testing.T) {
		testoutput.SuppressIfPasses(t)
		token, err := accessTokenStore.GetMostRecentValidTokenForSession(458)
		require.NoError(t, err)
		assert.Equal(t, database.MostRecentToken{
			Token:              "token5",
			SecondsUntilExpiry: -3,
			TooNewToRefresh:    false,
		}, token)
	})
}
