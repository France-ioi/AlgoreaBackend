// +build !unit

package database_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/testhelpers"
)

type groupItemsAggregatesResultRow struct {
	GroupID                    int64
	ItemID                     int64
	CachedFullAccessSince      *database.Time
	FullAccessSince            *database.Time
	CachedFullAccess           bool
	CachedPartialAccessSince   *database.Time
	PartialAccessSince         *database.Time
	CachedPartialAccess        bool
	ManagerAccess              bool
	CachedManagerAccess        bool
	CachedGrayedAccessSince    *database.Time
	CachedGrayedAccess         bool
	CachedSolutionsAccessSince *database.Time
	SolutionsAccessSince       *database.Time
	CachedAccessSolutions      bool
	CachedAccessReason         *string
	AccessReason               *string
}

var expectedRow14 = groupItemsAggregatesResultRow{
	GroupID:                    1,
	ItemID:                     4,
	CachedFullAccessSince:      (*database.Time)(ptrTime(time.Date(2017, 1, 1, 12, 13, 14, 0, time.UTC))),
	FullAccessSince:            (*database.Time)(ptrTime(time.Date(2017, 1, 1, 12, 13, 14, 0, time.UTC))),
	CachedFullAccess:           true,
	CachedPartialAccessSince:   (*database.Time)(ptrTime(time.Date(2017, 1, 1, 12, 13, 14, 0, time.UTC))),
	PartialAccessSince:         (*database.Time)(ptrTime(time.Date(2017, 1, 1, 12, 13, 14, 0, time.UTC))),
	CachedPartialAccess:        true,
	CachedManagerAccess:        true,
	ManagerAccess:              true,
	CachedGrayedAccess:         false,
	CachedSolutionsAccessSince: (*database.Time)(ptrTime(time.Date(2017, 1, 1, 12, 13, 14, 0, time.UTC))),
	SolutionsAccessSince:       (*database.Time)(ptrTime(time.Date(2017, 1, 1, 12, 13, 14, 0, time.UTC))),
	CachedAccessSolutions:      true,
	AccessReason:               ptrString("some reason"),
}

