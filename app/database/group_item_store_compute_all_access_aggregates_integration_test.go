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
	GroupID                   int64
	ItemID                    int64
	PropagateAccess           string
	CachedFullAccessDate      *database.Time
	FullAccessDate            *database.Time
	CachedFullAccess          bool
	CachedPartialAccessDate   *database.Time
	PartialAccessDate         *database.Time
	CachedPartialAccess       bool
	ManagerAccess             bool
	CachedManagerAccess       bool
	CachedGrayedAccessDate    *database.Time
	CachedGrayedAccess        bool
	CachedAccessSolutionsDate *database.Time
	AccessSolutionsDate       *database.Time
	CachedAccessSolutions     bool
	CachedAccessReason        *string
	AccessReason              *string
}

var expectedRow14 = groupItemsAggregatesResultRow{
	GroupID:                   1,
	ItemID:                    4,
	CachedFullAccessDate:      (*database.Time)(ptrTime(time.Date(2017, 1, 1, 12, 13, 14, 0, time.UTC))),
	FullAccessDate:            (*database.Time)(ptrTime(time.Date(2017, 1, 1, 12, 13, 14, 0, time.UTC))),
	CachedFullAccess:          true,
	CachedPartialAccessDate:   (*database.Time)(ptrTime(time.Date(2017, 1, 1, 12, 13, 14, 0, time.UTC))),
	PartialAccessDate:         (*database.Time)(ptrTime(time.Date(2017, 1, 1, 12, 13, 14, 0, time.UTC))),
	CachedPartialAccess:       true,
	CachedManagerAccess:       true,
	ManagerAccess:             true,
	CachedGrayedAccessDate:    (*database.Time)(ptrTime(time.Date(2017, 1, 1, 12, 13, 14, 0, time.UTC))),
	CachedGrayedAccess:        true,
	CachedAccessSolutionsDate: (*database.Time)(ptrTime(time.Date(2017, 1, 1, 12, 13, 14, 0, time.UTC))),
	AccessSolutionsDate:       (*database.Time)(ptrTime(time.Date(2017, 1, 1, 12, 13, 14, 0, time.UTC))),
	CachedAccessSolutions:     true,
	CachedAccessReason:        ptrString("some cached reason"),
	AccessReason:              ptrString("some reason"),
	PropagateAccess:           "done",
}

