// +build !unit

package database_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/testhelpers"
)

func TestPermissionGeneratedStore_TriggerAfterInsert_MarksAttemptsAsChanged(t *testing.T) {
	for _, test := range []struct {
		name            string
		groupID         int64
		itemID          int64
		canView         string
		expectedChanged []groupItemPair
	}{
		{
			name:            "make a parent item visible",
			groupID:         104,
			itemID:          2,
			canView:         "info",
			expectedChanged: []groupItemPair{{104, 3}, {105, 3}},
		},
		{
			name:            "make an ancestor item visible",
			groupID:         104,
			itemID:          1,
			canView:         "info",
			expectedChanged: []groupItemPair{{104, 2}, {104, 3}, {105, 2}, {105, 3}},
		},
		{
			name:            "make a parent item invisible",
			groupID:         104,
			itemID:          2,
			canView:         "none",
			expectedChanged: []groupItemPair{},
		},
		{
			name:            "make an item visible",
			groupID:         104,
			itemID:          3,
			canView:         "info",
			expectedChanged: []groupItemPair{},
		},
		{
			name:            "make a parent item visible for an expired membership",
			groupID:         108,
			itemID:          2,
			canView:         "none",
			expectedChanged: []groupItemPair{},
		},
	} {
		test := test
		t.Run(test.name, func(t *testing.T) {
			db := testhelpers.SetupDBWithFixtureString(groupGroupMarksAttemptsAsChangedFixture)
			defer func() { _ = db.Close() }()

			dataStore := database.NewDataStoreWithTable(db, "permissions_generated")
			assert.NoError(t, dataStore.InsertMap(map[string]interface{}{
				"group_id": test.groupID, "item_id": test.itemID, "can_view_generated": test.canView,
			}))

			assertAttemptsMarkedAsChanged(t, dataStore, test.expectedChanged)
		})
	}
}

func TestPermissionGeneratedStore_TriggerAfterUpdate_MarksAttemptsAsChanged(t *testing.T) {
	for _, test := range []struct {
		name            string
		groupID         int64
		itemID          int64
		newGroupID      int64
		newItemID       int64
		canView         string
		expectedChanged []groupItemPair
		noChanges       bool
	}{
		{
			name:            "make a parent item visible",
			groupID:         102,
			itemID:          1,
			newGroupID:      104,
			newItemID:       2,
			canView:         "info",
			expectedChanged: []groupItemPair{{104, 3}, {105, 3}},
		},
		{
			name:            "make an ancestor item visible",
			groupID:         102,
			itemID:          1,
			newGroupID:      104,
			newItemID:       1,
			canView:         "info",
			expectedChanged: []groupItemPair{{104, 2}, {104, 3}, {105, 2}, {105, 3}},
		},
		{
			name:            "make a parent item invisible",
			groupID:         102,
			itemID:          1,
			newGroupID:      104,
			newItemID:       2,
			canView:         "none",
			expectedChanged: []groupItemPair{},
		},
		{
			name:            "make an item visible",
			groupID:         107,
			itemID:          1,
			newGroupID:      104,
			newItemID:       3,
			canView:         "info",
			expectedChanged: []groupItemPair{},
		},
		{
			name:            "switch from invisible to visible",
			groupID:         107,
			itemID:          1,
			newGroupID:      107,
			newItemID:       1,
			canView:         "info",
			expectedChanged: []groupItemPair{{101, 2}, {101, 3}, {102, 2}, {102, 3}, {103, 2}, {103, 3}, {104, 2}, {104, 3}, {105, 2}, {105, 3}},
		},
		{
			name:            "make a parent item visible for an expired membership",
			groupID:         102,
			itemID:          1,
			newGroupID:      108,
			newItemID:       2,
			canView:         "none",
			expectedChanged: []groupItemPair{},
		},
		{
			name:            "no changes",
			groupID:         104,
			itemID:          3,
			newGroupID:      104,
			newItemID:       3,
			canView:         "info",
			expectedChanged: []groupItemPair{},
			noChanges:       true,
		},
	} {
		test := test
		t.Run(test.name, func(t *testing.T) {
			db := testhelpers.SetupDBWithFixtureString(groupGroupMarksAttemptsAsChangedFixture)
			defer func() { _ = db.Close() }()

			dataStore := database.NewDataStoreWithTable(db, "permissions_generated")
			result := dataStore.Where("group_id = ?", test.groupID).
				Where("item_id = ?", test.itemID).UpdateColumn(map[string]interface{}{
				"group_id":           test.newGroupID,
				"item_id":            test.newItemID,
				"can_view_generated": test.canView,
			})
			assert.NoError(t, result.Error())

			if test.noChanges {
				assert.Zero(t, result.RowsAffected())
			} else {
				assert.Equal(t, int64(1), result.RowsAffected())
			}
			assertAttemptsMarkedAsChanged(t, dataStore, test.expectedChanged)
		})
	}
}
