// +build !unit

package database_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/testhelpers"
)

func TestGroupGroupStore_DeleteRelation(t *testing.T) {
	tests := []struct {
		name                     string
		fixture                  string
		shouldDeleteOrphans      bool
		wantErr                  error
		remainingGroupIDs        []int64
		remainingGroupsGroups    []map[string]interface{}
		remainingGroupsAncestors []map[string]interface{}
	}{
		{
			name:                "refuses to produce orphans",
			fixture:             "orphan",
			shouldDeleteOrphans: false,
			wantErr:             database.ErrGroupBecomesOrphan,
			remainingGroupIDs:   []int64{1, 2},
			remainingGroupsGroups: []map[string]interface{}{
				{"parent_group_id": "1", "child_group_id": "2"},
			},
			remainingGroupsAncestors: []map[string]interface{}{
				{"ancestor_group_id": "1", "child_group_id": "1"},
				{"ancestor_group_id": "1", "child_group_id": "2"},
				{"ancestor_group_id": "2", "child_group_id": "2"},
			},
		},
		{
			name:                "no orphans",
			fixture:             "no_orphans",
			shouldDeleteOrphans: false,
			remainingGroupIDs:   []int64{1, 2, 3},
			remainingGroupsGroups: []map[string]interface{}{
				{"parent_group_id": "3", "child_group_id": "2"},
			},
			remainingGroupsAncestors: []map[string]interface{}{
				{"ancestor_group_id": "1", "child_group_id": "1"},
				{"ancestor_group_id": "2", "child_group_id": "2"},
				{"ancestor_group_id": "3", "child_group_id": "2"},
				{"ancestor_group_id": "3", "child_group_id": "3"},
			},
		},
		{
			name:                  "deletes orphans",
			fixture:               "orphan",
			shouldDeleteOrphans:   true,
			remainingGroupIDs:     []int64{1},
			remainingGroupsGroups: nil,
			remainingGroupsAncestors: []map[string]interface{}{
				{"ancestor_group_id": "1", "child_group_id": "1"},
			},
		},
		{
			name:                "deletes only orphans",
			fixture:             "deletes_only_orphans",
			shouldDeleteOrphans: true,
			remainingGroupIDs:   []int64{1, 3, 5, 10},
			remainingGroupsGroups: []map[string]interface{}{
				{"parent_group_id": "10", "child_group_id": "3"},
				{"parent_group_id": "10", "child_group_id": "5"},
			},
			remainingGroupsAncestors: []map[string]interface{}{
				{"ancestor_group_id": "1", "child_group_id": "1"},
				{"ancestor_group_id": "3", "child_group_id": "3"},
				{"ancestor_group_id": "5", "child_group_id": "5"},
				{"ancestor_group_id": "10", "child_group_id": "3"},
				{"ancestor_group_id": "10", "child_group_id": "5"},
				{"ancestor_group_id": "10", "child_group_id": "10"},
			},
		},
		{
			name:                  "deletes only orphans of accepted types",
			fixture:               "deletes_only_orphans_of_accepted_types",
			shouldDeleteOrphans:   true,
			remainingGroupIDs:     []int64{1, 8, 9, 11},
			remainingGroupsGroups: nil,
			remainingGroupsAncestors: []map[string]interface{}{
				{"ancestor_group_id": "1", "child_group_id": "1"},
				{"ancestor_group_id": "8", "child_group_id": "8"},
				{"ancestor_group_id": "9", "child_group_id": "9"},
				{"ancestor_group_id": "11", "child_group_id": "11"},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			db := testhelpers.SetupDBWithFixture("group_group_store/delete_relation/" + tt.fixture)
			defer func() { _ = db.Close() }()
			dataStore := database.NewDataStore(db)
			assert.NoError(t, dataStore.Table("groups_propagate").UpdateColumn("ancestors_computation_state", "done").Error())

			err := dataStore.InTransaction(func(s *database.DataStore) error {
				return s.GroupGroups().DeleteRelation(1, 2, tt.shouldDeleteOrphans)
			})
			assert.Equal(t, tt.wantErr, err)
			assertGroupRelations(t, dataStore, tt.remainingGroupIDs, tt.remainingGroupsGroups, tt.remainingGroupsAncestors)
			assertGroupLinkedObjects(t, dataStore, tt.remainingGroupIDs)
		})
	}
}