func TestGroupItemStore_ComputeAllAccess_AggregatesCachedFullAccessDate(t *testing.T) {
	db := testhelpers.SetupDBWithFixture("group_item_store/compute_all_access/_common")
	defer func() { _ = db.Close() }()

	groupItemStore := database.NewDataStore(db).GroupItems()
	currentDate := time.Now().UTC().Round(time.Second)
	assert.NoError(t, groupItemStore.Where("group_id=1 AND item_id=1").
		UpdateColumn("full_access_since", currentDate.AddDate(0, 0, -8)).Error())
	assert.NoError(t, groupItemStore.Where("group_id=1 AND item_id=2").
		UpdateColumn("full_access_since", currentDate.AddDate(0, 0, -9)).Error())
	assert.NoError(t, groupItemStore.Exec(`
		DELETE groups_items_propagate FROM groups_items_propagate JOIN groups_items USING (id)
			WHERE group_id=1 AND item_id=2`).Error())
	assert.NoError(t, groupItemStore.Where("group_id=1 AND item_id=3").
		UpdateColumn("cached_full_access_since", currentDate.AddDate(0, 0, -10)).Error())
	assert.NoError(t, groupItemStore.Exec(`
		DELETE groups_items_propagate FROM groups_items_propagate JOIN groups_items USING (id)
			WHERE group_id=1 AND item_id=3`).Error())
	assert.NoError(t, groupItemStore.Where("group_id=2 AND item_id=1").
		UpdateColumn("full_access_since", currentDate.AddDate(0, 0, -11)).Error())
	assert.NoError(t, groupItemStore.Where("group_id=2 AND item_id=11").
		UpdateColumn("full_access_since", currentDate.AddDate(0, 0, -12)).Error())
	assert.NoError(t, groupItemStore.UpdateColumn("cached_full_access", 1).Error())
	assert.NoError(t, groupItemStore.InTransaction(func(ds *database.DataStore) error {
		ds.GroupItems().ComputeAllAccess()
		return nil
	}))

	assertAllGroupItemsAreDone(t, groupItemStore)

	var result []groupItemsAggregatesResultRow
	assert.NoError(t, groupItemStore.Order("group_id, item_id").Scan(&result).Error())
	assertGroupsItemsAggregatesResultRowsEqual(t, []groupItemsAggregatesResultRow{
		{
			GroupID:               1,
			ItemID:                1,
			CachedFullAccessSince: (*database.Time)(ptrTime(currentDate.AddDate(0, 0, -8))),
			FullAccessSince:       (*database.Time)(ptrTime(currentDate.AddDate(0, 0, -8))),
			CachedFullAccess:      true,
		},
		{
			GroupID:               1,
			ItemID:                2,
			CachedFullAccessSince: nil, // since it has been already done
			FullAccessSince:       (*database.Time)(ptrTime(currentDate.AddDate(0, 0, -9))),
			CachedFullAccess:      true, // since it has been already done
		},
		{
			GroupID:               1,
			ItemID:                3,
			CachedFullAccessSince: (*database.Time)(ptrTime(currentDate.AddDate(0, 0, -10))), // since it has been already done
			FullAccessSince:       nil,
			CachedFullAccess:      true,
		},
		expectedRow14,
		{
			GroupID:               1,
			ItemID:                11,
			CachedFullAccessSince: (*database.Time)(ptrTime(currentDate.AddDate(0, 0, -10))), // least(1_1, 1_3)

			CachedFullAccess: false, // since it is a new record and we do not set cached_full_access=1 in ComputeAllAccess()
		},
		{
			GroupID:               1,
			ItemID:                12,
			CachedFullAccessSince: (*database.Time)(ptrTime(currentDate.AddDate(0, 0, -10))), // least(1_1, 1_2, 1_11)
		},
		{
			GroupID:               2,
			ItemID:                1,
			CachedFullAccessSince: (*database.Time)(ptrTime(currentDate.AddDate(0, 0, -11))),
			FullAccessSince:       (*database.Time)(ptrTime(currentDate.AddDate(0, 0, -11))),
			CachedFullAccess:      true,
		},
		{
			GroupID:               2,
			ItemID:                11,
			CachedFullAccessSince: (*database.Time)(ptrTime(currentDate.AddDate(0, 0, -12))),
			FullAccessSince:       (*database.Time)(ptrTime(currentDate.AddDate(0, 0, -12))),
			CachedFullAccess:      true,
		},
		{
			GroupID:               2,
			ItemID:                12,
			CachedFullAccessSince: (*database.Time)(ptrTime(currentDate.AddDate(0, 0, -12))),
			CachedFullAccess:      false, // since it is a new record and we do not set cached_full_access=1 in ComputeAllAccess()
		},
	}, result)
}