func TestGroupItemStore_ComputeAllAccess_AggregatesCachedFullAccessDate(t *testing.T) {
	db := testhelpers.SetupDBWithFixture("group_item_store/compute_all_access/_common")
	defer func() { _ = db.Close() }()

	groupItemStore := database.NewDataStore(db).GroupItems()
	currentDate := time.Now().UTC().Round(time.Second)
	assert.NoError(t, groupItemStore.Where("group_id=1 AND item_id=1").
		UpdateColumn("full_access_date", currentDate.AddDate(0, 0, -8)).Error())
	assert.NoError(t, groupItemStore.Where("group_id=1 AND item_id=2").
		UpdateColumn("full_access_date", currentDate.AddDate(0, 0, -9)).Error())
	assert.NoError(t, groupItemStore.Where("group_id=1 AND item_id=2").
		UpdateColumn("propagate_access", "done").Error())
	assert.NoError(t, groupItemStore.Where("group_id=1 AND item_id=3").
		UpdateColumn("cached_full_access_date", currentDate.AddDate(0, 0, -10)).Error())
	assert.NoError(t, groupItemStore.Where("group_id=1 AND item_id=3").
		UpdateColumn("propagate_access", "done").Error())
	assert.NoError(t, groupItemStore.Where("group_id=2 AND item_id=1").
		UpdateColumn("full_access_date", currentDate.AddDate(0, 0, -11)).Error())
	assert.NoError(t, groupItemStore.Where("group_id=2 AND item_id=11").
		UpdateColumn("full_access_date", currentDate.AddDate(0, 0, -12)).Error())
	assert.NoError(t, groupItemStore.UpdateColumn("cached_full_access", 1).Error())
	assert.NoError(t, groupItemStore.InTransaction(func(ds *database.DataStore) error {
		ds.GroupItems().ComputeAllAccess()
		return nil
	}))

	var result []groupItemsAggregatesResultRow

	assert.NoError(t, groupItemStore.Order("group_id, item_id").Scan(&result).Error())
	assertGroupsItemsAggregatesResultRowsEqual(t, []groupItemsAggregatesResultRow{
		{
			GroupID:              1,
			ItemID:               1,
			CachedFullAccessDate: (*database.Time)(ptrTime(currentDate.AddDate(0, 0, -8))),
			FullAccessDate:       (*database.Time)(ptrTime(currentDate.AddDate(0, 0, -8))),
			CachedFullAccess:     true,
			PropagateAccess:      "done",
		},
		{
			GroupID:              1,
			ItemID:               2,
			CachedFullAccessDate: nil, // since it has been already done
			FullAccessDate:       (*database.Time)(ptrTime(currentDate.AddDate(0, 0, -9))),
			CachedFullAccess:     true, // since it has been already done
			PropagateAccess:      "done",
		},
		{
			GroupID:              1,
			ItemID:               3,
			CachedFullAccessDate: (*database.Time)(ptrTime(currentDate.AddDate(0, 0, -10))), // since it has been already done
			FullAccessDate:       nil,
			CachedFullAccess:     true,
			PropagateAccess:      "done",
		},
		expectedRow14,
		{
			GroupID:              1,
			ItemID:               11,
			CachedFullAccessDate: (*database.Time)(ptrTime(currentDate.AddDate(0, 0, -10))), // least(1_1, 1_3)

			CachedFullAccess: false, // since it is a new record and we do not set cached_full_access=1 in ComputeAllAccess()
			PropagateAccess:  "done",
		},
		{
			GroupID:              1,
			ItemID:               12,
			CachedFullAccessDate: (*database.Time)(ptrTime(currentDate.AddDate(0, 0, -10))), // least(1_1, 1_2, 1_11)
			PropagateAccess:      "done",
		},
		{
			GroupID:              2,
			ItemID:               1,
			CachedFullAccessDate: (*database.Time)(ptrTime(currentDate.AddDate(0, 0, -11))),
			FullAccessDate:       (*database.Time)(ptrTime(currentDate.AddDate(0, 0, -11))),
			CachedFullAccess:     true,
			PropagateAccess:      "done",
		},
		{
			GroupID:              2,
			ItemID:               11,
			CachedFullAccessDate: (*database.Time)(ptrTime(currentDate.AddDate(0, 0, -12))),
			FullAccessDate:       (*database.Time)(ptrTime(currentDate.AddDate(0, 0, -12))),
			CachedFullAccess:     true,
			PropagateAccess:      "done",
		},
		{
			GroupID:              2,
			ItemID:               12,
			CachedFullAccessDate: (*database.Time)(ptrTime(currentDate.AddDate(0, 0, -12))),
			CachedFullAccess:     false, // since it is a new record and we do not set cached_full_access=1 in ComputeAllAccess()
			PropagateAccess:      "done",
		},
	}, result)
}

