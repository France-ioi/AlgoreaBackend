//go:build !unit

package items_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/France-ioi/AlgoreaBackend/v2/app/api/items"
	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/testhelpers"
	"github.com/France-ioi/AlgoreaBackend/v2/testhelpers/testoutput"
)

func Test_getDataForResultPathStart(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	type args struct {
		participantID int64
		ids           []int64
	}
	tests := []struct {
		name    string
		fixture string
		args    args
		want    []map[string]interface{}
	}{
		{
			name: "fails if not enough permissions for the first item",
			fixture: `
				permissions_generated:
					- {group_id: 200, item_id: 1, can_view_generated: info}
					- {group_id: 200, item_id: 2, can_view_generated: content}
			`,
			args: args{participantID: 101, ids: []int64{1, 2}},
		},
		{
			name: "fails if not enough permissions for the second item",
			fixture: `
				permissions_generated:
					- {group_id: 200, item_id: 1, can_view_generated: content}
					- {group_id: 200, item_id: 2, can_view_generated: info}
			`,
			args: args{participantID: 101, ids: []int64{1, 2}},
		},
		{
			name: "fails if the path is not a parent-child sequence",
			fixture: `
				permissions_generated:
					- {group_id: 200, item_id: 1, can_view_generated: content}
					- {group_id: 200, item_id: 3, can_view_generated: content}
			`,
			args: args{participantID: 101, ids: []int64{1, 3}},
		},
		{
			name: "fails if the first item is not a root activity/skill",
			fixture: `
				permissions_generated:
					- {group_id: 200, item_id: 2, can_view_generated: content}
					- {group_id: 200, item_id: 3, can_view_generated: content}
			`,
			args: args{participantID: 101, ids: []int64{2, 3}},
		},
		{
			name: "supports a root activity as a first item",
			fixture: `
				permissions_generated:
					- {group_id: 200, item_id: 1, can_view_generated: content}
					- {group_id: 200, item_id: 2, can_view_generated: content}
			`,
			args: args{participantID: 101, ids: []int64{1, 2}},
			want: []map[string]interface{}{{
				"attempt_id0": int64(0), "attempt_id1": int64(0),
				"has_started_result0": int64(0), "has_started_result1": int64(0),
			}},
		},
		{
			name: "supports a root skill as a first item",
			fixture: `
				permissions_generated:
					- {group_id: 200, item_id: 3, can_view_generated: content}
					- {group_id: 200, item_id: 4, can_view_generated: content}
			`,
			args: args{participantID: 101, ids: []int64{3, 4}},
			want: []map[string]interface{}{{
				"attempt_id0": int64(0), "attempt_id1": int64(0),
				"has_started_result0": int64(0), "has_started_result1": int64(0),
			}},
		},
		{
			name: "supports a root activity of a managed group as a first item",
			fixture: `
				groups: [{id: 102}, {id: 103, root_activity_id: 5, root_skill_id: 6}, {id: 104}, {id: 105}]
				groups_groups: [{parent_group_id: 102, child_group_id: 103}, {parent_group_id: 104, child_group_id: 105}]
				group_managers: [{manager_id: 104, group_id: 102}]
				items:
					- {id: 5, default_language_tag: fr}
					- {id: 6, default_language_tag: fr}
				items_items: [{parent_item_id: 5, child_item_id: 6, child_order: 1}]
				permissions_generated:
					- {group_id: 105, item_id: 5, can_view_generated: content}
					- {group_id: 105, item_id: 6, can_view_generated: content}
				attempts:
					- {participant_id: 105, id: 0}
			`,
			args: args{participantID: 105, ids: []int64{5, 6}},
			want: []map[string]interface{}{{
				"attempt_id0": int64(0), "attempt_id1": int64(0),
				"has_started_result0": int64(0), "has_started_result1": int64(0),
			}},
		},
		{
			name: "supports a root skill of a managed group as a first item",
			fixture: `
				groups: [{id: 102}, {id: 103, root_activity_id: 6, root_skill_id: 5}, {id: 104}, {id: 105}]
				groups_groups: [{parent_group_id: 102, child_group_id: 103}, {parent_group_id: 104, child_group_id: 105}]
				group_managers: [{manager_id: 104, group_id: 102}]
				items:
					- {id: 5, default_language_tag: fr}
					- {id: 6, default_language_tag: fr}
				items_items: [{parent_item_id: 5, child_item_id: 6, child_order: 1}]
				permissions_generated:
					- {group_id: 105, item_id: 5, can_view_generated: content}
					- {group_id: 105, item_id: 6, can_view_generated: content}
				attempts:
					- {participant_id: 105, id: 0}
			`,
			args: args{participantID: 105, ids: []int64{5, 6}},
			want: []map[string]interface{}{{
				"attempt_id0": int64(0), "attempt_id1": int64(0),
				"has_started_result0": int64(0), "has_started_result1": int64(0),
			}},
		},
		{
			name: "ignores results of other participants",
			fixture: `
				permissions_generated:
					- {group_id: 200, item_id: 1, can_view_generated: content}
					- {group_id: 200, item_id: 2, can_view_generated: content}
				results:
					- {participant_id: 101, attempt_id: 0, item_id: 1, started_at: 2019-05-30 11:00:00}
					- {participant_id: 101, attempt_id: 0, item_id: 2, started_at: 2019-05-30 11:00:00}
			`,
			args: args{participantID: 100, ids: []int64{1, 2}},
			want: []map[string]interface{}{
				{"attempt_id0": int64(0), "has_started_result0": int64(0), "attempt_id1": int64(0), "has_started_result1": int64(0)},
			},
		},
		{
			name: "supports permissions given directly",
			fixture: `
				permissions_generated:
					- {group_id: 100, item_id: 1, can_view_generated: content}
					- {group_id: 100, item_id: 2, can_view_generated: content}
			`,
			args: args{participantID: 100, ids: []int64{1, 2}},
			want: []map[string]interface{}{
				{"attempt_id0": int64(0), "has_started_result0": int64(0), "attempt_id1": int64(0), "has_started_result1": int64(0)},
			},
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
					- {participant_id: 100, attempt_id: 1, item_id: 22}
					- {participant_id: 100, attempt_id: 2, item_id: 22}
					- {participant_id: 100, attempt_id: 3, item_id: 22}
					- {participant_id: 101, attempt_id: 4, item_id: 22}
			`,
			args: args{participantID: 100, ids: []int64{1, 2, 22}},
			want: []map[string]interface{}{
				{
					"attempt_id0": int64(0), "attempt_id1": int64(0), "attempt_id2": int64(1),
					"has_started_result0": int64(0), "has_started_result1": int64(1), "has_started_result2": int64(0),
				},
			},
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
			args: args{participantID: 103, ids: []int64{22}},
			want: []map[string]interface{}{
				{"attempt_id0": int64(1), "has_started_result0": int64(0)},
			},
		},
		{
			name: "can find attempt chains without a result for the first item",
			fixture: `
				permissions_generated:
					- {group_id: 101, item_id: 1, can_view_generated: content}
					- {group_id: 101, item_id: 2, can_view_generated: content}
				attempts:
					- {participant_id: 101, id: 1}
				results:
					- {participant_id: 101, attempt_id: 1, item_id: 2}
			`,
			args: args{participantID: 101, ids: []int64{1, 2}},
			want: []map[string]interface{}{
				{"attempt_id0": int64(0), "has_started_result0": int64(0), "attempt_id1": int64(0), "has_started_result1": int64(0)},
			},
		},
		{
			name: "prefers the last (by id) existing attempt chain with started results",
			fixture: `
				permissions_generated:
					- {group_id: 101, item_id: 1, can_view_generated: content}
					- {group_id: 101, item_id: 2, can_view_generated: content}
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
					- {participant_id: 101, attempt_id: 1, item_id: 1, started_at: 2019-05-30 11:00:00}
					- {participant_id: 101, attempt_id: 1, item_id: 2, started_at: 2019-05-30 11:00:00}
					- {participant_id: 101, attempt_id: 2, item_id: 1, started_at: 2019-05-30 11:00:00}
					- {participant_id: 101, attempt_id: 2, item_id: 2, started_at: 2019-05-30 11:00:00}
					- {participant_id: 101, attempt_id: 2, item_id: 22}
					- {participant_id: 101, attempt_id: 3, item_id: 22, started_at: 2019-05-30 11:00:00}
					- {participant_id: 101, attempt_id: 4, item_id: 22, started_at: 2019-05-30 11:00:00}
					- {participant_id: 101, attempt_id: 5, item_id: 22, started_at: 2019-05-30 11:00:00}
			`,
			args: args{participantID: 101, ids: []int64{1, 2, 22}},
			want: []map[string]interface{}{
				{
					"attempt_id0": int64(1), "has_started_result0": int64(1),
					"attempt_id1": int64(1), "has_started_result1": int64(1),
					"attempt_id2": int64(3), "has_started_result2": int64(1),
				},
			},
		},
		{
			name: "prefers the attempt chain with the highest score",
			fixture: `
				permissions_generated:
					- {group_id: 101, item_id: 1, can_view_generated: content}
					- {group_id: 101, item_id: 2, can_view_generated: content}
					- {group_id: 101, item_id: 22, can_view_generated: content}
				attempts:
					- {participant_id: 101, id: 1}
					- {participant_id: 101, id: 2}
					- {participant_id: 101, id: 3, parent_attempt_id: 1, root_item_id: 22}
					- {participant_id: 101, id: 4, parent_attempt_id: 0, root_item_id: 22}
					- {participant_id: 101, id: 5, parent_attempt_id: 0, root_item_id: 22}
				results:
					- {participant_id: 101, attempt_id: 0, item_id: 1, started_at: 2019-05-30 11:00:00}
					- {participant_id: 101, attempt_id: 1, item_id: 1}
					- {participant_id: 101, attempt_id: 1, item_id: 2, started_at: 2019-05-30 11:00:00}
					- {participant_id: 101, attempt_id: 2, item_id: 1, started_at: 2019-05-30 11:00:00}
					- {participant_id: 101, attempt_id: 2, item_id: 2, started_at: 2019-05-30 11:00:00}
					- {participant_id: 101, attempt_id: 2, item_id: 22}
					- {participant_id: 101, attempt_id: 3, item_id: 22, started_at: 2019-05-30 11:00:00}
					- {participant_id: 101, attempt_id: 4, item_id: 22}
					- {participant_id: 101, attempt_id: 5, item_id: 22, started_at: 2019-05-30 11:00:00}
			`,
			args: args{participantID: 101, ids: []int64{1, 2, 22}},
			want: []map[string]interface{}{
				{
					"attempt_id0": int64(0), "has_started_result0": int64(1),
					"attempt_id1": int64(0), "has_started_result1": int64(0),
					"attempt_id2": int64(5), "has_started_result2": int64(1),
				},
			},
		},
		{
			name: "prefers the last (by id) attempt chain among all chains with started results for the same items",
			fixture: `
				permissions_generated:
					- {group_id: 101, item_id: 1, can_view_generated: content}
					- {group_id: 101, item_id: 2, can_view_generated: content}
					- {group_id: 101, item_id: 22, can_view_generated: content}
				attempts:
					- {participant_id: 101, id: 1}
					- {participant_id: 101, id: 2}
					- {participant_id: 101, id: 3, parent_attempt_id: 1, root_item_id: 22}
					- {participant_id: 101, id: 4, parent_attempt_id: 0, root_item_id: 22}
					- {participant_id: 101, id: 5, parent_attempt_id: 0, root_item_id: 22}
				results:
					- {participant_id: 101, attempt_id: 0, item_id: 1, started_at: 2019-05-30 11:00:00}
					- {participant_id: 101, attempt_id: 1, item_id: 1, started_at: 2019-05-30 11:00:00}
					- {participant_id: 101, attempt_id: 1, item_id: 2}
					- {participant_id: 101, attempt_id: 2, item_id: 1}
					- {participant_id: 101, attempt_id: 2, item_id: 2, started_at: 2019-05-30 11:00:00}
					- {participant_id: 101, attempt_id: 2, item_id: 22}
					- {participant_id: 101, attempt_id: 3, item_id: 22, started_at: 2019-05-30 11:00:00}
					- {participant_id: 101, attempt_id: 4, item_id: 22}
					- {participant_id: 101, attempt_id: 5, item_id: 22, started_at: 2019-05-30 11:00:00}
			`,
			args: args{participantID: 101, ids: []int64{1, 2, 22}},
			want: []map[string]interface{}{
				{
					"attempt_id0": int64(1), "has_started_result0": int64(1),
					"attempt_id1": int64(1), "has_started_result1": int64(0),
					"attempt_id2": int64(3), "has_started_result2": int64(1),
				},
			},
		},
		{
			name: "ignores attempt chains with missing results for items requiring explicit entry",
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
			args: args{participantID: 101, ids: []int64{1, 2, 22}},
		},
		{
			name: "ignores attempt chains with not started results below an attempt not allowing submissions",
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
			args: args{participantID: 101, ids: []int64{1, 2, 22}},
		},
		{
			name: "ignores attempt chains with not started results below an ended attempt",
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
			args: args{participantID: 101, ids: []int64{1, 2, 22}},
		},
		{
			name: "supports attempt chains with ended or not allowing submissions attempt",
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
			args: args{participantID: 101, ids: []int64{1, 2, 22}},
			want: []map[string]interface{}{
				{
					"attempt_id0": int64(1), "has_started_result0": int64(1),
					"attempt_id1": int64(1), "has_started_result1": int64(1),
					"attempt_id2": int64(2), "has_started_result2": int64(1),
				},
			},
		},
		{
			name: "ignores attempt chains with not started results for an attempt not allowing submissions",
			fixture: `
				groups: [{id: 103, root_activity_id: 1}]
				permissions_generated:
					- {group_id: 103, item_id: 1, can_view_generated: content}
				attempts:
					- {participant_id: 103, id: 1, allows_submissions_until: 2019-05-30 11:00:00}
				results:
					- {participant_id: 103, attempt_id: 1, item_id: 1}
			`,
			args: args{participantID: 103, ids: []int64{1}},
		},
		{
			name: "ignores attempt chains with not started results for an ended attempt",
			fixture: `
				groups: [{id: 103, root_activity_id: 1}]
				permissions_generated:
					- {group_id: 103, item_id: 1, can_view_generated: content}
				attempts:
					- {participant_id: 103, id: 1, ended_at: 2019-05-30 11:00:00}
				results:
					- {participant_id: 103, attempt_id: 1, item_id: 1}
			`,
			args: args{participantID: 103, ids: []int64{1}},
		},
		{
			name: "supports root explicit-entry single-item path with only a started result on attempt 0 (no rooted attempt)",
			fixture: `
				groups: [{id: 110, root_activity_id: 22}]
				groups_groups:
					- {parent_group_id: 110, child_group_id: 100}
				permissions_generated:
					- {group_id: 100, item_id: 22, can_view_generated: content}
				results:
					- {participant_id: 100, attempt_id: 0, item_id: 22, started_at: 2019-05-30 11:00:00}
			`,
			args: args{participantID: 100, ids: []int64{22}},
			want: []map[string]interface{}{
				{"attempt_id0": int64(0), "has_started_result0": int64(1)},
			},
		},
		{
			name: "supports paths through a non-final root explicit-entry item with only a started result on attempt 0 (no rooted attempt)",
			fixture: `
				groups: [{id: 110, root_activity_id: 22}]
				groups_groups:
					- {parent_group_id: 110, child_group_id: 100}
				items:
					- {id: 30, default_language_tag: fr}
				items_items:
					- {parent_item_id: 22, child_item_id: 30, child_order: 1}
				permissions_generated:
					- {group_id: 100, item_id: 22, can_view_generated: content}
					- {group_id: 100, item_id: 30, can_view_generated: content}
				results:
					- {participant_id: 100, attempt_id: 0, item_id: 22, started_at: 2019-05-30 11:00:00}
			`,
			args: args{participantID: 100, ids: []int64{22, 30}},
			want: []map[string]interface{}{
				{
					"attempt_id0": int64(0), "has_started_result0": int64(1),
					"attempt_id1": int64(0), "has_started_result1": int64(0),
				},
			},
		},
		{
			name: "supports paths through an intermediate explicit-entry item with only a started result on attempt 0 (no rooted attempt)",
			fixture: `
				items:
					- {id: 21, default_language_tag: fr, requires_explicit_entry: true}
					- {id: 30, default_language_tag: fr}
				items_items:
					- {parent_item_id: 1, child_item_id: 21, child_order: 5}
					- {parent_item_id: 21, child_item_id: 30, child_order: 1}
				permissions_generated:
					- {group_id: 101, item_id: 1, can_view_generated: content}
					- {group_id: 101, item_id: 21, can_view_generated: content}
					- {group_id: 101, item_id: 30, can_view_generated: content}
				results:
					- {participant_id: 101, attempt_id: 0, item_id: 1, started_at: 2019-05-30 11:00:00}
					- {participant_id: 101, attempt_id: 0, item_id: 21, started_at: 2019-05-30 11:00:00}
			`,
			args: args{participantID: 101, ids: []int64{1, 21, 30}},
			want: []map[string]interface{}{
				{
					"attempt_id0": int64(0), "has_started_result0": int64(1),
					"attempt_id1": int64(0), "has_started_result1": int64(1),
					"attempt_id2": int64(0), "has_started_result2": int64(0),
				},
			},
		},
		{
			// This pins the tie-break introduced by the relaxation: when an explicit-entry item has BOTH
			// a non-rooted attempt with a STARTED result AND rooted-but-not-started attempts available,
			// the non-rooted started attempt wins (lower score). Before the relaxation the rooted attempt
			// was the only viable candidate, so attempt_id2 used to be 1; under the new semantics it is 0.
			name: "prefers a non-rooted attempt with a started result over a rooted attempt with a not-started result for an explicit-entry item",
			fixture: `
				permissions_generated:
					- {group_id: 100, item_id: 1, can_view_generated: content}
					- {group_id: 100, item_id: 2, can_view_generated: content}
					- {group_id: 100, item_id: 22, can_view_generated: content}
				attempts:
					- {participant_id: 100, id: 1, parent_attempt_id: 0, root_item_id: 22}
					- {participant_id: 100, id: 2, parent_attempt_id: 1, root_item_id: 22}
				results:
					- {participant_id: 100, attempt_id: 0, item_id: 1, started_at: 2019-05-30 11:00:00}
					- {participant_id: 100, attempt_id: 0, item_id: 2, started_at: 2019-05-30 11:00:00}
					- {participant_id: 100, attempt_id: 0, item_id: 22, started_at: 2019-05-30 11:00:00}
					- {participant_id: 100, attempt_id: 1, item_id: 22}
					- {participant_id: 100, attempt_id: 2, item_id: 22}
			`,
			args: args{participantID: 100, ids: []int64{1, 2, 22}},
			want: []map[string]interface{}{
				{
					"attempt_id0": int64(0), "has_started_result0": int64(1),
					"attempt_id1": int64(0), "has_started_result1": int64(1),
					"attempt_id2": int64(0), "has_started_result2": int64(1),
				},
			},
		},
		// The four cases below verify the negative side of the relaxation: missing or NOT-STARTED results
		// on non-rooted attempts must NOT unlock chains for explicit-entry items. A result row whose
		// started_at is NULL can legitimately appear as a side effect of score propagation from descendants
		// (or as a placeholder created during attempt setup): it does not prove that the participant ever
		// actually started/entered the item on that attempt. Only a result with a non-NULL started_at
		// provides that evidence.
		{
			name: "still ignores paths through a root explicit-entry item without any result for it",
			fixture: `
				groups: [{id: 110, root_activity_id: 22}]
				groups_groups:
					- {parent_group_id: 110, child_group_id: 100}
				items:
					- {id: 30, default_language_tag: fr}
				items_items:
					- {parent_item_id: 22, child_item_id: 30, child_order: 1}
				permissions_generated:
					- {group_id: 100, item_id: 22, can_view_generated: content}
					- {group_id: 100, item_id: 30, can_view_generated: content}
			`,
			args: args{participantID: 100, ids: []int64{22, 30}},
		},
		{
			name: "still ignores paths through an intermediate explicit-entry item without any result for it",
			fixture: `
				items:
					- {id: 21, default_language_tag: fr, requires_explicit_entry: true}
					- {id: 30, default_language_tag: fr}
				items_items:
					- {parent_item_id: 1, child_item_id: 21, child_order: 5}
					- {parent_item_id: 21, child_item_id: 30, child_order: 1}
				permissions_generated:
					- {group_id: 101, item_id: 1, can_view_generated: content}
					- {group_id: 101, item_id: 21, can_view_generated: content}
					- {group_id: 101, item_id: 30, can_view_generated: content}
				results:
					- {participant_id: 101, attempt_id: 0, item_id: 1, started_at: 2019-05-30 11:00:00}
			`,
			args: args{participantID: 101, ids: []int64{1, 21, 30}},
		},
		{
			name: "still ignores paths through a root explicit-entry item with only a not-started result on a non-rooted attempt",
			fixture: `
				groups: [{id: 110, root_activity_id: 22}]
				groups_groups:
					- {parent_group_id: 110, child_group_id: 100}
				items:
					- {id: 30, default_language_tag: fr}
				items_items:
					- {parent_item_id: 22, child_item_id: 30, child_order: 1}
				permissions_generated:
					- {group_id: 100, item_id: 22, can_view_generated: content}
					- {group_id: 100, item_id: 30, can_view_generated: content}
				results:
					- {participant_id: 100, attempt_id: 0, item_id: 22}
			`,
			args: args{participantID: 100, ids: []int64{22, 30}},
		},
		{
			name: "still ignores paths through an intermediate explicit-entry item with only a not-started result on a non-rooted attempt",
			fixture: `
				items:
					- {id: 21, default_language_tag: fr, requires_explicit_entry: true}
					- {id: 30, default_language_tag: fr}
				items_items:
					- {parent_item_id: 1, child_item_id: 21, child_order: 5}
					- {parent_item_id: 21, child_item_id: 30, child_order: 1}
				permissions_generated:
					- {group_id: 101, item_id: 1, can_view_generated: content}
					- {group_id: 101, item_id: 21, can_view_generated: content}
					- {group_id: 101, item_id: 30, can_view_generated: content}
				results:
					- {participant_id: 101, attempt_id: 0, item_id: 1, started_at: 2019-05-30 11:00:00}
					- {participant_id: 101, attempt_id: 0, item_id: 21}
			`,
			args: args{participantID: 101, ids: []int64{1, 21, 30}},
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
	ctx := testhelpers.CreateTestContext()
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			testoutput.SuppressIfPasses(t)

			db := testhelpers.SetupDBWithFixtureString(ctx, globalFixture, tt.fixture)
			defer func() { _ = db.Close() }()
			store := database.NewDataStore(db)
			var got []map[string]interface{}
			require.NoError(t, store.InTransaction(func(s *database.DataStore) error {
				require.NoError(t, s.GroupGroups().CreateNewAncestors())
				got = items.GetDataForResultPathStart(s, tt.args.participantID, tt.args.ids)
				return nil
			}))
			assert.Equal(t, tt.want, got)
		})
	}
}
