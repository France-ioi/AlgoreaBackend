// +build !unit

package database_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/testhelpers"
)

func TestGroupItemStore_ComputeCached(t *testing.T) {
	db := testhelpers.SetupDBWithFixture("group_item_store/compute_cached")
	defer func() { _ = db.Close() }()

	groupItemStore := database.NewDataStore(db).GroupItems()
	assert.NoError(t, groupItemStore.ComputeCached())

	type groupItemCachedResult struct {
		ID                    int64 `gorm:"column:ID"`
		CachedFullAccess      bool  `gorm:"column:bCachedFullAccess"`
		CachedPartialAccess   bool  `gorm:"column:bCachedPartialAccess"`
		CachedAccessSolutions bool  `gorm:"column:bCachedAccessSolutions"`
		CachedGrayedAccess    bool  `gorm:"column:bCachedGrayedAccess"`
	}

	expected := []groupItemCachedResult{
		{ID: 11, CachedFullAccess: true},
		{ID: 12, CachedPartialAccess: true},
		{ID: 13, CachedAccessSolutions: true},
		{ID: 14, CachedGrayedAccess: true},

		{ID: 21, CachedFullAccess: false},
		{ID: 22, CachedPartialAccess: false},
		{ID: 23, CachedAccessSolutions: false},
		{ID: 24, CachedGrayedAccess: false},

		{ID: 31, CachedFullAccess: false},
		{ID: 32, CachedPartialAccess: false},
		{ID: 33, CachedAccessSolutions: false},
		{ID: 34, CachedGrayedAccess: false},

		{ID: 41, CachedFullAccess: false},
		{ID: 42, CachedPartialAccess: false},
		{ID: 43, CachedAccessSolutions: false},
		{ID: 44, CachedGrayedAccess: false},

		{ID: 51, CachedFullAccess: false},
		{ID: 52, CachedPartialAccess: false},
		{ID: 53, CachedAccessSolutions: false},
		{ID: 54, CachedGrayedAccess: false},

		{ID: 61, CachedFullAccess: true},
		{ID: 62, CachedPartialAccess: true},
		{ID: 63, CachedAccessSolutions: true},
		{ID: 64, CachedGrayedAccess: true},
	}
	var result []groupItemCachedResult
	assert.NoError(t, groupItemStore.Order("ID").Scan(&result).Error())
	assert.Equal(t, expected, result)
}
