//go:build !unit

package database_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/golang"
	"github.com/France-ioi/AlgoreaBackend/v2/testhelpers"
	"github.com/France-ioi/AlgoreaBackend/v2/testhelpers/testoutput"
)

func TestGroupGroupStore_DeleteRelation(t *testing.T) {
	tests := []struct {
		name                         string
		fixture                      string
		shouldDeleteOrphans          bool
		remainingGroupIDs            []int64
		remainingGroupsGroups        []map[string]interface{}
		remainingGroupsAncestors     []map[string]interface{}
		expectedGrantedPermissions   []grantedPermission
		expectedGeneratedPermissions []permissionsGeneratedResultRow
	}{
		{
			name:                "can produce orphans",
			fixture:             "orphan",
			shouldDeleteOrphans: false,
			remainingGroupIDs:   []int64{1, 2, 3, 4, 5},
			remainingGroupsGroups: []map[string]interface{}{
				{"parent_group_id": int64(2), "child_group_id": int64(5)},
			},
			remainingGroupsAncestors: []map[string]interface{}{
				{"ancestor_group_id": int64(1), "child_group_id": int64(1)},
				{"ancestor_group_id": int64(2), "child_group_id": int64(2)},
				{"ancestor_group_id": int64(2), "child_group_id": int64(5)},
				{"ancestor_group_id": int64(3), "child_group_id": int64(3)},
				{"ancestor_group_id": int64(4), "child_group_id": int64(4)},
				{"ancestor_group_id": int64(5), "child_group_id": int64(5)},
				{"ancestor_group_id": int64(111), "child_group_id": int64(111)},
			},
			expectedGrantedPermissions: []grantedPermission{
				{GroupID: 1, ItemID: 1, SourceGroupID: 1, Origin: "group_membership", CanView: "none"},
				{GroupID: 2, ItemID: 1, SourceGroupID: 2, Origin: "group_membership", CanView: "none"},
				{GroupID: 3, ItemID: 1, SourceGroupID: 3, Origin: "group_membership", CanView: "none"},
				{GroupID: 4, ItemID: 1, SourceGroupID: 4, Origin: "group_membership", CanView: "none"},
				{GroupID: 5, ItemID: 1, SourceGroupID: 5, Origin: "group_membership", CanView: "none"},
			},
			expectedGeneratedPermissions: []permissionsGeneratedResultRow{
				{GroupID: 1, ItemID: 1, CanViewGenerated: "none"},
				{GroupID: 2, ItemID: 1, CanViewGenerated: "none"},
				{GroupID: 3, ItemID: 1, CanViewGenerated: "none"},
				{GroupID: 4, ItemID: 1, CanViewGenerated: "none"},
				{GroupID: 5, ItemID: 1, CanViewGenerated: "none"},
			},
		},
		{
			name:                "no orphans",
			fixture:             "no_orphans",
			shouldDeleteOrphans: false,
			remainingGroupIDs:   []int64{1, 2, 3},
			remainingGroupsGroups: []map[string]interface{}{
				{"parent_group_id": int64(3), "child_group_id": int64(2)},
			},
			remainingGroupsAncestors: []map[string]interface{}{
				{"ancestor_group_id": int64(1), "child_group_id": int64(1)},
				{"ancestor_group_id": int64(2), "child_group_id": int64(2)},
				{"ancestor_group_id": int64(3), "child_group_id": int64(2)},
				{"ancestor_group_id": int64(3), "child_group_id": int64(3)},
				{"ancestor_group_id": int64(111), "child_group_id": int64(111)},
			},
			expectedGrantedPermissions: []grantedPermission{
				{GroupID: 1, ItemID: 1, SourceGroupID: 1, Origin: "group_membership", CanView: "none"},
				{GroupID: 2, ItemID: 1, SourceGroupID: 1, Origin: "item_unlocking", CanView: "content"},
				{GroupID: 2, ItemID: 1, SourceGroupID: 2, Origin: "group_membership", CanView: "info"},
				{GroupID: 3, ItemID: 1, SourceGroupID: 1, Origin: "group_membership", CanView: "info"},
				{GroupID: 3, ItemID: 1, SourceGroupID: 3, Origin: "group_membership", CanView: "none"},
			},
			expectedGeneratedPermissions: []permissionsGeneratedResultRow{
				{GroupID: 1, ItemID: 1, CanViewGenerated: "none"},
				{GroupID: 2, ItemID: 1, CanViewGenerated: "content"},
				{GroupID: 3, ItemID: 1, CanViewGenerated: "info"},
			},
		},
		{
			name:                  "deletes orphans",
			fixture:               "orphan",
			shouldDeleteOrphans:   true,
			remainingGroupIDs:     []int64{1, 3, 4},
			remainingGroupsGroups: nil,
			remainingGroupsAncestors: []map[string]interface{}{
				{"ancestor_group_id": int64(1), "child_group_id": int64(1)},
				{"ancestor_group_id": int64(3), "child_group_id": int64(3)},
				{"ancestor_group_id": int64(4), "child_group_id": int64(4)},
				{"ancestor_group_id": int64(111), "child_group_id": int64(111)},
			},
			expectedGrantedPermissions: []grantedPermission{
				{GroupID: 1, ItemID: 1, SourceGroupID: 1, Origin: "group_membership", CanView: "none"},
				{GroupID: 3, ItemID: 1, SourceGroupID: 3, Origin: "group_membership", CanView: "none"},
				{GroupID: 4, ItemID: 1, SourceGroupID: 4, Origin: "group_membership", CanView: "none"},
			},
			expectedGeneratedPermissions: []permissionsGeneratedResultRow{
				{GroupID: 1, ItemID: 1, CanViewGenerated: "none"},
				{GroupID: 3, ItemID: 1, CanViewGenerated: "none"},
				{GroupID: 4, ItemID: 1, CanViewGenerated: "none"},
			},
		},
		{
			name:                "deletes only orphans",
			fixture:             "deletes_only_orphans",
			shouldDeleteOrphans: true,
			remainingGroupIDs:   []int64{1, 3, 5, 10},
			remainingGroupsGroups: []map[string]interface{}{
				{"parent_group_id": int64(10), "child_group_id": int64(3)},
				{"parent_group_id": int64(10), "child_group_id": int64(5)},
			},
			remainingGroupsAncestors: []map[string]interface{}{
				{"ancestor_group_id": int64(1), "child_group_id": int64(1)},
				{"ancestor_group_id": int64(3), "child_group_id": int64(3)},
				{"ancestor_group_id": int64(5), "child_group_id": int64(5)},
				{"ancestor_group_id": int64(10), "child_group_id": int64(3)},
				{"ancestor_group_id": int64(10), "child_group_id": int64(5)},
				{"ancestor_group_id": int64(10), "child_group_id": int64(10)},
				{"ancestor_group_id": int64(111), "child_group_id": int64(111)},
			},
			expectedGrantedPermissions: []grantedPermission{
				{GroupID: 1, ItemID: 1, SourceGroupID: 1, Origin: "group_membership", CanView: "none"},
				{GroupID: 3, ItemID: 1, SourceGroupID: 3, Origin: "group_membership", CanView: "none"},
				{GroupID: 5, ItemID: 1, SourceGroupID: 5, Origin: "group_membership", CanView: "none"},
				{GroupID: 10, ItemID: 1, SourceGroupID: 10, Origin: "group_membership", CanView: "none"},
			},
			expectedGeneratedPermissions: []permissionsGeneratedResultRow{
				{GroupID: 1, ItemID: 1, CanViewGenerated: "none"},
				{GroupID: 3, ItemID: 1, CanViewGenerated: "none"},
				{GroupID: 5, ItemID: 1, CanViewGenerated: "none"},
				{GroupID: 10, ItemID: 1, CanViewGenerated: "none"},
			},
		},
		{
			name:                  "deletes only orphans of accepted types",
			fixture:               "deletes_only_orphans_of_accepted_types",
			shouldDeleteOrphans:   true,
			remainingGroupIDs:     []int64{1, 8, 9, 11},
			remainingGroupsGroups: nil,
			remainingGroupsAncestors: []map[string]interface{}{
				{"ancestor_group_id": int64(1), "child_group_id": int64(1)},
				{"ancestor_group_id": int64(8), "child_group_id": int64(8)},
				{"ancestor_group_id": int64(9), "child_group_id": int64(9)},
				{"ancestor_group_id": int64(11), "child_group_id": int64(11)},
				{"ancestor_group_id": int64(111), "child_group_id": int64(111)},
			},
			expectedGrantedPermissions: []grantedPermission{
				{GroupID: 1, ItemID: 1, SourceGroupID: 1, Origin: "group_membership", CanView: "none"},
				{GroupID: 8, ItemID: 1, SourceGroupID: 8, Origin: "group_membership", CanView: "none"},
				{GroupID: 9, ItemID: 1, SourceGroupID: 9, Origin: "group_membership", CanView: "none"},
				{GroupID: 11, ItemID: 1, SourceGroupID: 11, Origin: "group_membership", CanView: "none"},
			},
			expectedGeneratedPermissions: []permissionsGeneratedResultRow{
				{GroupID: 1, ItemID: 1, CanViewGenerated: "none"},
				{GroupID: 8, ItemID: 1, CanViewGenerated: "none"},
				{GroupID: 9, ItemID: 1, CanViewGenerated: "none"},
				{GroupID: 11, ItemID: 1, CanViewGenerated: "none"},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			testoutput.SuppressIfPasses(t)

			db := testhelpers.SetupDBWithFixture("group_group_store/delete_relation/" + tt.fixture)
			defer func() { _ = db.Close() }()
			dataStore := database.NewDataStore(db)
			assert.NoError(t, dataStore.Table("groups_propagate").UpdateColumn("ancestors_computation_state", "done").Error())

			assert.NoError(t, dataStore.InTransaction(func(s *database.DataStore) error {
				s.SchedulePermissionsPropagation()
				return nil
			}))
			err := dataStore.InTransaction(func(s *database.DataStore) error {
				return s.GroupGroups().DeleteRelation(1, 2, tt.shouldDeleteOrphans)
			})
			require.NoError(t, err)
			assertGroupRelations(t, dataStore, tt.remainingGroupIDs, tt.remainingGroupsGroups, tt.remainingGroupsAncestors)
			assertGroupLinkedObjects(t, dataStore, tt.remainingGroupIDs, tt.expectedGrantedPermissions, tt.expectedGeneratedPermissions)
		})
	}
}