func TestGroupItemStore_ComputeAllAccess_AggregatesCachedPartialAccessDate(t *testing.T) {
	db := testhelpers.SetupDBWithFixture("group_item_store/compute_all_access/_common")
	defer func() { _ = db.Close() }()

	groupItemStore := database.NewDataStore(db).GroupItems()
	currentDate := time.Now().UTC().Round(time.Second)
	assert.NoError(t, groupItemStore.Where("group_id=1 AND item_id=1").
		UpdateColumn("partial_access_since", currentDate.AddDate(0, 0, -8)).Error())
	assert.NoError(t, groupItemStore.Where("group_id=1 AND item_id=2").
		UpdateColumn("partial_access_since", currentDate.AddDate(0, 0, -9)).Error())
	assert.NoError(t, groupItemStore.Exec(`
		DELETE groups_items_propagate FROM groups_items_propagate JOIN groups_items USING (id)
			WHERE group_id=1 AND item_id=2`).Error())
	assert.NoError(t, groupItemStore.Where("group_id=1 AND item_id=3").
		UpdateColumn("cached_partial_access_since", currentDate.AddDate(0, 0, -10)).Error())
	assert.NoError(t, groupItemStore.Exec(`
		DELETE groups_items_propagate FROM groups_items_propagate JOIN groups_items USING (id)
			WHERE group_id=1 AND item_id=3`).Error())
	assert.NoError(t, groupItemStore.Where("group_id=2 AND item_id=1").
		UpdateColumn("partial_access_since", currentDate.AddDate(0, 0, -11)).Error())
	assert.NoError(t, groupItemStore.Where("group_id=2 AND item_id=11").
		UpdateColumn("partial_access_since", currentDate.AddDate(0, 0, -12)).Error())
	assert.NoError(t, groupItemStore.UpdateColumn("cached_partial_access", 1).Error())
	assert.NoError(t, groupItemStore.InTransaction(func(ds *database.DataStore) error {
		ds.GroupItems().ComputeAllAccess()
		return nil
	}))

	assertAllGroupItemsAreDone(t, groupItemStore)

	var result []groupItemsAggregatesResultRow
	assert.NoError(t, groupItemStore.Order("group_id, item_id").Scan(&result).Error())
	assertGroupsItemsAggregatesResultRowsEqual(t, []groupItemsAggregatesResultRow{
		{
			GroupID:                  1,
			ItemID:                   1,
			CachedPartialAccessSince: (*database.Time)(ptrTime(currentDate.AddDate(0, 0, -8))),
			PartialAccessSince:       (*database.Time)(ptrTime(currentDate.AddDate(0, 0, -8))),
			CachedPartialAccess:      true,
		},
		{
			GroupID:                  1,
			ItemID:                   2,
			CachedPartialAccessSince: nil, // since it has been already done
			PartialAccessSince:       (*database.Time)(ptrTime(currentDate.AddDate(0, 0, -9))),
			CachedPartialAccess:      true, // since it has been already done
		},
		{
			GroupID:                  1,
			ItemID:                   3,
			CachedPartialAccessSince: (*database.Time)(ptrTime(currentDate.AddDate(0, 0, -10))), // since it has been already done
			PartialAccessSince:       nil,
			CachedPartialAccess:      true,
		},
		expectedRow14,
		{
			GroupID:                  1,
			ItemID:                   11,
			CachedPartialAccessSince: (*database.Time)(ptrTime(currentDate.AddDate(0, 0, -10))), // least(1_1, 1_3)

			CachedPartialAccess: false, // since it is a new record and we do not set cached_partial_access=1 in ComputeAllAccess()
		},
		{
			GroupID:                  1,
			ItemID:                   12,
			CachedPartialAccessSince: nil, // since partial_access_propagation!='AsPartial'
		},
		{
			GroupID:                  2,
			ItemID:                   1,
			CachedPartialAccessSince: (*database.Time)(ptrTime(currentDate.AddDate(0, 0, -11))),
			PartialAccessSince:       (*database.Time)(ptrTime(currentDate.AddDate(0, 0, -11))),
			CachedPartialAccess:      true,
		},
		{
			GroupID:                  2,
			ItemID:                   11,
			CachedPartialAccessSince: (*database.Time)(ptrTime(currentDate.AddDate(0, 0, -12))),
			PartialAccessSince:       (*database.Time)(ptrTime(currentDate.AddDate(0, 0, -12))),
			CachedPartialAccess:      true,
		},
		{
			GroupID:                  2,
			ItemID:                   12,
			CachedPartialAccessSince: nil, // since partial_access_propagation!='AsPartial'
		},
	}, result)
}

