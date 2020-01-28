// +build !unit

package database_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/testhelpers"
)

func TestItemItemStore_TriggerAfterInsert_MarksAttemptsAsChanged(t *testing.T) {
	db := testhelpers.SetupDBWithFixtureString(groupGroupMarksAttemptsAsChangedFixture)
	defer func() { _ = db.Close() }()

	dataStore := database.NewDataStore(db)
	assert.NoError(t, dataStore.ItemItems().InsertMap(map[string]interface{}{
		"parent_item_id": 1, "child_item_id": 2, "child_order": 1,
	}))

	assertAttemptsMarkedAsChanged(t, dataStore, []groupItemPair{
		{101, 2}, {102, 2}, {103, 2},
		{104, 2}, {105, 2},
	})
}
