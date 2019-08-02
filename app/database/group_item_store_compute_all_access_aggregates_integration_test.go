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
	GroupID                   int64      `gorm:"column:idGroup"`
	ItemID                    int64      `gorm:"column:idItem"`
	PropagateAccess           string     `gorm:"column:sPropagateAccess"`
	CachedFullAccessDate      *time.Time `gorm:"column:sCachedFullAccessDate"`
	FullAccessDate            *time.Time `gorm:"column:sFullAccessDate"`
	CachedFullAccess          bool       `gorm:"column:bCachedFullAccess"`
	CachedPartialAccessDate   *time.Time `gorm:"column:sCachedPartialAccessDate"`
	PartialAccessDate         *time.Time `gorm:"column:sPartialAccessDate"`
	CachedPartialAccess       bool       `gorm:"column:bCachedPartialAccess"`
	ManagerAccess             bool       `gorm:"column:bManagerAccess"`
	CachedManagerAccess       bool       `gorm:"column:bCachedManagerAccess"`
	CachedGrayedAccessDate    *time.Time `gorm:"column:sCachedGrayedAccessDate"`
	CachedGrayedAccess        bool       `gorm:"column:bCachedGrayedAccess"`
	CachedAccessSolutionsDate *time.Time `gorm:"column:sCachedAccessSolutionsDate"`
	AccessSolutionsDate       *time.Time `gorm:"column:sAccessSolutionsDate"`
	CachedAccessSolutions     bool       `gorm:"column:bCachedAccessSolutions"`
	CachedAccessReason        *string    `gorm:"column:sCachedAccessReason"`
	AccessReason              *string    `gorm:"column:sAccessReason"`
}

