// +build !unit

package database_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/testhelpers"
)

type stateResultRow struct {
	ParticipantID          int64
	AttemptID              int64
	ItemID                 int64
	ResultPropagationState string
}

func TestResultStore_Propagate_WithCyclicGraph(t *testing.T) {
	db := testhelpers.SetupDBWithFixture("results_propagation/cyclic")
	defer func() { _ = db.Close() }()

	resultStore := database.NewDataStore(db).Results()

	err := resultStore.InTransaction(func(s *database.DataStore) error {
		return s.Results().Propagate()
	})
	assert.NoError(t, err)

	var result []stateResultRow
	assert.NoError(t, resultStore.Select("participant_id, attempt_id, item_id, result_propagation_state").
		Order("participant_id, item_id, attempt_id").Scan(&result).Error())
	assert.Equal(t, []stateResultRow{
		{ParticipantID: 101, AttemptID: 1, ItemID: 1, ResultPropagationState: "to_be_recomputed"},
		{ParticipantID: 101, AttemptID: 1, ItemID: 2, ResultPropagationState: "to_be_recomputed"},
		// another user
		{ParticipantID: 102, AttemptID: 1, ItemID: 2, ResultPropagationState: "done"},
		{ParticipantID: 102, AttemptID: 1, ItemID: 3, ResultPropagationState: "done"},
	}, result)
}
