// +build !unit

package currentuser_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/app/api/currentuser"
	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/service"
	"github.com/France-ioi/AlgoreaBackend/testhelpers"
)

func Test_checkPreconditionsForGroupRequests(t *testing.T) {
	tests := []struct {
		name         string
		fixture      string
		wantAPIError service.APIError
	}{
		{
			name: "parent group is not a team",
			fixture: `
				groups:
					- {ID: 1, bFreeAccess: 1, sType: "Class", idTeamItem: 1234}
					- {ID: 2, sType: Team, idTeamItem: 1234}
					- {ID: 3, sType: "Team", "idTeamItem": 1234}
					- {ID: 10, sType: UserSelf}
				groups_groups: [{idGroupParent: 2, idGroupChild: 10, sType: "joinedByCode"}]`,
			wantAPIError: service.NoError,
		},
		{
			name: "parent group is a team without idTeamItem",
			fixture: `
				groups:
					- {ID: 1, bFreeAccess: 1, sType: "Team"}
					- {ID: 2, sType: Team, idTeamItem: 1234}
					- {ID: 3, sType: "Team", "idTeamItem": 1234}
					- {ID: 10, sType: UserSelf}
				groups_groups: [{idGroupParent: 2, idGroupChild: 10, sType: "invitationAccepted"}]`,
			wantAPIError: service.NoError,
		},
		{
			name: "parent group is a team with idTeamItem, but the user is not on teams",
			fixture: `
				groups:
					- {ID: 1, bFreeAccess: 1, sType: "Team", idTeamItem: 1234}
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
					- {idGroupParent: 4, idGroupChild: 10, sType: "invitationAccepted"}
					- {idGroupParent: 5, idGroupChild: 10, sType: "requestAccepted"}
					- {idGroupParent: 6, idGroupChild: 10, sType: "joinedByCode"}
					- {idGroupParent: 7, idGroupChild: 10, sType: "direct"}`,
			wantAPIError: service.NoError,
		},
		{
			name: "parent group is a team with idTeamItem, but the user is on teams with mismatching idTeamItems",
			fixture: `
				groups:
					- {ID: 1, bFreeAccess: 1, sType: "Team", idTeamItem: 1234}
					- {ID: 2, sType: Team, idTeamItem: 2345}
					- {ID: 3, sType: Team}
					- {ID: 4, sType: Team, idTeamItem: 2345}
					- {ID: 5, sType: Team, idTeamItem: 2345}
					- {ID: 10, sType: UserSelf}
				groups_groups:
					- {idGroupParent: 2, idGroupChild: 10, sType: "invitationAccepted"}
					- {idGroupParent: 3, idGroupChild: 10, sType: "requestAccepted"}
					- {idGroupParent: 4, idGroupChild: 10, sType: "joinedByCode"}
					- {idGroupParent: 5, idGroupChild: 10, sType: "direct"}`,
			wantAPIError: service.NoError,
		},
		{
			name: "parent group is a team with idTeamItem and the user is on a team with the same idTeamItem (invitationAccepted)",
			fixture: `
				groups:
					- {ID: 1, bFreeAccess: 1, sType: "Team", idTeamItem: 1234}
					- {ID: 2, sType: Team, idTeamItem: 1234}
					- {ID: 10, sType: UserSelf}
				groups_groups:
					- {idGroupParent: 2, idGroupChild: 10, sType: "invitationAccepted"}`,
			wantAPIError: service.ErrUnprocessableEntity(errors.New("you are already on a team for this item")),
		},
		{
			name: "parent group is a team with idTeamItem and the user is on a team with the same idTeamItem (requestAccepted)",
			fixture: `
				groups:
					- {ID: 1, bFreeAccess: 1, sType: "Team", idTeamItem: 1234}
					- {ID: 2, sType: Team, idTeamItem: 1234}
					- {ID: 10, sType: UserSelf}
				groups_groups:
					- {idGroupParent: 2, idGroupChild: 10, sType: "requestAccepted"}`,
			wantAPIError: service.ErrUnprocessableEntity(errors.New("you are already on a team for this item")),
		},
		{
			name: "parent group is a team with idTeamItem and the user is on a team with the same idTeamItem (joinedByCode)",
			fixture: `
				groups:
					- {ID: 1, bFreeAccess: 1, sType: "Team", idTeamItem: 1234}
					- {ID: 2, sType: Team, idTeamItem: 1234}
					- {ID: 10, sType: UserSelf}
				groups_groups:
					- {idGroupParent: 2, idGroupChild: 10, sType: "joinedByCode"}`,
			wantAPIError: service.ErrUnprocessableEntity(errors.New("you are already on a team for this item")),
		},
		{
			name: "parent group is a team with idTeamItem and the user is on a team with the same idTeamItem (direct)",
			fixture: `
				groups:
					- {ID: 1, bFreeAccess: 1, sType: "Team", idTeamItem: 1234}
					- {ID: 2, sType: Team, idTeamItem: 1234}
					- {ID: 10, sType: UserSelf}
				groups_groups:
					- {idGroupParent: 2, idGroupChild: 10, sType: "direct"}`,
			wantAPIError: service.ErrUnprocessableEntity(errors.New("you are already on a team for this item")),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			db := testhelpers.SetupDBWithFixtureString(tt.fixture)
			defer func() { _ = db.Close() }()

			store := database.NewDataStore(db)
			var apiError service.APIError
			assert.NoError(t, store.InTransaction(func(transactionStore *database.DataStore) error {
				apiError = currentuser.CheckPreconditionsForGroupRequests(transactionStore,
					&database.User{SelfGroupID: ptrInt64(10)}, 1, true)
				return nil
			}))
			assert.Equal(t, tt.wantAPIError, apiError)
		})
	}
}

func ptrInt64(i int64) *int64 { return &i }
