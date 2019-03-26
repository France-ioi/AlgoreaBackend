// +build !unit

package database_test

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/logging"
	"github.com/France-ioi/AlgoreaBackend/app/loggingtest"
	"github.com/France-ioi/AlgoreaBackend/testhelpers"
)

type unlocksResultRow struct {
	IdGroup                 int64      `gorm:"column:idGroup"`
	IdItem                  int64      `gorm:"column:idItem"`
	PartialAccessDate       *time.Time `gorm:"column:sPartialAccessDate"`
	CachedPartialAccessDate *time.Time `gorm:"column:sCachedPartialAccessDate"`
	CachedPartialAccess     bool       `gorm:"column:bCachedPartialAccess"`
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
	assert.NoError(t, userItemStore.Where("ID=11").UpdateColumn(
		"bKeyObtained", 1,
	).Error())
	itemStore := database.NewDataStore(db).Items()
	assert.NoError(t, itemStore.Where("ID=1").UpdateColumn(
		"idItemUnlocked", "1001,abc",
	).Error())

	assert.NoError(t, userItemStore.ComputeAllUserItems())

	logs := strings.Split((&loggingtest.Hook{hook}).GetAllStructuredLogs(), "\n")
	assert.Len(t, logs, 1)
	assert.Contains(t, logs[0], `level=warning`)
	assert.Contains(t, logs[0], `msg="cannot parse items.idItemUnlocked"`)
	assert.Contains(t, logs[0], `error="strconv.ParseInt: parsing \"abc\": invalid syntax"`)
	assert.Contains(t, logs[0], `items.ID=1`)
	assert.Contains(t, logs[0], `items.idItemUnlocked="1001,abc"`)
}

func testUnlocks(db *database.DB, t *testing.T) {
	userItemStore := database.NewDataStore(db).UserItems()
	assert.NoError(t, userItemStore.Where("ID=11").UpdateColumn(
		"bKeyObtained", 1,
	).Error())
	assert.NoError(t, userItemStore.Where("ID=13").UpdateColumn(
		"bKeyObtained", 1,
	).Error())
	assert.NoError(t, userItemStore.Where("ID=14").UpdateColumn(
		"bKeyObtained", 1,
	).Error())
	itemStore := database.NewDataStore(db).Items()
	assert.NoError(t, itemStore.Where("ID=1").UpdateColumn(
		"idItemUnlocked", "1001,1002",
	).Error())
	assert.NoError(t, itemStore.Where("ID=3").UpdateColumn(
		"idItemUnlocked", "2001,2002",
	).Error())
	assert.NoError(t, itemStore.Where("ID=4").UpdateColumn(
		"idItemUnlocked", "4001,4002",
	).Error())
	err := userItemStore.ComputeAllUserItems()
	assert.NoError(t, err)
	var result []unlocksResultRow
	assert.NoError(t, database.NewDataStore(db).GroupItems().
		Select("idGroup, idItem, bCachedPartialAccess").
		Order("idGroup, idItem").
		Scan(&result).Error())
	assert.Equal(t, []unlocksResultRow{
		{IdGroup: 101, IdItem: 1001, CachedPartialAccess: true},
		{IdGroup: 101, IdItem: 1002, CachedPartialAccess: true},
		{IdGroup: 101, IdItem: 2001, CachedPartialAccess: true},
		{IdGroup: 101, IdItem: 2002, CachedPartialAccess: true},
		{IdGroup: 101, IdItem: 4001, CachedPartialAccess: true},
		{IdGroup: 101, IdItem: 4002, CachedPartialAccess: true},
	}, result)
	var count int64
	assert.NoError(t, database.NewDataStore(db).GroupItems().
		Where("TIMESTAMPDIFF(SECOND, sCachedPartialAccessDate, NOW()) > 1").Count(&count).Error())
	assert.Zero(t, count)
	assert.NoError(t, database.NewDataStore(db).GroupItems().
		Where("TIMESTAMPDIFF(SECOND, sPartialAccessDate, NOW()) > 1").Count(&count).Error())
	assert.Zero(t, count)
	assert.NoError(t, database.NewDataStore(db).GroupItems().
		Where("sCachedPartialAccessDate IS NULL").Count(&count).Error())
	assert.Zero(t, count)
	assert.NoError(t, database.NewDataStore(db).GroupItems().
		Where("sPartialAccessDate IS NULL").Count(&count).Error())
	assert.Zero(t, count)
}
