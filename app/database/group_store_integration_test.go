//go:build !unit

package database_test

import (
	"fmt"
	"reflect"
	"testing"

	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/testhelpers"
	"github.com/France-ioi/AlgoreaBackend/v2/testhelpers/testoutput"
)

func TestGroupStore_CreateNew(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	ctx := testhelpers.CreateTestContext()
	for _, test := range []struct {
		groupType            string
		shouldCreateAttempts bool
	}{
		{groupType: "Class", shouldCreateAttempts: false},
		{groupType: "Team", shouldCreateAttempts: true},
	} {
		test := test
		t.Run(test.groupType, func(t *testing.T) {
			testoutput.SuppressIfPasses(t)

			db := testhelpers.SetupDBWithFixtureString(ctx)
			defer func() { _ = db.Close() }()

			var newID int64
			var err error
			dataStore := database.NewDataStore(db)
			require.NoError(t, dataStore.InTransaction(func(store *database.DataStore) error {
				newID, err = store.Groups().CreateNew("Some group", test.groupType)
				return err
			}))
			assert.Positive(t, newID)
			type resultType struct {
				Name         string
				Type         string
				CreatedAtSet bool
			}
			var result resultType
			require.NoError(t, dataStore.Groups().ByID(newID).
				Select("name, type, ABS(TIMESTAMPDIFF(SECOND, created_at, NOW())) < 3 AS created_at_set").
				Take(&result).Error())
			assert.Equal(t, resultType{
				Name:         "Some group",
				Type:         test.groupType,
				CreatedAtSet: true,
			}, result)

			found, err := dataStore.GroupAncestors().
				Where("ancestor_group_id = ?", newID).
				Where("child_group_id = ?", newID).HasRows()
			require.NoError(t, err)
			assert.True(t, found)

			var attempts []map[string]interface{}
			require.NoError(t, dataStore.Attempts().
				Select(`
					participant_id, id, creator_id, parent_attempt_id, root_item_id,
					ABS(TIMESTAMPDIFF(SECOND, created_at, NOW())) < 3 AS created_at_set`).
				ScanIntoSliceOfMaps(&attempts).Error())
			var expectedAttempts []map[string]interface{}
			if test.shouldCreateAttempts {
				expectedAttempts = []map[string]interface{}{
					{
						"participant_id": newID, "id": int64(0), "creator_id": nil, "parent_attempt_id": nil,
						"root_item_id": nil, "created_at_set": int64(1),
					},
				}
			}
			assert.Equal(t, expectedAttempts, attempts)
		})
	}
}

func TestGroupStore_CreateNew_Duplicate(t *testing.T) {
	testoutput.SuppressIfPasses(t)
	db := testhelpers.SetupDBWithFixtureString(testhelpers.CreateTestContext(), `groups: [{id: 1, name: "Some group"}]`)
	defer func() { _ = db.Close() }()

	var nextID int64
	monkey.PatchInstanceMethod(reflect.TypeOf(&database.DataStore{}), "NewID", func(*database.DataStore) int64 {
		nextID++
		return nextID
	})
	defer monkey.UnpatchAll()

	var newID int64
	require.NoError(t, database.NewDataStore(db).InTransaction(func(store *database.DataStore) error {
		var err error
		newID, err = store.Groups().CreateNew("Some group", "Class")
		return err
	}))
	assert.Equal(t, int64(2), newID)
}

