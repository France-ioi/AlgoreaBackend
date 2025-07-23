//go:build !unit

package items_test

import (
	"reflect"
	"testing"
	_ "unsafe"

	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/testhelpers"
	"github.com/France-ioi/AlgoreaBackend/v2/testhelpers/testoutput"
)

func Test_insertItemRow_Duplicate(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db := testhelpers.SetupDBWithFixtureString(testhelpers.CreateTestContext(), `items: [{id: 1, default_language_tag: fr}]`)
	defer func() { _ = db.Close() }()

	var nextID int64
	monkey.PatchInstanceMethod(reflect.TypeOf(&database.DataStore{}), "NewID", func(*database.DataStore) int64 {
		nextID++
		return nextID
	})
	defer monkey.UnpatchAll()

	dataStore := database.NewDataStore(db)
	itemID, err := insertItemRow(dataStore, map[string]interface{}{"default_language_tag": "fr"})
	require.NoError(t, err)
	assert.Equal(t, int64(2), itemID)
}

//go:linkname insertItemRow github.com/France-ioi/AlgoreaBackend/v2/app/api/items.insertItemRow
func insertItemRow(*database.DataStore, map[string]interface{}) (int64, error)
