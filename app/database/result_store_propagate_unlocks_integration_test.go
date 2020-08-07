// +build !unit

package database_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/testhelpers"
)

type unlocksResultRow struct {
	GroupID       int64
	ItemID        int64
	CanView       string
	SourceGroupID int64
	Origin        string
}

func TestResultStore_Propagate_Unlocks(t *testing.T) {
	db := testhelpers.SetupDBWithFixture("results_propagation/_common", "results_propagation/unlocks")
	defer func() { _ = db.Close() }()

	testUnlocks(db, t)
}

func TestResultStore_Propagate_Unlocks_UpdatesOldRecords(t *testing.T) {
	db := testhelpers.SetupDBWithFixture(
		"results_propagation/_common",
		"results_propagation/unlocks",
		"results_propagation/unlocks_old_records")
	defer func() { _ = db.Close() }()

	testUnlocks(db, t)
}

func testUnlocks(db *database.DB, t *testing.T) {
	resultStore := database.NewDataStore(db).Results()
	for _, ids := range []struct {
		ParticipantID int64
		AttemptID     int64
		ItemID        int64
	}{
		{ParticipantID: 101, AttemptID: 1, ItemID: 1},
		{ParticipantID: 101, AttemptID: 1, ItemID: 3},
		{ParticipantID: 101, AttemptID: 1, ItemID: 4},
	} {
		assert.NoError(t, resultStore.Where("participant_id = ? AND attempt_id = ? AND item_id = ?",
			ids.ParticipantID, ids.AttemptID, ids.ItemID).UpdateColumn(
			"score_computed", 100,
		).Error())
	}
	itemUnlockingRuleStore := database.NewDataStore(db).ItemUnlockingRules()
	for unlockingItemID, unlockedItemIDs := range map[int64][]int64{1: {1001, 1002}, 3: {2001, 2002}, 4: {4001, 4002}} {
		for _, unlockedItemID := range unlockedItemIDs {
			assert.NoError(t, itemUnlockingRuleStore.InsertMap(map[string]interface{}{
				"unlocking_item_id": unlockingItemID, "unlocked_item_id": unlockedItemID,
			}))
		}
	}
	assert.NoError(t, itemUnlockingRuleStore.InsertMap(map[string]interface{}{
		"unlocking_item_id": 4, "unlocked_item_id": 4003, "score": 101,
	}))

	err := resultStore.InTransaction(func(s *database.DataStore) error {
		return s.Results().Propagate()
	})
	assert.NoError(t, err)

	var result []unlocksResultRow
	dataStore := database.NewDataStore(db)
	assert.NoError(t, dataStore.PermissionsGranted().
		Select("group_id, item_id, can_view, source_group_id, origin").
		Order("group_id, item_id").
		Scan(&result).Error())
	assert.Equal(t, []unlocksResultRow{
		{GroupID: 101, ItemID: 1001, CanView: "content", SourceGroupID: 101, Origin: "item_unlocking"},
		{GroupID: 101, ItemID: 1002, CanView: "content", SourceGroupID: 101, Origin: "item_unlocking"},
		{GroupID: 101, ItemID: 2001, CanView: "content", SourceGroupID: 101, Origin: "item_unlocking"},
		{GroupID: 101, ItemID: 2002, CanView: "content", SourceGroupID: 101, Origin: "item_unlocking"},
		{GroupID: 101, ItemID: 4001, CanView: "content", SourceGroupID: 101, Origin: "item_unlocking"},
		{GroupID: 101, ItemID: 4002, CanView: "content", SourceGroupID: 101, Origin: "item_unlocking"},
	}, result)
	var count int64
	assert.NoError(t, dataStore.PermissionsGranted().
		Where("TIMESTAMPDIFF(SECOND, latest_update_at, NOW()) > 1").Count(&count).Error())
	assert.Zero(t, count)
	assert.NoError(t, dataStore.Permissions().
		Where("can_view_generated != 'content'").Count(&count).Error())
	assert.Zero(t, count)
	assert.NoError(t, dataStore.PermissionsGranted().
		Where("can_view != 'content'").Count(&count).Error())
	assert.Zero(t, count)

	found, err := dataStore.Results().Where("participant_id = 101").Where("item_id = 2001").HasRows()
	assert.NoError(t, err)
	assert.True(t, found, "should have created a new result for the unlocked item 2001")
}
