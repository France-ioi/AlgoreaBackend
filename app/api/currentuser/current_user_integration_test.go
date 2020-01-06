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
					- {id: 1, free_access: 1, type: "Class", team_item_id: 1234}
					- {id: 2, type: Team, team_item_id: 1234}
					- {id: 3, type: "Team", "team_item_id": 1234}
					- {id: 10, type: UserSelf}
				groups_groups: [{parent_group_id: 2, child_group_id: 10}]`,
			wantAPIError: service.NoError,
		},
		{
			name: "parent group is a team without team_item_id",
			fixture: `
				groups:
					- {id: 1, free_access: 1, type: "Team"}
					- {id: 2, type: Team, team_item_id: 1234}
					- {id: 3, type: "Team", "team_item_id": 1234}
					- {id: 10, type: UserSelf}
				groups_groups: [{parent_group_id: 2, child_group_id: 10}]`,
			wantAPIError: service.NoError,
		},
		{
			name: "parent group is a team with team_item_id, but the user is not on teams",
			fixture: `
				groups:
					- {id: 1, free_access: 1, type: "Team", team_item_id: 1234}
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
					- {parent_group_id: 4, child_group_id: 10}
					- {parent_group_id: 5, child_group_id: 10}
					- {parent_group_id: 6, child_group_id: 10}
					- {parent_group_id: 7, child_group_id: 10}`,
			wantAPIError: service.NoError,
		},
		{
			name: "parent group is a team with team_item_id, but the user is on teams with mismatching team_item_id",
			fixture: `
				groups:
					- {id: 1, free_access: 1, type: "Team", team_item_id: 1234}
					- {id: 2, type: Team, team_item_id: 2345}
					- {id: 3, type: Team}
					- {id: 4, type: Team, team_item_id: 2345}
					- {id: 5, type: Team, team_item_id: 2345}
					- {id: 10, type: UserSelf}
				groups_groups:
					- {parent_group_id: 2, child_group_id: 10}
					- {parent_group_id: 3, child_group_id: 10}
					- {parent_group_id: 4, child_group_id: 10}
					- {parent_group_id: 5, child_group_id: 10}`,
			wantAPIError: service.NoError,
		},
		{
			name: "parent group is a team with team_item_id and the user is on a team with the same team_item_id",
			fixture: `
				groups:
					- {id: 1, free_access: 1, type: "Team", team_item_id: 1234}
					- {id: 2, type: Team, team_item_id: 1234}
					- {id: 10, type: UserSelf}
				groups_groups:
					- {parent_group_id: 2, child_group_id: 10}`,
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
					&database.User{GroupID: 10}, 1, "createJoinRequest")
				return nil
			}))
			assert.Equal(t, tt.wantAPIError, apiError)
		})
	}
}
