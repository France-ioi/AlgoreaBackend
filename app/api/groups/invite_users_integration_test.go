// +build !unit

package groups_test

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/app/api/groups"
	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/testhelpers"
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
					- {id: 1, type: "Class", team_item_id: 1234}
					- {id: 2, type: Team, team_item_id: 1234}
					- {id: 3, type: "Team", "team_item_id": 1234}
					- {id: 10, type: UserSelf}
				groups_groups: [{parent_group_id: 2, child_group_id: 10, type: "joinedByCode"}]`,
			groupsToInvite: []int64{10},
			want:           []int64{10},
		},
		{
			name: "parent group is a team without team_item_id",
			fixture: `
				groups:
					- {id: 1, type: "Team"}
					- {id: 2, type: Team, team_item_id: 1234}
					- {id: 3, type: "Team", "team_item_id": 1234}
					- {id: 10, type: UserSelf}
				groups_groups: [{parent_group_id: 2, child_group_id: 10, type: "invitationAccepted"}]`,
			groupsToInvite: []int64{10},
			want:           []int64{10},
		},
		{
			name: "parent group is a team with team_item_id, but children are not in teams",
			fixture: `
				groups:
					- {id: 1, type: "Team", team_item_id: 1234}
					- {id: 2, type: Team, team_item_id: 1234}
					- {id: 3, type: "Team", "team_item_id": 1234}
					- {id: 4, type: "Class", "team_item_id": 1234}
					- {id: 5, type: "Friends", "team_item_id": 1234}
					- {id: 6, type: "Other", "team_item_id": 1234}
					- {id: 7, type: "Club", "team_item_id": 1234}
					- {id: 10, type: UserSelf}
					- {id: 11, type: UserSelf}
					- {id: 12, type: UserSelf}
					- {id: 13, type: UserSelf}
					- {id: 14, type: UserSelf}
					- {id: 15, type: UserSelf}
				groups_groups:
					- {parent_group_id: 2, child_group_id: 10, type: "invitationSent"}
					- {parent_group_id: 3, child_group_id: 11, type: "requestSent"}
					- {parent_group_id: 2, child_group_id: 12, type: "invitationRefused"}
					- {parent_group_id: 3, child_group_id: 13, type: "requestRefused"}
					- {parent_group_id: 2, child_group_id: 14, type: "removed"}
					- {parent_group_id: 3, child_group_id: 15, type: "left"}
					- {parent_group_id: 4, child_group_id: 10, type: "invitationAccepted"}
					- {parent_group_id: 5, child_group_id: 11, type: "requestAccepted"}
					- {parent_group_id: 6, child_group_id: 12, type: "joinedByCode"}
					- {parent_group_id: 7, child_group_id: 13, type: "direct"}`,
			groupsToInvite: []int64{10, 11, 12, 13, 14, 15},
			want:           []int64{10, 11, 12, 13, 14, 15},
		},
		{
			name: "parent group is a team with team_item_id, but children groups are in teams with mismatching team_item_id",
			fixture: `
				groups:
					- {id: 1, type: "Team", team_item_id: 1234}
					- {id: 2, type: Team, team_item_id: 2345}
					- {id: 3, type: "Team"}
					- {id: 10, type: UserSelf}
					- {id: 11, type: UserSelf}
					- {id: 12, type: UserSelf}
					- {id: 13, type: UserSelf}
				groups_groups:
					- {parent_group_id: 2, child_group_id: 10, type: "invitationAccepted"}
					- {parent_group_id: 3, child_group_id: 11, type: "requestAccepted"}
					- {parent_group_id: 2, child_group_id: 12, type: "joinedByCode"}
					- {parent_group_id: 3, child_group_id: 13, type: "direct"}`,
			groupsToInvite: []int64{10, 11, 12, 13},
			want:           []int64{10, 11, 12, 13},
		},
		{
			name: "parent group is a team with team_item_id and children groups are in teams with the same team_item_id",
			fixture: `
				groups:
					- {id: 1, type: "Team", team_item_id: 1234}
					- {id: 2, type: Team, team_item_id: 1234}
					- {id: 3, type: "Team", team_item_id: 1234}
					- {id: 10, type: UserSelf}
					- {id: 11, type: UserSelf}
					- {id: 12, type: UserSelf}
					- {id: 13, type: UserSelf}
				groups_groups:
					- {parent_group_id: 2, child_group_id: 10, type: "invitationAccepted"}
					- {parent_group_id: 3, child_group_id: 11, type: "requestAccepted"}
					- {parent_group_id: 2, child_group_id: 12, type: "joinedByCode"}
					- {parent_group_id: 3, child_group_id: 13, type: "direct"}`,
			groupsToInvite: []int64{10, 11, 12, 13},
			want:           []int64{},
			wantWrongIDs:   []int64{10, 11, 12, 13},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
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
