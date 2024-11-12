//go:build !unit

package groups_test

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/v2/app/api/groups"
	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/testhelpers"
	"github.com/France-ioi/AlgoreaBackend/v2/testhelpers/testoutput"
)

func Test_filterOtherTeamsMembersOut(t *testing.T) {
	tests := []struct {
		name           string
		fixture        string
		groupsToInvite []int64
		want           []int64
		wantWrongIDs   []int64
	}{
		{
			name: "parent group is not a team",
			fixture: `
				groups:
					- {id: 1, type: Class}
					- {id: 2, type: Team}
					- {id: 3, type: Team}
					- {id: 10, type: User}
				groups_groups: [{parent_group_id: 2, child_group_id: 10}]
				items: [{id: 1234, default_language_tag: fr, allows_multiple_attempts: 1}]
				attempts:
					- {participant_id: 1, id: 1, root_item_id: 1234}
					- {participant_id: 2, id: 1, root_item_id: 1234}
					- {participant_id: 3, id: 1, root_item_id: 1234}`,
			groupsToInvite: []int64{10},
			want:           []int64{10},
		},
		{
			name: "parent group is a team without attempts",
			fixture: `
				groups:
					- {id: 1, type: Team}
					- {id: 2, type: Team}
					- {id: 3, type: Team}
					- {id: 10, type: User}
				groups_groups: [{parent_group_id: 2, child_group_id: 10}]
				items: [{id: 1234, default_language_tag: fr, allows_multiple_attempts: 1}]
				attempts:
					- {participant_id: 2, id: 1, root_item_id: 1234}
					- {participant_id: 3, id: 1, root_item_id: 1234}`,
			groupsToInvite: []int64{10},
			want:           []int64{10},
		},
		{
			name: "parent group is a team with attempts, but children are not in teams",
			fixture: `
				groups:
					- {id: 1, type: Team}
					- {id: 2, type: Team}
					- {id: 3, type: Team}
					- {id: 4, type: Class}
					- {id: 5, type: Friends}
					- {id: 6, type: Other}
					- {id: 7, type: Club}
					- {id: 10, type: User}
					- {id: 11, type: User}
					- {id: 12, type: User}
					- {id: 13, type: User}
					- {id: 14, type: User}
					- {id: 15, type: User}
				groups_groups:
					- {parent_group_id: 4, child_group_id: 10}
					- {parent_group_id: 5, child_group_id: 11}
					- {parent_group_id: 6, child_group_id: 12}
					- {parent_group_id: 7, child_group_id: 13}
				items: [{id: 1234, default_language_tag: fr, allows_multiple_attempts: 1}]
				attempts:
					- {participant_id: 1, id: 1, root_item_id: 1234}
					- {participant_id: 2, id: 1, root_item_id: 1234}
					- {participant_id: 3, id: 1, root_item_id: 1234}
					- {participant_id: 4, id: 1, root_item_id: 1234}
					- {participant_id: 5, id: 1, root_item_id: 1234}
					- {participant_id: 6, id: 1, root_item_id: 1234}
					- {participant_id: 7, id: 1, root_item_id: 1234}`,
			groupsToInvite: []int64{10, 11, 12, 13, 14, 15},
			want:           []int64{10, 11, 12, 13, 14, 15},
		},
		{
			name: "parent group is a team with attempts, but children groups are in teams participating in different contests",
			fixture: `
				groups:
					- {id: 1, type: Team}
					- {id: 2, type: Team}
					- {id: 3, type: Team}
					- {id: 10, type: User}
					- {id: 11, type: User}
					- {id: 12, type: User}
					- {id: 13, type: User}
				groups_groups:
					- {parent_group_id: 2, child_group_id: 10}
					- {parent_group_id: 3, child_group_id: 11}
					- {parent_group_id: 2, child_group_id: 12}
					- {parent_group_id: 3, child_group_id: 13}
				items:
					- {id: 1234, default_language_tag: fr, allows_multiple_attempts: 1}
					- {id: 2345, default_language_tag: fr, allows_multiple_attempts: 1}
				attempts:
					- {participant_id: 1, id: 1, root_item_id: 1234}
					- {participant_id: 2, id: 1, root_item_id: 2345}`,
			groupsToInvite: []int64{10, 11, 12, 13},
			want:           []int64{10, 11, 12, 13},
		},
		{
			name: "parent group is a team with attempts and children groups are in teams with attempts for the same contest",
			fixture: `
				groups:
					- {id: 1, type: Team}
					- {id: 2, type: Team}
					- {id: 3, type: Team}
					- {id: 10, type: User}
					- {id: 11, type: User}
					- {id: 12, type: User}
					- {id: 13, type: User}
				groups_groups:
					- {parent_group_id: 2, child_group_id: 10}
					- {parent_group_id: 3, child_group_id: 11}
					- {parent_group_id: 2, child_group_id: 12}
					- {parent_group_id: 3, child_group_id: 13}
				items: [{id: 1234, default_language_tag: fr, allows_multiple_attempts: 1}]
				attempts:
					- {participant_id: 1, id: 1, root_item_id: 1234}
					- {participant_id: 2, id: 1, root_item_id: 1234}
					- {participant_id: 3, id: 1, root_item_id: 1234}`,
			groupsToInvite: []int64{10, 11, 12, 13},
			want:           []int64{},
			wantWrongIDs:   []int64{10, 11, 12, 13},
		},
		{
			name: "parent group is a team with expired attempts and children groups are in teams with attempts for the same contest",
			fixture: `
				groups:
					- {id: 1, type: Team}
					- {id: 2, type: Team}
					- {id: 3, type: Team}
					- {id: 10, type: User}
					- {id: 11, type: User}
					- {id: 12, type: User}
					- {id: 13, type: User}
				groups_groups:
					- {parent_group_id: 2, child_group_id: 10}
					- {parent_group_id: 3, child_group_id: 11}
					- {parent_group_id: 2, child_group_id: 12}
					- {parent_group_id: 3, child_group_id: 13}
				items: [{id: 1234, default_language_tag: fr, allows_multiple_attempts: 1}]
				attempts:
					- {participant_id: 1, id: 1, root_item_id: 1234, allows_submissions_until: 2019-05-30 11:00:00}
					- {participant_id: 2, id: 1, root_item_id: 1234}
					- {participant_id: 3, id: 1, root_item_id: 1234}`,
			groupsToInvite: []int64{10, 11, 12, 13},
			want:           []int64{10, 11, 12, 13},
			wantWrongIDs:   []int64{},
		},
		{
			name: "parent group is a team with ended attempts and children groups are in teams with attempts for the same contest",
			fixture: `
				groups:
					- {id: 1, type: Team}
					- {id: 2, type: Team}
					- {id: 3, type: Team}
					- {id: 10, type: User}
					- {id: 11, type: User}
					- {id: 12, type: User}
					- {id: 13, type: User}
				groups_groups:
					- {parent_group_id: 2, child_group_id: 10}
					- {parent_group_id: 3, child_group_id: 11}
					- {parent_group_id: 2, child_group_id: 12}
					- {parent_group_id: 3, child_group_id: 13}
				items: [{id: 1234, default_language_tag: fr, allows_multiple_attempts: 1}]
				attempts:
					- {participant_id: 1, id: 1, root_item_id: 1234, ended_at: 2019-05-30 11:00:00}
					- {participant_id: 2, id: 1, root_item_id: 1234}
					- {participant_id: 3, id: 1, root_item_id: 1234}`,
			groupsToInvite: []int64{10, 11, 12, 13},
			want:           []int64{10, 11, 12, 13},
			wantWrongIDs:   []int64{},
		},
		{
			name: "parent group is a team with attempts and children groups are in teams with expired/ended attempts for the same contest",
			fixture: `
				groups:
					- {id: 1, type: Team}
					- {id: 2, type: Team}
					- {id: 3, type: Team}
					- {id: 10, type: User}
					- {id: 11, type: User}
					- {id: 12, type: User}
					- {id: 13, type: User}
				groups_groups:
					- {parent_group_id: 2, child_group_id: 10}
					- {parent_group_id: 3, child_group_id: 11}
					- {parent_group_id: 2, child_group_id: 12}
					- {parent_group_id: 3, child_group_id: 13}
				items: [{id: 1234, default_language_tag: fr, allows_multiple_attempts: 1}]
				attempts:
					- {participant_id: 1, id: 1, root_item_id: 1234}
					- {participant_id: 2, id: 1, root_item_id: 1234, allows_submissions_until: 2019-05-30 11:00:00}
					- {participant_id: 3, id: 1, root_item_id: 1234, ended_at: 2019-05-30 11:00:00}`,
			groupsToInvite: []int64{10, 11, 12, 13},
			want:           []int64{10, 11, 12, 13},
			wantWrongIDs:   []int64{},
		},
		{
			name: "parent group is a team with expired attempts and children groups are in teams with expired/ended attempts " +
				"for the same contest, but the contest doesn't allow multiple attempts",
			fixture: `
				groups:
					- {id: 1, type: Team}
					- {id: 2, type: Team}
					- {id: 3, type: Team}
					- {id: 10, type: User}
					- {id: 11, type: User}
					- {id: 12, type: User}
					- {id: 13, type: User}
				groups_groups:
					- {parent_group_id: 2, child_group_id: 10}
					- {parent_group_id: 3, child_group_id: 11}
					- {parent_group_id: 2, child_group_id: 12}
					- {parent_group_id: 3, child_group_id: 13}
				items: [{id: 1234, default_language_tag: fr, allows_multiple_attempts: 0}]
				attempts:
					- {participant_id: 1, id: 1, root_item_id: 1234}
					- {participant_id: 2, id: 1, root_item_id: 1234, ended_at: 2019-05-30 11:00:00}
					- {participant_id: 3, id: 1, root_item_id: 1234, allows_submissions_until: 2019-05-30 11:00:00}`,
			groupsToInvite: []int64{10, 11, 12, 13},
			want:           []int64{},
			wantWrongIDs:   []int64{10, 11, 12, 13},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			testoutput.SuppressIfPasses(t)

			db := testhelpers.SetupDBWithFixtureString(tt.fixture)
			defer func() { _ = db.Close() }()

			results := make(map[string]string, len(tt.groupsToInvite))
			groupIDToLoginMap := make(map[int64]string, len(tt.groupsToInvite))
			for _, id := range tt.groupsToInvite {
				login := strconv.FormatInt(id, 10)
				results[login] = "not_found"
				groupIDToLoginMap[id] = login
			}

			store := database.NewDataStore(db)
			var got []int64
			assert.NoError(t, store.InTransaction(func(transactionStore *database.DataStore) error {
				got = groups.FilterOtherTeamsMembersOutForLogins(transactionStore, 1, tt.groupsToInvite, results, groupIDToLoginMap)
				return nil
			}))
			assert.Equal(t, tt.want, got)
			for _, id := range tt.wantWrongIDs {
				assert.Equal(t, "in_another_team", results[groupIDToLoginMap[id]])
			}
		})
	}
}
