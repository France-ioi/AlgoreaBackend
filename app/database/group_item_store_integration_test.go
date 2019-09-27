// +build !unit

package database_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/testhelpers"
)

func TestGroupItemStore_RemovePartialAccess(t *testing.T) {
	db := testhelpers.SetupDBWithFixtureString(`
		groups_items:
			- {group_id: 10, item_id: 1234, manager_access: 0, partial_access_since: 2019-03-22 08:44:55,
				cached_partial_access_since: 2018-03-22 08:44:55, cached_partial_access: 1, creator_user_id: 1}
			- {group_id: 11, item_id: 1235, manager_access: 1, partial_access_since: 2019-03-22 08:44:55,
				cached_partial_access_since: 2018-03-22 08:44:55, cached_partial_access: 1, creator_user_id: 1}`)
	defer func() { _ = db.Close() }()

	groupItemStore := database.NewDataStore(db).GroupItems()
	groupItemStore.RemovePartialAccess(10, 1234)

	expected := []map[string]interface{}{
		{"group_id": "10", "item_id": "1234", "manager_access": "0",
			"partial_access_since": nil, "cached_partial_access_since": nil, "cached_partial_access": "0"},
		{"group_id": "11", "item_id": "1235", "manager_access": "1",
			"partial_access_since":        "2019-03-22 08:44:55",
			"cached_partial_access_since": "2018-03-22 08:44:55",
			"cached_partial_access":       "1"},
	}

	var got []map[string]interface{}
	assert.NoError(t, groupItemStore.
		Select("group_id, item_id, manager_access, partial_access_since, cached_partial_access_since, cached_partial_access").
		Order("group_id, item_id").ScanIntoSliceOfMaps(&got).Error())
	assert.Equal(t, expected, got)

	groupItemStore.RemovePartialAccess(11, 1235)
	assert.NoError(t, groupItemStore.
		Select("group_id, item_id, manager_access, partial_access_since, cached_partial_access_since, cached_partial_access").
		Order("group_id, item_id").ScanIntoSliceOfMaps(&got).Error())
	assert.Equal(t, expected, got)
}
