//go:build !unit

package database_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/testhelpers"
)

func TestUser_CanSeeAnswer(t *testing.T) {
	tests := []struct {
		name           string
		userID         int64
		participantID  int64
		itemID         int64
		expectedResult bool
	}{
		{
			name:           "okay (full access)",
			userID:         111,
			participantID:  111,
			itemID:         50,
			expectedResult: true,
		},
		{
			name:           "okay (content access)",
			userID:         101,
			participantID:  101,
			itemID:         50,
			expectedResult: true,
		},
		{
			name:           "okay (a team member)",
			userID:         101,
			participantID:  102,
			itemID:         60,
			expectedResult: true,
		},
		{
			name:           "user not found",
			userID:         404,
			participantID:  121,
			itemID:         50,
			expectedResult: false,
		},
		{
			name:           "user doesn't have access to the item",
			userID:         121,
			participantID:  121,
			itemID:         50,
			expectedResult: false,
		},
		{
			name:           "wrong item",
			userID:         101,
			participantID:  101,
			itemID:         51,
			expectedResult: false,
		},
		{
			name:           "user is not a member of the team",
			userID:         101,
			participantID:  103,
			itemID:         60,
			expectedResult: false,
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			db := testhelpers.SetupDBWithFixtureString(`
				groups: [{id: 101}, {id: 111}, {id: 121}]
				users:
					- {login: "john", group_id: 101}
					- {login: "jane", group_id: 111}
					- {login: "guest", group_id: 121}
				groups_groups:
					- {parent_group_id: 102, child_group_id: 101}
				groups_ancestors:
					- {ancestor_group_id: 101, child_group_id: 101}
					- {ancestor_group_id: 102, child_group_id: 101}
					- {ancestor_group_id: 102, child_group_id: 102}
					- {ancestor_group_id: 111, child_group_id: 111}
					- {ancestor_group_id: 121, child_group_id: 121}
				languages: [{tag: fr}]
				items:
					- {id: 10, default_language_tag: fr}
					- {id: 50, default_language_tag: fr}
					- {id: 60, default_language_tag: fr}
				permissions_generated:
					- {group_id: 101, item_id: 50, can_view_generated: content}
					- {group_id: 101, item_id: 60, can_view_generated: content}
					- {group_id: 111, item_id: 50, can_view_generated: content_with_descendants}
					- {group_id: 121, item_id: 50, can_view_generated: info}`)
			defer func() { _ = db.Close() }()
			store := database.NewDataStore(db)
			user := &database.User{}
			assert.NoError(t, user.LoadByID(store, test.userID))

			canSeeAnswer := user.CanSeeAnswer(store, test.participantID, test.itemID)
			assert.Equal(t, test.expectedResult, canSeeAnswer)
		})
	}
}
