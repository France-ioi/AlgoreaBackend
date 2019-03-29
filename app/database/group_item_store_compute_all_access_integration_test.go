// +build !unit

package database_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/testhelpers"
)

type groupItemsResultRow struct {
	GroupID         int64  `gorm:"column:idGroup"`
	ItemID          int64  `gorm:"column:idItem"`
	PropagateAccess string `gorm:"column:sPropagateAccess"`
}

func TestGroupItemStore_ComputeAllAccess_Concurrency(t *testing.T) {
	db := testhelpers.SetupDBWithFixture("group_item_store/compute_all_access/_common")
	defer func() { _ = db.Close() }()

	dataStore := database.NewDataStore(db)
	const threadsNumber = 30
	done := make(chan bool, threadsNumber)
	for i := 0; i < threadsNumber; i++ {
		go func() {
			defer func() {
				done <- true
			}()
			assert.NoError(t, dataStore.InTransaction(func(ds *database.DataStore) error {
				ds.GroupItems().ComputeAllAccess()
				return nil
			}))
		}()
	}
	for i := 0; i < threadsNumber; i++ {
		<-done
	}

	groupItemStore := dataStore.GroupItems()
	var result []groupItemsResultRow

	assert.NoError(t, groupItemStore.Order("idGroup, idItem").Scan(&result).Error())
	assert.Equal(t, []groupItemsResultRow{
		{GroupID: 1, ItemID: 1, PropagateAccess: "done"},
		{GroupID: 1, ItemID: 2, PropagateAccess: "done"},
		{GroupID: 1, ItemID: 3, PropagateAccess: "done"},
		{GroupID: 1, ItemID: 4, PropagateAccess: "done"},
		{GroupID: 1, ItemID: 11, PropagateAccess: "done"},
		{GroupID: 1, ItemID: 12, PropagateAccess: "done"},
		{GroupID: 2, ItemID: 1, PropagateAccess: "done"},
		{GroupID: 2, ItemID: 11, PropagateAccess: "done"},
		{GroupID: 2, ItemID: 12, PropagateAccess: "done"},
	}, result)

	assert.NoError(t, groupItemStore.Table("groups_items_propagate").Joins("LEFT JOIN groups_items USING(ID)").
		Order("idGroup, idItem").Select("idGroup, idItem, groups_items_propagate.sPropagateAccess").Scan(&result).Error())
	assert.Equal(t, []groupItemsResultRow{
		{GroupID: 1, ItemID: 1, PropagateAccess: "done"},
		{GroupID: 1, ItemID: 2, PropagateAccess: "done"},
		{GroupID: 1, ItemID: 3, PropagateAccess: "done"},
		{GroupID: 1, ItemID: 4, PropagateAccess: "done"},
		{GroupID: 1, ItemID: 11, PropagateAccess: "done"},
		{GroupID: 1, ItemID: 12, PropagateAccess: "done"},
		{GroupID: 2, ItemID: 1, PropagateAccess: "done"},
		{GroupID: 2, ItemID: 11, PropagateAccess: "done"},
		{GroupID: 2, ItemID: 12, PropagateAccess: "done"},
	}, result)
}