func TestGroupItemStore_ComputeAllAccess_AggregatesCachedManagerAccess(t *testing.T) {
	db := testhelpers.SetupDBWithFixture("group_item_store/compute_all_access/_common")
	defer func() { _ = db.Close() }()

	groupItemStore := database.NewDataStore(db).GroupItems()
	assert.NoError(t, groupItemStore.Where("group_id=1 AND item_id=1").
		UpdateColumn("manager_access", 1).Error())
	assert.NoError(t, groupItemStore.Where("group_id=1 AND item_id=2").
		UpdateColumn("manager_access", 1).Error())
	assert.NoError(t, groupItemStore.Exec(`
		DELETE groups_items_propagate FROM groups_items_propagate JOIN groups_items USING (id)
			WHERE group_id=1 AND item_id=2`).Error())
	assert.NoError(t, groupItemStore.Where("group_id=1 AND item_id=3").
		UpdateColumn("cached_manager_access", 1).Error())
	assert.NoError(t, groupItemStore.Exec(`
		DELETE groups_items_propagate FROM groups_items_propagate JOIN groups_items USING (id)
			WHERE group_id=1 AND item_id=3`).Error())
	assert.NoError(t, groupItemStore.Where("group_id=2 AND item_id=1").
		UpdateColumn("manager_access", 1).Error())
	assert.NoError(t, groupItemStore.Where("group_id=2 AND item_id=11").
		UpdateColumn("manager_access", 1).Error())
	assert.NoError(t, groupItemStore.InTransaction(func(ds *database.DataStore) error {
		ds.GroupItems().ComputeAllAccess()
		return nil
	}))

	assertAllGroupItemsAreDone(t, groupItemStore)

	var result []groupItemsAggregatesResultRow
	assert.NoError(t, groupItemStore.Order("group_id, item_id").Scan(&result).Error())
	assertGroupsItemsAggregatesResultRowsEqual(t, []groupItemsAggregatesResultRow{
		{
			GroupID:             1,
			ItemID:              1,
			CachedManagerAccess: true,
			ManagerAccess:       true,
		},
		{
			GroupID:             1,
			ItemID:              2,
			CachedManagerAccess: false, // since it has been already done
			ManagerAccess:       true,
		},
		{
			GroupID:             1,
			ItemID:              3,
			CachedManagerAccess: true,
			ManagerAccess:       false,
		},
		expectedRow14,
		{
			GroupID:             1,
			ItemID:              11,
			CachedManagerAccess: true,
		},
		{
			GroupID:             1,
			ItemID:              12,
			CachedManagerAccess: true,
		},
		{
			GroupID:             2,
			ItemID:              1,
			CachedManagerAccess: true,
			ManagerAccess:       true,
		},
		{
			GroupID:             2,
			ItemID:              11,
			CachedManagerAccess: true,
			ManagerAccess:       true,
		},
		{
			GroupID:             2,
			ItemID:              12,
			CachedManagerAccess: true,
		},
	}, result)
}

func TestGroupItemStore_ComputeAllAccess_AggregatesCachedGrayedAccessDate(t *testing.T) {
	db := testhelpers.SetupDBWithFixture("group_item_store/compute_all_access/_common")
	defer func() { _ = db.Close() }()

	groupItemStore := database.NewDataStore(db).GroupItems()
	currentDate := time.Now().UTC().Round(time.Second)
	assert.NoError(t, groupItemStore.Where("group_id=1 AND item_id=1").
		UpdateColumn("partial_access_since", currentDate.AddDate(0, 0, -8)).Error())
	assert.NoError(t, groupItemStore.Where("group_id=1 AND item_id=2").
		UpdateColumn("partial_access_since", currentDate.AddDate(0, 0, -9)).Error())
	assert.NoError(t, groupItemStore.Where("group_id=1 AND item_id=3").
		UpdateColumn("cached_partial_access_since", currentDate.AddDate(0, 0, -10)).Error())
	assert.NoError(t, groupItemStore.Exec(`
		DELETE groups_items_propagate FROM groups_items_propagate JOIN groups_items USING (id)
			WHERE group_id=1 AND item_id=3`).Error())
	assert.NoError(t, groupItemStore.Where("group_id=2 AND item_id=1").
		UpdateColumn("partial_access_since", currentDate.AddDate(0, 0, -11)).Error())
	assert.NoError(t, groupItemStore.Exec(`
		DELETE groups_items_propagate FROM groups_items_propagate JOIN groups_items USING (id)
			WHERE group_id=1 AND item_id=2`).Error())
	assert.NoError(t, groupItemStore.Where("group_id=2 AND item_id=11").
		UpdateColumn("partial_access_since", currentDate.AddDate(0, 0, -12)).Error())
	assert.NoError(t, groupItemStore.UpdateColumn("cached_grayed_access", 1).Error())
	assert.NoError(t, groupItemStore.ItemItems().Updates(map[string]interface{}{
		"partial_access_propagation": "AsGrayed",
	}).Error())
	assert.NoError(t, groupItemStore.InTransaction(func(ds *database.DataStore) error {
		ds.GroupItems().ComputeAllAccess()
		return nil
	}))

	assertAllGroupItemsAreDone(t, groupItemStore)

	var result []groupItemsAggregatesResultRow
	assert.NoError(t, groupItemStore.Order("group_id, item_id").Scan(&result).Error())
	assertGroupsItemsAggregatesResultRowsEqual(t, []groupItemsAggregatesResultRow{
		{
			GroupID:                  1,
			ItemID:                   1,
			CachedPartialAccessSince: (*database.Time)(ptrTime(currentDate.AddDate(0, 0, -8))),
			PartialAccessSince:       (*database.Time)(ptrTime(currentDate.AddDate(0, 0, -8))),
			CachedGrayedAccess:       false, // since we have partial access here
		},
		{
			GroupID:                  1,
			ItemID:                   2,
			CachedPartialAccessSince: nil, // since it has been already done
			PartialAccessSince:       (*database.Time)(ptrTime(currentDate.AddDate(0, 0, -9))),
			CachedGrayedAccess:       true, // since it has been already done
		},
		{
			GroupID:                  1,
			ItemID:                   3,
			CachedPartialAccessSince: (*database.Time)(ptrTime(currentDate.AddDate(0, 0, -10))), // since it has been already done
			PartialAccessSince:       nil,
			CachedGrayedAccess:       true,
		},
		expectedRow14,
		{
			GroupID:                  1,
			ItemID:                   11,
			CachedPartialAccessSince: nil, // since partial_access_propagation != "AsPartial"

			CachedGrayedAccessSince: (*database.Time)(ptrTime(currentDate.AddDate(0, 0, -10))), // least cached_partial_access_since(1_1, 1_3)

			CachedGrayedAccess: false, // since it is a new record and we do not set cached_grayed_access=1 in ComputeAllAccess()
		},
		{
			GroupID:                  1,
			ItemID:                   12,
			CachedPartialAccessSince: nil, // since partial_access_propagation != "AsPartial"

			CachedGrayedAccessSince: (*database.Time)(ptrTime(currentDate.AddDate(0, 0, -8))), // from 1_1

			CachedGrayedAccess: false, // since it is a new record and we do not set cached_grayed_access=1 in ComputeAllAccess()
		},
		{
			GroupID:                  2,
			ItemID:                   1,
			CachedPartialAccessSince: (*database.Time)(ptrTime(currentDate.AddDate(0, 0, -11))),
			PartialAccessSince:       (*database.Time)(ptrTime(currentDate.AddDate(0, 0, -11))),
		},
		{
			GroupID:                  2,
			ItemID:                   11,
			CachedPartialAccessSince: (*database.Time)(ptrTime(currentDate.AddDate(0, 0, -12))),
			PartialAccessSince:       (*database.Time)(ptrTime(currentDate.AddDate(0, 0, -12))),
			CachedGrayedAccessSince:  (*database.Time)(ptrTime(currentDate.AddDate(0, 0, -11))), // from 2_1

			CachedGrayedAccess: false, // since it is a new record and we do not set cached_grayed_access=1 in ComputeAllAccess()
		},
		{
			GroupID: 2,
			ItemID:  12,

			CachedPartialAccessSince: nil, // since partial_access_propagation != "AsPartial"

			CachedGrayedAccessSince: (*database.Time)(ptrTime(currentDate.AddDate(0, 0, -12))), // from 2_11

			CachedGrayedAccess: false, // since it is a new record and we do not set cached_grayed_access=1 in ComputeAllAccess()
		},
	}, result)
}