func TestGroupStore_CheckIfEntryConditionsStillSatisfiedForAllActiveParticipations(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	const mainFixture = `
		groups: [{id: 1}, {id: 2}, {id: 3}, {id: 4}, {id: 5}]
		groups_groups:
			- {parent_group_id: 1, child_group_id: 3}
			- {parent_group_id: 2, child_group_id: 3}
			- {parent_group_id: 2, child_group_id: 5}
		attempts: [{participant_id: 2, id: 100, root_item_id: 10}]
		results: [{participant_id: 2, attempt_id: 100, item_id: 10, started_at: 2019-05-30 11:00:00}]
	`

	ctx := testhelpers.CreateTestContext()

	type args struct {
		teamGroupID int64
		userID      int64
		isAddition  bool
	}
	tests := []struct {
		name    string
		fixture string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "addition breaks entry_min_admitted_members_ratio = All (can_enter_from)",
			fixture: `
				items: [{id: 10, default_language_tag: fr, entry_min_admitted_members_ratio: All, entry_max_team_size: 100}]
				permissions_granted:
					- {group_id: 1, item_id: 10, source_group_id: 1, can_enter_from: 2019-05-30 11:00:00, can_enter_until: 3019-05-30 11:00:00}
					- {group_id: 5, item_id: 10, source_group_id: 5, can_enter_from: 2019-05-30 11:00:00, can_enter_until: 3019-05-30 11:00:00}
					- {group_id: 4, item_id: 10, source_group_id: 4, can_enter_from: 2019-05-30 11:00:01, can_enter_until: 3019-05-30 11:00:00}`,
			args: args{teamGroupID: 2, userID: 4, isAddition: true},
		},
		{
			name: "addition breaks entry_min_admitted_members_ratio = All (can_enter_until)",
			fixture: `
				items: [{id: 10, default_language_tag: fr, entry_min_admitted_members_ratio: All, entry_max_team_size: 100}]
				permissions_granted:
					- {group_id: 1, item_id: 10, source_group_id: 1, can_enter_from: 2019-05-30 11:00:00, can_enter_until: 3019-05-30 11:00:00}
					- {group_id: 5, item_id: 10, source_group_id: 5, can_enter_from: 2019-05-30 11:00:00, can_enter_until: 3019-05-30 11:00:00}
					- {group_id: 4, item_id: 10, source_group_id: 4, can_enter_from: 2019-04-30 11:00:00, can_enter_until: 2019-05-30 11:00:00}`,
			args: args{teamGroupID: 2, userID: 4, isAddition: true},
		},
		{
			name: "addition breaks entry_min_admitted_members_ratio = Half (can_enter_from)",
			fixture: `
				items: [{id: 10, default_language_tag: fr, entry_min_admitted_members_ratio: Half, entry_max_team_size: 100}]
				permissions_granted:
					- {group_id: 1, item_id: 10, source_group_id: 1, can_enter_from: 2019-05-30 11:00:00, can_enter_until: 3019-05-30 11:00:00}
					- {group_id: 5, item_id: 10, source_group_id: 5, can_enter_from: 3019-05-30 11:00:00, can_enter_until: 3019-05-30 11:00:00}
					- {group_id: 4, item_id: 10, source_group_id: 4, can_enter_from: 2019-05-30 11:00:01, can_enter_until: 3019-05-30 11:00:00}`,
			args: args{teamGroupID: 2, userID: 4, isAddition: true},
		},
		{
			name: "addition breaks entry_min_admitted_members_ratio = Half (can_enter_until)",
			fixture: `
				items: [{id: 10, default_language_tag: fr, entry_min_admitted_members_ratio: Half, entry_max_team_size: 100}]
				permissions_granted:
					- {group_id: 1, item_id: 10, source_group_id: 1, can_enter_from: 2019-05-30 11:00:00, can_enter_until: 3019-05-30 11:00:00}
					- {group_id: 5, item_id: 10, source_group_id: 5, can_enter_from: 3019-05-30 11:00:00, can_enter_until: 3019-05-30 11:00:00}
					- {group_id: 4, item_id: 10, source_group_id: 4, can_enter_from: 2019-05-30 11:00:00, can_enter_until: 2019-05-30 11:00:00}`,
			args: args{teamGroupID: 2, userID: 4, isAddition: true},
		},
		{
			name: "addition breaks entry_min_admitted_members_ratio = One (can_enter_from)",
			fixture: `
				items: [{id: 10, default_language_tag: fr, entry_min_admitted_members_ratio: One, entry_max_team_size: 100}]
				permissions_granted:
					- {group_id: 4, item_id: 10, source_group_id: 4, can_enter_from: 2019-05-30 11:00:01, can_enter_until: 3019-05-30 11:00:00}`,
			args: args{teamGroupID: 2, userID: 4, isAddition: true},
		},
		{
			name: "addition breaks entry_min_admitted_members_ratio = One (can_enter_until)",
			fixture: `
				items: [{id: 10, default_language_tag: fr, entry_min_admitted_members_ratio: One, entry_max_team_size: 100}]
				permissions_granted:
					- {group_id: 4, item_id: 10, source_group_id: 4, can_enter_from: 2019-05-30 11:00:00, can_enter_until: 2019-05-30 11:00:00}`,
			args: args{teamGroupID: 2, userID: 4, isAddition: true},
		},
		{
			name: "removal breaks entry_min_admitted_members_ratio = All",
			fixture: `
				groups_groups: [{parent_group_id: 2, child_group_id: 4}]
				items: [{id: 10, default_language_tag: fr, entry_min_admitted_members_ratio: All, entry_max_team_size: 100}]
				permissions_granted:
					- {group_id: 1, item_id: 10, source_group_id: 1, can_enter_from: 2019-05-30 11:00:00, can_enter_until: 3019-05-30 11:00:00}`,
			args: args{teamGroupID: 2, userID: 4},
		},
		{
			name: "removal breaks entry_min_admitted_members_ratio = Half",
			fixture: `
				groups: [{id: 6}]
				groups_groups: [{parent_group_id: 2, child_group_id: 4}, {parent_group_id: 2, child_group_id: 6}]
				items: [{id: 10, default_language_tag: fr, entry_min_admitted_members_ratio: Half, entry_max_team_size: 100}]
				permissions_granted:
					- {group_id: 1, item_id: 10, source_group_id: 1, can_enter_from: 2019-05-30 11:00:00, can_enter_until: 3019-05-30 11:00:00}
					- {group_id: 4, item_id: 10, source_group_id: 4, can_enter_from: 2019-05-30 11:00:00, can_enter_until: 3019-05-30 11:00:00}`,
			args: args{teamGroupID: 2, userID: 4},
		},
		{
			name: "removal breaks entry_min_admitted_members_ratio = One",
			fixture: `
				groups_groups: [{parent_group_id: 2, child_group_id: 4}]
				items: [{id: 10, default_language_tag: fr, entry_min_admitted_members_ratio: One, entry_max_team_size: 100}]
				permissions_granted:
					- {group_id: 4, item_id: 10, source_group_id: 4, can_enter_from: 2019-05-30 11:00:00, can_enter_until: 3019-05-30 11:00:00}`,
			args: args{teamGroupID: 2, userID: 4},
		},
		{
			name: "addition satisfies entry_min_admitted_members_ratio = All",
			fixture: `
				items: [{id: 10, default_language_tag: fr, entry_min_admitted_members_ratio: All, entry_max_team_size: 100}]
				permissions_granted:
					- {group_id: 1, item_id: 10, source_group_id: 1, can_enter_from: 2019-05-30 11:00:00, can_enter_until: 3019-05-30 11:00:00}
					- {group_id: 5, item_id: 10, source_group_id: 5, can_enter_from: 2019-05-30 11:00:00, can_enter_until: 3019-05-30 11:00:00}
					- {group_id: 4, item_id: 10, source_group_id: 4, can_enter_from: 2019-05-30 11:00:00, can_enter_until: 2019-05-30 11:00:01}`,
			args: args{teamGroupID: 2, userID: 4, isAddition: true},
			want: true,
		},
		{
			name: "addition satisfies entry_min_admitted_members_ratio = Half",
			fixture: `
				items: [{id: 10, default_language_tag: fr, entry_min_admitted_members_ratio: Half, entry_max_team_size: 100}]
				permissions_granted:
					- {group_id: 1, item_id: 10, source_group_id: 1, can_enter_from: 2019-05-30 11:00:00, can_enter_until: 3019-05-30 11:00:00}
					- {group_id: 4, item_id: 10, source_group_id: 4, can_enter_from: 2019-05-30 11:00:00, can_enter_until: 2019-05-30 11:00:01}`,
			args: args{teamGroupID: 2, userID: 4, isAddition: true},
			want: true,
		},
		{
			name: "addition satisfies entry_min_admitted_members_ratio = One",
			fixture: `
				items: [{id: 10, default_language_tag: fr, entry_min_admitted_members_ratio: One, entry_max_team_size: 100}]
				permissions_granted:
					- {group_id: 1, item_id: 10, source_group_id: 1, can_enter_from: 2019-05-30 11:00:00, can_enter_until: 3019-05-30 11:00:00}`,
			args: args{teamGroupID: 2, userID: 4, isAddition: true},
			want: true,
		},
		{
			name: "addition satisfies entry_min_admitted_members_ratio = None",
			fixture: `
				items: [{id: 10, default_language_tag: fr, entry_min_admitted_members_ratio: None, entry_max_team_size: 100}]`,
			args: args{teamGroupID: 2, userID: 4, isAddition: true},
			want: true,
		},
		{
			name: "removal satisfies entry_min_admitted_members_ratio = All",
			fixture: `
				groups_groups: [{parent_group_id: 2, child_group_id: 4}]
				items: [{id: 10, default_language_tag: fr, entry_min_admitted_members_ratio: All, entry_max_team_size: 100}]
				permissions_granted:
					- {group_id: 1, item_id: 10, source_group_id: 1, can_enter_from: 2019-05-30 11:00:00, can_enter_until: 3019-05-30 11:00:00}
					- {group_id: 5, item_id: 10, source_group_id: 5, can_enter_from: 2019-05-30 11:00:00, can_enter_until: 3019-05-30 11:00:00}`,
			args: args{teamGroupID: 2, userID: 4},
			want: true,
		},
		{
			name: "removal satisfies entry_min_admitted_members_ratio = Half",
			fixture: `
				groups_groups: [{parent_group_id: 2, child_group_id: 4}]
				items: [{id: 10, default_language_tag: fr, entry_min_admitted_members_ratio: Half, entry_max_team_size: 100}]
				permissions_granted:
					- {group_id: 5, item_id: 10, source_group_id: 5, can_enter_from: 2019-05-30 11:00:00, can_enter_until: 3019-05-30 11:00:00}`,
			args: args{teamGroupID: 2, userID: 4},
			want: true,
		},
		{
			name: "removal satisfies entry_min_admitted_members_ratio = One",
			fixture: `
				groups_groups: [{parent_group_id: 2, child_group_id: 4}]
				items: [{id: 10, default_language_tag: fr, entry_min_admitted_members_ratio: One, entry_max_team_size: 100}]
				permissions_granted:
					- {group_id: 5, item_id: 10, source_group_id: 5, can_enter_from: 2019-05-30 11:00:00, can_enter_until: 3019-05-30 11:00:00}`,
			args: args{teamGroupID: 2, userID: 4},
			want: true,
		},
		{
			name: "removal satisfies entry_min_admitted_members_ratio = None",
			fixture: `
				groups_groups: [{parent_group_id: 2, child_group_id: 4}]
				items: [{id: 10, default_language_tag: fr, entry_min_admitted_members_ratio: None, entry_max_team_size: 100}]`,
			args: args{teamGroupID: 2, userID: 4},
			want: true,
		},
		{
			name: "addition breaks entry_max_team_size",
			fixture: `
				items: [{id: 10, default_language_tag: fr, entry_min_admitted_members_ratio: None, entry_max_team_size: 2}]`,
			args: args{teamGroupID: 2, userID: 4, isAddition: true},
		},
		{
			name: "removal breaks entry_max_team_size",
			fixture: `
				groups_groups: [{parent_group_id: 2, child_group_id: 4}]
				items: [{id: 10, default_language_tag: fr, entry_min_admitted_members_ratio: None, entry_max_team_size: 1}]`,
			args: args{teamGroupID: 2, userID: 4},
		},
		{
			name: "ignores participations without started_at",
			fixture: `
				items: [{id: 10, default_language_tag: fr, entry_min_admitted_members_ratio: Half, entry_max_team_size: 100}]
				permissions_granted:
					- {group_id: 1, item_id: 10, source_group_id: 1, can_enter_from: 2019-05-30 11:00:00, can_enter_until: 3019-05-30 11:00:00}
					- {group_id: 4, item_id: 10, source_group_id: 4, can_enter_from: 2019-05-30 11:00:00, can_enter_until: 2019-05-30 11:00:01}
				items: [{id: 11, default_language_tag: fr, entry_min_admitted_members_ratio: All, entry_max_team_size: 1}]
				attempts: [{participant_id: 2, id: 101, root_item_id: 11}]
				results: [{participant_id: 2, attempt_id: 101, item_id: 11, started_at: null}]`,
			args: args{teamGroupID: 2, userID: 4, isAddition: true},
			want: true,
		},
		{
			name: "ignores expired participations",
			fixture: `
				items: [{id: 10, default_language_tag: fr, entry_min_admitted_members_ratio: Half, entry_max_team_size: 100}]
				permissions_granted:
					- {group_id: 1, item_id: 10, source_group_id: 1, can_enter_from: 2019-05-30 11:00:00, can_enter_until: 3019-05-30 11:00:00}
					- {group_id: 4, item_id: 10, source_group_id: 4, can_enter_from: 2019-05-30 11:00:00, can_enter_until: 2019-05-30 11:00:01}
				items: [{id: 11, default_language_tag: fr, entry_min_admitted_members_ratio: All, entry_max_team_size: 1}]
				attempts: [{participant_id: 2, id: 101, root_item_id: 11, allows_submissions_until: 2019-05-30 12:00:00}]
				results: [{participant_id: 2, attempt_id: 101, item_id: 11, started_at: 2019-05-30 11:00:00}]`,
			args: args{teamGroupID: 2, userID: 4, isAddition: true},
			want: true,
		},
		{
			name: "ignores ended participations",
			fixture: `
				items: [{id: 10, default_language_tag: fr, entry_min_admitted_members_ratio: Half, entry_max_team_size: 100}]
				permissions_granted:
					- {group_id: 1, item_id: 10, source_group_id: 1, can_enter_from: 2019-05-30 11:00:00, can_enter_until: 3019-05-30 11:00:00}
					- {group_id: 4, item_id: 10, source_group_id: 4, can_enter_from: 2019-05-30 11:00:00, can_enter_until: 2019-05-30 11:00:01}
				items: [{id: 11, default_language_tag: fr, entry_min_admitted_members_ratio: All, entry_max_team_size: 1}]
				attempts: [{participant_id: 2, id: 101, root_item_id: 11, ended_at: 2019-05-30 12:00:00}]
				results: [{participant_id: 2, attempt_id: 101, item_id: 11, started_at: 2019-05-30 11:00:00}]`,
			args: args{teamGroupID: 2, userID: 4, isAddition: true},
			want: true,
		},
		{
			name: "multiple participations",
			fixture: `
				items: [{id: 10, default_language_tag: fr, entry_min_admitted_members_ratio: Half, entry_max_team_size: 100}]
				permissions_granted:
					- {group_id: 1, item_id: 10, source_group_id: 1, can_enter_from: 2019-05-30 11:00:00, can_enter_until: 3019-05-30 11:00:00}
					- {group_id: 4, item_id: 10, source_group_id: 4, can_enter_from: 2019-05-30 11:00:00, can_enter_until: 2019-05-30 11:00:01}
				items: [{id: 11, default_language_tag: fr, entry_min_admitted_members_ratio: All, entry_max_team_size: 1}]
				attempts: [{participant_id: 2, id: 101, root_item_id: 11}]
				results: [{participant_id: 2, attempt_id: 101, item_id: 11, started_at: 2019-05-30 11:00:00}]`,
			args: args{teamGroupID: 2, userID: 4, isAddition: true},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			testoutput.SuppressIfPasses(t)

			db := testhelpers.SetupDBWithFixtureString(ctx, mainFixture, tt.fixture)
			defer func() { _ = db.Close() }()
			for _, withLock := range []bool{true, false} {
				withLock := withLock
				t.Run(fmt.Sprintf(" withLock = %v", withLock), func(t *testing.T) {
					testoutput.SuppressIfPasses(t)

					assert.NoError(t, database.NewDataStore(db).InTransaction(func(store *database.DataStore) error {
						if err := store.GroupGroups().CreateNewAncestors(); err != nil {
							return err
						}
						got, err := store.Groups().CheckIfEntryConditionsStillSatisfiedForAllActiveParticipations(
							tt.args.teamGroupID, tt.args.userID, tt.args.isAddition, withLock)
						if (err != nil) != tt.wantErr {
							t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
							return nil
						}
						if got != tt.want {
							t.Errorf("got = %v, want %v", got, tt.want)
						}
						return nil
					}))
				})
			}
		})
	}
}

