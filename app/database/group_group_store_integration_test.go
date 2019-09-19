// +build !unit

package database_test

import (
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
				{"group_parent_id": "1", "group_child_id": "2"},
			},
			remainingGroupsAncestors: []map[string]interface{}{
				{"group_ancestor_id": "1", "group_child_id": "1"},
				{"group_ancestor_id": "1", "group_child_id": "2"},
				{"group_ancestor_id": "2", "group_child_id": "2"},
			},
		},
		{
			name:                "no orphans",
			fixture:             "no_orphans",
			shouldDeleteOrphans: false,
			remainingGroupIDs:   []int64{1, 2, 3},
			remainingGroupsGroups: []map[string]interface{}{
				{"group_parent_id": "3", "group_child_id": "2"},
			},
			remainingGroupsAncestors: []map[string]interface{}{
				{"group_ancestor_id": "1", "group_child_id": "1"},
				{"group_ancestor_id": "2", "group_child_id": "2"},
				{"group_ancestor_id": "3", "group_child_id": "2"},
				{"group_ancestor_id": "3", "group_child_id": "3"},
			},
		},
		{
			name:                  "deletes orphans",
			fixture:               "orphan",
			shouldDeleteOrphans:   true,
			remainingGroupIDs:     []int64{1},
			remainingGroupsGroups: nil,
			remainingGroupsAncestors: []map[string]interface{}{
				{"group_ancestor_id": "1", "group_child_id": "1"},
			},
		},
		{
			name:                "deletes only orphans",
			fixture:             "deletes_only_orphans",
			shouldDeleteOrphans: true,
			remainingGroupIDs:   []int64{1, 3, 5, 10},
			remainingGroupsGroups: []map[string]interface{}{
				{"group_parent_id": "10", "group_child_id": "3"},
				{"group_parent_id": "10", "group_child_id": "5"},
			},
			remainingGroupsAncestors: []map[string]interface{}{
				{"group_ancestor_id": "1", "group_child_id": "1"},
				{"group_ancestor_id": "3", "group_child_id": "3"},
				{"group_ancestor_id": "5", "group_child_id": "5"},
				{"group_ancestor_id": "10", "group_child_id": "3"},
				{"group_ancestor_id": "10", "group_child_id": "5"},
				{"group_ancestor_id": "10", "group_child_id": "10"},
			},
		},
		{
			name:                  "deletes only orphans of accepted types",
			fixture:               "deletes_only_orphans_of_accepted_types",
			shouldDeleteOrphans:   true,
			remainingGroupIDs:     []int64{1, 8, 9, 10, 11, 12},
			remainingGroupsGroups: nil,
			remainingGroupsAncestors: []map[string]interface{}{
				{"group_ancestor_id": "1", "group_child_id": "1"},
				{"group_ancestor_id": "8", "group_child_id": "8"},
				{"group_ancestor_id": "9", "group_child_id": "9"},
				{"group_ancestor_id": "10", "group_child_id": "10"},
				{"group_ancestor_id": "11", "group_child_id": "11"},
				{"group_ancestor_id": "12", "group_child_id": "12"},
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
	assert.NoError(t, dataStore.Groups().Order("id").Pluck("id", &ids).Error())
	assert.Equal(t, remainingGroupIDs, ids)
	assert.NoError(t, dataStore.GroupGroups().Select("group_parent_id, group_child_id").Order("group_parent_id, group_child_id").
		ScanIntoSliceOfMaps(&rows).Error())
	assert.Equal(t, remainingGroupsGroups, rows)
	assert.NoError(t, dataStore.GroupAncestors().Select("group_ancestor_id, group_child_id").Order("group_ancestor_id, group_child_id").
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
	assert.NoError(t, dataStore.GroupAttempts().Order("group_id").
		Pluck("group_id", &ids).Error())
	assert.Equal(t, remainingGroupIDs, ids)
	assert.NoError(t, dataStore.GroupItems().Order("group_id").
		Pluck("group_id", &ids).Error())
	assert.Equal(t, remainingGroupIDs, ids)
	assert.NoError(t, dataStore.Table("groups_items_propagate").Order("id").
		Pluck("id", &ids).Error())
	expectedGroupsItemsPropagateIDs := make([]int64, len(remainingGroupIDs))
	for index, id := range remainingGroupIDs {
		expectedGroupsItemsPropagateIDs[index] = id + 100
	}
	assert.Equal(t, expectedGroupsItemsPropagateIDs, ids)
	assert.NoError(t, dataStore.Table("groups_login_prefixes").Order("group_id").
		Pluck("group_id", &ids).Error())
	assert.Equal(t, remainingGroupIDs, ids)
	assert.NoError(t, dataStore.Table("groups_propagate").Order("id").
		Pluck("id", &ids).Error())
	assert.Equal(t, remainingGroupIDs, ids)
}
