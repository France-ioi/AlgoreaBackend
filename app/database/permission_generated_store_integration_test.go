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

func TestPermissionGeneratedStore_MatchingUserAncestors(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db := testhelpers.SetupDBWithFixtureString(`
		groups: [{id: 1}, {id: 2}, {id: 3}, {id: 4}, {id: 5}]
		users: [{group_id: 5}]
		groups_ancestors:
			- {ancestor_group_id: 1, child_group_id: 5}
			- {ancestor_group_id: 2, child_group_id: 5, expires_at: 2019-05-30 11:00:00}
			- {ancestor_group_id: 3, child_group_id: 5}
		items: [{id: 2, default_language_tag: 2}, {id: 3, default_language_tag: 2}]
		permissions_generated:
			- {group_id: 1, item_id: 2, can_view_generated: none}
			- {group_id: 2, item_id: 2, can_view_generated: info}
			- {group_id: 3, item_id: 3, can_view_generated: content}
			- {group_id: 4, item_id: 3, can_view_generated: content_with_descendants}
			- {group_id: 5, item_id: 2, can_view_generated: solution}
	`)
	defer func() { _ = db.Close() }()

	permissionsStore := database.NewDataStore(db).Permissions()
	var result []map[string]interface{}
	require.NoError(t, permissionsStore.MatchingUserAncestors(&database.User{GroupID: 5}).
		Select("item_id, can_view_generated").
		ScanIntoSliceOfMaps(&result).Error())
	assert.Equal(t, []map[string]interface{}{
		{"item_id": int64(2), "can_view_generated": "none"},
		{"item_id": int64(3), "can_view_generated": "content"},
		{"item_id": int64(2), "can_view_generated": "solution"},
	}, result)
}

func TestPermissionGeneratedStore_TriggerAfterInsert_MarksResultsAsChanged(t *testing.T) {
	for _, test := range []struct {
		name            string
		groupID         int64
		itemID          int64
		canView         string
		expectedChanged []resultPrimaryKeyAndState
	}{
		{
			name:    "make a parent item visible",
			groupID: 104,
			itemID:  2,
			canView: "info",
			expectedChanged: []resultPrimaryKeyAndState{
				{ResultPrimaryKey: ResultPrimaryKey{104, 1, 3}},
				{ResultPrimaryKey: ResultPrimaryKey{105, 1, 3}},
			},
		},
		{
			name:    "make an ancestor item visible",
			groupID: 104,
			itemID:  1,
			canView: "info",
			expectedChanged: []resultPrimaryKeyAndState{
				{ResultPrimaryKey: ResultPrimaryKey{104, 1, 2}},
				{ResultPrimaryKey: ResultPrimaryKey{104, 1, 3}},
				{ResultPrimaryKey: ResultPrimaryKey{105, 1, 2}},
				{ResultPrimaryKey: ResultPrimaryKey{105, 1, 3}},
			},
		},
		{
			name:            "make a parent item invisible",
			groupID:         104,
			itemID:          2,
			canView:         "none",
			expectedChanged: []resultPrimaryKeyAndState{},
		},
		{
			name:            "make an item visible",
			groupID:         104,
			itemID:          3,
			canView:         "info",
			expectedChanged: []resultPrimaryKeyAndState{},
		},
		{
			name:            "make a parent item visible for an expired membership",
			groupID:         108,
			itemID:          2,
			canView:         "none",
			expectedChanged: []resultPrimaryKeyAndState{},
		},
	} {
		test := test
		t.Run(test.name, func(t *testing.T) {
			testoutput.SuppressIfPasses(t)

			db := testhelpers.SetupDBWithFixtureString(groupGroupMarksResultsAsChangedFixture)
			defer func() { _ = db.Close() }()

			dataStore := database.NewDataStoreWithTable(db, "permissions_generated")
			assert.NoError(t, dataStore.InTransaction(func(store *database.DataStore) error {
				return store.GroupGroups().CreateNewAncestors()
			}))
			assert.NoError(t, dataStore.InsertMap(map[string]interface{}{
				"group_id": test.groupID, "item_id": test.itemID, "can_view_generated": test.canView,
			}))

			assertResultsMarkedAsChanged(t, dataStore, test.expectedChanged)
		})
	}
}

