// +build !unit

package database_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/testhelpers"
)

func TestGroupAttemptStore_CreateNew(t *testing.T) {
	db := testhelpers.SetupDBWithFixtureString(`
		groups_attempts:
			- {id: 1, group_id: 10, item_id: 20, order: 1}
			- {id: 2, group_id: 10, item_id: 30, order: 3}
			- {id: 3, group_id: 20, item_id: 20, order: 4}`)
	defer func() { _ = db.Close() }()

	var newID int64
	var err error
	assert.NoError(t, database.NewDataStore(db).InTransaction(func(store *database.DataStore) error {
		newID, err = store.GroupAttempts().CreateNew(10, 20)
		return err
	}))
	assert.True(t, newID > 0)
	type resultType struct {
		GroupID             int64
		ItemID              int64
		StartDateSet        bool
		LastActivityDateSet bool
		Order               int32
	}
	var result resultType
	assert.NoError(t, database.NewDataStore(db).GroupAttempts().ByID(newID).
		Select(`
			group_id, item_id, ABS(TIMESTAMPDIFF(SECOND, start_date, NOW())) < 3 AS start_date_set,
			ABS(TIMESTAMPDIFF(SECOND, last_activity_date, NOW())) < 3 AS last_activity_date_set, `+"`order`").
		Take(&result).Error())
	assert.Equal(t, resultType{
		GroupID:             10,
		ItemID:              20,
		StartDateSet:        true,
		LastActivityDateSet: true,
		Order:               2,
	}, result)
}

