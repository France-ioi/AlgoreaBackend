// +build !unit

package database_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/testhelpers"
)

type unlocksResultRow struct {
	GroupID       int64
	ItemID        int64
	CanView       string
	CanEnterFrom  database.Time
	CanEnterUntil database.Time
	SourceGroupID int64
	Origin        string
}

func TestResultStore_Propagate_Unlocks(t *testing.T) {
	db := testhelpers.SetupDBWithFixture("results_propagation/_common", "results_propagation/unlocks")
	defer func() { _ = db.Close() }()

	testRegularUnlocks(db, t)
}

func TestResultStore_Propagate_Unlocks_UpdatesOldRecords(t *testing.T) {
	db := testhelpers.SetupDBWithFixture(
		"results_propagation/_common",
		"results_propagation/unlocks",
		"results_propagation/unlocks_old_records")
	defer func() { _ = db.Close() }()

	testRegularUnlocks(db, t)
}

func TestResultStore_Propagate_Unlocks_KeepsOldGrants(t *testing.T) {
	db := testhelpers.SetupDBWithFixture(
		"results_propagation/_common",
		"results_propagation/unlocks")
	defer func() { _ = db.Close() }()

	oldTS := time.Now().UTC().Add(-time.Minute).Format("2006-01-02 15:04:05")
	grantedPermissions := []map[string]interface{}{
		generateGrantedPermissionsRow("1001", "content", oldTS, "9999-12-31 23:59:58", oldTS),
		generateGrantedPermissionsRow("1002", "content_with_descendants", oldTS, "9999-12-31 23:59:58", oldTS),
		generateGrantedPermissionsRow("2001", "content", oldTS, "9999-12-31 23:59:58", oldTS),
		generateGrantedPermissionsRow("2002", "info", oldTS, "9999-12-31 23:59:58", oldTS),
		generateGrantedPermissionsRow("4001", "none", oldTS, "9999-12-31 23:59:58", oldTS),
		generateGrantedPermissionsRow("4002", "content", oldTS, "9999-12-31 23:59:58", oldTS),
	}
	assert.NoError(t, database.NewDataStore(db).PermissionsGranted().InsertMaps(grantedPermissions))

	prepareDependencies(db, t)
	dataStore := database.NewDataStore(db)
	err := dataStore.InTransaction(func(s *database.DataStore) error {
		s.ScheduleResultsPropagation()
		return nil
	})
	assert.NoError(t, err)

	for i := range grantedPermissions {
		grantedPermissions[i]["updated"] = "0"
		delete(grantedPermissions[i], "latest_update_at")
	}
	const content = "content"
	grantedPermissions[3]["can_view"] = content
	grantedPermissions[4]["can_view"] = content
	grantedPermissions[3]["updated"] = "1"
	grantedPermissions[4]["updated"] = "1"

	var result []map[string]interface{}
	assert.NoError(t, dataStore.PermissionsGranted().
		Select(`
			group_id, item_id, can_view, can_enter_from, can_enter_until, source_group_id, origin,
			ABS(TIMESTAMPDIFF(SECOND, latest_update_at, NOW())) <= 1 AS updated`).
		Order("group_id, item_id").
		ScanIntoSliceOfMaps(&result).Error())
	assert.Equal(t, grantedPermissions, result)
}

func generateGrantedPermissionsRow(itemID, canView, canEnterFrom, canEnterUntil, latestUpdateAt string) map[string]interface{} {
	return map[string]interface{}{
		"group_id": "101", "item_id": itemID, "can_view": canView, "can_enter_from": canEnterFrom,
		"can_enter_until": canEnterUntil, "source_group_id": "101", "origin": "item_unlocking", "latest_update_at": latestUpdateAt,
	}
}

func TestResultStore_Propagate_Unlocks_ItemsRequiringExplicitEntry(t *testing.T) {
	db := testhelpers.SetupDBWithFixture("results_propagation/_common", "results_propagation/unlocks")
	defer func() { _ = db.Close() }()
	assert.NoError(t, db.Exec("UPDATE items SET requires_explicit_entry=1").Error())

	testExplicitEntryUnlocks(db, t)
}

