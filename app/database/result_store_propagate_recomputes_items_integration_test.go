//go:build !unit

package database_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/testhelpers"
	"github.com/France-ioi/AlgoreaBackend/v2/testhelpers/testoutput"
)

func TestResultStore_Propagate_RecomputesResultsForItemsFromTableResultsRecomputeForItems(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db := testhelpers.SetupDBWithFixtureString(testhelpers.CreateTestContext(), `
		groups: [{id: 1}, {id: 2}, {id: 3}]
		items:
			- {id: 111, default_language_tag: fr}
			- {id: 222, default_language_tag: fr}
			- {id: 333, default_language_tag: fr}
			- {id: 444, default_language_tag: fr, requires_explicit_entry: 1}
		items_items:
			- {parent_item_id: 111, child_item_id: 222, child_order: 1}
			- {parent_item_id: 222, child_item_id: 333, child_order: 1}
			- {parent_item_id: 444, child_item_id: 333, child_order: 1}
		items_ancestors:
			- {ancestor_item_id: 111, child_item_id: 222}
			- {ancestor_item_id: 111, child_item_id: 333}
			- {ancestor_item_id: 222, child_item_id: 333}
			- {ancestor_item_id: 444, child_item_id: 333}
		results:
			- {participant_id: 1, attempt_id: 1, item_id: 111, latest_activity_at: '2019-05-30 11:00:00',  validated_at: '2019-05-30 11:00:00'}
			- {participant_id: 2, attempt_id: 1, item_id: 222, latest_activity_at: '2019-05-30 11:00:00', validated_at: '2019-05-30 11:00:00'}
			- {participant_id: 3, attempt_id: 2, item_id: 111, latest_activity_at: '2019-05-30 11:00:00', validated_at: '2019-05-30 11:00:00'}
			- {participant_id: 3, attempt_id: 2, item_id: 333, latest_activity_at: '2019-05-30 11:00:00', validated_at: '2019-05-30 11:00:00'}
			- {participant_id: 1, attempt_id: 1, item_id: 444, latest_activity_at: '2019-05-30 11:00:00', validated_at: '2019-05-30 11:00:00'}
			- {participant_id: 2, attempt_id: 1, item_id: 444, latest_activity_at: '2019-05-30 11:00:00', validated_at: '2019-05-30 11:00:00'}
		attempts:
			- {participant_id: 1, id: 1}
			- {participant_id: 2, id: 1}
			- {participant_id: 3, id: 2}
		results_recompute_for_items:
			- {item_id: 111}
			- {item_id: 222}
	`)
	defer func() { _ = db.Close() }()
	dataStore := database.NewDataStore(db)

	require.NoError(t,
		dataStore.InTransaction(func(dataStore *database.DataStore) error {
			dataStore.ScheduleResultsPropagation()
			return nil
		}), "Results Propagation failed")

	hasRows, err := dataStore.Table("results_recompute_for_items").HasRows()
	require.NoError(t, err)
	assert.False(t, hasRows)
	hasRows, err = dataStore.Table("results_propagate").HasRows()
	require.NoError(t, err)
	assert.False(t, hasRows)

	expectedTime, _ := time.Parse(time.DateTime, "2019-05-30 11:00:00")
	expectedDBTime := database.Time(expectedTime)
	var result []validationDateResultRow
	queryResultsAndStatesForTests(t, dataStore.Results(), &result, "validated_at")
	assert.Equal(t, []validationDateResultRow{
		{ParticipantID: 1, AttemptID: 1, ItemID: 111, State: "done", ValidatedAt: nil},
		{ParticipantID: 1, AttemptID: 1, ItemID: 444, State: "done", ValidatedAt: &expectedDBTime},
		{ParticipantID: 2, AttemptID: 1, ItemID: 222, State: "done", ValidatedAt: nil},
		{ParticipantID: 2, AttemptID: 1, ItemID: 444, State: "done", ValidatedAt: &expectedDBTime},
		{ParticipantID: 3, AttemptID: 2, ItemID: 111, State: "done", ValidatedAt: nil},
		{ParticipantID: 3, AttemptID: 2, ItemID: 333, State: "done", ValidatedAt: &expectedDBTime},
	}, result)
}