func TestGroupAttemptStore_GetAttemptItemIDIfUserHasAccess(t *testing.T) {
	tests := []struct {
		name           string
		fixture        string
		attemptID      int64
		userID         int64
		expectedFound  bool
		expectedItemID int64
	}{
		{
			name: "okay (full access)",
			fixture: `
				users_items: [{user_id: 11, item_id: 50}]
				groups_attempts: [{id: 100, group_id: 111, item_id: 50, order: 0}]`,
			attemptID:      100,
			userID:         11,
			expectedFound:  true,
			expectedItemID: 50,
		},
		{
			name: "okay (partial access)",
			fixture: `
				users_items: [{user_id: 10, item_id: 50}]
				groups_attempts: [{id: 100, group_id: 101, item_id: 50, order: 0}]`,
			attemptID:      100,
			userID:         10,
			expectedFound:  true,
			expectedItemID: 50,
		},
		{
			name:      "okay (has_attempts=1, groups_groups.type=requestAccepted)",
			userID:    10,
			attemptID: 200,
			fixture: `
				users_items:
					- {user_id: 10, item_id: 60}
				groups_attempts:
					- {id: 200, group_id: 102, item_id: 60, order: 0}`,
			expectedFound:  true,
			expectedItemID: 60,
		},
		{
			name:      "okay (has_attempts=1, groups_groups.type=joinedByCode)",
			userID:    10,
			attemptID: 200,
			fixture: `
				users_items:
					- {user_id: 10, item_id: 60}
				groups_attempts:
					- {id: 200, group_id: 120, item_id: 60, order: 0}`,
			expectedFound:  true,
			expectedItemID: 60,
		},
		{
			name:      "okay (has_attempts=1, groups_groups.type=invitationAccepted)",
			userID:    10,
			attemptID: 200,
			fixture: `
				users_items:
					- {user_id: 10, item_id: 60}
				groups_attempts:
					- {id: 200, group_id: 110, item_id: 60, order: 0}`,
			expectedFound:  true,
			expectedItemID: 60,
		},
		{
			name:          "user not found",
			fixture:       `groups_attempts: [{id: 100, group_id: 121, item_id: 50, order: 0}]`,
			userID:        404,
			attemptID:     100,
			expectedFound: false,
		},
		{
			name:      "user doesn't have access to the item",
			userID:    12,
			attemptID: 100,
			fixture: `
				users_items: [{user_id: 12, item_id: 50}]
				groups_attempts: [{id: 100, group_id: 121, item_id: 50, order: 0}]`,
			expectedFound: false,
		},
		{
			name:          "no groups_attempts",
			userID:        10,
			attemptID:     100,
			fixture:       `users_items: [{user_id: 10, item_id: 50}]`,
			expectedFound: false,
		},
		{
			name:      "wrong item in groups_attempts",
			userID:    10,
			attemptID: 100,
			fixture: `
				users_items: [{user_id: 10, item_id: 50}]
				groups_attempts: [{id: 100, group_id: 101, item_id: 51, order: 0}]`,
			expectedFound: false,
		},
		{
			name:      "no users_items",
			userID:    10,
			attemptID: 100,
			fixture: `
				users_items:
					- {user_id: 10, item_id: 51}
					- {user_id: 11, item_id: 50}
				groups_attempts: [{id: 100, group_id: 101, item_id: 50, order: 0}]`,
			expectedFound: false,
		},
		{
			name:      "user is not a member of the team (invitationSent)",
			userID:    10,
			attemptID: 100,
			fixture: `
				users_items: [{user_id: 10, item_id: 60}]
				groups_attempts: [{id: 100, group_id: 103, item_id: 60, order: 0}]`,
			expectedFound: false,
		},
		{
			name:      "user is not a member of the team (requestSent)",
			userID:    10,
			attemptID: 100,
			fixture: `
				users_items: [{user_id: 10, item_id: 60}]
				groups_attempts: [{id: 100, group_id: 104, item_id: 60, order: 0}]`,
			expectedFound: false,
		},
		{
			name:      "user is not a member of the team (invitationRefused)",
			userID:    10,
			attemptID: 100,
			fixture: `
				users_items: [{user_id: 10, item_id: 60}]
				groups_attempts: [{id: 100, group_id: 105, item_id: 60, order: 0}]`,
			expectedFound: false,
		},
		{
			name:      "user is not a member of the team (requestRefused)",
			userID:    10,
			attemptID: 100,
			fixture: `
				users_items: [{user_id: 10, item_id: 60}]
				groups_attempts: [{id: 100, group_id: 106, item_id: 60, order: 0}]`,
			expectedFound: false,
		},
		{
			name:      "user is not a member of the team (removed)",
			userID:    10,
			attemptID: 100,
			fixture: `
				users_items: [{user_id: 10, item_id: 60}]
				groups_attempts: [{id: 100, group_id: 107, item_id: 60, order: 0}]`,
			expectedFound: false,
		},
		{
			name:      "user is not a member of the team (left)",
			userID:    10,
			attemptID: 100,
			fixture: `
				users_items: [{user_id: 10, item_id: 60}]
				groups_attempts: [{id: 100, group_id: 108, item_id: 60, order: 0}]`,
			expectedFound: false,
		},
		{
			name:      "user is a member of the team (direct)",
			userID:    10,
			attemptID: 100,
			fixture: `
				users_items: [{user_id: 10, item_id: 60}]
				groups_attempts: [{id: 100, group_id: 109, item_id: 60, order: 0}]`,
			expectedFound:  true,
			expectedItemID: 60,
		},
		{
			name:      "groups_attempts.group_id is not user's self group",
			userID:    10,
			attemptID: 100,
			fixture: `
				users_items: [{user_id: 10, item_id: 50}]
				groups_attempts: [{id: 100, group_id: 102, item_id: 50, order: 0}]`,
			expectedFound: false,
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			db := testhelpers.SetupDBWithFixtureString(`
				users:
					- {id: 10, login: "john", group_self_id: 101}
					- {id: 11, login: "jane", group_self_id: 111}
					- {id: 12, login: "guest", group_self_id: 121}
				groups_groups:
					- {group_parent_id: 102, group_child_id: 101, type: requestAccepted}
					- {group_parent_id: 103, group_child_id: 101, type: invitationSent}
					- {group_parent_id: 104, group_child_id: 101, type: requestSent}
					- {group_parent_id: 105, group_child_id: 101, type: invitationRefused}
					- {group_parent_id: 106, group_child_id: 101, type: requestRefused}
					- {group_parent_id: 107, group_child_id: 101, type: removed}
					- {group_parent_id: 108, group_child_id: 101, type: left}
					- {group_parent_id: 109, group_child_id: 101, type: direct}
					- {group_parent_id: 110, group_child_id: 101, type: invitationAccepted}
					- {group_parent_id: 120, group_child_id: 101, type: joinedByCode}
				groups_ancestors:
					- {group_ancestor_id: 101, group_child_id: 101, is_self: 1}
					- {group_ancestor_id: 102, group_child_id: 101, is_self: 0}
					- {group_ancestor_id: 102, group_child_id: 102, is_self: 1}
					- {group_ancestor_id: 111, group_child_id: 111, is_self: 1}
					- {group_ancestor_id: 121, group_child_id: 121, is_self: 1}
					- {group_ancestor_id: 109, group_child_id: 101, is_self: 0}
					- {group_ancestor_id: 109, group_child_id: 109, is_self: 1}
					- {group_ancestor_id: 110, group_child_id: 101, is_self: 0}
					- {group_ancestor_id: 110, group_child_id: 110, is_self: 1}
					- {group_ancestor_id: 120, group_child_id: 101, is_self: 0}
					- {group_ancestor_id: 120, group_child_id: 120, is_self: 1}
				items:
					- {id: 10, has_attempts: 0}
					- {id: 50, has_attempts: 0}
					- {id: 60, has_attempts: 1}
				groups_items:
					- {group_id: 101, item_id: 50, cached_partial_access_date: "2017-05-29 06:38:38", user_created_id: 1}
					- {group_id: 101, item_id: 60, cached_partial_access_date: "2017-05-29 06:38:38", user_created_id: 1}
					- {group_id: 111, item_id: 50, cached_full_access_date: "2017-05-29 06:38:38", user_created_id: 1}
					- {group_id: 121, item_id: 50, cached_grayed_access_date: "2017-05-29 06:38:38", user_created_id: 1}`,
				test.fixture)
			defer func() { _ = db.Close() }()
			store := database.NewDataStore(db)
			user := &database.User{}
			assert.NoError(t, user.LoadByID(store, test.userID))
			found, itemID, err := store.GroupAttempts().GetAttemptItemIDIfUserHasAccess(test.attemptID, user)
			assert.NoError(t, err)
			assert.Equal(t, test.expectedFound, found)
			assert.Equal(t, test.expectedItemID, itemID)
		})
	}
}

