//go:build !unit

package database_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/testhelpers"
	"github.com/France-ioi/AlgoreaBackend/v2/testhelpers/testoutput"
)

func TestItemItemStore_TriggerAfterInsert_MarksResultsAsChanged(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db := testhelpers.SetupDBWithFixtureString(groupGroupMarksResultsAsChangedFixture)
	defer func() { _ = db.Close() }()

	dataStore := database.NewDataStore(db)
	assertResultsMarkedAsChanged(t, dataStore, []resultPrimaryKeyAndState{})

	assert.NoError(t, dataStore.ItemItems().InsertMap(map[string]interface{}{
		"parent_item_id": 1, "child_item_id": 2, "child_order": 1,
	}))

	assertResultsMarkedAsChanged(t, dataStore, []resultPrimaryKeyAndState{
		{ResultPrimaryKey: ResultPrimaryKey{101, 1, 1}, State: "to_be_recomputed"},
		{ResultPrimaryKey: ResultPrimaryKey{102, 1, 1}, State: "to_be_recomputed"},
		{ResultPrimaryKey: ResultPrimaryKey{103, 1, 1}, State: "to_be_recomputed"},
		{ResultPrimaryKey: ResultPrimaryKey{104, 1, 1}, State: "to_be_recomputed"},
		{ResultPrimaryKey: ResultPrimaryKey{105, 1, 1}, State: "to_be_recomputed"},
		{ResultPrimaryKey: ResultPrimaryKey{106, 1, 1}, State: "to_be_recomputed"},
		{ResultPrimaryKey: ResultPrimaryKey{107, 1, 1}, State: "to_be_recomputed"},
		{ResultPrimaryKey: ResultPrimaryKey{108, 1, 1}, State: "to_be_recomputed"},
		{ResultPrimaryKey: ResultPrimaryKey{109, 1, 1}, State: "to_be_recomputed"},

		{ResultPrimaryKey: ResultPrimaryKey{101, 1, 2}},
		{ResultPrimaryKey: ResultPrimaryKey{102, 1, 2}},
		{ResultPrimaryKey: ResultPrimaryKey{103, 1, 2}},
		{ResultPrimaryKey: ResultPrimaryKey{104, 1, 2}},
		{ResultPrimaryKey: ResultPrimaryKey{105, 1, 2}},
		{ResultPrimaryKey: ResultPrimaryKey{106, 1, 2}},
		{ResultPrimaryKey: ResultPrimaryKey{107, 1, 2}},
		{ResultPrimaryKey: ResultPrimaryKey{108, 1, 2}},
		{ResultPrimaryKey: ResultPrimaryKey{109, 1, 2}},
	})
}

func TestItemItemStore_TriggerAfterInsert_MarksPermissionsForRecomputing(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db := testhelpers.SetupDBWithFixture("permission_granted_store/compute_all_access/_common")
	defer func() { _ = db.Close() }()

	dataStore := database.NewDataStore(db)
	require.NoError(t, dataStore.InTransaction(func(dataStore *database.DataStore) error {
		dataStore.SchedulePermissionsPropagation()
		return nil
	}))

	var markedPermissions []groupItemsResultRow
	require.NoError(t, dataStore.Table("permissions_propagate").
		Select("group_id, item_id, propagate_to").
		Order("group_id, item_id").Scan(&markedPermissions).Error())
	require.Empty(t, markedPermissions)

	assert.NoError(t, dataStore.ItemItems().InsertMap(map[string]interface{}{
		"parent_item_id": 1, "child_item_id": 2, "child_order": 1,
	}))

	require.NoError(t, dataStore.Table("permissions_propagate").
		Select("group_id, item_id, propagate_to").
		Order("group_id, item_id").Scan(&markedPermissions).Error())
	assert.Equal(t, []groupItemsResultRow{
		{GroupID: 1, ItemID: 2, PropagateTo: "self"},
		{GroupID: 2, ItemID: 2, PropagateTo: "self"},
	}, markedPermissions)
}

func TestItemItemStore_TriggerAfterUpdate_MarksPermissionsForRecomputing(t *testing.T) {
	for _, test := range []struct {
		column   string
		newValue interface{}
	}{
		{column: "content_view_propagation", newValue: "as_info"},
		{column: "upper_view_levels_propagation", newValue: "as_is"},
		{column: "grant_view_propagation", newValue: 1},
		{column: "watch_propagation", newValue: 1},
		{column: "edit_propagation", newValue: 1},
	} {
		test := test
		t.Run(fmt.Sprintf("%s=%v", test.column, test.newValue), func(t *testing.T) {
			testoutput.SuppressIfPasses(t)

			db := testhelpers.SetupDBWithFixture("permission_granted_store/compute_all_access/_common")
			defer func() { _ = db.Close() }()

			dataStore := database.NewDataStore(db)
			require.NoError(t, dataStore.InTransaction(func(dataStore *database.DataStore) error {
				dataStore.SchedulePermissionsPropagation()
				return nil
			}))

			var markedPermissions []groupItemsResultRow
			require.NoError(t, dataStore.Table("permissions_propagate").
				Select("group_id, item_id, propagate_to").
				Order("group_id, item_id").Scan(&markedPermissions).Error())
			require.Empty(t, markedPermissions)

			assert.NoError(t, dataStore.ItemItems().Where("parent_item_id=4 AND child_item_id=1").
				UpdateColumn(test.column, test.newValue).Error())

			require.NoError(t, dataStore.Table("permissions_propagate").
				Select("group_id, item_id, propagate_to").
				Order("group_id, item_id").Scan(&markedPermissions).Error())
			assert.Equal(t, []groupItemsResultRow{
				{GroupID: 1, ItemID: 1, PropagateTo: "self"},
				{GroupID: 2, ItemID: 1, PropagateTo: "self"},
			}, markedPermissions)
		})
	}
}

func TestItemItemStore_TriggerBeforeDelete_MarksPermissionsForRecomputing(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db := testhelpers.SetupDBWithFixture("permission_granted_store/compute_all_access/_common")
	defer func() { _ = db.Close() }()

	dataStore := database.NewDataStore(db)
	require.NoError(t, dataStore.InTransaction(func(dataStore *database.DataStore) error {
		dataStore.SchedulePermissionsPropagation()
		return nil
	}))

	var markedPermissions []groupItemsResultRow
	require.NoError(t, dataStore.Table("permissions_propagate").
		Select("group_id, item_id, propagate_to").
		Order("group_id, item_id").Scan(&markedPermissions).Error())
	require.Empty(t, markedPermissions)

	assert.NoError(t, dataStore.ItemItems().Delete("parent_item_id=4 AND child_item_id=1").Error())

	require.NoError(t, dataStore.Table("permissions_propagate").
		Select("group_id, item_id, propagate_to").
		Order("group_id, item_id").Scan(&markedPermissions).Error())
	assert.Equal(t, []groupItemsResultRow{
		{GroupID: 1, ItemID: 1, PropagateTo: "self"},
		{GroupID: 2, ItemID: 1, PropagateTo: "self"},
	}, markedPermissions)
}