func TestResultStore_Propagate_Unlocks_ItemsRequiringExplicitEntry_EverythingHasBeenSetAlready(t *testing.T) {
	db := testhelpers.SetupDBWithFixture("results_propagation/_common", "results_propagation/unlocks")
	defer func() { _ = db.Close() }()
	assert.NoError(t, db.Exec("UPDATE items SET requires_explicit_entry=1").Error())
	oldTS := time.Now().UTC().Add(-time.Minute).Format("2006-01-02 15:04:05")
	grantedPermissions := []map[string]interface{}{
		generateGrantedPermissionsRow("1001", "content", oldTS, "9999-12-31 23:59:59", oldTS),
		generateGrantedPermissionsRow("1002", "content", oldTS, "9999-12-31 23:59:59", oldTS),
		generateGrantedPermissionsRow("2001", "content", oldTS, "9999-12-31 23:59:59", oldTS),
		generateGrantedPermissionsRow("2002", "content", oldTS, "9999-12-31 23:59:59", oldTS),
		generateGrantedPermissionsRow("4001", "content", oldTS, "9999-12-31 23:59:59", oldTS),
		generateGrantedPermissionsRow("4002", "content", oldTS, "9999-12-31 23:59:59", oldTS),
	}
	assert.NoError(t, database.NewDataStore(db).PermissionsGranted().InsertMaps(grantedPermissions))

	prepareDependencies(db, t)
	dataStore := database.NewDataStore(db)
	err := dataStore.InTransaction(func(s *database.DataStore) error {
		s.ScheduleResultsPropagation()
		return nil
	})
	assert.NoError(t, err)
	var result []map[string]interface{}
	assert.NoError(t, dataStore.PermissionsGranted().
		Select("group_id, item_id, can_view, can_enter_from, can_enter_until, source_group_id, origin, latest_update_at").
		Order("group_id, item_id").
		ScanIntoSliceOfMaps(&result).Error())
	assert.Equal(t, grantedPermissions, result)
}

func TestResultStore_Propagate_Unlocks_ItemsRequiringExplicitEntry_CanEnterFromIsInTheFuture(t *testing.T) {
	db := testhelpers.SetupDBWithFixture("results_propagation/_common", "results_propagation/unlocks")
	defer func() { _ = db.Close() }()
	assert.NoError(t, db.Exec("UPDATE items SET requires_explicit_entry=1").Error())
	oldTS := time.Now().UTC().Add(-time.Minute).Format("2006-01-02 15:04:05")
	futureTS := time.Now().UTC().Add(time.Minute).Format("2006-01-02 15:04:05")
	grantedPermissions := []map[string]interface{}{
		generateGrantedPermissionsRow("1001", "none", futureTS, "9999-12-31 23:59:59", oldTS),
		generateGrantedPermissionsRow("1002", "none", futureTS, "9999-12-31 23:59:59", oldTS),
		generateGrantedPermissionsRow("2001", "none", futureTS, "9999-12-31 23:59:59", oldTS),
		generateGrantedPermissionsRow("2002", "none", futureTS, "9999-12-31 23:59:59", oldTS),
		generateGrantedPermissionsRow("4001", "none", futureTS, "9999-12-31 23:59:59", oldTS),
		generateGrantedPermissionsRow("4002", "none", futureTS, "9999-12-31 23:59:59", oldTS),
	}
	assert.NoError(t, database.NewDataStore(db).PermissionsGranted().InsertMaps(grantedPermissions))

	testExplicitEntryUnlocks(db, t)
}

var maxTime = database.Time(time.Date(9999, 12, 31, 23, 59, 59, 0, time.UTC))

func testRegularUnlocks(db *database.DB, t *testing.T) {
	prepareDependencies(db, t)

	dataStore := database.NewDataStore(db)
	err := dataStore.InTransaction(func(s *database.DataStore) error {
		s.ScheduleResultsPropagation()
		return nil
	})
	assert.NoError(t, err)

	var result []unlocksResultRow
	assert.NoError(t, dataStore.PermissionsGranted().
		Select("group_id, item_id, can_view, can_enter_from, can_enter_until, source_group_id, origin").
		Order("group_id, item_id").
		Scan(&result).Error())
	assert.Equal(t, []unlocksResultRow{
		{
			GroupID: 101, ItemID: 1001, CanView: "content", CanEnterFrom: maxTime, CanEnterUntil: maxTime,
			SourceGroupID: 101, Origin: "item_unlocking",
		},
		{
			GroupID: 101, ItemID: 1002, CanView: "content", CanEnterFrom: maxTime, CanEnterUntil: maxTime,
			SourceGroupID: 101, Origin: "item_unlocking",
		},
		{
			GroupID: 101, ItemID: 2001, CanView: "content", CanEnterFrom: maxTime, CanEnterUntil: maxTime,
			SourceGroupID: 101, Origin: "item_unlocking",
		},
		{
			GroupID: 101, ItemID: 2002, CanView: "content", CanEnterFrom: maxTime, CanEnterUntil: maxTime,
			SourceGroupID: 101, Origin: "item_unlocking",
		},
		{
			GroupID: 101, ItemID: 4001, CanView: "content", CanEnterFrom: maxTime, CanEnterUntil: maxTime,
			SourceGroupID: 101, Origin: "item_unlocking",
		},
		{
			GroupID: 101, ItemID: 4002, CanView: "content", CanEnterFrom: maxTime, CanEnterUntil: maxTime,
			SourceGroupID: 101, Origin: "item_unlocking",
		},
	}, result)
	var count int64
	assert.NoError(t, dataStore.PermissionsGranted().
		Where("TIMESTAMPDIFF(SECOND, latest_update_at, NOW()) > 1").Count(&count).Error())
	assert.Zero(t, count)
	assert.NoError(t, dataStore.Permissions().
		Where("can_view_generated != 'content'").Count(&count).Error())
	assert.Zero(t, count)

	found, err := dataStore.Results().Where("participant_id = 101").Where("item_id = 2001").HasRows()
	assert.NoError(t, err)
	assert.True(t, found, "should have created a new result for the unlocked item 2001")
}

