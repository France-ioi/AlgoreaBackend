// +build !unit

package database_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/testhelpers"
)

type existingAttemptsRow struct {
	GroupID                int64
	ItemID                 int64
	ResultPropagationState string
}

func testAttemptStoreComputeAllAttemptsCreatesNew(t *testing.T, fixtures []string,
	expectedNewAttempts []existingAttemptsRow) {
	mergedFixtures := make([]string, 0, len(fixtures)+1)
	mergedFixtures = append(mergedFixtures, `
		groups: [{id: 1}, {id: 2}, {id: 3}, {id: 4}]
		groups_ancestors:
			- {ancestor_group_id: 1, child_group_id: 1}
			- {ancestor_group_id: 2, child_group_id: 2}
			- {ancestor_group_id: 3, child_group_id: 3}
			- {ancestor_group_id: 1, child_group_id: 2}
			- {ancestor_group_id: 1, child_group_id: 3}
			- {ancestor_group_id: 2, child_group_id: 3}
			- {ancestor_group_id: 4, child_group_id: 4}
			- {ancestor_group_id: 4, child_group_id: 3, expires_at: 2019-05-30 11:00:00}
		items:
			- {id: 111, default_language_tag: fr}
			- {id: 222, default_language_tag: fr}
			- {id: 333, default_language_tag: fr}
		items_items:
			- {parent_item_id: 111, child_item_id: 222, child_order: 1}
			- {parent_item_id: 222, child_item_id: 333, child_order: 1}
		items_ancestors:
			- {ancestor_item_id: 111, child_item_id: 222}
			- {ancestor_item_id: 111, child_item_id: 333}
			- {ancestor_item_id: 222, child_item_id: 333}
		attempts:
			- {group_id: 3, item_id: 333, order: 1, result_propagation_state: to_be_propagated}
	`)
	mergedFixtures = append(mergedFixtures, fixtures...)
	db := testhelpers.SetupDBWithFixtureString(mergedFixtures...)
	defer func() { _ = db.Close() }()

	attemptStore := database.NewDataStore(db).Attempts()
	err := attemptStore.InTransaction(func(s *database.DataStore) error {
		return s.Attempts().ComputeAllAttempts()
	})
	assert.NoError(t, err)

	for i := range expectedNewAttempts {
		expectedNewAttempts[i].ResultPropagationState = "done"
	}
	expectedNewAttempts = append(expectedNewAttempts,
		existingAttemptsRow{GroupID: 3, ItemID: 333, ResultPropagationState: "done"})
	var result []existingAttemptsRow
	assert.NoError(t, attemptStore.Select("group_id, item_id, result_propagation_state").
		Order("group_id, item_id").Scan(&result).Error())
	assert.Equal(t, expectedNewAttempts, result)
}

func TestAttemptStore_ComputeAllAttempts_CreatesNew(t *testing.T) {
	for _, test := range []struct {
		name                string
		fixtures            []string
		expectedNewAttempts []existingAttemptsRow
	}{
		{name: "should not create new attempts if no permissions for parent items"},
		{
			name:     "should not create new attempts if can_view_generated = none for ancestor items",
			fixtures: []string{`permissions_generated: [{group_id: 3, item_id: 111, can_view_generated: none}]`},
		},
		{
			name:     "should not create new attempts if can_view_generated > none only for the item (not for its ancestor)",
			fixtures: []string{`permissions_generated: [{group_id: 3, item_id: 333, can_view_generated: info}]`},
		},
		{
			name:     "should not create new attempts if can_view_generated > none for an ancestor items and the group's expired ancestor",
			fixtures: []string{`permissions_generated: [{group_id: 4, item_id: 111, can_view_generated: info}]`},
		},
		{
			name:                "creates new attempts if can_view_generated > none for an ancestor items and the group itself",
			fixtures:            []string{`permissions_generated: [{group_id: 3, item_id: 111, can_view_generated: info}]`},
			expectedNewAttempts: []existingAttemptsRow{{GroupID: 3, ItemID: 111}, {GroupID: 3, ItemID: 222}},
		},
		{
			name:                "creates new attempts if can_view_generated > none for an ancestor items and the group's ancestor",
			fixtures:            []string{`permissions_generated: [{group_id: 1, item_id: 111, can_view_generated: info}]`},
			expectedNewAttempts: []existingAttemptsRow{{GroupID: 3, ItemID: 111}, {GroupID: 3, ItemID: 222}},
		},
		{
			name: "creates new attempts if can_view_generated > none for an ancestor items and the group itself, " +
				"but only for visible items's descendants",
			fixtures:            []string{`permissions_generated: [{group_id: 3, item_id: 222, can_view_generated: info}]`},
			expectedNewAttempts: []existingAttemptsRow{{GroupID: 3, ItemID: 222}},
		},
		{
			name: "creates new attempts if can_view_generated > none for an ancestor items and the group's ancestor, " +
				"but only for visible items's descendants",
			fixtures:            []string{`permissions_generated: [{group_id: 1, item_id: 222, can_view_generated: info}]`},
			expectedNewAttempts: []existingAttemptsRow{{GroupID: 3, ItemID: 222}},
		},
	} {
		test := test
		t.Run(test.name, func(t *testing.T) {
			testAttemptStoreComputeAllAttemptsCreatesNew(t, test.fixtures, test.expectedNewAttempts)
		})
	}
}