func assertGroupRelations(t *testing.T, dataStore *database.DataStore,
	remainingGroupIDs []int64, remainingGroupsGroups, remainingGroupsAncestors []map[string]interface{}) {
	var rows []map[string]interface{}
	var ids []int64

	totalRemainingGroupIDs := make([]int64, 0, len(remainingGroupIDs)+1)
	totalRemainingGroupIDs = append(totalRemainingGroupIDs, remainingGroupIDs...)
	totalRemainingGroupIDs = append(totalRemainingGroupIDs, int64(111))
	assert.NoError(t, dataStore.Groups().Order("id").Pluck("id", &ids).Error())
	assert.Equal(t, totalRemainingGroupIDs, ids)
	assert.NoError(t, dataStore.GroupGroups().Select("parent_group_id, child_group_id").Order("parent_group_id, child_group_id").
		ScanIntoSliceOfMaps(&rows).Error())
	assert.Equal(t, remainingGroupsGroups, rows)
	assert.NoError(t, dataStore.GroupAncestors().Select("ancestor_group_id, child_group_id").Order("ancestor_group_id, child_group_id").
		ScanIntoSliceOfMaps(&rows).Error())
	assert.Equal(t, remainingGroupsAncestors, rows)
	var count int64
	assert.NoError(t, dataStore.Table("groups_propagate").Where("ancestors_computation_state != 'done'").Count(&count).Error())
	assert.Zero(t, count)
}

func assertGroupLinkedObjects(t *testing.T, dataStore *database.DataStore, remainingGroupIDs []int64) {
	var ids []int64
	assert.NoError(t, dataStore.Table("filters").Order("group_id").
		Pluck("group_id", &ids).Error())
	assert.Equal(t, remainingGroupIDs, ids)
	assert.NoError(t, dataStore.Attempts().Order("group_id").
		Pluck("group_id", &ids).Error())
	assert.Equal(t, remainingGroupIDs, ids)
	assert.NoError(t, dataStore.PermissionsGranted().Order("group_id").
		Pluck("group_id", &ids).Error())
	assert.Equal(t, remainingGroupIDs, ids)
	assert.NoError(t, dataStore.Permissions().Order("group_id").
		Pluck("group_id", &ids).Error())
	assert.Equal(t, remainingGroupIDs, ids)
	assert.NoError(t, dataStore.Table("groups_login_prefixes").Order("group_id").
		Pluck("group_id", &ids).Error())
	assert.Equal(t, remainingGroupIDs, ids)

	totalRemainingGroupIDs := make([]int64, 0, len(remainingGroupIDs)+1)
	totalRemainingGroupIDs = append(totalRemainingGroupIDs, remainingGroupIDs...)
	totalRemainingGroupIDs = append(totalRemainingGroupIDs, int64(111))
	assert.NoError(t, dataStore.Table("groups_propagate").Order("id").
		Pluck("id", &ids).Error())
	assert.Equal(t, totalRemainingGroupIDs, ids)
}

const done = "done"
const groupGroupMarksAttemptsAsChangedFixture = `
	items:
		- {id: 1, default_language_tag: fr}
		- {id: 2, default_language_tag: fr}
		- {id: 3, default_language_tag: fr}
	items_ancestors:
		- {ancestor_item_id: 1, child_item_id: 2}
		- {ancestor_item_id: 1, child_item_id: 3}
		- {ancestor_item_id: 2, child_item_id: 3}
	groups: [{id: 101}, {id: 102}, {id: 103}, {id: 104}, {id: 105}, {id: 106}, {id: 107}, {id: 108}]
	groups_groups:
		- {parent_group_id: 101, child_group_id: 102, child_order: 1}
		- {parent_group_id: 101, child_group_id: 103, child_order: 2}
		- {parent_group_id: 101, child_group_id: 104, child_order: 3}
		- {parent_group_id: 101, child_group_id: 105, child_order: 4}
		- {parent_group_id: 101, child_group_id: 106, child_order: 5}
		- {parent_group_id: 102, child_group_id: 103, child_order: 1}
		- {parent_group_id: 104, child_group_id: 105, child_order: 1}
		- {parent_group_id: 107, child_group_id: 105, child_order: 1}
		- {parent_group_id: 108, child_group_id: 104, child_order: 1, expires_at: 2019-05-30 11:00:00}
	permissions_generated:
		- {group_id: 102, item_id: 1, can_view_generated: info}
		- {group_id: 101, item_id: 1, can_view_generated: none}
		- {group_id: 105, item_id: 3, can_view_generated: content}
		- {group_id: 107, item_id: 1, can_view_generated: none}
		- {group_id: 108, item_id: 1, can_view_generated: solution}
	attempts:
		- {group_id: 101, item_id: 1, order: 1}
		- {group_id: 102, item_id: 1, order: 1}
		- {group_id: 103, item_id: 1, order: 1}
		- {group_id: 104, item_id: 1, order: 1}
		- {group_id: 105, item_id: 1, order: 1}
		- {group_id: 106, item_id: 1, order: 1}
		- {group_id: 107, item_id: 1, order: 1}
		- {group_id: 108, item_id: 1, order: 1}
		- {group_id: 101, item_id: 2, order: 1}
		- {group_id: 102, item_id: 2, order: 1}
		- {group_id: 103, item_id: 2, order: 1}
		- {group_id: 104, item_id: 2, order: 1}
		- {group_id: 105, item_id: 2, order: 1}
		- {group_id: 106, item_id: 2, order: 1}
		- {group_id: 107, item_id: 2, order: 1}
		- {group_id: 108, item_id: 2, order: 1}
		- {group_id: 101, item_id: 3, order: 1}
		- {group_id: 102, item_id: 3, order: 1}
		- {group_id: 103, item_id: 3, order: 1}
		- {group_id: 104, item_id: 3, order: 1}
		- {group_id: 105, item_id: 3, order: 1}
		- {group_id: 106, item_id: 3, order: 1}
		- {group_id: 107, item_id: 3, order: 1}
		- {group_id: 108, item_id: 3, order: 1}`