func TestGroupAttemptStore_VisibleAndByItemID(t *testing.T) {
	tests := []struct {
		name        string
		fixture     string
		attemptID   int64
		userID      int64
		itemID      int64
		expectedIDs []int64
		expectedErr error
	}{
		{
			name: "okay (full access)",
			fixture: `groups_attempts:
			                - {id: 100, group_id: 111, item_id: 50, order: 0}
			                - {id: 101, group_id: 111, item_id: 50, order: 1}
			                - {id: 102, group_id: 111, item_id: 70, order: 0}`,
			attemptID:   100,
			userID:      11,
			expectedIDs: []int64{100, 101},
			itemID:      50,
		},
		{
			name:        "okay (partial access)",
			fixture:     `groups_attempts: [{id: 100, group_id: 101, item_id: 50, order: 0}]`,
			attemptID:   100,
			userID:      10,
			expectedIDs: []int64{100},
			itemID:      50,
		},
		{
			name:        "okay (has_attempts=1, groups_groups.type=requestAccepted)",
			userID:      10,
			attemptID:   200,
			fixture:     `groups_attempts: [{id: 200, group_id: 102, item_id: 60, order: 0},{id: 201, group_id: 102, item_id: 60, order: 1}]`,
			expectedIDs: []int64{200, 201},
			itemID:      60,
		},
		{
			name:        "okay (has_attempts=1, groups_groups.type=joinedByCode)",
			userID:      10,
			attemptID:   200,
			fixture:     `groups_attempts: [{id: 200, group_id: 120, item_id: 60, order: 0},{id: 201, group_id: 120, item_id: 60, order: 1}]`,
			expectedIDs: []int64{200, 201},
			itemID:      60,
		},
		{
			name:        "okay (has_attempts=1, groups_groups.type=invitationAccepted)",
			userID:      10,
			attemptID:   200,
			fixture:     `groups_attempts: [{id: 200, group_id: 110, item_id: 60, order: 0}]`,
			expectedIDs: []int64{200},
			itemID:      60,
		},
		{
			name:        "user not found",
			fixture:     `groups_attempts: [{id: 100, group_id: 121, item_id: 50, order: 0}]`,
			userID:      404,
			attemptID:   100,
			expectedIDs: []int64(nil),
		},
		{
			name:        "user doesn't have access to the item",
			userID:      12,
			attemptID:   100,
			fixture:     `groups_attempts: [{id: 100, group_id: 121, item_id: 50, order: 0}]`,
			expectedIDs: []int64(nil),
		},
		{
			name:        "no groups_attempts",
			userID:      10,
			attemptID:   100,
			fixture:     "",
			expectedIDs: nil,
		},
		{
			name:        "wrong item in groups_attempts",
			userID:      10,
			attemptID:   100,
			fixture:     `groups_attempts: [{id: 100, group_id: 101, item_id: 51, order: 0}]`,
			expectedIDs: nil,
		},
		{
			name:        "user is not a member of the team (invitationSent)",
			userID:      10,
			attemptID:   100,
			fixture:     `groups_attempts: [{id: 100, group_id: 103, item_id: 60, order: 0}]`,
			expectedIDs: nil,
		},
		{
			name:        "user is not a member of the team (requestSent)",
			userID:      10,
			attemptID:   100,
			fixture:     `groups_attempts: [{id: 100, group_id: 104, item_id: 60, order: 0}]`,
			expectedIDs: nil,
		},
		{
			name:        "user is not a member of the team (invitationRefused)",
			userID:      10,
			attemptID:   100,
			fixture:     `groups_attempts: [{id: 100, group_id: 105, item_id: 60, order: 0}]`,
			expectedIDs: nil,
		},
		{
			name:        "user is not a member of the team (requestRefused)",
			userID:      10,
			attemptID:   100,
			fixture:     `groups_attempts: [{id: 100, group_id: 106, item_id: 60, order: 0}]`,
			expectedIDs: nil,
		},
		{
			name:        "user is not a member of the team (removed)",
			userID:      10,
			attemptID:   100,
			fixture:     `groups_attempts: [{id: 100, group_id: 107, item_id: 60, order: 0}]`,
			expectedIDs: nil,
		},
		{
			name:        "user is not a member of the team (left)",
			userID:      10,
			attemptID:   100,
			fixture:     `groups_attempts: [{id: 100, group_id: 108, item_id: 60, order: 0}]`,
			expectedIDs: nil,
		},
		{
			name:        "user is not a member of the team (direct)",
			userID:      10,
			attemptID:   100,
			fixture:     `groups_attempts: [{id: 100, group_id: 109, item_id: 60, order: 0}]`,
			expectedIDs: nil,
		},
		{
			name:        "groups_attempts.group_id is not user's self group",
			userID:      10,
			attemptID:   100,
			fixture:     `groups_attempts: [{id: 100, group_id: 102, item_id: 50, order: 0}]`,
			expectedIDs: nil,
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			db := testhelpers.SetupDBWithFixtureString(`
				users:
					- {id: 10, login: "john", group_self_id: 101}
					- {id: 11, login: "jane", group_self_id: 111}
					- {id: 12, login: "guest", group_self_id: 121}
				groups_groups:
					- {group_parent_id: 102, group_child_id: 101, type: requestAccepted}
					- {group_parent_id: 103, group_child_id: 101, type: invitationSent}
					- {group_parent_id: 104, group_child_id: 101, type: requestSent}
					- {group_parent_id: 105, group_child_id: 101, type: invitationRefused}
					- {group_parent_id: 106, group_child_id: 101, type: requestRefused}
					- {group_parent_id: 107, group_child_id: 101, type: removed}
					- {group_parent_id: 108, group_child_id: 101, type: left}
					- {group_parent_id: 109, group_child_id: 101, type: direct}
					- {group_parent_id: 110, group_child_id: 101, type: invitationAccepted}
					- {group_parent_id: 120, group_child_id: 101, type: joinedByCode}
				groups_ancestors:
					- {group_ancestor_id: 101, group_child_id: 101, is_self: 1}
					- {group_ancestor_id: 102, group_child_id: 101, is_self: 0}
					- {group_ancestor_id: 102, group_child_id: 102, is_self: 1}
					- {group_ancestor_id: 111, group_child_id: 111, is_self: 1}
					- {group_ancestor_id: 121, group_child_id: 121, is_self: 1}
					- {group_ancestor_id: 109, group_child_id: 101, is_self: 0}
					- {group_ancestor_id: 109, group_child_id: 109, is_self: 1}
					- {group_ancestor_id: 110, group_child_id: 101, is_self: 0}
					- {group_ancestor_id: 110, group_child_id: 110, is_self: 1}
					- {group_ancestor_id: 120, group_child_id: 101, is_self: 0}
					- {group_ancestor_id: 120, group_child_id: 120, is_self: 1}
				items:
					- {id: 10, has_attempts: 0}
					- {id: 50, has_attempts: 0}
					- {id: 60, has_attempts: 1}
					- {id: 70, has_attempts: 0}
				groups_items:
					- {group_id: 101, item_id: 50, cached_partial_access_date: "2017-05-29 06:38:38", user_created_id: 10}
					- {group_id: 101, item_id: 60, cached_partial_access_date: "2017-05-29 06:38:38", user_created_id: 10}
					- {group_id: 101, item_id: 70, cached_partial_access_date: "2017-05-29 06:38:38", user_created_id: 10}
					- {group_id: 111, item_id: 50, cached_full_access_date: "2017-05-29 06:38:38", user_created_id: 10}
					- {group_id: 111, item_id: 70, cached_full_access_date: "2017-05-29 06:38:38", user_created_id: 10}
					- {group_id: 121, item_id: 50, cached_grayed_access_date: "2017-05-29 06:38:38", user_created_id: 10}
					- {group_id: 121, item_id: 70, cached_grayed_access_date: "2017-05-29 06:38:38", user_created_id: 10}`,
				test.fixture)
			defer func() { _ = db.Close() }()
			store := database.NewDataStore(db)
			user := &database.User{}
			assert.NoError(t, user.LoadByID(store, test.userID))
			var ids []int64
			err := store.GroupAttempts().VisibleAndByItemID(user, test.itemID).Pluck("groups_attempts.id", &ids).Error()
			assert.Equal(t, test.expectedErr, err)
			assert.Equal(t, test.expectedIDs, ids)
		})
	}
}
