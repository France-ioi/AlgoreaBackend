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
	GroupID                 int64
	ItemID                  int64
	PartialAccessDate       *database.Time
	CachedPartialAccessDate *database.Time
	CachedPartialAccess     bool
}

func TestUserItemStore_ComputeAllUserItems_Unlocks(t *testing.T) {
	db := testhelpers.SetupDBWithFixture("users_items_propagation/_common", "users_items_propagation/unlocks")
	defer func() { _ = db.Close() }()

	testUnlocks(db, t)
}

func TestUserItemStore_ComputeAllUserItems_Unlocks_UpdatesOldRecords(t *testing.T) {
	db := testhelpers.SetupDBWithFixture(
		"users_items_propagation/_common",
		"users_items_propagation/unlocks",
		"users_items_propagation/unlocks_old_records")
	defer func() { _ = db.Close() }()

	testUnlocks(db, t)
}

func TestUserItemStore_ComputeAllUserItems_Unlocks_WarnsWhenIdIsNotInteger(t *testing.T) {
	db := testhelpers.SetupDBWithFixture(
		"users_items_propagation/_common",
		"users_items_propagation/unlocks",
	)
	defer func() { _ = db.Close() }()

	hook, restoreFunc := logging.MockSharedLoggerHook()
	defer restoreFunc()

	userItemStore := database.NewDataStore(db).UserItems()
	assert.NoError(t, userItemStore.Where("id=11").UpdateColumn(
		"key_obtained", 1,
	).Error())
	itemStore := database.NewDataStore(db).Items()
	assert.NoError(t, itemStore.Where("id=1").UpdateColumn(
		"item_unlocked_id", "1001,abc",
	).Error())

	err := userItemStore.InTransaction(func(s *database.DataStore) error {
		return s.UserItems().ComputeAllUserItems()
	})
	assert.NoError(t, err)

	logs := strings.Split((&loggingtest.Hook{Hook: hook}).GetAllStructuredLogs(), "\n")
	assert.Len(t, logs, 1)
	assert.Contains(t, logs[0], `level=warning`)
	assert.Contains(t, logs[0], `msg="cannot parse items.item_unlocked_id"`)
	assert.Contains(t, logs[0], `error="strconv.ParseInt: parsing \"abc\": invalid syntax"`)
	assert.Contains(t, logs[0], `items.id=1`)
	assert.Contains(t, logs[0], `items.item_unlocked_id="1001,abc"`)
}

func testUnlocks(db *database.DB, t *testing.T) {
	userItemStore := database.NewDataStore(db).UserItems()
	assert.NoError(t, userItemStore.Where("id=11").UpdateColumn(
		"key_obtained", 1,
	).Error())
	assert.NoError(t, userItemStore.Where("id=13").UpdateColumn(
		"key_obtained", 1,
	).Error())
	assert.NoError(t, userItemStore.Where("id=14").UpdateColumn(
		"key_obtained", 1,
	).Error())
	itemStore := database.NewDataStore(db).Items()
	assert.NoError(t, itemStore.Where("id=1").UpdateColumn(
		"item_unlocked_id", "1001,1002",
	).Error())
	assert.NoError(t, itemStore.Where("id=3").UpdateColumn(
		"item_unlocked_id", "2001,2002",
	).Error())
	assert.NoError(t, itemStore.Where("id=4").UpdateColumn(
		"item_unlocked_id", "4001,4002",
	).Error())

	err := userItemStore.InTransaction(func(s *database.DataStore) error {
		return s.UserItems().ComputeAllUserItems()
	})
	assert.NoError(t, err)

	var result []unlocksResultRow
	assert.NoError(t, database.NewDataStore(db).GroupItems().
		Select("group_id, item_id, cached_partial_access").
		Order("group_id, item_id").
		Scan(&result).Error())
	assert.Equal(t, []unlocksResultRow{
		{GroupID: 101, ItemID: 1001, CachedPartialAccess: true},
		{GroupID: 101, ItemID: 1002, CachedPartialAccess: true},
		{GroupID: 101, ItemID: 2001, CachedPartialAccess: true},
		{GroupID: 101, ItemID: 2002, CachedPartialAccess: true},
		{GroupID: 101, ItemID: 4001, CachedPartialAccess: true},
		{GroupID: 101, ItemID: 4002, CachedPartialAccess: true},
	}, result)
	var count int64
	assert.NoError(t, database.NewDataStore(db).GroupItems().
		Where("TIMESTAMPDIFF(SECOND, cached_partial_access_date, NOW()) > 1").Count(&count).Error())
	assert.Zero(t, count)
	assert.NoError(t, database.NewDataStore(db).GroupItems().
		Where("TIMESTAMPDIFF(SECOND, partial_access_date, NOW()) > 1").Count(&count).Error())
	assert.Zero(t, count)
	assert.NoError(t, database.NewDataStore(db).GroupItems().
		Where("cached_partial_access_date IS NULL").Count(&count).Error())
	assert.Zero(t, count)
	assert.NoError(t, database.NewDataStore(db).GroupItems().
		Where("partial_access_date IS NULL").Count(&count).Error())
	assert.Zero(t, count)
}
