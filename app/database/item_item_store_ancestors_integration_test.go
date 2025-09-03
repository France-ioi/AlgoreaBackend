//go:build !unit

package database_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/testhelpers"
	"github.com/France-ioi/AlgoreaBackend/v2/testhelpers/testoutput"
)

type itemAncestorsResultRow struct {
	AncestorItemID int64
	ChildItemID    int64
}

type itemPropagateResultRow struct {
	ID                        int64
	AncestorsComputationState string
}

func TestItemItemStore_CreateNewAncestors_Concurrent(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db := testhelpers.SetupDBWithFixture(testhelpers.CreateTestContext(), "item_item_store/ancestors/_common")
	defer func() { _ = db.Close() }()

	testhelpers.RunConcurrently(func() {
		dataStore := database.NewDataStoreWithContext(db.GetContext(), db)
		assert.NoError(t, dataStore.InTransaction(func(ds *database.DataStore) error {
			return ds.ItemItems().CreateNewAncestors()
		}))
	}, 30)

	itemItemStore := database.NewDataStore(db).ItemItems()
	var result []itemAncestorsResultRow
	require.NoError(t, itemItemStore.ItemAncestors().Order("child_item_id, ancestor_item_id").Scan(&result).Error())

	assert.Equal(t, []itemAncestorsResultRow{
		{ChildItemID: 2, AncestorItemID: 1},
		{ChildItemID: 3, AncestorItemID: 1},
		{ChildItemID: 3, AncestorItemID: 2},
		{ChildItemID: 4, AncestorItemID: 1},
		{ChildItemID: 4, AncestorItemID: 2},
		{ChildItemID: 4, AncestorItemID: 3},
	}, result)

	var propagateResult []itemPropagateResultRow
	require.NoError(t, itemItemStore.Table("items_propagate").Order("id").Scan(&propagateResult).Error())
	assert.Equal(t, []itemPropagateResultRow{
		{ID: 2, AncestorsComputationState: "done"},
		{ID: 3, AncestorsComputationState: "done"},
		{ID: 4, AncestorsComputationState: "done"},
	}, propagateResult)
}

func TestItemItemStore_CreateNewAncestors_Cyclic(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db := testhelpers.SetupDBWithFixture(testhelpers.CreateTestContext(),
		"item_item_store/ancestors/_common", "item_item_store/ancestors/cyclic")
	defer func() { _ = db.Close() }()

	itemItemStore := database.NewDataStore(db).ItemItems()
	assert.NoError(t, itemItemStore.InTransaction(func(ds *database.DataStore) error {
		return ds.ItemItems().CreateNewAncestors()
	}))

	var result []itemAncestorsResultRow
	assert.NoError(t, itemItemStore.ItemAncestors().Order("child_item_id, ancestor_item_id").Scan(&result).Error())

	assert.Equal(t, []itemAncestorsResultRow{
		{ChildItemID: 3, AncestorItemID: 2}, // this one has already been there
	}, result)

	var propagateResult []itemPropagateResultRow
	require.NoError(t, itemItemStore.Table("items_propagate").Order("id").Scan(&propagateResult).Error())
	assert.Equal(t, []itemPropagateResultRow{
		{ID: 1, AncestorsComputationState: "todo"},
		{ID: 2, AncestorsComputationState: "todo"},
		{ID: 3, AncestorsComputationState: "todo"},
		{ID: 4, AncestorsComputationState: "todo"},
	}, propagateResult)
}

func TestItemItemStore_CreateNewAncestors_IgnoresDoneItems(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db := testhelpers.SetupDBWithFixture(testhelpers.CreateTestContext(), "item_item_store/ancestors/_common")
	defer func() { _ = db.Close() }()

	itemItemStore := database.NewDataStore(db).ItemItems()

	for i := 1; i <= 4; i++ {
		require.NoError(t, itemItemStore.Exec(
			"INSERT INTO items_propagate (id, ancestors_computation_state) VALUES (?, 'done') "+
				"ON DUPLICATE KEY UPDATE ancestors_computation_state='done'", i).
			Error())
	}

	assert.NoError(t, itemItemStore.InTransaction(func(ds *database.DataStore) error {
		return ds.ItemItems().CreateNewAncestors()
	}))

	var result []itemAncestorsResultRow
	assert.NoError(t, itemItemStore.ItemAncestors().Order("child_item_id, ancestor_item_id").Scan(&result).Error())

	assert.Equal(t, []itemAncestorsResultRow{
		{ChildItemID: 3, AncestorItemID: 2}, // this one has already been there
	}, result)

	var propagateResult []itemPropagateResultRow
	require.NoError(t, itemItemStore.Table("items_propagate").Order("id").Scan(&propagateResult).Error())
	assert.Equal(t, []itemPropagateResultRow{
		{ID: 1, AncestorsComputationState: "done"},
		{ID: 2, AncestorsComputationState: "done"},
		{ID: 3, AncestorsComputationState: "done"},
		{ID: 4, AncestorsComputationState: "done"},
	}, propagateResult)
}
