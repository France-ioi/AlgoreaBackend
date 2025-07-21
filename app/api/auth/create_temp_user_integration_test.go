//go:build !unit

package auth_test

import (
	"reflect"
	"testing"
	"time"
	_ "unsafe"

	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/app/rand"
	"github.com/France-ioi/AlgoreaBackend/v2/golang"
	"github.com/France-ioi/AlgoreaBackend/v2/testhelpers"
	"github.com/France-ioi/AlgoreaBackend/v2/testhelpers/testoutput"
)

const expectedTimestamp = "2019-05-30 11:00:00"

func Test_createTempUserGroup_Duplicate(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db := testhelpers.SetupDBWithFixtureString(testhelpers.CreateTestContext(), `groups: [{id: 1, type: "User"}]`)
	defer func() { _ = db.Close() }()

	var nextID int64
	monkey.PatchInstanceMethod(reflect.TypeOf(&database.DataStore{}), "NewID", func(*database.DataStore) int64 {
		nextID++
		return nextID
	})
	defer monkey.UnpatchAll()

	expectedTime, _ := time.Parse(time.DateTime, expectedTimestamp)
	testhelpers.MockDBTime(expectedTimestamp)
	defer testhelpers.RestoreDBTime()

	dataStore := database.NewDataStore(db)
	selfGroupID := createTempUserGroup(dataStore)
	assert.Equal(t, int64(2), selfGroupID)

	type groupData struct {
		Type       string
		CreatedAt  database.Time
		IsOpen     bool
		SendEmails bool
	}
	var group groupData
	require.NoError(t, dataStore.Groups().Take(&group, "id = ?", selfGroupID).Error())

	assert.Equal(t, groupData{
		Type:       "User",
		CreatedAt:  database.Time(expectedTime),
		IsOpen:     false,
		SendEmails: false,
	}, group)
}

func Test_createTempUser_Duplicate(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db := testhelpers.SetupDBWithFixtureString(testhelpers.CreateTestContext(), `
		groups: [{id: 1, type: "User"}, {id: 2, type: "User"}]
		users: [{group_id: 1, login: "tmp-50000000"}]`)
	defer func() { _ = db.Close() }()

	nextInt31 := int32(40000000) - 1
	monkey.Patch(rand.Int31n, func(n int32) int32 {
		require.Equal(t, int32(99999999-10000000+1), n)
		nextInt31++
		return nextInt31
	})
	defer monkey.UnpatchAll()

	expectedTime, _ := time.Parse(time.DateTime, expectedTimestamp)
	testhelpers.MockDBTime(expectedTimestamp)
	defer testhelpers.RestoreDBTime()

	dataStore := database.NewDataStore(db)
	expectedIP := "1.2.3.4"
	expectedDefaultLanguage := "en"
	login := createTempUser(dataStore, 2, expectedDefaultLanguage, expectedIP)
	assert.Equal(t, "tmp-50000001", login)

	type userData struct {
		LoginID         *int64
		Login           string
		TempUser        bool
		RegisteredAt    *database.Time
		GroupID         int64
		DefaultLanguage string
		LastIP          *string
	}
	var user userData
	require.NoError(t, dataStore.Users().Take(&user, "login = ?", login).Error())
	assert.Equal(t, userData{
		LoginID:         golang.Ptr(int64(0)),
		Login:           login,
		TempUser:        true,
		RegisteredAt:    golang.Ptr(database.Time(expectedTime)),
		GroupID:         2,
		DefaultLanguage: expectedDefaultLanguage,
		LastIP:          golang.Ptr(expectedIP),
	}, user)
}

//go:linkname createTempUser github.com/France-ioi/AlgoreaBackend/v2/app/api/auth.createTempUser
func createTempUser(*database.DataStore, int64, interface{}, string) string

//go:linkname createTempUserGroup github.com/France-ioi/AlgoreaBackend/v2/app/api/auth.createTempUserGroup
func createTempUserGroup(*database.DataStore) int64
