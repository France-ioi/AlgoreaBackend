// +build !unit

package database_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/testhelpers"
)

type stateResultRow struct {
	ID                     int64
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
	assert.NoError(t, attemptStore.Select("id, result_propagation_state").Order("id").Scan(&result).Error())
	assert.Equal(t, []stateResultRow{
		{ID: 11, ResultPropagationState: "to_be_recomputed"},
		{ID: 12, ResultPropagationState: "to_be_recomputed"},
		// another user
		{ID: 22, ResultPropagationState: "done"},
	}, result)
}