var expectedRow14 = groupItemsAggregatesResultRow{
	GroupID:                   1,
	ItemID:                    4,
	CachedFullAccessDate:      ptrTime(time.Date(2017, 1, 1, 12, 13, 14, 0, time.UTC)),
	FullAccessDate:            ptrTime(time.Date(2017, 1, 1, 12, 13, 14, 0, time.UTC)),
	CachedFullAccess:          true,
	CachedPartialAccessDate:   ptrTime(time.Date(2017, 1, 1, 12, 13, 14, 0, time.UTC)),
	PartialAccessDate:         ptrTime(time.Date(2017, 1, 1, 12, 13, 14, 0, time.UTC)),
	CachedPartialAccess:       true,
	CachedManagerAccess:       true,
	ManagerAccess:             true,
	CachedGrayedAccessDate:    ptrTime(time.Date(2017, 1, 1, 12, 13, 14, 0, time.UTC)),
	CachedGrayedAccess:        true,
	CachedAccessSolutionsDate: ptrTime(time.Date(2017, 1, 1, 12, 13, 14, 0, time.UTC)),
	AccessSolutionsDate:       ptrTime(time.Date(2017, 1, 1, 12, 13, 14, 0, time.UTC)),
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
	assert.NoError(t, groupItemStore.Where("idGroup=1 AND idItem=1").
		UpdateColumn("sFullAccessDate", currentDate.AddDate(0, 0, -8)).Error())
	assert.NoError(t, groupItemStore.Where("idGroup=1 AND idItem=2").
		UpdateColumn("sFullAccessDate", currentDate.AddDate(0, 0, -9)).Error())
	assert.NoError(t, groupItemStore.Where("idGroup=1 AND idItem=2").
		UpdateColumn("sPropagateAccess", "done").Error())
	assert.NoError(t, groupItemStore.Where("idGroup=1 AND idItem=3").
		UpdateColumn("sCachedFullAccessDate", currentDate.AddDate(0, 0, -10)).Error())
	assert.NoError(t, groupItemStore.Where("idGroup=1 AND idItem=3").
		UpdateColumn("sPropagateAccess", "done").Error())
	assert.NoError(t, groupItemStore.Where("idGroup=2 AND idItem=1").
		UpdateColumn("sFullAccessDate", currentDate.AddDate(0, 0, -11)).Error())
	assert.NoError(t, groupItemStore.Where("idGroup=2 AND idItem=11").
		UpdateColumn("sFullAccessDate", currentDate.AddDate(0, 0, -12)).Error())
	assert.NoError(t, groupItemStore.UpdateColumn("bCachedFullAccess", 1).Error())
	assert.NoError(t, groupItemStore.InTransaction(func(ds *database.DataStore) error {
		ds.GroupItems().ComputeAllAccess()
		return nil
	}))

	var result []groupItemsAggregatesResultRow

	assert.NoError(t, groupItemStore.Order("idGroup, idItem").Scan(&result).Error())
	assertGroupsItemsAggregatesResultRowsEqual(t, []groupItemsAggregatesResultRow{
		{
			GroupID:              1,
			ItemID:               1,
			CachedFullAccessDate: ptrTime(currentDate.AddDate(0, 0, -8)),
			FullAccessDate:       ptrTime(currentDate.AddDate(0, 0, -8)),
			CachedFullAccess:     true,
			PropagateAccess:      "done",
		},
		{
			GroupID:              1,
			ItemID:               2,
			CachedFullAccessDate: nil, // since it has been already done
			FullAccessDate:       ptrTime(currentDate.AddDate(0, 0, -9)),
			CachedFullAccess:     true, // since it has been already done
			PropagateAccess:      "done",
		},
		{
			GroupID:              1,
			ItemID:               3,
			CachedFullAccessDate: ptrTime(currentDate.AddDate(0, 0, -10)), // since it has been already done
			FullAccessDate:       nil,
			CachedFullAccess:     true,
			PropagateAccess:      "done",
		},
		expectedRow14,
		{
			GroupID:              1,
			ItemID:               11,
			CachedFullAccessDate: ptrTime(currentDate.AddDate(0, 0, -10)), // least(1_1, 1_3)

			CachedFullAccess: false, // since it is a new record and we do not set bCachedFullAccess=1 in ComputeAllAccess()
			PropagateAccess:  "done",
		},
		{
			GroupID:              1,
			ItemID:               12,
			CachedFullAccessDate: ptrTime(currentDate.AddDate(0, 0, -10)), // least(1_1, 1_2, 1_11)
			PropagateAccess:      "done",
		},
		{
			GroupID:              2,
			ItemID:               1,
			CachedFullAccessDate: ptrTime(currentDate.AddDate(0, 0, -11)),
			FullAccessDate:       ptrTime(currentDate.AddDate(0, 0, -11)),
			CachedFullAccess:     true,
			PropagateAccess:      "done",
		},
		{
			GroupID:              2,
			ItemID:               11,
			CachedFullAccessDate: ptrTime(currentDate.AddDate(0, 0, -12)),
			FullAccessDate:       ptrTime(currentDate.AddDate(0, 0, -12)),
			CachedFullAccess:     true,
			PropagateAccess:      "done",
		},
		{
			GroupID:              2,
			ItemID:               12,
			CachedFullAccessDate: ptrTime(currentDate.AddDate(0, 0, -12)),
			CachedFullAccess:     false, // since it is a new record and we do not set bCachedFullAccess=1 in ComputeAllAccess()
			PropagateAccess:      "done",
		},
	}, result)
}

