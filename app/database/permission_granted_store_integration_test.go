// +build !unit

package database_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/testhelpers"
)

func TestPermissionGrantedStore_RemoveContentAccess(t *testing.T) {
	db := testhelpers.SetupDBWithFixtureString(`
		groups: [{id: 10}, {id: 11}, {id: 12}]
		items: [{id: 1234}, {id: 1235}, {id: 1236}]
		permissions_generated:
			- {group_id: 10, item_id: 1234, can_view_generated: content}
			- {group_id: 11, item_id: 1235, can_view_generated: solution}
			- {group_id: 12, item_id: 1236, can_view_generated: solution}
		permissions_granted:
			- {group_id: 10, item_id: 1234, can_view: content, giver_group_id: -1}
			- {group_id: 11, item_id: 1235, can_view: solution, giver_group_id: -1}
			- {group_id: 12, item_id: 1236, can_view: content, giver_group_id: 12}`)
	defer func() { _ = db.Close() }()

	permissionGrantedStore := database.NewDataStore(db).PermissionsGranted()
	permissionGrantedStore.RemoveContentAccess(10, 1234)

	expected := []map[string]interface{}{
		{"group_id": "10", "item_id": "1234", "can_view": "none"},
		{"group_id": "11", "item_id": "1235", "can_view": "solution"},
		{"group_id": "12", "item_id": "1236", "can_view": "content"},
	}
	expectedGenerated := []map[string]interface{}{
		{"group_id": "10", "item_id": "1234", "can_view_generated": "content"},
		{"group_id": "11", "item_id": "1235", "can_view_generated": "solution"},
		{"group_id": "12", "item_id": "1236", "can_view_generated": "solution"},
	}

	var got []map[string]interface{}
	assert.NoError(t, permissionGrantedStore.
		Select("group_id, item_id, can_view").
		Order("group_id, item_id").ScanIntoSliceOfMaps(&got).Error())
	assert.Equal(t, expected, got)

	permissionGeneratedStore := permissionGrantedStore.Permissions()
	assert.NoError(t, permissionGeneratedStore.
		Select("group_id, item_id, can_view_generated").
		Order("group_id, item_id").ScanIntoSliceOfMaps(&got).Error())
	assert.Equal(t, expectedGenerated, got)

	permissionGrantedStore.RemoveContentAccess(11, 1235)
	assert.NoError(t, permissionGrantedStore.
		Select("group_id, item_id, can_view").
		Order("group_id, item_id").ScanIntoSliceOfMaps(&got).Error())
	assert.Equal(t, expected, got)
	assert.NoError(t, permissionGeneratedStore.
		Select("group_id, item_id, can_view_generated").
		Order("group_id, item_id").ScanIntoSliceOfMaps(&got).Error())
	assert.Equal(t, expectedGenerated, got)

	permissionGrantedStore.RemoveContentAccess(12, 1236)
	assert.NoError(t, permissionGrantedStore.
		Select("group_id, item_id, can_view").
		Order("group_id, item_id").ScanIntoSliceOfMaps(&got).Error())
	assert.Equal(t, expected, got)
	assert.NoError(t, permissionGeneratedStore.
		Select("group_id, item_id, can_view_generated").
		Order("group_id, item_id").ScanIntoSliceOfMaps(&got).Error())
	assert.Equal(t, expectedGenerated, got)
}