func TestGroupItemStore_ComputeAllAccess_AggregatesCachedAccessSolutionsDate(t *testing.T) {
	db := testhelpers.SetupDBWithFixture("group_item_store/compute_all_access/_common")
	defer func() { _ = db.Close() }()

	groupItemStore := database.NewDataStore(db).GroupItems()
	currentDate := time.Now().UTC().Round(time.Second)
	assert.NoError(t, groupItemStore.Where("group_id=1 AND item_id=1").
		UpdateColumn("solutions_access_since", currentDate.AddDate(0, 0, -8)).Error())
	assert.NoError(t, groupItemStore.Where("group_id=1 AND item_id=2").
		UpdateColumn("solutions_access_since", currentDate.AddDate(0, 0, -9)).Error())
	assert.NoError(t, groupItemStore.Exec(`
		DELETE groups_items_propagate FROM groups_items_propagate JOIN groups_items USING (id)
			WHERE group_id=1 AND item_id=2`).Error())
	assert.NoError(t, groupItemStore.Where("group_id=1 AND item_id=3").
		UpdateColumn("cached_solutions_access_since", currentDate.AddDate(0, 0, -10)).Error())
	assert.NoError(t, groupItemStore.Exec(`
		DELETE groups_items_propagate FROM groups_items_propagate JOIN groups_items USING (id)
			WHERE group_id=1 AND item_id=3`).Error())
	assert.NoError(t, groupItemStore.Where("group_id=2 AND item_id=1").
		UpdateColumn("solutions_access_since", currentDate.AddDate(0, 0, -11)).Error())
	assert.NoError(t, groupItemStore.Where("group_id=2 AND item_id=11").
		UpdateColumn("solutions_access_since", currentDate.AddDate(0, 0, -12)).Error())
	assert.NoError(t, groupItemStore.UpdateColumn("cached_access_solutions", 1).Error())
	assert.NoError(t, groupItemStore.InTransaction(func(ds *database.DataStore) error {
		ds.GroupItems().ComputeAllAccess()
		return nil
	}))

	assertAllGroupItemsAreDone(t, groupItemStore)

	var result []groupItemsAggregatesResultRow
	assert.NoError(t, groupItemStore.Order("group_id, item_id").Scan(&result).Error())
	assertGroupsItemsAggregatesResultRowsEqual(t, []groupItemsAggregatesResultRow{
		{
			GroupID:                    1,
			ItemID:                     1,
			CachedSolutionsAccessSince: (*database.Time)(ptrTime(currentDate.AddDate(0, 0, -8))),
			SolutionsAccessSince:       (*database.Time)(ptrTime(currentDate.AddDate(0, 0, -8))),
			CachedAccessSolutions:      true,
		},
		{
			GroupID:                    1,
			ItemID:                     2,
			CachedSolutionsAccessSince: nil, // since it has been already done
			SolutionsAccessSince:       (*database.Time)(ptrTime(currentDate.AddDate(0, 0, -9))),
			CachedAccessSolutions:      true, // since it has been already done
		},
		{
			GroupID:                    1,
			ItemID:                     3,
			CachedSolutionsAccessSince: (*database.Time)(ptrTime(currentDate.AddDate(0, 0, -10))), // since it has been already done
			SolutionsAccessSince:       nil,
			CachedAccessSolutions:      true,
		},
		expectedRow14,
		{
			GroupID:                    1,
			ItemID:                     11,
			CachedSolutionsAccessSince: (*database.Time)(ptrTime(currentDate.AddDate(0, 0, -10))), // least(1_1, 1_3)

			CachedAccessSolutions: false, // since it is a new record and we do not set cached_access_solutions=1 in ComputeAllAccess()
		},
		{
			GroupID:                    1,
			ItemID:                     12,
			CachedSolutionsAccessSince: (*database.Time)(ptrTime(currentDate.AddDate(0, 0, -10))), // least(1_1, 1_2, 1_11)

			CachedAccessSolutions: false, // since it is a new record and we do not set cached_access_solutions=1 in ComputeAllAccess()
		},
		{
			GroupID:                    2,
			ItemID:                     1,
			CachedSolutionsAccessSince: (*database.Time)(ptrTime(currentDate.AddDate(0, 0, -11))),
			SolutionsAccessSince:       (*database.Time)(ptrTime(currentDate.AddDate(0, 0, -11))),
			CachedAccessSolutions:      true,
		},
		{
			GroupID:                    2,
			ItemID:                     11,
			CachedSolutionsAccessSince: (*database.Time)(ptrTime(currentDate.AddDate(0, 0, -12))),
			SolutionsAccessSince:       (*database.Time)(ptrTime(currentDate.AddDate(0, 0, -12))),
			CachedAccessSolutions:      true,
		},
		{
			GroupID:                    2,
			ItemID:                     12,
			CachedSolutionsAccessSince: (*database.Time)(ptrTime(currentDate.AddDate(0, 0, -12))),
			CachedAccessSolutions:      false, // since it is a new record and we do not set cached_access_solutions=1 in ComputeAllAccess()
		},
	}, result)
}