func assertGroupRelations(t *testing.T, dataStore *database.DataStore,
	remainingGroupIDs []int64, remainingGroupsGroups, remainingGroupsAncestors []map[string]interface{},
) {
	t.Helper()

	var rows []map[string]interface{}
	var ids []int64

	totalRemainingGroupIDs := make([]int64, 0, len(remainingGroupIDs)+1)
	totalRemainingGroupIDs = append(totalRemainingGroupIDs, remainingGroupIDs...)
	totalRemainingGroupIDs = append(totalRemainingGroupIDs, int64(111))
	require.NoError(t, dataStore.Groups().Order("id").Pluck("id", &ids).Error())
	assert.Equal(t, totalRemainingGroupIDs, ids)
	require.NoError(t, dataStore.ActiveGroupGroups().Select("parent_group_id, child_group_id").Order("parent_group_id, child_group_id").
		ScanIntoSliceOfMaps(&rows).Error())
	assert.Equal(t, remainingGroupsGroups, rows)
	require.NoError(t, dataStore.ActiveGroupAncestors().Select("ancestor_group_id, child_group_id").Order("ancestor_group_id, child_group_id").
		ScanIntoSliceOfMaps(&rows).Error())
	assert.Equal(t, remainingGroupsAncestors, rows)
	var count int64
	require.NoError(t, dataStore.Table("groups_propagate").Where("ancestors_computation_state != 'done'").Count(&count).Error())
	assert.Zero(t, count)
}