func TestGroupItemStore_ComputeAllAccess_AggregatesCachedPartialAccessDate(t *testing.T) {
	db := testhelpers.SetupDBWithFixture("group_item_store/compute_all_access/_common")
	defer func() { _ = db.Close() }()

	groupItemStore := database.NewDataStore(db).GroupItems()
	currentDate := time.Now().UTC().Round(time.Second)
	assert.NoError(t, groupItemStore.Where("idGroup=1 AND idItem=1").
		UpdateColumn("sPartialAccessDate", currentDate.AddDate(0, 0, -8)).Error())
	assert.NoError(t, groupItemStore.Where("idGroup=1 AND idItem=2").
		UpdateColumn("sPartialAccessDate", currentDate.AddDate(0, 0, -9)).Error())
	assert.NoError(t, groupItemStore.Where("idGroup=1 AND idItem=2").
		UpdateColumn("sPropagateAccess", "done").Error())
	assert.NoError(t, groupItemStore.Where("idGroup=1 AND idItem=3").
		UpdateColumn("sCachedPartialAccessDate", currentDate.AddDate(0, 0, -10)).Error())
	assert.NoError(t, groupItemStore.Where("idGroup=1 AND idItem=3").
		UpdateColumn("sPropagateAccess", "done").Error())
	assert.NoError(t, groupItemStore.Where("idGroup=2 AND idItem=1").
		UpdateColumn("sPartialAccessDate", currentDate.AddDate(0, 0, -11)).Error())
	assert.NoError(t, groupItemStore.Where("idGroup=2 AND idItem=11").
		UpdateColumn("sPartialAccessDate", currentDate.AddDate(0, 0, -12)).Error())
	assert.NoError(t, groupItemStore.UpdateColumn("bCachedPartialAccess", 1).Error())
	assert.NoError(t, groupItemStore.InTransaction(func(ds *database.DataStore) error {
		ds.GroupItems().ComputeAllAccess()
		return nil
	}))

	var result []groupItemsAggregatesResultRow

	assert.NoError(t, groupItemStore.Order("idGroup, idItem").Scan(&result).Error())
	assertGroupsItemsAggregatesResultRowsEqual(t, []groupItemsAggregatesResultRow{
		{
			GroupID:                 1,
			ItemID:                  1,
			CachedPartialAccessDate: ptrTime(currentDate.AddDate(0, 0, -8)),
			PartialAccessDate:       ptrTime(currentDate.AddDate(0, 0, -8)),
			CachedPartialAccess:     true,
			PropagateAccess:         "done",
		},
		{
			GroupID:                 1,
			ItemID:                  2,
			CachedPartialAccessDate: nil, // since it has been already done
			PartialAccessDate:       ptrTime(currentDate.AddDate(0, 0, -9)),
			CachedPartialAccess:     true, // since it has been already done
			PropagateAccess:         "done",
		},
		{
			GroupID:                 1,
			ItemID:                  3,
			CachedPartialAccessDate: ptrTime(currentDate.AddDate(0, 0, -10)), // since it has been already done
			PartialAccessDate:       nil,
			CachedPartialAccess:     true,
			PropagateAccess:         "done",
		},
		expectedRow14,
		{
			GroupID:                 1,
			ItemID:                  11,
			CachedPartialAccessDate: ptrTime(currentDate.AddDate(0, 0, -10)), // least(1_1, 1_3)

			CachedPartialAccess: false, // since it is a new record and we do not set bCachedPartialAccess=1 in ComputeAllAccess()
			PropagateAccess:     "done",
		},
		{
			GroupID:                 1,
			ItemID:                  12,
			CachedPartialAccessDate: nil, // since bAccessRestricted=1
			PropagateAccess:         "done",
		},
		{
			GroupID:                 2,
			ItemID:                  1,
			CachedPartialAccessDate: ptrTime(currentDate.AddDate(0, 0, -11)),
			PartialAccessDate:       ptrTime(currentDate.AddDate(0, 0, -11)),
			CachedPartialAccess:     true,
			PropagateAccess:         "done",
		},
		{
			GroupID:                 2,
			ItemID:                  11,
			CachedPartialAccessDate: ptrTime(currentDate.AddDate(0, 0, -12)),
			PartialAccessDate:       ptrTime(currentDate.AddDate(0, 0, -12)),
			CachedPartialAccess:     true,
			PropagateAccess:         "done",
		},
		{
			GroupID:                 2,
			ItemID:                  12,
			CachedPartialAccessDate: nil, // since bAccessRestricted=1
			PropagateAccess:         "done",
		},
	}, result)
}

