// +build !unit

package database_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/testhelpers"
)

type itemAncestorsResultRow struct {
	ItemAncestorID int64 `gorm:"column:idItemAncestor"`
	ItemChildID    int64 `gorm:"column:idItemChild"`
}

type itemPropagateResultRow struct {
	ID                        int64  `gorm:"column:ID"`
	AncestorsComputationState string `gorm:"column:sAncestorsComputationState"`
}

func TestItemItemStore_CreateNewAncestors_Concurrent(t *testing.T) {
	db := testhelpers.SetupDBWithFixture("item_item_store/ancestors/_common")
	defer func() { _ = db.Close() }()

	itemItemStore := database.NewDataStore(db).ItemItems()
	testhelpers.RunConcurrently(func() {
		assert.NoError(t, itemItemStore.InTransaction(func(ds *database.DataStore) error {
			ds.ItemItems().CreateNewAncestors()
			return nil
		}))
	}, 30)

	var result []itemAncestorsResultRow
	assert.NoError(t, itemItemStore.ItemAncestors().Order("idItemChild, idItemAncestor").Scan(&result).Error())

	assert.Equal(t, []itemAncestorsResultRow{
		{ItemChildID: 2, ItemAncestorID: 1},
		{ItemChildID: 3, ItemAncestorID: 1},
		{ItemChildID: 3, ItemAncestorID: 2},
		{ItemChildID: 4, ItemAncestorID: 1},
		{ItemChildID: 4, ItemAncestorID: 2},
		{ItemChildID: 4, ItemAncestorID: 3},
	}, result)

	var propagateResult []itemPropagateResultRow
	assert.NoError(t, itemItemStore.Table("items_propagate").Order("ID").Scan(&propagateResult).Error())
	assert.Equal(t, []itemPropagateResultRow{
		{ID: 2, AncestorsComputationState: "done"},
		{ID: 3, AncestorsComputationState: "done"},
		{ID: 4, AncestorsComputationState: "done"},
	}, propagateResult)
}

func TestItemItemStore_CreateNewAncestors_Cyclic(t *testing.T) {
	db := testhelpers.SetupDBWithFixture("item_item_store/ancestors/_common", "item_item_store/ancestors/cyclic")
	defer func() { _ = db.Close() }()

	itemItemStore := database.NewDataStore(db).ItemItems()
	assert.NoError(t, itemItemStore.InTransaction(func(ds *database.DataStore) error {
		ds.ItemItems().CreateNewAncestors()
		return nil
	}))

	var result []itemAncestorsResultRow
	assert.NoError(t, itemItemStore.ItemAncestors().Order("idItemChild, idItemAncestor").Scan(&result).Error())

	assert.Equal(t, []itemAncestorsResultRow{
		{ItemChildID: 3, ItemAncestorID: 2}, // this one has already been there
	}, result)

	var propagateResult []itemPropagateResultRow
	assert.NoError(t, itemItemStore.Table("items_propagate").Order("ID").Scan(&propagateResult).Error())
	assert.Equal(t, []itemPropagateResultRow{
		{ID: 1, AncestorsComputationState: "todo"},
		{ID: 2, AncestorsComputationState: "todo"},
		{ID: 3, AncestorsComputationState: "todo"},
		{ID: 4, AncestorsComputationState: "todo"},
	}, propagateResult)
}

func TestItemItemStore_CreateNewAncestors_IgnoresDoneItems(t *testing.T) {
	db := testhelpers.SetupDBWithFixture("item_item_store/ancestors/_common")
	defer func() { _ = db.Close() }()

	itemItemStore := database.NewDataStore(db).ItemItems()

	for i := 1; i <= 4; i++ {
		assert.NoError(t, itemItemStore.Exec(
			"INSERT INTO items_propagate (ID, sAncestorsComputationState) VALUES (?, 'done') "+
				"ON DUPLICATE KEY UPDATE sAncestorsComputationState='done'", i).
			Error())
	}

	assert.NoError(t, itemItemStore.InTransaction(func(ds *database.DataStore) error {
		ds.ItemItems().CreateNewAncestors()
		return nil
	}))

	var result []itemAncestorsResultRow
	assert.NoError(t, itemItemStore.ItemAncestors().Order("idItemChild, idItemAncestor").Scan(&result).Error())

	assert.Equal(t, []itemAncestorsResultRow{
		{ItemChildID: 3, ItemAncestorID: 2}, // this one has already been there
	}, result)

	var propagateResult []itemPropagateResultRow
	assert.NoError(t, itemItemStore.Table("items_propagate").Order("ID").Scan(&propagateResult).Error())
	assert.Equal(t, []itemPropagateResultRow{
		{ID: 1, AncestorsComputationState: "done"},
		{ID: 2, AncestorsComputationState: "done"},
		{ID: 3, AncestorsComputationState: "done"},
		{ID: 4, AncestorsComputationState: "done"},
	}, propagateResult)
}