type grantedPermission struct {
	GroupID       int64
	ItemID        int64
	SourceGroupID int64
	Origin        string
	CanView       string
}

func assertGroupLinkedObjects(t *testing.T, dataStore *database.DataStore, remainingGroupIDs []int64,
	expectedGrantedPermissions []grantedPermission, expectedGeneratedPermissions []permissionsGeneratedResultRow,
) {
	t.Helper()

	var ids []int64
	require.NoError(t, dataStore.Table("filters").Order("group_id").
		Pluck("group_id", &ids).Error())
	assert.Equal(t, remainingGroupIDs, ids)
	require.NoError(t, dataStore.Attempts().Order("participant_id").
		Pluck("participant_id", &ids).Error())
	assert.Equal(t, remainingGroupIDs, ids)
	require.NoError(t, dataStore.Results().Order("participant_id").
		Pluck("participant_id", &ids).Error())
	assert.Equal(t, remainingGroupIDs, ids)

	var grantedPermissions []grantedPermission
	require.NoError(t, dataStore.PermissionsGranted().Order("group_id, item_id, source_group_id, origin").
		Select("group_id, item_id, source_group_id, origin, can_view").Scan(&grantedPermissions).Error())
	assert.Equal(t, expectedGrantedPermissions, grantedPermissions)

	var generatedPermissions []permissionsGeneratedResultRow
	require.NoError(t, dataStore.Permissions().Order("group_id, item_id").
		Select("group_id, item_id, can_view_generated").Scan(&generatedPermissions).Error())
	assert.Equal(t, expectedGeneratedPermissions, generatedPermissions)

	assert.Equal(t, remainingGroupIDs, ids)
	require.NoError(t, dataStore.Permissions().Order("group_id").
		Pluck("group_id", &ids).Error())
	assert.Equal(t, remainingGroupIDs, ids)

	var cnt int64
	require.NoError(t, dataStore.Table("groups_propagate").
		Where("ancestors_computation_state <> 'done'").Count(&cnt).Error())
	assert.Zero(t, cnt)

	require.NoError(t, dataStore.Table("permissions_propagate").Count(&cnt).Error())
	assert.Zero(t, cnt)
}

