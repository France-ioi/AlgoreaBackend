//go:build !unit

package currentuser_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/v2/app/api/currentuser"
	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/app/service"
	"github.com/France-ioi/AlgoreaBackend/v2/testhelpers"
	"github.com/France-ioi/AlgoreaBackend/v2/testhelpers/testoutput"
)

func Test_checkPreconditionsForGroupRequests(t *testing.T) {
	tests := []struct {
		name         string
		fixture      string
		wantAPIError *service.APIError
	}{
		{
			name: "parent group is not a team",
			fixture: `
				groups:
					- {id: 1, is_public: 1, type: "Class"}
					- {id: 2, type: Team}
					- {id: 3, type: "Team"}
					- {id: 10, type: User}
				items: [{id: 1234, default_language_tag: fr}]
				groups_groups: [{parent_group_id: 2, child_group_id: 10}]
				attempts:
					- {participant_id: 1, id: 1, root_item_id: 1234}
					- {participant_id: 2, id: 1, root_item_id: 1234}
					- {participant_id: 3, id: 1, root_item_id: 1234}`,
			wantAPIError: service.NoError,
		},
		{
			name: "parent group is a team",
			fixture: `
				groups:
					- {id: 1, is_public: 1, type: "Team"}
					- {id: 2, type: Team}
					- {id: 3, type: "Team"}
					- {id: 10, type: User}
				groups_groups: [{parent_group_id: 2, child_group_id: 10}]
				items: [{id: 1234, default_language_tag: fr}]
				attempts:
					- {participant_id: 2, id: 1, root_item_id: 1234}`,
			wantAPIError: service.NoError,
		},
		{
			name: "parent group is a team with attempts for the given item requiring explicit entry, but the user is not on teams",
			fixture: `
				groups:
					- {id: 1, is_public: 1, type: "Team"}
					- {id: 2, type: Team}
					- {id: 3, type: "Team"}
					- {id: 4, type: "Class"}
					- {id: 5, type: "Friends"}
					- {id: 6, type: "Other"}
					- {id: 7, type: "Club"}
					- {id: 10, type: User}
					- {id: 11, type: User}
					- {id: 12, type: User}
					- {id: 13, type: User}
					- {id: 14, type: User}
					- {id: 15, type: User}
				groups_groups:
					- {parent_group_id: 4, child_group_id: 10}
					- {parent_group_id: 5, child_group_id: 10}
					- {parent_group_id: 6, child_group_id: 10}
					- {parent_group_id: 7, child_group_id: 10}
				items: [{id: 1234, default_language_tag: fr}]
				attempts:
					- {participant_id: 1, id: 1, root_item_id: 1234}
					- {participant_id: 4, id: 1, root_item_id: 1234}`,
			wantAPIError: service.NoError,
		},
		{
			name: "parent group is a team with attempts, but the user is on teams with attempts for other items requiring explicit entry",
			fixture: `
				groups:
					- {id: 1, is_public: 1, type: "Team"}
					- {id: 2, type: Team}
					- {id: 3, type: Team}
					- {id: 4, type: Team}
					- {id: 5, type: Team}
					- {id: 10, type: User}
				groups_groups:
					- {parent_group_id: 2, child_group_id: 10}
					- {parent_group_id: 3, child_group_id: 10}
					- {parent_group_id: 4, child_group_id: 10}
					- {parent_group_id: 5, child_group_id: 10}
				items:
					- {id: 1234, default_language_tag: fr}
					- {id: 2345, default_language_tag: fr}
				attempts:
					- {participant_id: 1, id: 1, root_item_id: 1234}
					- {participant_id: 2, id: 1, root_item_id: 2345}`,
			wantAPIError: service.NoError,
		},
		{
			name: "parent group is a team with attempts and the user is on a team with attempts for the same item requiring explicit entry",
			fixture: `
				groups:
					- {id: 1, is_public: 1, type: "Team"}
					- {id: 2, type: Team}
					- {id: 10, type: User}
				items: [{id: 1234, default_language_tag: fr}]
				groups_groups:
					- {parent_group_id: 2, child_group_id: 10}
				attempts:
					- {participant_id: 1, id: 1, root_item_id: 1234}
					- {participant_id: 2, id: 1, root_item_id: 1234}`,
			wantAPIError: service.ErrUnprocessableEntity(errors.New("team's participations are in conflict with the user's participations")),
		},
		{
			name: "parent group is a team with attempts and the user is on a team with attempts " +
				"for the same item requiring explicit entry (the item allows multiple attempts)",
			fixture: `
				groups:
					- {id: 1, is_public: 1, type: "Team"}
					- {id: 2, type: Team}
					- {id: 10, type: User}
				items: [{id: 1234, default_language_tag: fr, allows_multiple_attempts: 1}]
				groups_groups:
					- {parent_group_id: 2, child_group_id: 10}
				attempts:
					- {participant_id: 1, id: 1, root_item_id: 1234}
					- {participant_id: 2, id: 1, root_item_id: 1234}`,
			wantAPIError: service.ErrUnprocessableEntity(errors.New("team's participations are in conflict with the user's participations")),
		},
		{
			name: "parent group is a team with attempts and the user is on a team with expired attempts for " +
				"the same item requiring explicit entry (the item allows multiple attempts)",
			fixture: `
				groups:
					- {id: 1, is_public: 1, type: "Team"}
					- {id: 2, type: Team}
					- {id: 10, type: User}
				items: [{id: 1234, default_language_tag: fr, allows_multiple_attempts: 1}]
				groups_groups:
					- {parent_group_id: 2, child_group_id: 10}
				attempts:
					- {participant_id: 1, id: 1, root_item_id: 1234}
					- {participant_id: 2, id: 1, root_item_id: 1234, allows_submissions_until: 2019-05-30 11:00:00}`,
			wantAPIError: service.NoError,
		},
		{
			name: "parent group is a team with expired attempts and the user is on a team with expired attempts for " +
				"the same item requiring explicit entry (the item allows multiple attempts)",
			fixture: `
				groups:
					- {id: 1, is_public: 1, type: "Team"}
					- {id: 2, type: Team}
					- {id: 10, type: User}
				items: [{id: 1234, default_language_tag: fr, allows_multiple_attempts: 1}]
				groups_groups:
					- {parent_group_id: 2, child_group_id: 10}
				attempts:
					- {participant_id: 1, id: 1, root_item_id: 1234, allows_submissions_until: 2019-05-30 11:00:00}
					- {participant_id: 2, id: 1, root_item_id: 1234, allows_submissions_until: 2019-05-30 11:00:00}`,
			wantAPIError: service.NoError,
		},
		{
			name: "parent group is a team with expired attempts and the user is on a team with expired attempts " +
				"for the same item requiring explicit entry (the item doesn't allow multiple attempts)",
			fixture: `
				groups:
					- {id: 1, is_public: 1, type: "Team"}
					- {id: 2, type: Team}
					- {id: 10, type: User}
				items: [{id: 1234, default_language_tag: fr, allows_multiple_attempts: 0}]
				groups_groups:
					- {parent_group_id: 2, child_group_id: 10}
				attempts:
					- {participant_id: 1, id: 1, root_item_id: 1234, allows_submissions_until: 2019-05-30 11:00:00}
					- {participant_id: 2, id: 1, root_item_id: 1234, allows_submissions_until: 2019-05-30 11:00:00}`,
			wantAPIError: service.ErrUnprocessableEntity(errors.New("team's participations are in conflict with the user's participations")),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			testoutput.SuppressIfPasses(t)

			db := testhelpers.SetupDBWithFixtureString(tt.fixture)
			defer func() { _ = db.Close() }()

			store := database.NewDataStore(db)
			var apiError *service.APIError
			assert.NoError(t, store.InTransaction(func(transactionStore *database.DataStore) error {
				apiError = currentuser.CheckPreconditionsForGroupRequests(transactionStore,
					&database.User{GroupID: 10}, 1, "createJoinRequest")
				return nil
			}))
			assert.Equal(t, tt.wantAPIError, apiError)
		})
	}
}
