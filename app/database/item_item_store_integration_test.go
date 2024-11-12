//go:build !unit

package database_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/testhelpers"
	"github.com/France-ioi/AlgoreaBackend/v2/testhelpers/testoutput"
)

func TestItemItemStore_TriggerAfterInsert_MarksResultsAsChanged(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db := testhelpers.SetupDBWithFixtureString(groupGroupMarksResultsAsChangedFixture)
	defer func() { _ = db.Close() }()

	dataStore := database.NewDataStore(db)
	assert.NoError(t, dataStore.ItemItems().InsertMap(map[string]interface{}{
		"parent_item_id": 1, "child_item_id": 2, "child_order": 1,
	}))

	assertResultsMarkedAsChanged(t, dataStore, []resultPrimaryKey{
		{101, 1, 2},
		{102, 1, 2},
		{103, 1, 2},
		{104, 1, 2},
		{105, 1, 2},
		{106, 1, 2},
		{107, 1, 2},
		{108, 1, 2},
		{109, 1, 2},
	})
}
