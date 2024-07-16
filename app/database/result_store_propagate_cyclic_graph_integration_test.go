//go:build !unit

package database_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/testhelpers"
)

type stateResultRow struct {
	ParticipantID int64
	AttemptID     int64
	ItemID        int64
	State         string
}

func TestResultStore_Propagate_WithCyclicGraph(t *testing.T) {
	db := testhelpers.SetupDBWithFixture("results_propagation/cyclic")
	defer func() { _ = db.Close() }()

	resultStore := database.NewDataStore(db).Results()

	err := resultStore.InTransaction(func(s *database.DataStore) error {
		s.ScheduleResultsPropagation()
		return nil
	})
	assert.NoError(t, err)

	var result []stateResultRow
	assert.NoError(t, resultStore.Table("results_propagate").
		Select("participant_id, attempt_id, item_id, state").
		Order("participant_id, item_id, attempt_id").Scan(&result).Error())
	assert.Equal(t, []stateResultRow{
		{ParticipantID: 101, AttemptID: 1, ItemID: 1, State: "to_be_recomputed"},
		{ParticipantID: 101, AttemptID: 1, ItemID: 2, State: "to_be_recomputed"},
	}, result)
}