const (
	done                                   = "done"
	groupGroupMarksResultsAsChangedFixture = `
	items:
		- {id: 1, default_language_tag: fr}
		- {id: 2, default_language_tag: fr}
		- {id: 3, default_language_tag: fr}
	items_ancestors:
		- {ancestor_item_id: 1, child_item_id: 2}
		- {ancestor_item_id: 1, child_item_id: 3}
		- {ancestor_item_id: 2, child_item_id: 3}
	groups: [{id: 101}, {id: 102}, {id: 103}, {id: 104}, {id: 105}, {id: 106}, {id: 107}, {id: 108}, {id: 109, type: Team}]
	groups_groups:
		- {parent_group_id: 101, child_group_id: 102}
		- {parent_group_id: 101, child_group_id: 103}
		- {parent_group_id: 101, child_group_id: 104}
		- {parent_group_id: 101, child_group_id: 105}
		- {parent_group_id: 101, child_group_id: 106}
		- {parent_group_id: 101, child_group_id: 109}
		- {parent_group_id: 102, child_group_id: 103}
		- {parent_group_id: 102, child_group_id: 109}
		- {parent_group_id: 104, child_group_id: 105}
		- {parent_group_id: 107, child_group_id: 105}
		- {parent_group_id: 108, child_group_id: 104, expires_at: 2019-05-30 11:00:00}
	permissions_generated:
		- {group_id: 102, item_id: 1, can_view_generated: info}
		- {group_id: 101, item_id: 1, can_view_generated: none}
		- {group_id: 105, item_id: 3, can_view_generated: content}
		- {group_id: 107, item_id: 1, can_view_generated: none}
		- {group_id: 108, item_id: 1, can_view_generated: solution}
	attempts:
		- {id: 1, participant_id: 101}
		- {id: 1, participant_id: 102}
		- {id: 1, participant_id: 103}
		- {id: 1, participant_id: 104}
		- {id: 1, participant_id: 105}
		- {id: 1, participant_id: 106}
		- {id: 1, participant_id: 107}
		- {id: 1, participant_id: 108}
		- {id: 1, participant_id: 109}
	results:
		- {attempt_id: 1, participant_id: 101, item_id: 1}
		- {attempt_id: 1, participant_id: 102, item_id: 1}
		- {attempt_id: 1, participant_id: 103, item_id: 1}
		- {attempt_id: 1, participant_id: 104, item_id: 1}
		- {attempt_id: 1, participant_id: 105, item_id: 1}
		- {attempt_id: 1, participant_id: 106, item_id: 1}
		- {attempt_id: 1, participant_id: 107, item_id: 1}
		- {attempt_id: 1, participant_id: 108, item_id: 1}
		- {attempt_id: 1, participant_id: 109, item_id: 1}
		- {attempt_id: 1, participant_id: 101, item_id: 2}
		- {attempt_id: 1, participant_id: 102, item_id: 2}
		- {attempt_id: 1, participant_id: 103, item_id: 2}
		- {attempt_id: 1, participant_id: 104, item_id: 2}
		- {attempt_id: 1, participant_id: 105, item_id: 2}
		- {attempt_id: 1, participant_id: 106, item_id: 2}
		- {attempt_id: 1, participant_id: 107, item_id: 2}
		- {attempt_id: 1, participant_id: 108, item_id: 2}
		- {attempt_id: 1, participant_id: 109, item_id: 2}
		- {attempt_id: 1, participant_id: 101, item_id: 3}
		- {attempt_id: 1, participant_id: 102, item_id: 3}
		- {attempt_id: 1, participant_id: 103, item_id: 3}
		- {attempt_id: 1, participant_id: 104, item_id: 3}
		- {attempt_id: 1, participant_id: 105, item_id: 3}
		- {attempt_id: 1, participant_id: 106, item_id: 3}
		- {attempt_id: 1, participant_id: 107, item_id: 3}
		- {attempt_id: 1, participant_id: 108, item_id: 3}
		- {attempt_id: 1, participant_id: 109, item_id: 3}`
)

