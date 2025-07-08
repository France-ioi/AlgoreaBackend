//go:build !unit

package database_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/golang"
	"github.com/France-ioi/AlgoreaBackend/v2/testhelpers"
	"github.com/France-ioi/AlgoreaBackend/v2/testhelpers/testoutput"
)

type existingResultsRow struct {
	ParticipantID    int64
	AttemptID        int64
	ItemID           int64
	LatestActivityAt string
	State            string
}

type resultStorePropagateCreatesNewTestCase struct {
	name               string
	fixtures           []string
	expectedNewResults []existingResultsRow
	rootItemID         *int64
}

func testResultStorePropagateCreatesNew(t *testing.T, testCase *resultStorePropagateCreatesNewTestCase) {
	mergedFixtures := make([]string, 0, len(testCase.fixtures)+1)
	mergedFixtures = append(mergedFixtures, `
		groups: [{id: 1}, {id: 2}, {id: 3}, {id: 4}]
		groups_ancestors:
			- {ancestor_group_id: 1, child_group_id: 2}
			- {ancestor_group_id: 1, child_group_id: 3}
			- {ancestor_group_id: 2, child_group_id: 3}
			- {ancestor_group_id: 4, child_group_id: 3, expires_at: 2019-05-30 11:00:00}
		items:
			- {id: 111, default_language_tag: fr}
			- {id: 222, default_language_tag: fr}
			- {id: 333, default_language_tag: fr}
			- {id: 444, default_language_tag: fr, requires_explicit_entry: 1}
			- {id: 555, default_language_tag: fr}
		items_items:
			- {parent_item_id: 111, child_item_id: 222, child_order: 1}
			- {parent_item_id: 222, child_item_id: 333, child_order: 1}
			- {parent_item_id: 444, child_item_id: 333, child_order: 1}
			- {parent_item_id: 555, child_item_id: 444, child_order: 1}
		items_ancestors:
			- {ancestor_item_id: 111, child_item_id: 222}
			- {ancestor_item_id: 111, child_item_id: 333}
			- {ancestor_item_id: 222, child_item_id: 333}
			- {ancestor_item_id: 444, child_item_id: 333}
			- {ancestor_item_id: 555, child_item_id: 333}
			- {ancestor_item_id: 555, child_item_id: 444}
		attempts:
			- {participant_id: 3, id: 1}
		results:
			- {participant_id: 3, attempt_id: 1, item_id: 333, latest_activity_at: "2019-05-30 11:00:00"}
		results_propagate:
			- {participant_id: 3, attempt_id: 1, item_id: 333, state: to_be_propagated}
	`)
	mergedFixtures = append(mergedFixtures, testCase.fixtures...)
	db := testhelpers.SetupDBWithFixtureString(mergedFixtures...)
	defer func() { _ = db.Close() }()

	if testCase.rootItemID != nil {
		assert.NoError(t, database.NewDataStore(db).Attempts().Where("participant_id = 3 AND id = 1").
			UpdateColumn("root_item_id", testCase.rootItemID).Error())
	}
	resultStore := database.NewDataStore(db).Results()
	err := resultStore.InTransaction(func(s *database.DataStore) error {
		s.ScheduleResultsPropagation()
		return nil
	})
	assert.NoError(t, err)

	const expectedDate = "2019-05-30 11:00:00"
	for i := range testCase.expectedNewResults {
		testCase.expectedNewResults[i].State = "done"
		testCase.expectedNewResults[i].LatestActivityAt = expectedDate
	}
	testCase.expectedNewResults = append(testCase.expectedNewResults,
		existingResultsRow{ParticipantID: 3, AttemptID: 1, ItemID: 333, LatestActivityAt: expectedDate, State: "done"})
	var result []existingResultsRow
	queryResultsAndStatesForTests(t, resultStore, &result, "latest_activity_at")
	assert.Equal(t, testCase.expectedNewResults, result)
}

func TestResultStore_Propagate_CreatesNew(t *testing.T) {
	for _, test := range []resultStorePropagateCreatesNewTestCase{
		{name: "should not create new results if no permissions for parent items"},
		{
			name:     "should not create new results if can_view_generated = none for ancestor items",
			fixtures: []string{`permissions_generated: [{group_id: 3, item_id: 111, can_view_generated: none}]`},
		},
		{
			name:     "should not create new results if can_view_generated > none only for the item (not for its ancestor)",
			fixtures: []string{`permissions_generated: [{group_id: 3, item_id: 333, can_view_generated: info}]`},
		},
		{
			name:     "should not create new results if can_view_generated > none for an ancestor items and the group's expired ancestor",
			fixtures: []string{`permissions_generated: [{group_id: 4, item_id: 111, can_view_generated: info}]`},
		},
		{
			name:     "creates new results if can_view_generated > none for an ancestor items and the group itself",
			fixtures: []string{`permissions_generated: [{group_id: 3, item_id: 111, can_view_generated: info}]`},
			expectedNewResults: []existingResultsRow{
				{ParticipantID: 3, AttemptID: 1, ItemID: 111}, {ParticipantID: 3, AttemptID: 1, ItemID: 222},
			},
		},
		{
			name:     "creates new results if can_view_generated > none for an ancestor items and the group's ancestor",
			fixtures: []string{`permissions_generated: [{group_id: 1, item_id: 111, can_view_generated: info}]`},
			expectedNewResults: []existingResultsRow{
				{ParticipantID: 3, AttemptID: 1, ItemID: 111}, {ParticipantID: 3, AttemptID: 1, ItemID: 222},
			},
		},
		{
			name: "creates new results if can_view_generated > none for an ancestor items and the group itself, " +
				"but only for visible items's descendants",
			fixtures:           []string{`permissions_generated: [{group_id: 3, item_id: 222, can_view_generated: info}]`},
			expectedNewResults: []existingResultsRow{{ParticipantID: 3, AttemptID: 1, ItemID: 222}},
		},
		{
			name: "creates new results if can_view_generated > none for an ancestor items and the group's ancestor, " +
				"but only for visible items's descendants",
			fixtures:           []string{`permissions_generated: [{group_id: 1, item_id: 222, can_view_generated: info}]`},
			expectedNewResults: []existingResultsRow{{ParticipantID: 3, AttemptID: 1, ItemID: 222}},
		},
		{
			name:     "should not create new results for items requiring explicit entry",
			fixtures: []string{`permissions_generated: [{group_id: 1, item_id: 555, can_view_generated: info}]`},
		},
		{
			name:               "should not create new results for items above the root_item_id",
			fixtures:           []string{`permissions_generated: [{group_id: 3, item_id: 111, can_view_generated: info}]`},
			expectedNewResults: []existingResultsRow{{ParticipantID: 3, AttemptID: 1, ItemID: 222}},
			rootItemID:         golang.Ptr(int64(222)),
		},
	} {
		test := test
		t.Run(test.name, func(t *testing.T) {
			testoutput.SuppressIfPasses(t)
			testResultStorePropagateCreatesNew(t, &test)
		})
	}
}
