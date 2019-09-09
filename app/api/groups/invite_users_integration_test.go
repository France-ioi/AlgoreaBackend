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
					- {ID: 1, sType: "Class", idTeamItem: 1234}
					- {ID: 2, sType: Team, idTeamItem: 1234}
					- {ID: 3, sType: "Team", "idTeamItem": 1234}
					- {ID: 10, sType: UserSelf}
				groups_groups: [{idGroupParent: 2, idGroupChild: 10, sType: "joinedByCode"}]`,
			groupsToInvite: []int64{10},
			want:           []int64{10},
		},
		{
			name: "parent group is a team without idTeamItem",
			fixture: `
				groups:
					- {ID: 1, sType: "Team"}
					- {ID: 2, sType: Team, idTeamItem: 1234}
					- {ID: 3, sType: "Team", "idTeamItem": 1234}
					- {ID: 10, sType: UserSelf}
				groups_groups: [{idGroupParent: 2, idGroupChild: 10, sType: "invitationAccepted"}]`,
			groupsToInvite: []int64{10},
			want:           []int64{10},
		},
		{
			name: "parent group is a team with idTeamItem, but children are not in teams",
			fixture: `
				groups:
					- {ID: 1, sType: "Team", idTeamItem: 1234}
					- {ID: 2, sType: Team, idTeamItem: 1234}
					- {ID: 3, sType: "Team", "idTeamItem": 1234}
					- {ID: 4, sType: "Class", "idTeamItem": 1234}
					- {ID: 5, sType: "Friends", "idTeamItem": 1234}
					- {ID: 6, sType: "Other", "idTeamItem": 1234}
					- {ID: 7, sType: "Club", "idTeamItem": 1234}
					- {ID: 10, sType: UserSelf}
					- {ID: 11, sType: UserSelf}
					- {ID: 12, sType: UserSelf}
					- {ID: 13, sType: UserSelf}
					- {ID: 14, sType: UserSelf}
					- {ID: 15, sType: UserSelf}
				groups_groups:
					- {idGroupParent: 2, idGroupChild: 10, sType: "invitationSent"}
					- {idGroupParent: 3, idGroupChild: 11, sType: "requestSent"}
					- {idGroupParent: 2, idGroupChild: 12, sType: "invitationRefused"}
					- {idGroupParent: 3, idGroupChild: 13, sType: "requestRefused"}
					- {idGroupParent: 2, idGroupChild: 14, sType: "removed"}
					- {idGroupParent: 3, idGroupChild: 15, sType: "left"}
					- {idGroupParent: 4, idGroupChild: 10, sType: "invitationAccepted"}
					- {idGroupParent: 5, idGroupChild: 11, sType: "requestAccepted"}
					- {idGroupParent: 6, idGroupChild: 12, sType: "joinedByCode"}
					- {idGroupParent: 7, idGroupChild: 13, sType: "direct"}`,
			groupsToInvite: []int64{10, 11, 12, 13, 14, 15},
			want:           []int64{10, 11, 12, 13, 14, 15},
		},
		{
			name: "parent group is a team with idTeamItem, but children groups are in teams with mismatching idTeamItems",
			fixture: `
				groups:
					- {ID: 1, sType: "Team", idTeamItem: 1234}
					- {ID: 2, sType: Team, idTeamItem: 2345}
					- {ID: 3, sType: "Team"}
					- {ID: 10, sType: UserSelf}
					- {ID: 11, sType: UserSelf}
					- {ID: 12, sType: UserSelf}
					- {ID: 13, sType: UserSelf}
				groups_groups:
					- {idGroupParent: 2, idGroupChild: 10, sType: "invitationAccepted"}
					- {idGroupParent: 3, idGroupChild: 11, sType: "requestAccepted"}
					- {idGroupParent: 2, idGroupChild: 12, sType: "joinedByCode"}
					- {idGroupParent: 3, idGroupChild: 13, sType: "direct"}`,
			groupsToInvite: []int64{10, 11, 12, 13},
			want:           []int64{10, 11, 12, 13},
		},
		{
			name: "parent group is a team with idTeamItem and children groups are in teams with the same idTeamItem",
			fixture: `
				groups:
					- {ID: 1, sType: "Team", idTeamItem: 1234}
					- {ID: 2, sType: Team, idTeamItem: 1234}
					- {ID: 3, sType: "Team", idTeamItem: 1234}
					- {ID: 10, sType: UserSelf}
					- {ID: 11, sType: UserSelf}
					- {ID: 12, sType: UserSelf}
					- {ID: 13, sType: UserSelf}
				groups_groups:
					- {idGroupParent: 2, idGroupChild: 10, sType: "invitationAccepted"}
					- {idGroupParent: 3, idGroupChild: 11, sType: "requestAccepted"}
					- {idGroupParent: 2, idGroupChild: 12, sType: "joinedByCode"}
					- {idGroupParent: 3, idGroupChild: 13, sType: "direct"}`,
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
