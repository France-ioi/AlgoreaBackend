//go:build !unit

package groups_test

import (
	"reflect"
	"testing"
	"time"
	_ "unsafe"

	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/golang"
	"github.com/France-ioi/AlgoreaBackend/v2/testhelpers"
	"github.com/France-ioi/AlgoreaBackend/v2/testhelpers/testoutput"
)

func Test_createUserGroup_Duplicate(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db := testhelpers.SetupDBWithFixtureString(`groups: [{id: 1, type: "User"}]`)
	defer func() { _ = db.Close() }()

	expectedTimestamp := "2019-05-30 11:00:00"
	expectedTime, _ := time.Parse(time.DateTime, expectedTimestamp)
	testhelpers.MockDBTime(expectedTimestamp)
	defer testhelpers.RestoreDBTime()

	var nextID int64
	monkey.PatchInstanceMethod(reflect.TypeOf(&database.DataStore{}), "NewID", func(*database.DataStore) int64 {
		nextID++
		return nextID
	})
	defer monkey.UnpatchAll()

	dataStore := database.NewDataStore(db)
	expectedLogin := "login"
	selfGroupID := createUserGroup(dataStore, expectedLogin)
	assert.Equal(t, int64(2), selfGroupID)

	type groupData struct {
		Name        string
		Type        string
		Description *string
		CreatedAt   database.Time
		IsOpen      bool
		SendEmails  bool
	}
	var group groupData
	require.NoError(t, dataStore.Groups().Take(&group, "id = ?", selfGroupID).Error())
	assert.Equal(t, groupData{
		Name:        expectedLogin,
		Type:        "User",
		CreatedAt:   database.Time(expectedTime),
		IsOpen:      false,
		SendEmails:  false,
		Description: golang.Ptr(expectedLogin),
	}, group)
}

//go:linkname createUserGroup github.com/France-ioi/AlgoreaBackend/v2/app/api/groups.createUserGroup
func createUserGroup(store *database.DataStore, login string) int64