func testExplicitEntryUnlocks(db *database.DB, t *testing.T) {
	prepareDependencies(db, t)
	dataStore := database.NewDataStore(db)
	err := dataStore.InTransaction(func(s *database.DataStore) error {
		s.ScheduleResultsPropagation()
		return nil
	})
	assert.NoError(t, err)

	var result []unlocksResultRow
	assert.NoError(t, dataStore.PermissionsGranted().
		Select("group_id, item_id, can_view, can_enter_until, source_group_id, origin").
		Order("group_id, item_id").
		Scan(&result).Error())
	assert.Equal(t, []unlocksResultRow{
		{GroupID: 101, ItemID: 1001, CanView: "none", CanEnterUntil: maxTime, SourceGroupID: 101, Origin: "item_unlocking"},
		{GroupID: 101, ItemID: 1002, CanView: "none", CanEnterUntil: maxTime, SourceGroupID: 101, Origin: "item_unlocking"},
		{GroupID: 101, ItemID: 2001, CanView: "none", CanEnterUntil: maxTime, SourceGroupID: 101, Origin: "item_unlocking"},
		{GroupID: 101, ItemID: 2002, CanView: "none", CanEnterUntil: maxTime, SourceGroupID: 101, Origin: "item_unlocking"},
		{GroupID: 101, ItemID: 4001, CanView: "none", CanEnterUntil: maxTime, SourceGroupID: 101, Origin: "item_unlocking"},
		{GroupID: 101, ItemID: 4002, CanView: "none", CanEnterUntil: maxTime, SourceGroupID: 101, Origin: "item_unlocking"},
	}, result)
	var count int64
	assert.NoError(t, dataStore.PermissionsGranted().
		Where("TIMESTAMPDIFF(SECOND, latest_update_at, NOW()) > 1").Count(&count).Error())
	assert.Zero(t, count)
	assert.NoError(t, dataStore.PermissionsGranted().
		Where("TIMESTAMPDIFF(SECOND, can_enter_from, NOW()) > 1").Count(&count).Error())
	assert.Zero(t, count)
	assert.NoError(t, dataStore.Permissions().
		Where("can_view_generated != 'none'").Count(&count).Error())
	assert.Zero(t, count)

	found, err := dataStore.Results().Where("participant_id = 101").Where("item_id = 2001").HasRows()
	assert.NoError(t, err)
	assert.False(t, found, "should not have created a new result for the unlocked item 2001")
}

func prepareDependencies(db *database.DB, t *testing.T) {
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
	itemDependencyStore := database.NewDataStore(db).ItemDependencies()
	for itemID, dependentItemIDs := range map[int64][]int64{1: {1001, 1002}, 3: {2001, 2002}, 4: {4001, 4002}} {
		for _, dependentItemID := range dependentItemIDs {
			assert.NoError(t, itemDependencyStore.InsertMap(map[string]interface{}{
				"item_id": itemID, "dependent_item_id": dependentItemID, "score": 100,
			}))
		}
	}
	assert.NoError(t, itemDependencyStore.InsertMap(map[string]interface{}{
		"item_id": 4, "dependent_item_id": 4003, "score": 101,
	}))
	assert.NoError(t, itemDependencyStore.InsertMap(map[string]interface{}{
		"item_id": 4, "dependent_item_id": 4004, "score": 1, "grant_content_view": false,
	}))
}
