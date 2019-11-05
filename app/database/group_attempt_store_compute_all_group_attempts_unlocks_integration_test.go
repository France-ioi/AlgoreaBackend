// +build !unit

package database_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/logging"
	"github.com/France-ioi/AlgoreaBackend/app/loggingtest"
	"github.com/France-ioi/AlgoreaBackend/testhelpers"
)

type unlocksResultRow struct {
	GroupID int64
	ItemID  int64
	CanView string
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

func TestGroupAttemptStore_ComputeAllGroupAttempts_Unlocks_WarnsWhenIdIsNotInteger(t *testing.T) {
	db := testhelpers.SetupDBWithFixture(
		"groups_attempts_propagation/_common",
		"groups_attempts_propagation/unlocks",
	)
	defer func() { _ = db.Close() }()

	hook, restoreFunc := logging.MockSharedLoggerHook()
	defer restoreFunc()

	groupAttemptStore := database.NewDataStore(db).GroupAttempts()
	assert.NoError(t, groupAttemptStore.Where("id=11").UpdateColumn(
		"key_obtained", 1,
	).Error())
	itemStore := database.NewDataStore(db).Items()
	assert.NoError(t, itemStore.Where("id=1").UpdateColumn(
		"unlocked_item_ids", "1001,abc",
	).Error())

	err := groupAttemptStore.InTransaction(func(s *database.DataStore) error {
		return s.GroupAttempts().ComputeAllGroupAttempts()
	})
	assert.NoError(t, err)

	logs := strings.Split((&loggingtest.Hook{Hook: hook}).GetAllStructuredLogs(), "\n")
	assert.Len(t, logs, 1)
	assert.Contains(t, logs[0], `level=warning`)
	assert.Contains(t, logs[0], `msg="cannot parse items.unlocked_item_ids"`)
	assert.Contains(t, logs[0], `error="strconv.ParseInt: parsing \"abc\": invalid syntax"`)
	assert.Contains(t, logs[0], `items.id=1`)
	assert.Contains(t, logs[0], `items.unlocked_item_ids="1001,abc"`)
}

func testUnlocks(db *database.DB, t *testing.T) {
	groupAttemptStore := database.NewDataStore(db).GroupAttempts()
	assert.NoError(t, groupAttemptStore.Where("id=11").UpdateColumn(
		"key_obtained", 1,
	).Error())
	assert.NoError(t, groupAttemptStore.Where("id=13").UpdateColumn(
		"key_obtained", 1,
	).Error())
	assert.NoError(t, groupAttemptStore.Where("id=14").UpdateColumn(
		"key_obtained", 1,
	).Error())
	itemStore := database.NewDataStore(db).Items()
	assert.NoError(t, itemStore.Where("id=1").UpdateColumn(
		"unlocked_item_ids", "1001,1002",
	).Error())
	assert.NoError(t, itemStore.Where("id=3").UpdateColumn(
		"unlocked_item_ids", "2001,2002",
	).Error())
	assert.NoError(t, itemStore.Where("id=4").UpdateColumn(
		"unlocked_item_ids", "4001,4002",
	).Error())

	err := groupAttemptStore.InTransaction(func(s *database.DataStore) error {
		return s.GroupAttempts().ComputeAllGroupAttempts()
	})
	assert.NoError(t, err)

	var result []unlocksResultRow
	assert.NoError(t, database.NewDataStore(db).PermissionsGranted().
		Select("group_id, item_id, can_view").
		Order("group_id, item_id").
		Scan(&result).Error())
	assert.Equal(t, []unlocksResultRow{
		{GroupID: 101, ItemID: 1001, CanView: "content"},
		{GroupID: 101, ItemID: 1002, CanView: "content"},
		{GroupID: 101, ItemID: 2001, CanView: "content"},
		{GroupID: 101, ItemID: 2002, CanView: "content"},
		{GroupID: 101, ItemID: 4001, CanView: "content"},
		{GroupID: 101, ItemID: 4002, CanView: "content"},
	}, result)
	var count int64
	assert.NoError(t, database.NewDataStore(db).PermissionsGranted().
		Where("TIMESTAMPDIFF(SECOND, latest_update_on, NOW()) > 1").Count(&count).Error())
	assert.Zero(t, count)
	assert.NoError(t, database.NewDataStore(db).PermissionsGenerated().
		Where("can_view_generated != 'content'").Count(&count).Error())
	assert.Zero(t, count)
	assert.NoError(t, database.NewDataStore(db).PermissionsGranted().
		Where("can_view != 'content'").Count(&count).Error())
	assert.Zero(t, count)
}