func TestGroupGroupStore_TriggerAfterInsert_MarksResultsAsChanged(t *testing.T) {
	for _, test := range []struct {
		name            string
		parentGroupID   int64
		childGroupID    int64
		expiresAt       string
		expectedChanged []resultPrimaryKeyAndState
	}{
		{
			name:          "group joins another group",
			parentGroupID: 103,
			childGroupID:  104,
			expiresAt:     "9999-12-31 23:59:59",
			expectedChanged: []resultPrimaryKeyAndState{
				{ResultPrimaryKey: ResultPrimaryKey{104, 1, 2}},
				{ResultPrimaryKey: ResultPrimaryKey{104, 1, 3}},
				{ResultPrimaryKey: ResultPrimaryKey{105, 1, 2}},
				{ResultPrimaryKey: ResultPrimaryKey{105, 1, 3}},
			},
		},
		{
			name:            "group joins a group, but the relation is expired",
			parentGroupID:   103,
			childGroupID:    104,
			expiresAt:       "2019-05-30 11:00:00",
			expectedChanged: []resultPrimaryKeyAndState{},
		},
		{
			name:            "group joins a group, but the relation is a team membership",
			parentGroupID:   109,
			childGroupID:    104,
			expiresAt:       "9999-12-31 23:59:59",
			expectedChanged: []resultPrimaryKeyAndState{},
		},
		{
			name:            "group having no results joins another group",
			parentGroupID:   105,
			childGroupID:    103,
			expiresAt:       "9999-12-31 23:59:59",
			expectedChanged: []resultPrimaryKeyAndState{},
		},
		{
			name:            "parent group has no permissions on ancestor items",
			parentGroupID:   106,
			childGroupID:    104,
			expiresAt:       "9999-12-31 23:59:59",
			expectedChanged: []resultPrimaryKeyAndState{},
		},
		{
			name:            "no new visible item ancestors after joining a group",
			parentGroupID:   105,
			childGroupID:    102,
			expiresAt:       "9999-12-31 23:59:59",
			expectedChanged: []resultPrimaryKeyAndState{},
		},
	} {
		test := test
		t.Run(test.name, func(t *testing.T) {
			testoutput.SuppressIfPasses(t)

			db := testhelpers.SetupDBWithFixtureString(groupGroupMarksResultsAsChangedFixture)
			defer func() { _ = db.Close() }()

			dataStore := database.NewDataStore(db)
			assert.NoError(t, dataStore.InTransaction(func(store *database.DataStore) error {
				return store.GroupGroups().CreateNewAncestors()
			}))
			assert.NoError(t, dataStore.GroupGroups().InsertMap(map[string]interface{}{
				"parent_group_id": test.parentGroupID, "child_group_id": test.childGroupID, "expires_at": test.expiresAt,
			}))

			assertResultsMarkedAsChanged(t, dataStore, test.expectedChanged)
		})
	}
}

