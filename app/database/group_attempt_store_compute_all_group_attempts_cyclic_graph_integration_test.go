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

func TestGroupAttemptStore_ComputeAllGroupAttempts_WithCyclicGraph(t *testing.T) {
	db := testhelpers.SetupDBWithFixture("groups_attempts_propagation/cyclic")
	defer func() { _ = db.Close() }()

	groupAttemptStore := database.NewDataStore(db).GroupAttempts()

	err := groupAttemptStore.InTransaction(func(s *database.DataStore) error {
		return s.GroupAttempts().ComputeAllGroupAttempts()
	})
	assert.NoError(t, err)

	var result []stateResultRow
	assert.NoError(t, groupAttemptStore.Select("id, result_propagation_state").Order("id").Scan(&result).Error())
	assert.Equal(t, []stateResultRow{
		{ID: 11, ResultPropagationState: "to_be_recomputed"},
		{ID: 12, ResultPropagationState: "to_be_recomputed"},
		// another user
		{ID: 22, ResultPropagationState: "done"},
	}, result)
}
