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
				{"idGroupParent": "1", "idGroupChild": "2"},
			},
			remainingGroupsAncestors: []map[string]interface{}{
				{"idGroupAncestor": "1", "idGroupChild": "1"},
				{"idGroupAncestor": "1", "idGroupChild": "2"},
				{"idGroupAncestor": "2", "idGroupChild": "2"},
			},
		},
		{
			name:                "no orphans",
			fixture:             "no_orphans",
			shouldDeleteOrphans: false,
			remainingGroupIDs:   []int64{1, 2, 3},
			remainingGroupsGroups: []map[string]interface{}{
				{"idGroupParent": "3", "idGroupChild": "2"},
			},
			remainingGroupsAncestors: []map[string]interface{}{
				{"idGroupAncestor": "1", "idGroupChild": "1"},
				{"idGroupAncestor": "2", "idGroupChild": "2"},
				{"idGroupAncestor": "3", "idGroupChild": "2"},
				{"idGroupAncestor": "3", "idGroupChild": "3"},
			},
		},
		{
			name:                  "deletes orphans",
			fixture:               "orphan",
			shouldDeleteOrphans:   true,
			remainingGroupIDs:     []int64{1},
			remainingGroupsGroups: nil,
			remainingGroupsAncestors: []map[string]interface{}{
				{"idGroupAncestor": "1", "idGroupChild": "1"},
			},
		},
		{
			name:                "deletes only orphans",
			fixture:             "deletes_only_orphans",
			shouldDeleteOrphans: true,
			remainingGroupIDs:   []int64{1, 3, 5, 10},
			remainingGroupsGroups: []map[string]interface{}{
				{"idGroupParent": "10", "idGroupChild": "3"},
				{"idGroupParent": "10", "idGroupChild": "5"},
			},
			remainingGroupsAncestors: []map[string]interface{}{
				{"idGroupAncestor": "1", "idGroupChild": "1"},
				{"idGroupAncestor": "3", "idGroupChild": "3"},
				{"idGroupAncestor": "5", "idGroupChild": "5"},
				{"idGroupAncestor": "10", "idGroupChild": "3"},
				{"idGroupAncestor": "10", "idGroupChild": "5"},
				{"idGroupAncestor": "10", "idGroupChild": "10"},
			},
		},
		{
			name:                  "deletes only orphans of accepted types",
			fixture:               "deletes_only_orphans_of_accepted_types",
			shouldDeleteOrphans:   true,
			remainingGroupIDs:     []int64{1, 8, 9, 10, 11, 12},
			remainingGroupsGroups: nil,
			remainingGroupsAncestors: []map[string]interface{}{
				{"idGroupAncestor": "1", "idGroupChild": "1"},
				{"idGroupAncestor": "8", "idGroupChild": "8"},
				{"idGroupAncestor": "9", "idGroupChild": "9"},
				{"idGroupAncestor": "10", "idGroupChild": "10"},
				{"idGroupAncestor": "11", "idGroupChild": "11"},
				{"idGroupAncestor": "12", "idGroupChild": "12"},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			db := testhelpers.SetupDBWithFixture("group_group_store/delete_relation/" + tt.fixture)
			defer func() { _ = db.Close() }()
			dataStore := database.NewDataStore(db)
			assert.NoError(t, dataStore.Table("groups_propagate").UpdateColumn("sAncestorsComputationState", "done").Error())

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
	assert.NoError(t, dataStore.Groups().Order("ID").Pluck("ID", &ids).Error())
	assert.Equal(t, remainingGroupIDs, ids)
	assert.NoError(t, dataStore.GroupGroups().Select("idGroupParent, idGroupChild").Order("idGroupParent, idGroupChild").
		ScanIntoSliceOfMaps(&rows).Error())
	assert.Equal(t, remainingGroupsGroups, rows)
	assert.NoError(t, dataStore.GroupAncestors().Select("idGroupAncestor, idGroupChild").Order("idGroupAncestor, idGroupChild").
		ScanIntoSliceOfMaps(&rows).Error())
	assert.Equal(t, remainingGroupsAncestors, rows)
	var count int64
	assert.NoError(t, dataStore.Table("groups_propagate").Where("sAncestorsComputationState != 'done'").Count(&count).Error())
	assert.Zero(t, count)
}

func assertGroupLinkedObjects(t *testing.T, dataStore *database.DataStore, remainingGroupIDs []int64) {
	var ids []int64
	assert.NoError(t, dataStore.Table("filters").Order("idGroup").
		Pluck("idGroup", &ids).Error())
	assert.Equal(t, remainingGroupIDs, ids)
	assert.NoError(t, dataStore.GroupAttempts().Order("idGroup").
		Pluck("idGroup", &ids).Error())
	assert.Equal(t, remainingGroupIDs, ids)
	assert.NoError(t, dataStore.GroupItems().Order("idGroup").
		Pluck("idGroup", &ids).Error())
	assert.Equal(t, remainingGroupIDs, ids)
	assert.NoError(t, dataStore.Table("groups_items_propagate").Order("ID").
		Pluck("ID", &ids).Error())
	expectedGroupsItemsPropagateIDs := make([]int64, len(remainingGroupIDs))
	for index, id := range remainingGroupIDs {
		expectedGroupsItemsPropagateIDs[index] = id + 100
	}
	assert.Equal(t, expectedGroupsItemsPropagateIDs, ids)
	assert.NoError(t, dataStore.Table("groups_login_prefixes").Order("idGroup").
		Pluck("idGroup", &ids).Error())
	assert.Equal(t, remainingGroupIDs, ids)
	assert.NoError(t, dataStore.Table("groups_propagate").Order("ID").
		Pluck("ID", &ids).Error())
	assert.Equal(t, remainingGroupIDs, ids)
}
