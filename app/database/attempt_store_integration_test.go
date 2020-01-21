// +build !unit

package database_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/testhelpers"
)

func TestAttemptStore_CreateNew(t *testing.T) {
	db := testhelpers.SetupDBWithFixtureString(`
		groups:
			- {id: 100}
		users:
			- {group_id: 100}
		attempts:
			- {id: 1, group_id: 10, item_id: 20, order: 1}
			- {id: 2, group_id: 10, item_id: 30, order: 3}
			- {id: 3, group_id: 20, item_id: 20, order: 4}`)
	defer func() { _ = db.Close() }()

	testhelpers.MockDBTime("2019-05-30 11:00:00")
	defer testhelpers.RestoreDBTime()

	var newID int64
	var err error
	assert.NoError(t, database.NewDataStore(db).InTransaction(func(store *database.DataStore) error {
		newID, err = store.Attempts().CreateNew(10, 20, 100)
		return err
	}))
	assert.True(t, newID > 0)
	type resultType struct {
		GroupID          int64
		ItemID           int64
		CreatorID        int64
		StartedAt        *database.Time
		LatestActivityAt database.Time
		Order            int32
	}
	var result resultType
	expectedTime := database.Time(time.Date(2019, 5, 30, 11, 0, 0, 0, time.UTC))
	assert.NoError(t, database.NewDataStore(db).Attempts().ByID(newID).
		Select("group_id, item_id, creator_id, started_at, latest_activity_at, `order`").Take(&result).Error())
	assert.Equal(t, resultType{
		GroupID:          10,
		ItemID:           20,
		CreatorID:        100,
		StartedAt:        &expectedTime,
		LatestActivityAt: expectedTime,
		Order:            2,
	}, result)
}

func TestAttemptStore_GetAttemptItemIDIfUserHasAccess(t *testing.T) {
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
				attempts: [{id: 100, group_id: 111, item_id: 50, order: 0}]`,
			attemptID:      100,
			userID:         111,
			expectedFound:  true,
			expectedItemID: 50,
		},
		{
			name: "okay (content access)",
			fixture: `
				attempts: [{id: 100, group_id: 101, item_id: 50, order: 0}]`,
			attemptID:      100,
			userID:         101,
			expectedFound:  true,
			expectedItemID: 50,
		},
		{
			name:      "okay (as a team member)",
			userID:    101,
			attemptID: 200,
			fixture: `
				attempts:
					- {id: 200, group_id: 102, item_id: 60, order: 0}`,
			expectedFound:  true,
			expectedItemID: 60,
		},
		{
			name:          "user not found",
			fixture:       `attempts: [{id: 100, group_id: 121, item_id: 50, order: 0}]`,
			userID:        404,
			attemptID:     100,
			expectedFound: false,
		},
		{
			name:      "user doesn't have access to the item",
			userID:    121,
			attemptID: 100,
			fixture: `
				attempts: [{id: 100, group_id: 121, item_id: 50, order: 0}]`,
			expectedFound: false,
		},
		{
			name:          "no attempts",
			userID:        101,
			attemptID:     100,
			fixture:       ``,
			expectedFound: false,
		},
		{
			name:      "wrong item in attempts",
			userID:    101,
			attemptID: 100,
			fixture: `
				attempts: [{id: 100, group_id: 101, item_id: 51, order: 0}]`,
			expectedFound: false,
		},
		{
			name:      "user is not a member of the team",
			userID:    101,
			attemptID: 100,
			fixture: `
				attempts: [{id: 100, group_id: 103, item_id: 60, order: 0}]`,
			expectedFound: false,
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			db := testhelpers.SetupDBWithFixtureString(`
				groups: [{id: 101}, {id: 111}, {id: 121}]`, `
				users:
					- {login: "john", group_id: 101}
					- {login: "jane", group_id: 111}
					- {login: "guest", group_id: 121}
				groups_groups:
					- {parent_group_id: 102, child_group_id: 101}
				groups_ancestors:
					- {ancestor_group_id: 101, child_group_id: 101, is_self: 1}
					- {ancestor_group_id: 102, child_group_id: 101, is_self: 0}
					- {ancestor_group_id: 102, child_group_id: 102, is_self: 1}
					- {ancestor_group_id: 111, child_group_id: 111, is_self: 1}
					- {ancestor_group_id: 121, child_group_id: 121, is_self: 1}
				items:
					- {id: 10, default_language_tag: fr}
					- {id: 50, default_language_tag: fr}
					- {id: 60, default_language_tag: fr}
				permissions_generated:
					- {group_id: 101, item_id: 50, can_view_generated: content}
					- {group_id: 101, item_id: 60, can_view_generated: content}
					- {group_id: 111, item_id: 50, can_view_generated: content_with_descendants}
					- {group_id: 121, item_id: 50, can_view_generated: info}`,
				test.fixture)
			defer func() { _ = db.Close() }()
			store := database.NewDataStore(db)
			user := &database.User{}
			assert.NoError(t, user.LoadByID(store, test.userID))
			found, itemID, err := store.Attempts().GetAttemptItemIDIfUserHasAccess(test.attemptID, user)
			assert.NoError(t, err)
			assert.Equal(t, test.expectedFound, found)
			assert.Equal(t, test.expectedItemID, itemID)
		})
	}
}

func TestAttemptStore_ComputeAllAttempts(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{name: "basic", wantErr: false},
	}

	db := testhelpers.SetupDBWithFixture("attempts_propagation/main")
	defer func() { _ = db.Close() }()

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			err := database.NewDataStore(db).InTransaction(func(s *database.DataStore) error {
				return s.Attempts().ComputeAllAttempts()
			})
			if (err != nil) != tt.wantErr {
				t.Errorf("AttemptStore.computeAllAttempts() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAttemptStore_ComputeAllAttempts_Concurrent(t *testing.T) {
	db := testhelpers.SetupDBWithFixture("attempts_propagation/main")
	defer func() { _ = db.Close() }()

	testhelpers.RunConcurrently(func() {
		s := database.NewDataStore(db)
		err := s.InTransaction(func(st *database.DataStore) error {
			return st.Attempts().ComputeAllAttempts()
		})
		assert.NoError(t, err)
	}, 30)
}