func TestGroupStore_DeleteGroup(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db := testhelpers.SetupDBWithFixtureString(testhelpers.CreateTestContext(), `groups: [{id: 1234}]`)
	defer func() { _ = db.Close() }()
	groupStore := database.NewDataStore(db).Groups()
	assert.NoError(t, groupStore.InTransaction(func(store *database.DataStore) error {
		return store.Groups().DeleteGroup(1234)
	}))
	var ids []int64
	require.NoError(t, groupStore.Pluck("id", &ids).Error())
	assert.Empty(t, ids)
	require.NoError(t, groupStore.Table("groups_propagate").Pluck("id", &ids).Error())
	assert.Empty(t, ids)
}

func TestGroupStore_DeleteGroup_RecomputesAccess(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	ctx := testhelpers.CreateTestContext()
	for _, test := range []struct {
		name    string
		fixture string
	}{
		{
			name: "Common",
			fixture: `
				groups: [{id: 1234}, {id: 1235}]
				items: [{id: 10, default_language_tag: fr}]
				permissions_granted:
					- {group_id: 1234, item_id: 10, source_group_id: 1235, can_view: content}
					- {group_id: 1234, item_id: 10, source_group_id: 1234, can_view: info}
				permissions_generated: [{group_id: 1234, item_id: 10, can_view_generated: content}]`,
		},
		{
			name: "OrphanedSourceGroups",
			fixture: `
				groups: [{id: 1234}, {id: 1235}, {id: 1236}]
				groups_groups: [{parent_group_id: 1235, child_group_id: 1236}]
				groups_ancestors: [{ancestor_group_id: 1235, child_group_id: 1236}]
				items: [{id: 10, default_language_tag: fr}]
				permissions_granted:
					- {group_id: 1234, item_id: 10, source_group_id: 1236, can_view: content}
					- {group_id: 1234, item_id: 10, source_group_id: 1234, can_view: info}
				permissions_generated: [{group_id: 1234, item_id: 10, can_view_generated: content}]`,
		},
	} {
		test := test
		t.Run(test.name, func(t *testing.T) {
			testoutput.SuppressIfPasses(t)

			db := testhelpers.SetupDBWithFixtureString(ctx, test.fixture)
			defer func() { _ = db.Close() }()

			store := database.NewDataStore(db)
			require.NoError(t, store.InTransaction(func(store *database.DataStore) error {
				return store.Groups().DeleteGroup(1235)
			}))
			var newPermission string
			require.NoError(t, store.Permissions().Where("group_id = 1234 AND item_id = 10").
				PluckFirst("can_view_generated", &newPermission).Error())
			assert.Equal(t, "info", newPermission)
		})
	}
}

