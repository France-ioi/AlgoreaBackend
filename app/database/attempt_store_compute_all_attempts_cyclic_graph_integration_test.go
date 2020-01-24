// +build !unit

package database_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/testhelpers"
)

type stateResultRow struct {
	GroupID                int64
	ItemID                 int64
	ResultPropagationState string
}

func TestAttemptStore_ComputeAllAttempts_WithCyclicGraph(t *testing.T) {
	db := testhelpers.SetupDBWithFixture("attempts_propagation/cyclic")
	defer func() { _ = db.Close() }()

	attemptStore := database.NewDataStore(db).Attempts()

	err := attemptStore.InTransaction(func(s *database.DataStore) error {
		return s.Attempts().ComputeAllAttempts()
	})
	assert.NoError(t, err)

	var result []stateResultRow
	assert.NoError(t, attemptStore.Select("group_id, item_id, result_propagation_state").
		Order("group_id, item_id, `order`").Scan(&result).Error())
	assert.Equal(t, []stateResultRow{
		{GroupID: 101, ItemID: 1, ResultPropagationState: "to_be_recomputed"},
		{GroupID: 101, ItemID: 2, ResultPropagationState: "to_be_recomputed"},
		// another user
		{GroupID: 102, ItemID: 1, ResultPropagationState: "to_be_recomputed"},
		{GroupID: 102, ItemID: 2, ResultPropagationState: "to_be_recomputed"},
		{GroupID: 102, ItemID: 3, ResultPropagationState: "done"},
	}, result)
}
