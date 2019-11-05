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
			remainingGroupIDs:     []int64{1, 8, 9, 10, 11, 12},
			remainingGroupsGroups: nil,
			remainingGroupsAncestors: []map[string]interface{}{
				{"ancestor_group_id": "1", "child_group_id": "1"},
				{"ancestor_group_id": "8", "child_group_id": "8"},
				{"ancestor_group_id": "9", "child_group_id": "9"},
				{"ancestor_group_id": "10", "child_group_id": "10"},
				{"ancestor_group_id": "11", "child_group_id": "11"},
				{"ancestor_group_id": "12", "child_group_id": "12"},
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
	assert.NoError(t, dataStore.GroupAttempts().Order("group_id").
		Pluck("group_id", &ids).Error())
	assert.Equal(t, remainingGroupIDs, ids)
	assert.NoError(t, dataStore.PermissionsGranted().Order("group_id").
		Pluck("group_id", &ids).Error())
	assert.Equal(t, remainingGroupIDs, ids)
	assert.NoError(t, dataStore.PermissionsGenerated().Order("group_id").
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
