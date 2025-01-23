//go:build !unit

package items_test

import (
	"reflect"
	"testing"
	_ "unsafe"

	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/testhelpers"
	"github.com/France-ioi/AlgoreaBackend/v2/testhelpers/testoutput"
)

func Test_createContestParticipantsGroup_Duplicate(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db := testhelpers.SetupDBWithFixtureString(`
		groups: [{id: 1, type: "ContestParticipants"}]
		items: [{id: 123, default_language_tag: fr}]`)
	defer func() { _ = db.Close() }()

	var nextID int64
	monkey.PatchInstanceMethod(reflect.TypeOf(&database.DataStore{}), "NewID", func(*database.DataStore) int64 {
		nextID++
		return nextID
	})
	defer monkey.UnpatchAll()

	dataStore := database.NewDataStore(db)
	groupID := createContestParticipantsGroup(dataStore, 123)
	assert.Equal(t, int64(2), groupID)

	type groupData struct {
		Type string
		Name string
	}
	var group groupData
	assert.NoError(t, dataStore.Groups().Take(&group, "id = ?", groupID).Error())
	assert.Equal(t, groupData{
		Type: "ContestParticipants",
		Name: "123-participants",
	}, group)
}

//go:linkname createContestParticipantsGroup github.com/France-ioi/AlgoreaBackend/v2/app/api/items.createContestParticipantsGroup
func createContestParticipantsGroup(*database.DataStore, int64) int64
