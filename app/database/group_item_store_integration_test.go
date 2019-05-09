// +build !unit

package database_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/testhelpers"
)

func TestGroupItemStore_RemovePartialAccess(t *testing.T) {
	db := testhelpers.SetupDBWithFixtureString(`
		groups_items:
			- {idGroup: 10, idItem: 1234, bManagerAccess: 0, sPartialAccessDate: 2019-03-22T08:44:55Z,
				sCachedPartialAccessDate: 2018-03-22T08:44:55Z, bCachedPartialAccess: 1}
			- {idGroup: 11, idItem: 1235, bManagerAccess: 1, sPartialAccessDate: 2019-03-22T08:44:55Z,
				sCachedPartialAccessDate: 2018-03-22T08:44:55Z, bCachedPartialAccess: 1}`)
	defer func() { _ = db.Close() }()

	groupItemStore := database.NewDataStore(db).GroupItems()
	groupItemStore.RemovePartialAccess(10, 1234)

	date1, _ := time.Parse(time.RFC3339, "2019-03-22T08:44:55Z")
	date2, _ := time.Parse(time.RFC3339, "2018-03-22T08:44:55Z")
	expected := []map[string]interface{}{
		{"idGroup": "10", "idItem": "1234", "bManagerAccess": "0",
			"sPartialAccessDate": nil, "sCachedPartialAccessDate": nil, "bCachedPartialAccess": "0"},
		{"idGroup": "11", "idItem": "1235", "bManagerAccess": "1",
			"sPartialAccessDate":       date1,
			"sCachedPartialAccessDate": date2,
			"bCachedPartialAccess":     "1"},
	}

	var got []map[string]interface{}
	assert.NoError(t, groupItemStore.
		Select("idGroup, idItem, bManagerAccess, sPartialAccessDate, sCachedPartialAccessDate, bCachedPartialAccess").
		Order("idGroup, idItem").ScanIntoSliceOfMaps(&got).Error())
	assert.Equal(t, expected, got)

	groupItemStore.RemovePartialAccess(11, 1235)
	assert.NoError(t, groupItemStore.
		Select("idGroup, idItem, bManagerAccess, sPartialAccessDate, sCachedPartialAccessDate, bCachedPartialAccess").
		Order("idGroup, idItem").ScanIntoSliceOfMaps(&got).Error())
	assert.Equal(t, expected, got)
}