func TestGroupGroupStore_TriggerAfterUpdate_MarksResultsAsChanged(t *testing.T) {
	for _, test := range []struct {
		name            string
		parentGroupID   int64
		childGroupID    int64
		expiresAt       string
		expectedChanged []resultPrimaryKeyAndState
		doNotSetExpired bool
		noChanges       bool
	}{
		{
			name:          "restore an expired relation",
			parentGroupID: 103,
			childGroupID:  104,
			expiresAt:     "9999-12-31 23:59:59",
			expectedChanged: []resultPrimaryKeyAndState{
				{ResultPrimaryKey: ResultPrimaryKey{104, 1, 2}},
				{ResultPrimaryKey: ResultPrimaryKey{104, 1, 3}},
				{ResultPrimaryKey: ResultPrimaryKey{105, 1, 2}},
				{ResultPrimaryKey: ResultPrimaryKey{105, 1, 3}},
			},
		},
		{
			name:            "restore an expired relation for a team membership",
			parentGroupID:   109,
			childGroupID:    104,
			expiresAt:       "9999-12-31 23:59:59",
			expectedChanged: []resultPrimaryKeyAndState{},
		},
		{
			name:            "expire the relation",
			parentGroupID:   103,
			childGroupID:    104,
			doNotSetExpired: true,
			expiresAt:       "2019-05-30 11:00:00",
			expectedChanged: []resultPrimaryKeyAndState{},
		},
		{
			name:            "group having no results joins another group",
			parentGroupID:   105,
			childGroupID:    103,
			expiresAt:       "9999-12-31 23:59:59",
			expectedChanged: []resultPrimaryKeyAndState{},
		},
		{
			name:            "parent group has no permissions on ancestor items",
			parentGroupID:   106,
			childGroupID:    104,
			expiresAt:       "9999-12-31 23:59:59",
			expectedChanged: []resultPrimaryKeyAndState{},
		},
		{
			name:            "no new visible item ancestors after joining a group",
			parentGroupID:   105,
			childGroupID:    102,
			expiresAt:       "9999-12-31 23:59:59",
			expectedChanged: []resultPrimaryKeyAndState{},
		},
		{
			name:            "no changes",
			parentGroupID:   105,
			childGroupID:    102,
			doNotSetExpired: true,
			expiresAt:       "9999-12-31 23:59:59",
			expectedChanged: []resultPrimaryKeyAndState{},
			noChanges:       true,
		},
	} {
		test := test
		t.Run(test.name, func(t *testing.T) {
			testoutput.SuppressIfPasses(t)

			expiresAt := "2019-05-30 11:00:00"
			if test.doNotSetExpired {
				expiresAt = maxDateTime
			}
			db := testhelpers.SetupDBWithFixtureString(
				groupGroupMarksResultsAsChangedFixture,
				fmt.Sprintf("groups_groups: [{parent_group_id: %d, child_group_id: %d, expires_at: %s}]",
					test.parentGroupID, test.childGroupID, expiresAt))
			defer func() { _ = db.Close() }()

			dataStore := database.NewDataStore(db)
			require.NoError(t, dataStore.InTransaction(func(store *database.DataStore) error {
				return store.GroupGroups().CreateNewAncestors()
			}))
			groupGroupStore := dataStore.GroupGroups()
			result := groupGroupStore.Where("parent_group_id = ?", test.parentGroupID).
				Where("child_group_id = ?", test.childGroupID).
				UpdateColumn(map[string]interface{}{
					"expires_at": test.expiresAt,
				})
			require.NoError(t, result.Error())
			if test.noChanges {
				assert.Zero(t, result.RowsAffected())
			} else {
				assert.Equal(t, int64(1), result.RowsAffected())
			}
			assertResultsMarkedAsChanged(t, dataStore, test.expectedChanged)
		})
	}
}

func TestGroupGroupStore_TriggerBeforeUpdate_RefusesToModifyParentGroupIDOrChildGroupIDOrIsTeamMembership(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db := testhelpers.SetupDBWithFixtureString(`
		groups: [{id: 1}, {id: 2}]
		groups_groups: [{parent_group_id: 1, child_group_id: 2}]
	`)
	defer func() { _ = db.Close() }()

	const expectedErrorMessage = "Error 1644 (45000): Unable to change immutable columns of groups_groups " +
		"(parent_group_id/child_group_id/is_team_membership)"

	groupGroupStore := database.NewDataStore(db).GroupGroups()
	result := groupGroupStore.Where("parent_group_id = 1 AND child_group_id = 2").
		UpdateColumn("parent_group_id", 3)
	require.EqualError(t, result.Error(), expectedErrorMessage)
	result = groupGroupStore.Where("parent_group_id = 1 AND child_group_id = 2").
		UpdateColumn("child_group_id", 3)
	require.EqualError(t, result.Error(), expectedErrorMessage)
	result = groupGroupStore.Where("parent_group_id = 1 AND child_group_id = 2").
		UpdateColumn("is_team_membership", 1)
	require.EqualError(t, result.Error(), expectedErrorMessage)
}

