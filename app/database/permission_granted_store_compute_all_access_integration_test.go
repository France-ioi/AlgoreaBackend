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

func TestPermissionGrantedStore_ComputeAllAccess_Concurrency(t *testing.T) {
	db := testhelpers.SetupDBWithFixture("permission_granted_store/compute_all_access/_common")
	defer func() { _ = db.Close() }()

	dataStore := database.NewDataStore(db)
	testhelpers.RunConcurrently(func() {
		assert.NoError(t, dataStore.InTransaction(func(ds *database.DataStore) error {
			ds.PermissionsGranted().ComputeAllAccess()
			return nil
		}))
	}, 30)

	permissionsGeneratedStore := dataStore.PermissionsGenerated()
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

	assert.NoError(t, permissionsGeneratedStore.
		Order("group_id, item_id").
		Select(`group_id, item_id, propagate_access`).
		Scan(&result).Error())
	assert.Equal(t, allDone, result)
}