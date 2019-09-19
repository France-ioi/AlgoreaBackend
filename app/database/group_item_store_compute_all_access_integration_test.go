// +build !unit

package database_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/testhelpers"
)

type groupItemsResultRow struct {
	GroupID         int64
	ItemID          int64
	PropagateAccess string
}

func TestGroupItemStore_ComputeAllAccess_Concurrency(t *testing.T) {
	db := testhelpers.SetupDBWithFixture("group_item_store/compute_all_access/_common")
	defer func() { _ = db.Close() }()

	dataStore := database.NewDataStore(db)
	testhelpers.RunConcurrently(func() {
		assert.NoError(t, dataStore.InTransaction(func(ds *database.DataStore) error {
			ds.GroupItems().ComputeAllAccess()
			return nil
		}))
	}, 30)

	groupItemStore := dataStore.GroupItems()
	var result []groupItemsResultRow

	allDone := []groupItemsResultRow{
		{GroupID: 1, ItemID: 1, PropagateAccess: "done"},
		{GroupID: 1, ItemID: 2, PropagateAccess: "done"},
		{GroupID: 1, ItemID: 3, PropagateAccess: "done"},
		{GroupID: 1, ItemID: 4, PropagateAccess: "done"},
		{GroupID: 1, ItemID: 11, PropagateAccess: "done"},
		{GroupID: 1, ItemID: 12, PropagateAccess: "done"},
		{GroupID: 2, ItemID: 1, PropagateAccess: "done"},
		{GroupID: 2, ItemID: 11, PropagateAccess: "done"},
		{GroupID: 2, ItemID: 12, PropagateAccess: "done"},
	}

	assert.NoError(t, groupItemStore.Order("group_id, item_id").Scan(&result).Error())
	assert.Equal(t, allDone, result)

	assert.NoError(t, groupItemStore.Table("groups_items_propagate").Joins("LEFT JOIN groups_items USING(id)").
		Order("group_id, item_id").Select("group_id, item_id, groups_items_propagate.propagate_access").Scan(&result).Error())
	assert.Equal(t, allDone, result)
}