func TestGroupGroupStore_TriggerAfterInsert_MarksAttemptsAsChanged(t *testing.T) {
	for _, test := range []struct {
		name            string
		parentGroupID   int64
		childGroupID    int64
		expiresAt       string
		expectedChanged []groupItemPair
	}{
		{
			name:            "group joins another group",
			parentGroupID:   103,
			childGroupID:    104,
			expiresAt:       "9999-12-31 23:59:59",
			expectedChanged: []groupItemPair{{104, 2}, {104, 3}, {105, 2}, {105, 3}},
		},
		{
			name:            "group joins a group, but the relation is expired",
			parentGroupID:   103,
			childGroupID:    104,
			expiresAt:       "2019-05-30 11:00:00",
			expectedChanged: []groupItemPair{},
		},
		{
			name:            "group having no attempts joins another group",
			parentGroupID:   105,
			childGroupID:    103,
			expiresAt:       "9999-12-31 23:59:59",
			expectedChanged: []groupItemPair{},
		},
		{
			name:            "parent group has no permissions on ancestor items",
			parentGroupID:   106,
			childGroupID:    104,
			expiresAt:       "9999-12-31 23:59:59",
			expectedChanged: []groupItemPair{},
		},
		{
			name:            "no new visible item ancestors after joining a group",
			parentGroupID:   105,
			childGroupID:    102,
			expiresAt:       "9999-12-31 23:59:59",
			expectedChanged: []groupItemPair{},
		},
	} {
		test := test
		t.Run(test.name, func(t *testing.T) {
			db := testhelpers.SetupDBWithFixtureString(groupGroupMarksAttemptsAsChangedFixture)
			defer func() { _ = db.Close() }()

			dataStore := database.NewDataStore(db)
			assert.NoError(t, dataStore.InTransaction(func(store *database.DataStore) error {
				store.GroupGroups().CreateNewAncestors()
				return nil
			}))
			assert.NoError(t, dataStore.GroupGroups().InsertMap(map[string]interface{}{
				"parent_group_id": test.parentGroupID, "child_group_id": test.childGroupID, "expires_at": test.expiresAt, "child_order": 1,
			}))

			assertAttemptsMarkedAsChanged(t, dataStore, test.expectedChanged)
		})
	}
}

