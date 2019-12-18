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

func TestGroupAttemptStore_ComputeAllGroupAttempts_Unlocks(t *testing.T) {
	db := testhelpers.SetupDBWithFixture("groups_attempts_propagation/_common", "groups_attempts_propagation/unlocks")
	defer func() { _ = db.Close() }()

	testUnlocks(db, t)
}

func TestGroupAttemptStore_ComputeAllGroupAttempts_Unlocks_UpdatesOldRecords(t *testing.T) {
	db := testhelpers.SetupDBWithFixture(
		"groups_attempts_propagation/_common",
		"groups_attempts_propagation/unlocks",
		"groups_attempts_propagation/unlocks_old_records")
	defer func() { _ = db.Close() }()

	testUnlocks(db, t)
}

func testUnlocks(db *database.DB, t *testing.T) {
	groupAttemptStore := database.NewDataStore(db).GroupAttempts()
	for _, id := range []int64{11, 13, 14} {
		assert.NoError(t, groupAttemptStore.Where("id = ?", id).UpdateColumn(
			"score", 100,
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

	err := groupAttemptStore.InTransaction(func(s *database.DataStore) error {
		return s.GroupAttempts().ComputeAllGroupAttempts()
	})
	assert.NoError(t, err)

	var result []unlocksResultRow
	assert.NoError(t, database.NewDataStore(db).PermissionsGranted().
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
	assert.NoError(t, database.NewDataStore(db).PermissionsGranted().
		Where("TIMESTAMPDIFF(SECOND, latest_update_on, NOW()) > 1").Count(&count).Error())
	assert.Zero(t, count)
	assert.NoError(t, database.NewDataStore(db).Permissions().
		Where("can_view_generated != 'content'").Count(&count).Error())
	assert.Zero(t, count)
	assert.NoError(t, database.NewDataStore(db).PermissionsGranted().
		Where("can_view != 'content'").Count(&count).Error())
	assert.Zero(t, count)
}