func TestGroupItemStore_ComputeAllAccess_AggregatesCachedPartialAccessDate(t *testing.T) {
	db := testhelpers.SetupDBWithFixture("group_item_store/compute_all_access/_common")
	defer func() { _ = db.Close() }()

	groupItemStore := database.NewDataStore(db).GroupItems()
	currentDate := time.Now().UTC().Round(time.Second)
	assert.NoError(t, groupItemStore.Where("group_id=1 AND item_id=1").
		UpdateColumn("partial_access_date", currentDate.AddDate(0, 0, -8)).Error())
	assert.NoError(t, groupItemStore.Where("group_id=1 AND item_id=2").
		UpdateColumn("partial_access_date", currentDate.AddDate(0, 0, -9)).Error())
	assert.NoError(t, groupItemStore.Where("group_id=1 AND item_id=2").
		UpdateColumn("propagate_access", "done").Error())
	assert.NoError(t, groupItemStore.Where("group_id=1 AND item_id=3").
		UpdateColumn("cached_partial_access_date", currentDate.AddDate(0, 0, -10)).Error())
	assert.NoError(t, groupItemStore.Where("group_id=1 AND item_id=3").
		UpdateColumn("propagate_access", "done").Error())
	assert.NoError(t, groupItemStore.Where("group_id=2 AND item_id=1").
		UpdateColumn("partial_access_date", currentDate.AddDate(0, 0, -11)).Error())
	assert.NoError(t, groupItemStore.Where("group_id=2 AND item_id=11").
		UpdateColumn("partial_access_date", currentDate.AddDate(0, 0, -12)).Error())
	assert.NoError(t, groupItemStore.UpdateColumn("cached_partial_access", 1).Error())
	assert.NoError(t, groupItemStore.InTransaction(func(ds *database.DataStore) error {
		ds.GroupItems().ComputeAllAccess()
		return nil
	}))

	var result []groupItemsAggregatesResultRow

	assert.NoError(t, groupItemStore.Order("group_id, item_id").Scan(&result).Error())
	assertGroupsItemsAggregatesResultRowsEqual(t, []groupItemsAggregatesResultRow{
		{
			GroupID:                 1,
			ItemID:                  1,
			CachedPartialAccessDate: (*database.Time)(ptrTime(currentDate.AddDate(0, 0, -8))),
			PartialAccessDate:       (*database.Time)(ptrTime(currentDate.AddDate(0, 0, -8))),
			CachedPartialAccess:     true,
			PropagateAccess:         "done",
		},
		{
			GroupID:                 1,
			ItemID:                  2,
			CachedPartialAccessDate: nil, // since it has been already done
			PartialAccessDate:       (*database.Time)(ptrTime(currentDate.AddDate(0, 0, -9))),
			CachedPartialAccess:     true, // since it has been already done
			PropagateAccess:         "done",
		},
		{
			GroupID:                 1,
			ItemID:                  3,
			CachedPartialAccessDate: (*database.Time)(ptrTime(currentDate.AddDate(0, 0, -10))), // since it has been already done
			PartialAccessDate:       nil,
			CachedPartialAccess:     true,
			PropagateAccess:         "done",
		},
		expectedRow14,
		{
			GroupID:                 1,
			ItemID:                  11,
			CachedPartialAccessDate: (*database.Time)(ptrTime(currentDate.AddDate(0, 0, -10))), // least(1_1, 1_3)

			CachedPartialAccess: false, // since it is a new record and we do not set cached_partial_access=1 in ComputeAllAccess()
			PropagateAccess:     "done",
		},
		{
			GroupID:                 1,
			ItemID:                  12,
			CachedPartialAccessDate: nil, // since access_restricted=1
			PropagateAccess:         "done",
		},
		{
			GroupID:                 2,
			ItemID:                  1,
			CachedPartialAccessDate: (*database.Time)(ptrTime(currentDate.AddDate(0, 0, -11))),
			PartialAccessDate:       (*database.Time)(ptrTime(currentDate.AddDate(0, 0, -11))),
			CachedPartialAccess:     true,
			PropagateAccess:         "done",
		},
		{
			GroupID:                 2,
			ItemID:                  11,
			CachedPartialAccessDate: (*database.Time)(ptrTime(currentDate.AddDate(0, 0, -12))),
			PartialAccessDate:       (*database.Time)(ptrTime(currentDate.AddDate(0, 0, -12))),
			CachedPartialAccess:     true,
			PropagateAccess:         "done",
		},
		{
			GroupID:                 2,
			ItemID:                  12,
			CachedPartialAccessDate: nil, // since access_restricted=1
			PropagateAccess:         "done",
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
	assert.NoError(t, groupItemStore.Where("group_id=1 AND item_id=2").
		UpdateColumn("propagate_access", "done").Error())
	assert.NoError(t, groupItemStore.Where("group_id=1 AND item_id=3").
		UpdateColumn("cached_manager_access", 1).Error())
	assert.NoError(t, groupItemStore.Where("group_id=1 AND item_id=3").
		UpdateColumn("propagate_access", "done").Error())
	assert.NoError(t, groupItemStore.Where("group_id=2 AND item_id=1").
		UpdateColumn("manager_access", 1).Error())
	assert.NoError(t, groupItemStore.Where("group_id=2 AND item_id=11").
		UpdateColumn("manager_access", 1).Error())
	assert.NoError(t, groupItemStore.InTransaction(func(ds *database.DataStore) error {
		ds.GroupItems().ComputeAllAccess()
		return nil
	}))

	var result []groupItemsAggregatesResultRow

	assert.NoError(t, groupItemStore.Order("group_id, item_id").Scan(&result).Error())
	assertGroupsItemsAggregatesResultRowsEqual(t, []groupItemsAggregatesResultRow{
		{
			GroupID:             1,
			ItemID:              1,
			CachedManagerAccess: true,
			ManagerAccess:       true,
			PropagateAccess:     "done",
		},
		{
			GroupID:             1,
			ItemID:              2,
			CachedManagerAccess: false, // since it has been already done
			ManagerAccess:       true,
			PropagateAccess:     "done",
		},
		{
			GroupID:             1,
			ItemID:              3,
			CachedManagerAccess: true,
			ManagerAccess:       false,
			PropagateAccess:     "done",
		},
		expectedRow14,
		{
			GroupID:             1,
			ItemID:              11,
			CachedManagerAccess: true,
			PropagateAccess:     "done",
		},
		{
			GroupID:             1,
			ItemID:              12,
			CachedManagerAccess: true,
			PropagateAccess:     "done",
		},
		{
			GroupID:             2,
			ItemID:              1,
			CachedManagerAccess: true,
			ManagerAccess:       true,
			PropagateAccess:     "done",
		},
		{
			GroupID:             2,
			ItemID:              11,
			CachedManagerAccess: true,
			ManagerAccess:       true,
			PropagateAccess:     "done",
		},
		{
			GroupID:             2,
			ItemID:              12,
			CachedManagerAccess: true,
			PropagateAccess:     "done",
		},
	}, result)
}

func TestGroupItemStore_ComputeAllAccess_AggregatesCachedGrayedAccessDate(t *testing.T) {
	db := testhelpers.SetupDBWithFixture("group_item_store/compute_all_access/_common")
	defer func() { _ = db.Close() }()

	groupItemStore := database.NewDataStore(db).GroupItems()
	currentDate := time.Now().UTC().Round(time.Second)
	assert.NoError(t, groupItemStore.Where("group_id=1 AND item_id=1").
		UpdateColumn("partial_access_date", currentDate.AddDate(0, 0, -8)).Error())
	assert.NoError(t, groupItemStore.Where("group_id=1 AND item_id=2").
		UpdateColumn("partial_access_date", currentDate.AddDate(0, 0, -9)).Error())
	assert.NoError(t, groupItemStore.Where("group_id=1 AND item_id=2").
		UpdateColumn("propagate_access", "done").Error())
	assert.NoError(t, groupItemStore.Where("group_id=1 AND item_id=3").
		UpdateColumn("cached_partial_access_date", currentDate.AddDate(0, 0, -10)).Error())
	assert.NoError(t, groupItemStore.Where("group_id=1 AND item_id=3").
		UpdateColumn("propagate_access", "done").Error())
	assert.NoError(t, groupItemStore.Where("group_id=2 AND item_id=1").
		UpdateColumn("partial_access_date", currentDate.AddDate(0, 0, -11)).Error())
	assert.NoError(t, groupItemStore.Where("group_id=2 AND item_id=11").
		UpdateColumn("partial_access_date", currentDate.AddDate(0, 0, -12)).Error())
	assert.NoError(t, groupItemStore.UpdateColumn("cached_grayed_access", 1).Error())
	assert.NoError(t, groupItemStore.ItemItems().Updates(map[string]interface{}{
		"always_visible":    true,
		"access_restricted": true,
	}).Error())
	assert.NoError(t, groupItemStore.InTransaction(func(ds *database.DataStore) error {
		ds.GroupItems().ComputeAllAccess()
		return nil
	}))

	var result []groupItemsAggregatesResultRow

	assert.NoError(t, groupItemStore.Order("group_id, item_id").Scan(&result).Error())
	assertGroupsItemsAggregatesResultRowsEqual(t, []groupItemsAggregatesResultRow{
		{
			GroupID:                 1,
			ItemID:                  1,
			CachedPartialAccessDate: (*database.Time)(ptrTime(currentDate.AddDate(0, 0, -8))),
			PartialAccessDate:       (*database.Time)(ptrTime(currentDate.AddDate(0, 0, -8))),
			CachedGrayedAccess:      false, // since we have partial access here
			PropagateAccess:         "done",
		},
		{
			GroupID:                 1,
			ItemID:                  2,
			CachedPartialAccessDate: nil, // since it has been already done
			PartialAccessDate:       (*database.Time)(ptrTime(currentDate.AddDate(0, 0, -9))),
			CachedGrayedAccess:      true, // since it has been already done
			PropagateAccess:         "done",
		},
		{
			GroupID:                 1,
			ItemID:                  3,
			CachedPartialAccessDate: (*database.Time)(ptrTime(currentDate.AddDate(0, 0, -10))), // since it has been already done
			PartialAccessDate:       nil,
			CachedGrayedAccess:      true,
			PropagateAccess:         "done",
		},
		expectedRow14,
		{
			GroupID:                 1,
			ItemID:                  11,
			CachedPartialAccessDate: nil, // since access_restricted = 1

			CachedGrayedAccessDate: (*database.Time)(ptrTime(currentDate.AddDate(0, 0, -10))), // least cached_partial_access_date(1_1, 1_3)

			CachedGrayedAccess: false, // since it is a new record and we do not set cached_grayed_access=1 in ComputeAllAccess()
			PropagateAccess:    "done",
		},
		{
			GroupID:                 1,
			ItemID:                  12,
			CachedPartialAccessDate: nil, // since access_restricted=1

			CachedGrayedAccessDate: (*database.Time)(ptrTime(currentDate.AddDate(0, 0, -8))), // from 1_1

			CachedGrayedAccess: false, // since it is a new record and we do not set cached_grayed_access=1 in ComputeAllAccess()
			PropagateAccess:    "done",
		},
		{
			GroupID:                 2,
			ItemID:                  1,
			CachedPartialAccessDate: (*database.Time)(ptrTime(currentDate.AddDate(0, 0, -11))),
			PartialAccessDate:       (*database.Time)(ptrTime(currentDate.AddDate(0, 0, -11))),
			PropagateAccess:         "done",
		},
		{
			GroupID:                 2,
			ItemID:                  11,
			CachedPartialAccessDate: (*database.Time)(ptrTime(currentDate.AddDate(0, 0, -12))),
			PartialAccessDate:       (*database.Time)(ptrTime(currentDate.AddDate(0, 0, -12))),
			CachedGrayedAccessDate:  (*database.Time)(ptrTime(currentDate.AddDate(0, 0, -11))), // from 2_1

			CachedGrayedAccess: false, // since it is a new record and we do not set cached_grayed_access=1 in ComputeAllAccess()
			PropagateAccess:    "done",
		},
		{
			GroupID: 2,
			ItemID:  12,

			CachedPartialAccessDate: nil, // since access_restricted=1

			CachedGrayedAccessDate: (*database.Time)(ptrTime(currentDate.AddDate(0, 0, -12))), // from 2_11

			CachedGrayedAccess: false, // since it is a new record and we do not set cached_grayed_access=1 in ComputeAllAccess()
			PropagateAccess:    "done",
		},
	}, result)
}

func TestGroupItemStore_ComputeAllAccess_AggregatesCachedAccessSolutionsDate(t *testing.T) {
	db := testhelpers.SetupDBWithFixture("group_item_store/compute_all_access/_common")
	defer func() { _ = db.Close() }()

	groupItemStore := database.NewDataStore(db).GroupItems()
	currentDate := time.Now().UTC().Round(time.Second)
	assert.NoError(t, groupItemStore.Where("group_id=1 AND item_id=1").
		UpdateColumn("access_solutions_date", currentDate.AddDate(0, 0, -8)).Error())
	assert.NoError(t, groupItemStore.Where("group_id=1 AND item_id=2").
		UpdateColumn("access_solutions_date", currentDate.AddDate(0, 0, -9)).Error())
	assert.NoError(t, groupItemStore.Where("group_id=1 AND item_id=2").
		UpdateColumn("propagate_access", "done").Error())
	assert.NoError(t, groupItemStore.Where("group_id=1 AND item_id=3").
		UpdateColumn("cached_access_solutions_date", currentDate.AddDate(0, 0, -10)).Error())
	assert.NoError(t, groupItemStore.Where("group_id=1 AND item_id=3").
		UpdateColumn("propagate_access", "done").Error())
	assert.NoError(t, groupItemStore.Where("group_id=2 AND item_id=1").
		UpdateColumn("access_solutions_date", currentDate.AddDate(0, 0, -11)).Error())
	assert.NoError(t, groupItemStore.Where("group_id=2 AND item_id=11").
		UpdateColumn("access_solutions_date", currentDate.AddDate(0, 0, -12)).Error())
	assert.NoError(t, groupItemStore.UpdateColumn("cached_access_solutions", 1).Error())
	assert.NoError(t, groupItemStore.InTransaction(func(ds *database.DataStore) error {
		ds.GroupItems().ComputeAllAccess()
		return nil
	}))

	var result []groupItemsAggregatesResultRow

	assert.NoError(t, groupItemStore.Order("group_id, item_id").Scan(&result).Error())
	assertGroupsItemsAggregatesResultRowsEqual(t, []groupItemsAggregatesResultRow{
		{
			GroupID:                   1,
			ItemID:                    1,
			CachedAccessSolutionsDate: (*database.Time)(ptrTime(currentDate.AddDate(0, 0, -8))),
			AccessSolutionsDate:       (*database.Time)(ptrTime(currentDate.AddDate(0, 0, -8))),
			CachedAccessSolutions:     true,
			PropagateAccess:           "done",
		},
		{
			GroupID:                   1,
			ItemID:                    2,
			CachedAccessSolutionsDate: nil, // since it has been already done
			AccessSolutionsDate:       (*database.Time)(ptrTime(currentDate.AddDate(0, 0, -9))),
			CachedAccessSolutions:     true, // since it has been already done
			PropagateAccess:           "done",
		},
		{
			GroupID:                   1,
			ItemID:                    3,
			CachedAccessSolutionsDate: (*database.Time)(ptrTime(currentDate.AddDate(0, 0, -10))), // since it has been already done
			AccessSolutionsDate:       nil,
			CachedAccessSolutions:     true,
			PropagateAccess:           "done",
		},
		expectedRow14,
		{
			GroupID:                   1,
			ItemID:                    11,
			CachedAccessSolutionsDate: (*database.Time)(ptrTime(currentDate.AddDate(0, 0, -10))), // least(1_1, 1_3)

			CachedAccessSolutions: false, // since it is a new record and we do not set cached_access_solutions=1 in ComputeAllAccess()
			PropagateAccess:       "done",
		},
		{
			GroupID:                   1,
			ItemID:                    12,
			CachedAccessSolutionsDate: (*database.Time)(ptrTime(currentDate.AddDate(0, 0, -10))), // least(1_1, 1_2, 1_11)

			CachedAccessSolutions: false, // since it is a new record and we do not set cached_access_solutions=1 in ComputeAllAccess()
			PropagateAccess:       "done",
		},
		{
			GroupID:                   2,
			ItemID:                    1,
			CachedAccessSolutionsDate: (*database.Time)(ptrTime(currentDate.AddDate(0, 0, -11))),
			AccessSolutionsDate:       (*database.Time)(ptrTime(currentDate.AddDate(0, 0, -11))),
			CachedAccessSolutions:     true,
			PropagateAccess:           "done",
		},
		{
			GroupID:                   2,
			ItemID:                    11,
			CachedAccessSolutionsDate: (*database.Time)(ptrTime(currentDate.AddDate(0, 0, -12))),
			AccessSolutionsDate:       (*database.Time)(ptrTime(currentDate.AddDate(0, 0, -12))),
			CachedAccessSolutions:     true,
			PropagateAccess:           "done",
		},
		{
			GroupID:                   2,
			ItemID:                    12,
			CachedAccessSolutionsDate: (*database.Time)(ptrTime(currentDate.AddDate(0, 0, -12))),
			CachedAccessSolutions:     false, // since it is a new record and we do not set cached_access_solutions=1 in ComputeAllAccess()
			PropagateAccess:           "done",
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
	assert.NoError(t, groupItemStore.Where("group_id=1 AND item_id=2").
		UpdateColumn("propagate_access", "done").Error())
	assert.NoError(t, groupItemStore.Where("group_id=1 AND item_id=3").
		UpdateColumn("cached_access_reason", "cached reason 1_3").Error())
	assert.NoError(t, groupItemStore.Where("group_id=1 AND item_id=3").
		UpdateColumn("propagate_access", "done").Error())
	assert.NoError(t, groupItemStore.Where("group_id=2 AND item_id=1").
		UpdateColumn("access_reason", "reason 2_1").Error())
	assert.NoError(t, groupItemStore.Where("group_id=2 AND item_id=11").
		UpdateColumn("access_reason", "reason 2_11").Error())
	assert.NoError(t, groupItemStore.InTransaction(func(ds *database.DataStore) error {
		ds.GroupItems().ComputeAllAccess()
		return nil
	}))

	var result []groupItemsAggregatesResultRow

	assert.NoError(t, groupItemStore.Order("group_id, item_id").Scan(&result).Error())
	assertGroupsItemsAggregatesResultRowsEqual(t, []groupItemsAggregatesResultRow{
		{
			GroupID:            1,
			ItemID:             1,
			CachedAccessReason: nil,
			AccessReason:       ptrString("reason 1_1"),
			PropagateAccess:    "done",
		},
		{
			GroupID:            1,
			ItemID:             2,
			CachedAccessReason: nil,
			AccessReason:       ptrString("reason 1_2"),
			PropagateAccess:    "done",
		},
		{
			GroupID:            1,
			ItemID:             3,
			CachedAccessReason: ptrString("cached reason 1_3"), // since it has been 'done'
			AccessReason:       nil,
			PropagateAccess:    "done",
		},
		expectedRow14,
		{
			GroupID:            1,
			ItemID:             11,
			CachedAccessReason: nil,
			PropagateAccess:    "done",
		},
		{
			GroupID:            1,
			ItemID:             12,
			CachedAccessReason: nil,
			PropagateAccess:    "done",
		},
		{
			GroupID:            2,
			ItemID:             1,
			CachedAccessReason: nil,
			AccessReason:       ptrString("reason 2_1"),
			PropagateAccess:    "done",
		},
		{
			GroupID:            2,
			ItemID:             11,
			CachedAccessReason: nil,
			AccessReason:       ptrString("reason 2_11"),
			PropagateAccess:    "done",
		},
		{
			GroupID:            2,
			ItemID:             12,
			CachedAccessReason: nil,
			PropagateAccess:    "done",
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
		UpdateColumn("full_access_date", currentDate.AddDate(0, 0, -8)).Error())
	assert.NoError(t, groupItemStore.Where("group_id=1 AND item_id=2").
		UpdateColumn("access_reason", "reason 1_2").Error())
	assert.NoError(t, groupItemStore.Where("group_id=1 AND item_id=2").
		UpdateColumn("manager_access", 1).Error())
	assert.NoError(t, groupItemStore.Where("group_id=1 AND item_id=3").
		UpdateColumn("cached_access_reason", "cached reason 1_3").Error())
	assert.NoError(t, groupItemStore.Where("group_id=1 AND item_id=3").
		UpdateColumn("propagate_access", "done").Error())
	assert.NoError(t, groupItemStore.Where("group_id=2 AND item_id=1").
		UpdateColumn("access_reason", "reason 2_1").Error())
	assert.NoError(t, groupItemStore.Where("group_id=2 AND item_id=1").
		UpdateColumn("partial_access_date", currentDate.AddDate(0, 0, -11)).Error())
	assert.NoError(t, groupItemStore.Where("group_id=2 AND item_id=11").
		UpdateColumn("access_reason", "reason 2_11").Error())
	assert.NoError(t, groupItemStore.Where("group_id=2 AND item_id=11").
		UpdateColumn("access_solutions_date", currentDate.AddDate(0, 0, -12)).Error())
	assert.NoError(t, groupItemStore.InTransaction(func(ds *database.DataStore) error {
		ds.GroupItems().ComputeAllAccess()
		return nil
	}))

	var result []groupItemsAggregatesResultRow

	assert.NoError(t, groupItemStore.Order("group_id, item_id").Scan(&result).Error())
	assertGroupsItemsAggregatesResultRowsEqual(t, []groupItemsAggregatesResultRow{
		{
			GroupID:              1,
			ItemID:               1,
			CachedFullAccessDate: (*database.Time)(ptrTime(currentDate.AddDate(0, 0, -8))),
			FullAccessDate:       (*database.Time)(ptrTime(currentDate.AddDate(0, 0, -8))),
			CachedAccessReason:   nil,
			AccessReason:         ptrString("reason 1_1"),
			PropagateAccess:      "done",
		},
		{
			GroupID:             1,
			ItemID:              2,
			CachedManagerAccess: true,
			ManagerAccess:       true,
			CachedAccessReason:  nil,
			AccessReason:        ptrString("reason 1_2"),
			PropagateAccess:     "done",
		},
		{
			GroupID:            1,
			ItemID:             3,
			CachedAccessReason: ptrString("cached reason 1_3"), // since it has been 'done'
			AccessReason:       nil,
			PropagateAccess:    "done",
		},
		expectedRow14,
		{
			GroupID:              1,
			ItemID:               11,
			CachedFullAccessDate: (*database.Time)(ptrTime(currentDate.AddDate(0, 0, -8))),
			CachedManagerAccess:  true,
			CachedAccessReason:   ptrString("From ancestor group(s) reason 1_1, reason 1_2"), // concat(1_1, 1_2, 1_3)
			PropagateAccess:      "done",
		},
		{
			GroupID:              1,
			ItemID:               12,
			CachedFullAccessDate: (*database.Time)(ptrTime(currentDate.AddDate(0, 0, -8))),
			CachedManagerAccess:  true,
			CachedAccessReason:   ptrString("From ancestor group(s) reason 1_1, reason 1_2"), // concat(1_1, 1_2, 1_11)
			PropagateAccess:      "done",
		},
		{
			GroupID:                 2,
			ItemID:                  1,
			CachedPartialAccessDate: (*database.Time)(ptrTime(currentDate.AddDate(0, 0, -11))),
			PartialAccessDate:       (*database.Time)(ptrTime(currentDate.AddDate(0, 0, -11))),
			CachedAccessReason:      nil,
			AccessReason:            ptrString("reason 2_1"),
			PropagateAccess:         "done",
		},
		{
			GroupID:                   2,
			ItemID:                    11,
			CachedPartialAccessDate:   (*database.Time)(ptrTime(currentDate.AddDate(0, 0, -11))),
			CachedAccessSolutionsDate: (*database.Time)(ptrTime(currentDate.AddDate(0, 0, -12))),
			AccessSolutionsDate:       (*database.Time)(ptrTime(currentDate.AddDate(0, 0, -12))),
			CachedAccessReason:        ptrString("From ancestor group(s) reason 2_1"),
			AccessReason:              ptrString("reason 2_11"),
			PropagateAccess:           "done",
		},
		{
			GroupID:                   2,
			ItemID:                    12,
			CachedAccessSolutionsDate: (*database.Time)(ptrTime(currentDate.AddDate(0, 0, -12))),
			CachedAccessReason:        ptrString("From ancestor group(s) reason 2_11"),
			PropagateAccess:           "done",
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