func TestGroupItemStore_ComputeAllAccess_AggregatesCachedAccessReason_IsNotSetWhenNothingIsChanged(t *testing.T) {
	db := testhelpers.SetupDBWithFixture("group_item_store/compute_all_access/_common")
	defer func() { _ = db.Close() }()

	groupItemStore := database.NewDataStore(db).GroupItems()
	assert.NoError(t, groupItemStore.Where("group_id=1 AND item_id=1").
		UpdateColumn("access_reason", "reason 1_1").Error())
	assert.NoError(t, groupItemStore.Where("group_id=1 AND item_id=2").
		UpdateColumn("access_reason", "reason 1_2").Error())
	assert.NoError(t, groupItemStore.Exec(`
		DELETE groups_items_propagate FROM groups_items_propagate JOIN groups_items USING (id)
			WHERE group_id=1 AND item_id=2`).Error())
	assert.NoError(t, groupItemStore.Where("group_id=1 AND item_id=3").
		UpdateColumn("cached_access_reason", "cached reason 1_3").Error())
	assert.NoError(t, groupItemStore.Exec(`
		DELETE groups_items_propagate FROM groups_items_propagate JOIN groups_items USING (id)
			WHERE group_id=1 AND item_id=3`).Error())
	assert.NoError(t, groupItemStore.Where("group_id=2 AND item_id=1").
		UpdateColumn("access_reason", "reason 2_1").Error())
	assert.NoError(t, groupItemStore.Where("group_id=2 AND item_id=11").
		UpdateColumn("access_reason", "reason 2_11").Error())
	assert.NoError(t, groupItemStore.InTransaction(func(ds *database.DataStore) error {
		ds.GroupItems().ComputeAllAccess()
		return nil
	}))

	assertAllGroupItemsAreDone(t, groupItemStore)

	var result []groupItemsAggregatesResultRow
	assert.NoError(t, groupItemStore.Order("group_id, item_id").Scan(&result).Error())
	assertGroupsItemsAggregatesResultRowsEqual(t, []groupItemsAggregatesResultRow{
		{
			GroupID:            1,
			ItemID:             1,
			CachedAccessReason: nil,
			AccessReason:       ptrString("reason 1_1"),
		},
		{
			GroupID:            1,
			ItemID:             2,
			CachedAccessReason: nil,
			AccessReason:       ptrString("reason 1_2"),
		},
		{
			GroupID:            1,
			ItemID:             3,
			CachedAccessReason: ptrString("cached reason 1_3"), // since it has been 'done'
			AccessReason:       nil,
		},
		expectedRow14,
		{
			GroupID:            1,
			ItemID:             11,
			CachedAccessReason: nil,
		},
		{
			GroupID:            1,
			ItemID:             12,
			CachedAccessReason: nil,
		},
		{
			GroupID:            2,
			ItemID:             1,
			CachedAccessReason: nil,
			AccessReason:       ptrString("reason 2_1"),
		},
		{
			GroupID:            2,
			ItemID:             11,
			CachedAccessReason: nil,
			AccessReason:       ptrString("reason 2_11"),
		},
		{
			GroupID:            2,
			ItemID:             12,
			CachedAccessReason: nil,
		},
	}, result)
}