func TestGroupItemStore_ComputeAllAccess_AggregatesCachedManagerAccess(t *testing.T) {
	db := testhelpers.SetupDBWithFixture("group_item_store/compute_all_access/_common")
	defer func() { _ = db.Close() }()

	groupItemStore := database.NewDataStore(db).GroupItems()
	assert.NoError(t, groupItemStore.Where("idGroup=1 AND idItem=1").
		UpdateColumn("bManagerAccess", 1).Error())
	assert.NoError(t, groupItemStore.Where("idGroup=1 AND idItem=2").
		UpdateColumn("bManagerAccess", 1).Error())
	assert.NoError(t, groupItemStore.Where("idGroup=1 AND idItem=2").
		UpdateColumn("sPropagateAccess", "done").Error())
	assert.NoError(t, groupItemStore.Where("idGroup=1 AND idItem=3").
		UpdateColumn("bCachedManagerAccess", 1).Error())
	assert.NoError(t, groupItemStore.Where("idGroup=1 AND idItem=3").
		UpdateColumn("sPropagateAccess", "done").Error())
	assert.NoError(t, groupItemStore.Where("idGroup=2 AND idItem=1").
		UpdateColumn("bManagerAccess", 1).Error())
	assert.NoError(t, groupItemStore.Where("idGroup=2 AND idItem=11").
		UpdateColumn("bManagerAccess", 1).Error())
	assert.NoError(t, groupItemStore.InTransaction(func(ds *database.DataStore) error {
		ds.GroupItems().ComputeAllAccess()
		return nil
	}))

	var result []groupItemsAggregatesResultRow

	assert.NoError(t, groupItemStore.Order("idGroup, idItem").Scan(&result).Error())
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
	assert.NoError(t, groupItemStore.Where("idGroup=1 AND idItem=1").
		UpdateColumn("sPartialAccessDate", currentDate.AddDate(0, 0, -8)).Error())
	assert.NoError(t, groupItemStore.Where("idGroup=1 AND idItem=2").
		UpdateColumn("sPartialAccessDate", currentDate.AddDate(0, 0, -9)).Error())
	assert.NoError(t, groupItemStore.Where("idGroup=1 AND idItem=2").
		UpdateColumn("sPropagateAccess", "done").Error())
	assert.NoError(t, groupItemStore.Where("idGroup=1 AND idItem=3").
		UpdateColumn("sCachedPartialAccessDate", currentDate.AddDate(0, 0, -10)).Error())
	assert.NoError(t, groupItemStore.Where("idGroup=1 AND idItem=3").
		UpdateColumn("sPropagateAccess", "done").Error())
	assert.NoError(t, groupItemStore.Where("idGroup=2 AND idItem=1").
		UpdateColumn("sPartialAccessDate", currentDate.AddDate(0, 0, -11)).Error())
	assert.NoError(t, groupItemStore.Where("idGroup=2 AND idItem=11").
		UpdateColumn("sPartialAccessDate", currentDate.AddDate(0, 0, -12)).Error())
	assert.NoError(t, groupItemStore.UpdateColumn("bCachedGrayedAccess", 1).Error())
	assert.NoError(t, groupItemStore.ItemItems().Updates(map[string]interface{}{
		"bAlwaysVisible":    true,
		"bAccessRestricted": true,
	}).Error())
	assert.NoError(t, groupItemStore.InTransaction(func(ds *database.DataStore) error {
		ds.GroupItems().ComputeAllAccess()
		return nil
	}))

	var result []groupItemsAggregatesResultRow

	assert.NoError(t, groupItemStore.Order("idGroup, idItem").Scan(&result).Error())
	assertGroupsItemsAggregatesResultRowsEqual(t, []groupItemsAggregatesResultRow{
		{
			GroupID:                 1,
			ItemID:                  1,
			CachedPartialAccessDate: ptrTime(currentDate.AddDate(0, 0, -8)),
			PartialAccessDate:       ptrTime(currentDate.AddDate(0, 0, -8)),
			CachedGrayedAccess:      false, // since we have partial access here
			PropagateAccess:         "done",
		},
		{
			GroupID:                 1,
			ItemID:                  2,
			CachedPartialAccessDate: nil, // since it has been already done
			PartialAccessDate:       ptrTime(currentDate.AddDate(0, 0, -9)),
			CachedGrayedAccess:      true, // since it has been already done
			PropagateAccess:         "done",
		},
		{
			GroupID:                 1,
			ItemID:                  3,
			CachedPartialAccessDate: ptrTime(currentDate.AddDate(0, 0, -10)), // since it has been already done
			PartialAccessDate:       nil,
			CachedGrayedAccess:      true,
			PropagateAccess:         "done",
		},
		expectedRow14,
		{
			GroupID:                 1,
			ItemID:                  11,
			CachedPartialAccessDate: nil, // since bAccessRestricted = 1

			CachedGrayedAccessDate: ptrTime(currentDate.AddDate(0, 0, -10)), // least sCachedPartialAccessDate(1_1, 1_3)

			CachedGrayedAccess: false, // since it is a new record and we do not set bCachedGrayedAccess=1 in ComputeAllAccess()
			PropagateAccess:    "done",
		},
		{
			GroupID:                 1,
			ItemID:                  12,
			CachedPartialAccessDate: nil, // since bAccessRestricted=1

			CachedGrayedAccessDate: ptrTime(currentDate.AddDate(0, 0, -8)), // from 1_1

			CachedGrayedAccess: false, // since it is a new record and we do not set bCachedGrayedAccess=1 in ComputeAllAccess()
			PropagateAccess:    "done",
		},
		{
			GroupID:                 2,
			ItemID:                  1,
			CachedPartialAccessDate: ptrTime(currentDate.AddDate(0, 0, -11)),
			PartialAccessDate:       ptrTime(currentDate.AddDate(0, 0, -11)),
			PropagateAccess:         "done",
		},
		{
			GroupID:                 2,
			ItemID:                  11,
			CachedPartialAccessDate: ptrTime(currentDate.AddDate(0, 0, -12)),
			PartialAccessDate:       ptrTime(currentDate.AddDate(0, 0, -12)),
			CachedGrayedAccessDate:  ptrTime(currentDate.AddDate(0, 0, -11)), // from 2_1

			CachedGrayedAccess: false, // since it is a new record and we do not set bCachedGrayedAccess=1 in ComputeAllAccess()
			PropagateAccess:    "done",
		},
		{
			GroupID: 2,
			ItemID:  12,

			CachedPartialAccessDate: nil, // since bAccessRestricted=1

			CachedGrayedAccessDate: ptrTime(currentDate.AddDate(0, 0, -12)), // from 2_11

			CachedGrayedAccess: false, // since it is a new record and we do not set bCachedGrayedAccess=1 in ComputeAllAccess()
			PropagateAccess:    "done",
		},
	}, result)
}