func TestPermissionGeneratedStore_TriggerAfterUpdate_MarksResultsAsChanged(t *testing.T) {
	for _, test := range []struct {
		name            string
		groupID         int64
		itemID          int64
		canView         string
		expectedChanged []resultPrimaryKeyAndState
		noChanges       bool
		updateExisting  bool
	}{
		{
			name:    "make a parent item visible",
			groupID: 104,
			itemID:  2,
			canView: "info",
			expectedChanged: []resultPrimaryKeyAndState{
				{ResultPrimaryKey: ResultPrimaryKey{104, 1, 3}},
				{ResultPrimaryKey: ResultPrimaryKey{105, 1, 3}},
			},
		},
		{
			name:    "make an ancestor item visible",
			groupID: 104,
			itemID:  1,
			canView: "info",
			expectedChanged: []resultPrimaryKeyAndState{
				{ResultPrimaryKey: ResultPrimaryKey{104, 1, 2}},
				{ResultPrimaryKey: ResultPrimaryKey{104, 1, 3}},
				{ResultPrimaryKey: ResultPrimaryKey{105, 1, 2}},
				{ResultPrimaryKey: ResultPrimaryKey{105, 1, 3}},
			},
		},
		{
			name:            "make an ancestor item invisible",
			groupID:         108,
			itemID:          1,
			canView:         "none",
			expectedChanged: []resultPrimaryKeyAndState{},
			updateExisting:  true,
		},
		{
			name:            "make an item visible",
			groupID:         104,
			itemID:          3,
			canView:         "info",
			expectedChanged: []resultPrimaryKeyAndState{},
		},
		{
			name:           "switch ancestor from invisible to visible",
			groupID:        107,
			itemID:         1,
			canView:        "info",
			updateExisting: true,
			expectedChanged: []resultPrimaryKeyAndState{
				{ResultPrimaryKey: ResultPrimaryKey{105, 1, 2}},
				{ResultPrimaryKey: ResultPrimaryKey{105, 1, 3}},
				{ResultPrimaryKey: ResultPrimaryKey{107, 1, 2}},
				{ResultPrimaryKey: ResultPrimaryKey{107, 1, 3}},
			},
		},
		{
			name:            "make a parent item visible for an expired membership",
			groupID:         108,
			itemID:          2,
			canView:         "info",
			expectedChanged: []resultPrimaryKeyAndState{{ResultPrimaryKey: ResultPrimaryKey{108, 1, 3}}},
		},
		{
			name:            "no changes",
			groupID:         102,
			itemID:          1,
			canView:         "info",
			updateExisting:  true,
			expectedChanged: []resultPrimaryKeyAndState{},
			noChanges:       true,
		},
	} {
		test := test
		t.Run(test.name, func(t *testing.T) {
			testoutput.SuppressIfPasses(t)

			fixures := make([]string, 0, 2)
			if !test.updateExisting {
				fixures = append(fixures,
					fmt.Sprintf("permissions_generated: [{group_id: %d, item_id: %d}]", test.groupID, test.itemID))
			}
			fixures = append(fixures, groupGroupMarksResultsAsChangedFixture)
			db := testhelpers.SetupDBWithFixtureString(fixures...)
			defer func() { _ = db.Close() }()

			dataStore := database.NewDataStoreWithTable(db, "permissions_generated")
			require.NoError(t, dataStore.InTransaction(func(store *database.DataStore) error {
				return store.GroupGroups().CreateNewAncestors()
			}))
			result := dataStore.Where("group_id = ?", test.groupID).
				Where("item_id = ?", test.itemID).UpdateColumn(map[string]interface{}{
				"can_view_generated": test.canView,
			})
			require.NoError(t, result.Error())

			if test.noChanges {
				assert.Zero(t, result.RowsAffected())
			} else {
				assert.Equal(t, int64(1), result.RowsAffected())
			}
			assertResultsMarkedAsChanged(t, dataStore, test.expectedChanged)
		})
	}
}

func TestPermissionGeneratedStore_TriggerBeforeUpdate_RefusesToModifyGroupIDOrItemID(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db := testhelpers.SetupDBWithFixtureString(`
		groups: [{id: 1}]
		items: [{id: 2, default_language_tag: 2}]
		permissions_generated: [{group_id: 1, item_id: 2, can_view_generated: none}]
	`)
	defer func() { _ = db.Close() }()

	const expectedErrorMessage = "Error 1644 (45000): Unable to change immutable " +
		"permissions_generated.group_id and/or permissions_generated.child_item_id"

	dataStore := database.NewDataStoreWithTable(db, "permissions_generated")
	result := dataStore.Where("group_id = 1 AND item_id = 2").
		UpdateColumn("group_id", 3)
	require.EqualError(t, result.Error(), expectedErrorMessage)
	result = dataStore.Where("group_id = 1 AND item_id = 2").
		UpdateColumn("item_id", 3)
	require.EqualError(t, result.Error(), expectedErrorMessage)
}
