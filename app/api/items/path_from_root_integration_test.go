// +build !unit

package items_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/app/api/items"
	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/testhelpers"
)

func Test_findItemPath(t *testing.T) {
	type args struct {
		participantID int64
		itemID        int64
	}
	tests := []struct {
		name    string
		fixture string
		args    args
		want    []string
	}{
		{
			name: "fails if not enough permissions for the first item",
			fixture: `
				permissions_generated:
					- {group_id: 200, item_id: 1, can_view_generated: none}
					- {group_id: 200, item_id: 2, can_view_generated: info}
			`,
			args: args{participantID: 101, itemID: 2},
		},
		{
			name: "fails if not enough permissions for the second item",
			fixture: `
				permissions_generated:
					- {group_id: 200, item_id: 1, can_view_generated: info}
					- {group_id: 200, item_id: 2, can_view_generated: none}
			`,
			args: args{participantID: 101, itemID: 2},
		},
		{
			name: "supports a root activity as a first item",
			fixture: `
				permissions_generated:
					- {group_id: 200, item_id: 1, can_view_generated: content}
					- {group_id: 200, item_id: 2, can_view_generated: content}
			`,
			args: args{participantID: 101, itemID: 2},
			want: []string{"1", "2"},
		},
		{
			name: "supports a root skill as a first item",
			fixture: `
				permissions_generated:
					- {group_id: 200, item_id: 3, can_view_generated: content}
					- {group_id: 200, item_id: 4, can_view_generated: content}
			`,
			args: args{participantID: 101, itemID: 4},
			want: []string{"3", "4"},
		},
		{
			name: "supports permissions given directly",
			fixture: `
				permissions_generated:
					- {group_id: 100, item_id: 1, can_view_generated: content}
					- {group_id: 100, item_id: 2, can_view_generated: content}
			`,
			args: args{participantID: 100, itemID: 2},
			want: []string{"1", "2"},
		},
		{
			name: "steps into child attempts for items requiring explicit entry",
			fixture: `
				permissions_generated:
					- {group_id: 100, item_id: 1, can_view_generated: content}
					- {group_id: 100, item_id: 2, can_view_generated: content}
					- {group_id: 100, item_id: 22, can_view_generated: content}
				attempts:
					- {participant_id: 100, id: 1, parent_attempt_id: 0, root_item_id: 22}
					- {participant_id: 100, id: 2, parent_attempt_id: 1, root_item_id: 22}
					- {participant_id: 100, id: 3, parent_attempt_id: 0, root_item_id: 4}
					- {participant_id: 101, id: 4, parent_attempt_id: 0, root_item_id: 22}
				results:
					- {participant_id: 100, attempt_id: 0, started_at: 2019-05-30 11:00:00, item_id: 2}
					- {participant_id: 100, attempt_id: 0, started_at: 2019-05-30 11:00:00, item_id: 22}
					- {participant_id: 100, attempt_id: 1, item_id: 22}
					- {participant_id: 100, attempt_id: 2, item_id: 22}
					- {participant_id: 100, attempt_id: 3, item_id: 22}
					- {participant_id: 101, attempt_id: 4, item_id: 22}
			`,
			args: args{participantID: 100, itemID: 22},
			want: []string{"1", "2", "22"},
		},
		{
			name: "supports paths starting with an item requiring explicit entry",
			fixture: `
				groups: [{id: 103, root_activity_id: 22}]
				permissions_generated:
					- {group_id: 103, item_id: 22, can_view_generated: content}
				attempts:
					- {participant_id: 103, id: 0}
					- {participant_id: 103, id: 1, parent_attempt_id: 0, root_item_id: 22}
				results:
					- {participant_id: 103, attempt_id: 1, item_id: 22}
			`,
			args: args{participantID: 103, itemID: 22},
			want: []string{"22"},
		},
		{
			name: "can find a path without a result for the first item",
			fixture: `
				permissions_generated:
					- {group_id: 101, item_id: 1, can_view_generated: content}
					- {group_id: 101, item_id: 2, can_view_generated: content}
				attempts:
					- {participant_id: 101, id: 1}
				results:
					- {participant_id: 101, attempt_id: 1, item_id: 2}
			`,
			args: args{participantID: 101, itemID: 2},
			want: []string{"1", "2"},
		},
		{
			name: "prefers the path for the last (by id) existing attempt chain with started results",
			fixture: `
				items:
					- {id: 21, default_language_tag: fr}
				items_items:
					- {parent_item_id: 1, child_item_id: 21, child_order: 1}
					- {parent_item_id: 21, child_item_id: 22, child_order: 1}
				permissions_generated:
					- {group_id: 101, item_id: 1, can_view_generated: content}
					- {group_id: 101, item_id: 2, can_view_generated: content}
					- {group_id: 101, item_id: 21, can_view_generated: info}
					- {group_id: 101, item_id: 22, can_view_generated: content}
				attempts:
					- {participant_id: 101, id: 1}
					- {participant_id: 101, id: 2}
					- {participant_id: 101, id: 3, parent_attempt_id: 1, root_item_id: 22}
					- {participant_id: 101, id: 4, parent_attempt_id: 0, root_item_id: 22}
					- {participant_id: 101, id: 5, parent_attempt_id: 2, root_item_id: 22}
				results:
					- {participant_id: 101, attempt_id: 0, item_id: 1, started_at: 2019-05-30 11:00:00}
					- {participant_id: 101, attempt_id: 0, item_id: 2, started_at: 2019-05-30 11:00:00}
					- {participant_id: 101, attempt_id: 1, item_id: 1, started_at: 2019-05-30 11:00:00}
					- {participant_id: 101, attempt_id: 1, item_id: 2, started_at: 2019-05-30 11:00:00}
					- {participant_id: 101, attempt_id: 2, item_id: 1, started_at: 2019-05-30 11:00:00}
					- {participant_id: 101, attempt_id: 2, item_id: 2, started_at: 2019-05-30 11:00:00}
					- {participant_id: 101, attempt_id: 2, item_id: 21, started_at: 2019-05-30 11:00:00}
					- {participant_id: 101, attempt_id: 2, item_id: 22}
					- {participant_id: 101, attempt_id: 3, item_id: 22, started_at: 2019-05-30 11:00:00}
					- {participant_id: 101, attempt_id: 4, item_id: 22, started_at: 2019-05-30 11:00:00}
					- {participant_id: 101, attempt_id: 5, item_id: 22, started_at: 2019-05-30 11:00:00}
			`,
			args: args{participantID: 101, itemID: 22},
			want: []string{"1", "21", "22"},
		},
		{
			name: "prefers the path for the attempt chain with the highest score",
			fixture: `
				items:
					- {id: 21, default_language_tag: fr}
				items_items:
					- {parent_item_id: 1, child_item_id: 21, child_order: 1}
					- {parent_item_id: 21, child_item_id: 22, child_order: 1}
				permissions_generated:
					- {group_id: 101, item_id: 1, can_view_generated: content}
					- {group_id: 101, item_id: 2, can_view_generated: content}
					- {group_id: 101, item_id: 21, can_view_generated: info}
					- {group_id: 101, item_id: 22, can_view_generated: content}
				attempts:
					- {participant_id: 101, id: 1}
					- {participant_id: 101, id: 2}
					- {participant_id: 101, id: 3, parent_attempt_id: 1, root_item_id: 22}
					- {participant_id: 101, id: 4, parent_attempt_id: 0, root_item_id: 22}
					- {participant_id: 101, id: 5, parent_attempt_id: 0, root_item_id: 22}
				results:
					- {participant_id: 101, attempt_id: 0, item_id: 1, started_at: 2019-05-30 11:00:00}
					- {participant_id: 101, attempt_id: 0, item_id: 2, started_at: 2019-05-30 11:00:00}
					- {participant_id: 101, attempt_id: 1, item_id: 1}
					- {participant_id: 101, attempt_id: 1, item_id: 2, started_at: 2019-05-30 11:00:00}
					- {participant_id: 101, attempt_id: 2, item_id: 1, started_at: 2019-05-30 11:00:00}
					- {participant_id: 101, attempt_id: 2, item_id: 2, started_at: 2019-05-30 11:00:00}
					- {participant_id: 101, attempt_id: 2, item_id: 22}
					- {participant_id: 101, attempt_id: 3, item_id: 22, started_at: 2019-05-30 11:00:00}
					- {participant_id: 101, attempt_id: 4, item_id: 22}
					- {participant_id: 101, attempt_id: 5, item_id: 22, started_at: 2019-05-30 11:00:00}
			`,
			args: args{participantID: 101, itemID: 22},
			want: []string{"1", "2", "22"},
		},
		{
			name: "prefers the path for the last (by id) attempt chain among all chains with started results for the same items",
			fixture: `
				items:
					- {id: 21, default_language_tag: fr}
				items_items:
					- {parent_item_id: 1, child_item_id: 21, child_order: 1}
					- {parent_item_id: 21, child_item_id: 22, child_order: 1}
				permissions_generated:
					- {group_id: 101, item_id: 1, can_view_generated: content}
					- {group_id: 101, item_id: 2, can_view_generated: content}
					- {group_id: 101, item_id: 21, can_view_generated: info}
					- {group_id: 101, item_id: 22, can_view_generated: content}
				attempts:
					- {participant_id: 101, id: 1}
					- {participant_id: 101, id: 2}
					- {participant_id: 101, id: 3, parent_attempt_id: 1, root_item_id: 22}
					- {participant_id: 101, id: 4, parent_attempt_id: 0, root_item_id: 22}
					- {participant_id: 101, id: 5, parent_attempt_id: 0, root_item_id: 22}
					- {participant_id: 101, id: 6}
					- {participant_id: 101, id: 7, parent_attempt_id: 6, root_item_id: 22}
				results:
					- {participant_id: 101, attempt_id: 0, item_id: 1}
					- {participant_id: 101, attempt_id: 1, item_id: 1}
					- {participant_id: 101, attempt_id: 6, item_id: 1}
					- {participant_id: 101, attempt_id: 1, item_id: 2, started_at: 2019-05-30 11:00:00}
					- {participant_id: 101, attempt_id: 2, item_id: 1, started_at: 2019-05-30 11:00:00}
					- {participant_id: 101, attempt_id: 2, item_id: 2, started_at: 2019-05-30 11:00:00}
					- {participant_id: 101, attempt_id: 6, item_id: 21, started_at: 2019-05-30 11:00:00}
					- {participant_id: 101, attempt_id: 2, item_id: 22}
					- {participant_id: 101, attempt_id: 3, item_id: 22, started_at: 2019-05-30 11:00:00}
					- {participant_id: 101, attempt_id: 4, item_id: 22}
					- {participant_id: 101, attempt_id: 5, item_id: 22, started_at: 2019-05-30 11:00:00}
					- {participant_id: 101, attempt_id: 7, item_id: 22, started_at: 2019-05-30 11:00:00}
			`,
			args: args{participantID: 101, itemID: 22},
			want: []string{"1", "21", "22"},
		},
		{
			name: "ignores paths whose attempt chains have missing results for items requiring explicit entry",
			fixture: `
				permissions_generated:
					- {group_id: 200, item_id: 1, can_view_generated: content}
					- {group_id: 200, item_id: 2, can_view_generated: content}
					- {group_id: 200, item_id: 22, can_view_generated: content}
				attempts:
					- {participant_id: 101, id: 1, root_item_id: 22, parent_attempt_id: 0}
				results:
					- {participant_id: 101, attempt_id: 0, item_id: 1, started_at: 2019-05-30 11:00:00}
					- {participant_id: 101, attempt_id: 0, item_id: 2, started_at: 2019-05-30 11:00:00}
			`,
			args: args{participantID: 101, itemID: 22},
		},
		{
			name: "ignores paths whose attempt chains have not started results below an attempt not allowing submissions",
			fixture: `
				permissions_generated:
					- {group_id: 200, item_id: 1, can_view_generated: content}
					- {group_id: 200, item_id: 2, can_view_generated: content}
					- {group_id: 200, item_id: 22, can_view_generated: content}
				attempts:
					- {participant_id: 101, id: 1, allows_submissions_until: 2019-05-30 11:00:00}
					- {participant_id: 101, id: 2, root_item_id: 22, parent_attempt_id: 1}
				results:
					- {participant_id: 101, attempt_id: 1, item_id: 1, started_at: 2019-05-30 11:00:00}
					- {participant_id: 101, attempt_id: 1, item_id: 2, started_at: 2019-05-30 11:00:00}
					- {participant_id: 101, attempt_id: 2, item_id: 22}
			`,
			args: args{participantID: 101, itemID: 22},
		},
		{
			name: "ignores paths whose attempt chains have not started results below an ended attempt",
			fixture: `
				permissions_generated:
					- {group_id: 200, item_id: 1, can_view_generated: content}
					- {group_id: 200, item_id: 2, can_view_generated: content}
					- {group_id: 200, item_id: 22, can_view_generated: content}
				attempts:
					- {participant_id: 101, id: 1, ended_at: 2019-05-30 11:00:00}
					- {participant_id: 101, id: 2, root_item_id: 22, parent_attempt_id: 1}
				results:
					- {participant_id: 101, attempt_id: 1, item_id: 1, started_at: 2019-05-30 11:00:00}
					- {participant_id: 101, attempt_id: 1, item_id: 2, started_at: 2019-05-30 11:00:00}
					- {participant_id: 101, attempt_id: 2, item_id: 22}
			`,
			args: args{participantID: 101, itemID: 22},
		},
		{
			name: "supports path with attempt chains having ended or not allowing submissions attempts",
			fixture: `
				permissions_generated:
					- {group_id: 200, item_id: 1, can_view_generated: content}
					- {group_id: 200, item_id: 2, can_view_generated: content}
					- {group_id: 200, item_id: 22, can_view_generated: content}
				attempts:
					- {participant_id: 101, id: 1, ended_at: 2019-05-30 11:00:00, allows_submissions_until: 2019-05-30 11:00:00}
					- {participant_id: 101, id: 2, root_item_id: 22, parent_attempt_id: 1}
				results:
					- {participant_id: 101, attempt_id: 1, item_id: 1, started_at: 2019-05-30 11:00:00}
					- {participant_id: 101, attempt_id: 1, item_id: 2, started_at: 2019-05-30 11:00:00}
					- {participant_id: 101, attempt_id: 2, item_id: 22, started_at: 2019-05-30 11:00:00}
			`,
			args: args{participantID: 101, itemID: 22},
			want: []string{"1", "2", "22"},
		},
		{
			name: "ignores paths whose attempt chains have not started results for an attempt not allowing submissions",
			fixture: `
				groups: [{id: 103, root_activity_id: 1}]
				permissions_generated:
					- {group_id: 103, item_id: 1, can_view_generated: content}
				attempts:
					- {participant_id: 103, id: 1, allows_submissions_until: 2019-05-30 11:00:00}
				results:
					- {participant_id: 103, attempt_id: 1, item_id: 1}
			`,
			args: args{participantID: 103, itemID: 1},
		},
		{
			name: "ignores paths whose attempt chains have not started results for an ended attempt",
			fixture: `
				groups: [{id: 103, root_activity_id: 1}]
				permissions_generated:
					- {group_id: 103, item_id: 1, can_view_generated: content}
				attempts:
					- {participant_id: 103, id: 1, ended_at: 2019-05-30 11:00:00}
				results:
					- {participant_id: 103, attempt_id: 1, item_id: 1}
			`,
			args: args{participantID: 103, itemID: 1},
		},
	}
	const globalFixture = `
		groups: [{id: 100}, {id: 101}, {id: 200, root_activity_id: 1, root_skill_id: 3}]
		groups_groups: [{parent_group_id: 200, child_group_id: 100}, {parent_group_id: 200, child_group_id: 101}]
		items:
			- {id: 1, default_language_tag: fr}
			- {id: 2, default_language_tag: fr}
			- {id: 3, default_language_tag: fr}
			- {id: 4, default_language_tag: fr}
			- {id: 22, default_language_tag: fr, requires_explicit_entry: true}
		items_items:
			- {parent_item_id: 1, child_item_id: 2, child_order: 1}
			- {parent_item_id: 2, child_item_id: 3, child_order: 1}
			- {parent_item_id: 2, child_item_id: 22, child_order: 2}
			- {parent_item_id: 3, child_item_id: 4, child_order: 1}
		attempts:
			- {participant_id: 100, id: 0}
			- {participant_id: 101, id: 0}
	`
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			db := testhelpers.SetupDBWithFixtureString(globalFixture, tt.fixture)
			defer func() { _ = db.Close() }()
			store := database.NewDataStore(db)
			var got []string
			assert.NoError(t, store.InTransaction(func(s *database.DataStore) error {
				assert.NoError(t, s.GroupGroups().After())
				assert.NoError(t, s.ItemItems().After())
				return nil
			}))
			got = items.FindItemPath(store, tt.args.participantID, tt.args.itemID)
			assert.Equal(t, tt.want, got)
		})
	}
}