func TestGroupItemStore_ComputeAllAccess_AggregatesCachedAccessSolutionsDate(t *testing.T) {
	db := testhelpers.SetupDBWithFixture("group_item_store/compute_all_access/_common")
	defer func() { _ = db.Close() }()

	groupItemStore := database.NewDataStore(db).GroupItems()
	currentDate := time.Now().UTC().Round(time.Second)
	assert.NoError(t, groupItemStore.Where("idGroup=1 AND idItem=1").
		UpdateColumn("sAccessSolutionsDate", currentDate.AddDate(0, 0, -8)).Error())
	assert.NoError(t, groupItemStore.Where("idGroup=1 AND idItem=2").
		UpdateColumn("sAccessSolutionsDate", currentDate.AddDate(0, 0, -9)).Error())
	assert.NoError(t, groupItemStore.Where("idGroup=1 AND idItem=2").
		UpdateColumn("sPropagateAccess", "done").Error())
	assert.NoError(t, groupItemStore.Where("idGroup=1 AND idItem=3").
		UpdateColumn("sCachedAccessSolutionsDate", currentDate.AddDate(0, 0, -10)).Error())
	assert.NoError(t, groupItemStore.Where("idGroup=1 AND idItem=3").
		UpdateColumn("sPropagateAccess", "done").Error())
	assert.NoError(t, groupItemStore.Where("idGroup=2 AND idItem=1").
		UpdateColumn("sAccessSolutionsDate", currentDate.AddDate(0, 0, -11)).Error())
	assert.NoError(t, groupItemStore.Where("idGroup=2 AND idItem=11").
		UpdateColumn("sAccessSolutionsDate", currentDate.AddDate(0, 0, -12)).Error())
	assert.NoError(t, groupItemStore.UpdateColumn("bCachedAccessSolutions", 1).Error())
	assert.NoError(t, groupItemStore.InTransaction(func(ds *database.DataStore) error {
		ds.GroupItems().ComputeAllAccess()
		return nil
	}))

	var result []groupItemsAggregatesResultRow

	assert.NoError(t, groupItemStore.Order("idGroup, idItem").Scan(&result).Error())
	assertGroupsItemsAggregatesResultRowsEqual(t, []groupItemsAggregatesResultRow{
		{
			GroupID:                   1,
			ItemID:                    1,
			CachedAccessSolutionsDate: ptrTime(currentDate.AddDate(0, 0, -8)),
			AccessSolutionsDate:       ptrTime(currentDate.AddDate(0, 0, -8)),
			CachedAccessSolutions:     true,
			PropagateAccess:           "done",
		},
		{
			GroupID:                   1,
			ItemID:                    2,
			CachedAccessSolutionsDate: nil, // since it has been already done
			AccessSolutionsDate:       ptrTime(currentDate.AddDate(0, 0, -9)),
			CachedAccessSolutions:     true, // since it has been already done
			PropagateAccess:           "done",
		},
		{
			GroupID:                   1,
			ItemID:                    3,
			CachedAccessSolutionsDate: ptrTime(currentDate.AddDate(0, 0, -10)), // since it has been already done
			AccessSolutionsDate:       nil,
			CachedAccessSolutions:     true,
			PropagateAccess:           "done",
		},
		expectedRow14,
		{
			GroupID:                   1,
			ItemID:                    11,
			CachedAccessSolutionsDate: ptrTime(currentDate.AddDate(0, 0, -10)), // least(1_1, 1_3)

			CachedAccessSolutions: false, // since it is a new record and we do not set bCachedAccessSolutions=1 in ComputeAllAccess()
			PropagateAccess:       "done",
		},
		{
			GroupID:                   1,
			ItemID:                    12,
			CachedAccessSolutionsDate: ptrTime(currentDate.AddDate(0, 0, -10)), // least(1_1, 1_2, 1_11)

			CachedAccessSolutions: false, // since it is a new record and we do not set bCachedAccessSolutions=1 in ComputeAllAccess()
			PropagateAccess:       "done",
		},
		{
			GroupID:                   2,
			ItemID:                    1,
			CachedAccessSolutionsDate: ptrTime(currentDate.AddDate(0, 0, -11)),
			AccessSolutionsDate:       ptrTime(currentDate.AddDate(0, 0, -11)),
			CachedAccessSolutions:     true,
			PropagateAccess:           "done",
		},
		{
			GroupID:                   2,
			ItemID:                    11,
			CachedAccessSolutionsDate: ptrTime(currentDate.AddDate(0, 0, -12)),
			AccessSolutionsDate:       ptrTime(currentDate.AddDate(0, 0, -12)),
			CachedAccessSolutions:     true,
			PropagateAccess:           "done",
		},
		{
			GroupID:                   2,
			ItemID:                    12,
			CachedAccessSolutionsDate: ptrTime(currentDate.AddDate(0, 0, -12)),
			CachedAccessSolutions:     false, // since it is a new record and we do not set bCachedAccessSolutions=1 in ComputeAllAccess()
			PropagateAccess:           "done",
		},
	}, result)
}