func TestGroupGroupStore_TriggerBeforeUpdate_MarksAttemptsAsChanged(t *testing.T) {
	for _, test := range []struct {
		name            string
		parentGroupID   int64
		childGroupID    int64
		expiresAt       string
		expectedChanged []groupItemPair
		doNotSetExpired bool
		noChanges       bool
	}{
		{
			name:            "restore an expired relation",
			parentGroupID:   103,
			childGroupID:    104,
			expiresAt:       "9999-12-31 23:59:59",
			expectedChanged: []groupItemPair{{104, 2}, {104, 3}, {105, 2}, {105, 3}},
		},
		{
			name:            "expire the relation",
			parentGroupID:   103,
			childGroupID:    104,
			doNotSetExpired: true,
			expiresAt:       "2019-05-30 11:00:00",
			expectedChanged: []groupItemPair{},
		},
		{
			name:            "group having no attempts joins another group",
			parentGroupID:   105,
			childGroupID:    103,
			expiresAt:       "9999-12-31 23:59:59",
			expectedChanged: []groupItemPair{},
		},
		{
			name:            "parent group has no permissions on ancestor items",
			parentGroupID:   106,
			childGroupID:    104,
			expiresAt:       "9999-12-31 23:59:59",
			expectedChanged: []groupItemPair{},
		},
		{
			name:            "no new visible item ancestors after joining a group",
			parentGroupID:   105,
			childGroupID:    102,
			expiresAt:       "9999-12-31 23:59:59",
			expectedChanged: []groupItemPair{},
		},
		{
			name:            "no changes",
			parentGroupID:   10,
			childGroupID:    102,
			doNotSetExpired: true,
			expiresAt:       "9999-12-31 23:59:59",
			expectedChanged: []groupItemPair{},
			noChanges:       true,
		},
	} {
		test := test
		t.Run(test.name, func(t *testing.T) {
			expiresAt := "2019-05-30 11:00:00"
			if test.doNotSetExpired {
				expiresAt = maxDateTime
			}
			db := testhelpers.SetupDBWithFixtureString(
				fmt.Sprintf("groups_groups: [{parent_group_id: %d, child_group_id: %d, expires_at: %s}]",
					test.parentGroupID, test.childGroupID, expiresAt),
				groupGroupMarksAttemptsAsChangedFixture)
			defer func() { _ = db.Close() }()

			dataStore := database.NewDataStore(db)
			assert.NoError(t, dataStore.InTransaction(func(store *database.DataStore) error {
				store.GroupGroups().CreateNewAncestors()
				return nil
			}))
			groupGroupStore := dataStore.GroupGroups()
			result := groupGroupStore.Where("parent_group_id = ?", test.parentGroupID).
				Where("child_group_id = ?", test.childGroupID).
				UpdateColumn(map[string]interface{}{
					"expires_at": test.expiresAt,
				})
			assert.NoError(t, result.Error())
			if test.noChanges {
				assert.Zero(t, result.RowsAffected())
			} else {
				assert.Equal(t, int64(1), result.RowsAffected())
			}
			assertAttemptsMarkedAsChanged(t, dataStore, test.expectedChanged)
		})
	}
}

func TestGroupGroupStore_TriggerBeforeUpdate_RefusesToModifyParentGroupIDOrChildGroupID(t *testing.T) {
	db := testhelpers.SetupDBWithFixtureString(`
		groups_groups: [{parent_group_id: 1, child_group_id: 2}]
	`)
	defer func() { _ = db.Close() }()

	const expectedErrorMessage = "Error 1644: Unable to change immutable " +
		"groups_groups.parent_group_id and/or groups_groups.child_group_id"

	groupGroupStore := database.NewDataStore(db).GroupGroups()
	result := groupGroupStore.Where("parent_group_id = 1 AND child_group_id = 2").
		UpdateColumn("parent_group_id", 3)
	assert.EqualError(t, result.Error(), expectedErrorMessage)
	result = groupGroupStore.Where("parent_group_id = 1 AND child_group_id = 2").
		UpdateColumn("child_group_id", 3)
	assert.EqualError(t, result.Error(), expectedErrorMessage)
}

type groupItemPair struct {
	GroupID int64
	ItemID  int64
}

func assertAttemptsMarkedAsChanged(t *testing.T, dataStore *database.DataStore, expectedChanged []groupItemPair) {
	type attemptRow struct {
		GroupID                int64
		ItemID                 int64
		ResultPropagationState string
	}
	var attempts []attemptRow
	assert.NoError(t, dataStore.Attempts().Select("group_id, item_id, result_propagation_state").
		Order("group_id, item_id").Scan(&attempts).Error())

	expectedChangedPairsMap := make(map[groupItemPair]bool, len(expectedChanged))
	for _, attempt := range expectedChanged {
		expectedChangedPairsMap[attempt] = true
	}
	for _, dbAttempt := range attempts {
		expectedState := done
		if expectedChangedPairsMap[groupItemPair{GroupID: dbAttempt.GroupID, ItemID: dbAttempt.ItemID}] {
			expectedState = "changed"
		}
		assert.Equal(t, expectedState, dbAttempt.ResultPropagationState,
			"Wrong result propagation state for attempt with group_id=%d, item_id=%d",
			dbAttempt.GroupID, dbAttempt.ItemID)
		delete(expectedChangedPairsMap, groupItemPair{GroupID: dbAttempt.GroupID, ItemID: dbAttempt.ItemID})
	}
	assert.Empty(t, expectedChangedPairsMap, "Cannot find attempts that should be marked as 'changed': %#v", expectedChangedPairsMap)
}