type ResultPrimaryKey struct {
	ParticipantID int64
	AttemptID     int64
	ItemID        int64
}

type resultPrimaryKeyAndState struct {
	ResultPrimaryKey
	State string
}

func assertResultsMarkedAsChanged(t *testing.T, dataStore *database.DataStore, expectedChanged []resultPrimaryKeyAndState) {
	t.Helper()

	var results []resultPrimaryKeyAndState
	queryResultsAndStatesForTests(t, dataStore.Results(), &results, "")

	expectedChangedResultsMap := make(map[ResultPrimaryKey]string, len(expectedChanged))
	for _, result := range expectedChanged {
		expectedChangedResultsMap[result.ResultPrimaryKey] = golang.IfElse(result.State == "", "to_be_propagated", result.State)
	}
	for _, dbResult := range results {
		expectedState := done
		if state, ok := expectedChangedResultsMap[dbResult.ResultPrimaryKey]; ok {
			expectedState = state
		}
		assert.Equal(t, expectedState, dbResult.State,
			"Wrong result propagation state for result with participant_id=%d, attempt_id=%d, item_id=%d",
			dbResult.ParticipantID, dbResult.AttemptID, dbResult.ItemID)
		delete(expectedChangedResultsMap, dbResult.ResultPrimaryKey)
	}
	assert.Empty(t, expectedChangedResultsMap,
		"Cannot find results that should be marked as 'to_be_propagated': %#v", expectedChangedResultsMap)
}

func queryResultsAndStatesForTests(t *testing.T, resultStore *database.ResultStore, result interface{}, customColumns string) {
	t.Helper()

	columns := "participant_id, attempt_id, item_id, IFNULL(state, 'done') AS state"
	if customColumns != "" {
		columns += "," + customColumns
	}
	assert.NoError(t,
		resultStore.Select(columns).
			Joins("LEFT JOIN results_propagate USING(participant_id, attempt_id, item_id)").
			Union(resultStore.Select(columns).
				Joins("RIGHT JOIN results_propagate USING(participant_id, attempt_id, item_id)")).
			Order("participant_id, attempt_id, item_id").Scan(result).Error())
}

func TestGroupGroupStore_TriggerBeforeDelete(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db := testhelpers.SetupDBWithFixtureString(`
		groups: [{id: 1}, {id: 2}, {id: 3, type: Team}, {id: 4}]
		groups_groups: [{parent_group_id: 1, child_group_id: 2}, {parent_group_id: 3, child_group_id: 4}]`)
	defer func() { _ = db.Close() }()

	dataStore := database.NewDataStore(db)
	require.NoError(t, dataStore.InTransaction(func(store *database.DataStore) error {
		return store.GroupGroups().CreateNewAncestors()
	}))

	found, err := dataStore.Table("groups_propagate").Where("ancestors_computation_state != 'done'").HasRows()
	require.NoError(t, err)
	require.False(t, found)

	require.NoError(t, dataStore.GroupGroups().Delete("parent_group_id = 1 AND child_group_id = 2").Error())

	type id struct {
		ID int64
	}
	var marked []id
	require.NoError(t, dataStore.Table("groups_propagate").Where("ancestors_computation_state = 'todo'").Scan(&marked).Error())
	assert.Equal(t, []id{{2}}, marked)
	require.NoError(t, dataStore.Table("groups_propagate").
		Where("ancestors_computation_state = 'todo'").
		UpdateColumn("ancestors_computation_state", "done").Error())
	found, err = dataStore.Table("groups_propagate").Where("ancestors_computation_state != 'done'").HasRows()
	require.NoError(t, err)
	require.False(t, found)

	require.NoError(t, dataStore.GroupGroups().Delete("parent_group_id = 3 AND child_group_id = 4").Error())
	found, err = dataStore.Table("groups_propagate").Where("ancestors_computation_state != 'done'").HasRows()
	require.NoError(t, err)
	assert.False(t, found)
}