func TestGroupItemStore_ComputeAllAccess_AggregatesCachedAccessReason_IsNotSetWhenNothingIsChanged(t *testing.T) {
	db := testhelpers.SetupDBWithFixture("group_item_store/compute_all_access/_common")
	defer func() { _ = db.Close() }()

	groupItemStore := database.NewDataStore(db).GroupItems()
	assert.NoError(t, groupItemStore.Where("idGroup=1 AND idItem=1").
		UpdateColumn("sAccessReason", "reason 1_1").Error())
	assert.NoError(t, groupItemStore.Where("idGroup=1 AND idItem=2").
		UpdateColumn("sAccessReason", "reason 1_2").Error())
	assert.NoError(t, groupItemStore.Where("idGroup=1 AND idItem=2").
		UpdateColumn("sPropagateAccess", "done").Error())
	assert.NoError(t, groupItemStore.Where("idGroup=1 AND idItem=3").
		UpdateColumn("sCachedAccessReason", "cached reason 1_3").Error())
	assert.NoError(t, groupItemStore.Where("idGroup=1 AND idItem=3").
		UpdateColumn("sPropagateAccess", "done").Error())
	assert.NoError(t, groupItemStore.Where("idGroup=2 AND idItem=1").
		UpdateColumn("sAccessReason", "reason 2_1").Error())
	assert.NoError(t, groupItemStore.Where("idGroup=2 AND idItem=11").
		UpdateColumn("sAccessReason", "reason 2_11").Error())
	assert.NoError(t, groupItemStore.InTransaction(func(ds *database.DataStore) error {
		ds.GroupItems().ComputeAllAccess()
		return nil
	}))

	var result []groupItemsAggregatesResultRow

	assert.NoError(t, groupItemStore.Order("idGroup, idItem").Scan(&result).Error())
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
	assert.NoError(t, groupItemStore.Where("idGroup=1 AND idItem=1").
		UpdateColumn("sAccessReason", "reason 1_1").Error())
	assert.NoError(t, groupItemStore.Where("idGroup=1 AND idItem=1").
		UpdateColumn("sFullAccessDate", currentDate.AddDate(0, 0, -8)).Error())
	assert.NoError(t, groupItemStore.Where("idGroup=1 AND idItem=2").
		UpdateColumn("sAccessReason", "reason 1_2").Error())
	assert.NoError(t, groupItemStore.Where("idGroup=1 AND idItem=2").
		UpdateColumn("bManagerAccess", 1).Error())
	assert.NoError(t, groupItemStore.Where("idGroup=1 AND idItem=3").
		UpdateColumn("sCachedAccessReason", "cached reason 1_3").Error())
	assert.NoError(t, groupItemStore.Where("idGroup=1 AND idItem=3").
		UpdateColumn("sPropagateAccess", "done").Error())
	assert.NoError(t, groupItemStore.Where("idGroup=2 AND idItem=1").
		UpdateColumn("sAccessReason", "reason 2_1").Error())
	assert.NoError(t, groupItemStore.Where("idGroup=2 AND idItem=1").
		UpdateColumn("sPartialAccessDate", currentDate.AddDate(0, 0, -11)).Error())
	assert.NoError(t, groupItemStore.Where("idGroup=2 AND idItem=11").
		UpdateColumn("sAccessReason", "reason 2_11").Error())
	assert.NoError(t, groupItemStore.Where("idGroup=2 AND idItem=11").
		UpdateColumn("sAccessSolutionsDate", currentDate.AddDate(0, 0, -12)).Error())
	assert.NoError(t, groupItemStore.InTransaction(func(ds *database.DataStore) error {
		ds.GroupItems().ComputeAllAccess()
		return nil
	}))

	var result []groupItemsAggregatesResultRow

	assert.NoError(t, groupItemStore.Order("idGroup, idItem").Scan(&result).Error())
	assertGroupsItemsAggregatesResultRowsEqual(t, []groupItemsAggregatesResultRow{
		{
			GroupID:              1,
			ItemID:               1,
			CachedFullAccessDate: ptrTime(currentDate.AddDate(0, 0, -8)),
			FullAccessDate:       ptrTime(currentDate.AddDate(0, 0, -8)),
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
			CachedFullAccessDate: ptrTime(currentDate.AddDate(0, 0, -8)),
			CachedManagerAccess:  true,
			CachedAccessReason:   ptrString("From ancestor group(s) reason 1_1, reason 1_2"), // concat(1_1, 1_2, 1_3)
			PropagateAccess:      "done",
		},
		{
			GroupID:              1,
			ItemID:               12,
			CachedFullAccessDate: ptrTime(currentDate.AddDate(0, 0, -8)),
			CachedManagerAccess:  true,
			CachedAccessReason:   ptrString("From ancestor group(s) reason 1_1, reason 1_2"), // concat(1_1, 1_2, 1_11)
			PropagateAccess:      "done",
		},
		{
			GroupID:                 2,
			ItemID:                  1,
			CachedPartialAccessDate: ptrTime(currentDate.AddDate(0, 0, -11)),
			PartialAccessDate:       ptrTime(currentDate.AddDate(0, 0, -11)),
			CachedAccessReason:      nil,
			AccessReason:            ptrString("reason 2_1"),
			PropagateAccess:         "done",
		},
		{
			GroupID:                   2,
			ItemID:                    11,
			CachedPartialAccessDate:   ptrTime(currentDate.AddDate(0, 0, -11)),
			CachedAccessSolutionsDate: ptrTime(currentDate.AddDate(0, 0, -12)),
			AccessSolutionsDate:       ptrTime(currentDate.AddDate(0, 0, -12)),
			CachedAccessReason:        ptrString("From ancestor group(s) reason 2_1"),
			AccessReason:              ptrString("reason 2_11"),
			PropagateAccess:           "done",
		},
		{
			GroupID:                   2,
			ItemID:                    12,
			CachedAccessSolutionsDate: ptrTime(currentDate.AddDate(0, 0, -12)),
			CachedAccessReason:        ptrString("From ancestor group(s) reason 2_11"),
			PropagateAccess:           "done",
		},
	}, result)
}

func ptrTime(t time.Time) *time.Time { return &t }
func ptrString(s string) *string     { return &s }
func ptrInt64(i int64) *int64        { return &i }

func assertGroupsItemsAggregatesResultRowsEqual(t *testing.T, expected, got []groupItemsAggregatesResultRow) {
	if len(got) != len(expected) {
		assert.ElementsMatch(t, expected, got)
		return
	}

	for i := 0; i < len(expected); i++ {
		assert.Equal(t, expected[i], got[i])
	}
}