func TestGroupStore_TriggerBeforeUpdate_RefusesToModifyType(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	const expectedErrorMessage = "Error 1644 (45000): Unable to change groups.type from/to User/Team"

	ctx := testhelpers.CreateTestContext()
	for _, test := range []struct {
		oldType     string
		newType     string
		expectError bool
	}{
		{oldType: "User", newType: "Team", expectError: true},
		{oldType: "Team", newType: "User", expectError: true},
		{oldType: "Class", newType: "User", expectError: true},
		{oldType: "Class", newType: "Team", expectError: true},
		{oldType: "User", newType: "Class", expectError: true},
		{oldType: "Team", newType: "Class", expectError: true},
		{oldType: "Team", newType: "Team", expectError: false},
		{oldType: "User", newType: "User", expectError: false},
		{oldType: "Other", newType: "Club", expectError: false},
	} {
		test := test
		t.Run(fmt.Sprintf("%s to %s", test.oldType, test.newType), func(t *testing.T) {
			testoutput.SuppressIfPasses(t)

			db := testhelpers.SetupDBWithFixtureString(ctx, `groups: [{id: 1, type: `+test.oldType+`}]`)
			defer func() { _ = db.Close() }()

			groupGroupStore := database.NewDataStore(db).Groups()
			result := groupGroupStore.ByID(1).UpdateColumn("type", test.newType)
			if test.expectError {
				assert.EqualError(t, result.Error(), expectedErrorMessage)
			} else {
				assert.NoError(t, result.Error())
			}
		})
	}
}