func TestGroupItemStore_ComputeAllAccess_AggregatesCachedAccessReason_IsSetWhenAccessIsChanged(t *testing.T) {
	db := testhelpers.SetupDBWithFixture("group_item_store/compute_all_access/_common")
	defer func() { _ = db.Close() }()

	currentDate := time.Now().UTC().Round(time.Second)

	groupItemStore := database.NewDataStore(db).GroupItems()
	assert.NoError(t, groupItemStore.Where("group_id=1 AND item_id=1").
		UpdateColumn("access_reason", "reason 1_1").Error())
	assert.NoError(t, groupItemStore.Where("group_id=1 AND item_id=1").
		UpdateColumn("full_access_since", currentDate.AddDate(0, 0, -8)).Error())
	assert.NoError(t, groupItemStore.Where("group_id=1 AND item_id=2").
		UpdateColumn("access_reason", "reason 1_2").Error())
	assert.NoError(t, groupItemStore.Where("group_id=1 AND item_id=2").
		UpdateColumn("manager_access", 1).Error())
	assert.NoError(t, groupItemStore.Where("group_id=1 AND item_id=3").
		UpdateColumn("cached_access_reason", "cached reason 1_3").Error())
	assert.NoError(t, groupItemStore.Exec(`
		DELETE groups_items_propagate FROM groups_items_propagate JOIN groups_items USING (id)
			WHERE group_id=1 AND item_id=3`).Error())
	assert.NoError(t, groupItemStore.Where("group_id=2 AND item_id=1").
		UpdateColumn("access_reason", "reason 2_1").Error())
	assert.NoError(t, groupItemStore.Where("group_id=2 AND item_id=1").
		UpdateColumn("partial_access_since", currentDate.AddDate(0, 0, -11)).Error())
	assert.NoError(t, groupItemStore.Where("group_id=2 AND item_id=11").
		UpdateColumn("access_reason", "reason 2_11").Error())
	assert.NoError(t, groupItemStore.Where("group_id=2 AND item_id=11").
		UpdateColumn("solutions_access_since", currentDate.AddDate(0, 0, -12)).Error())
	assert.NoError(t, groupItemStore.InTransaction(func(ds *database.DataStore) error {
		ds.GroupItems().ComputeAllAccess()
		return nil
	}))

	assertAllGroupItemsAreDone(t, groupItemStore)

	var result []groupItemsAggregatesResultRow
	assert.NoError(t, groupItemStore.Order("group_id, item_id").Scan(&result).Error())
	assertGroupsItemsAggregatesResultRowsEqual(t, []groupItemsAggregatesResultRow{
		{
			GroupID:               1,
			ItemID:                1,
			CachedFullAccessSince: (*database.Time)(ptrTime(currentDate.AddDate(0, 0, -8))),
			FullAccessSince:       (*database.Time)(ptrTime(currentDate.AddDate(0, 0, -8))),
			CachedAccessReason:    nil,
			AccessReason:          ptrString("reason 1_1"),
		},
		{
			GroupID:             1,
			ItemID:              2,
			CachedManagerAccess: true,
			ManagerAccess:       true,
			CachedAccessReason:  nil,
			AccessReason:        ptrString("reason 1_2"),
		},
		{
			GroupID:            1,
			ItemID:             3,
			CachedAccessReason: ptrString("cached reason 1_3"), // since it has been 'done'
			AccessReason:       nil,
		},
		expectedRow14,
		{
			GroupID:               1,
			ItemID:                11,
			CachedFullAccessSince: (*database.Time)(ptrTime(currentDate.AddDate(0, 0, -8))),
			CachedManagerAccess:   true,
			CachedAccessReason:    ptrString("From ancestor group(s) reason 1_1, reason 1_2"), // concat(1_1, 1_2, 1_3)
		},
		{
			GroupID:               1,
			ItemID:                12,
			CachedFullAccessSince: (*database.Time)(ptrTime(currentDate.AddDate(0, 0, -8))),
			CachedManagerAccess:   true,
			CachedAccessReason:    ptrString("From ancestor group(s) reason 1_1, reason 1_2"), // concat(1_1, 1_2, 1_11)
		},
		{
			GroupID:                  2,
			ItemID:                   1,
			CachedPartialAccessSince: (*database.Time)(ptrTime(currentDate.AddDate(0, 0, -11))),
			PartialAccessSince:       (*database.Time)(ptrTime(currentDate.AddDate(0, 0, -11))),
			CachedAccessReason:       nil,
			AccessReason:             ptrString("reason 2_1"),
		},
		{
			GroupID:                    2,
			ItemID:                     11,
			CachedPartialAccessSince:   (*database.Time)(ptrTime(currentDate.AddDate(0, 0, -11))),
			CachedSolutionsAccessSince: (*database.Time)(ptrTime(currentDate.AddDate(0, 0, -12))),
			SolutionsAccessSince:       (*database.Time)(ptrTime(currentDate.AddDate(0, 0, -12))),
			CachedAccessReason:         ptrString("From ancestor group(s) reason 2_1"),
			AccessReason:               ptrString("reason 2_11"),
		},
		{
			GroupID:                    2,
			ItemID:                     12,
			CachedSolutionsAccessSince: (*database.Time)(ptrTime(currentDate.AddDate(0, 0, -12))),
			CachedAccessReason:         ptrString("From ancestor group(s) reason 2_11"),
		},
	}, result)
}

func ptrTime(t time.Time) *time.Time { return &t }
func ptrString(s string) *string     { return &s }

func assertGroupsItemsAggregatesResultRowsEqual(t *testing.T, expected, got []groupItemsAggregatesResultRow) {
	if len(got) != len(expected) {
		assert.ElementsMatch(t, expected, got)
		return
	}

	for i := 0; i < len(expected); i++ {
		assert.Equal(t, expected[i], got[i])
	}
}

func assertAllGroupItemsAreDone(t *testing.T, groupItemStore *database.GroupItemStore) {
	var cnt int
	assert.NoError(t, groupItemStore.Table("groups_items_propagate").Count(&cnt).Error())
	assert.Zero(t, cnt, "found not done groups_items")
}
