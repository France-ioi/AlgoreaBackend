// +build !unit

package database_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/testhelpers"
)

type resultType struct {
	ParticipantID    int64
	AttemptID        int64
	ItemID           int64
	StartedAt        *database.Time
	LatestActivityAt database.Time
}

type attemptType struct {
	ParticipantID   int64
	ID              int64
	ParentAttemptID *int64
	RootItemID      *int64
	CreatorID       int64
	CreatedAt       *database.Time
}

func TestAttemptStore_CreateNew_CreatesNewResult(t *testing.T) {
	db := testhelpers.SetupDBWithFixtureString(`
		groups:
			- {id: 10}
			- {id: 100}
		users:
			- {group_id: 100}
		items: [{id: 20, default_language_tag: fr}, {id: 30, default_language_tag: fr}]
		attempts:
			- {id: 0, participant_id: 10}
			- {id: 0, participant_id: 20}
		results:
			- {attempt_id: 0, participant_id: 10, item_id: 30}
			- {attempt_id: 0, participant_id: 20, item_id: 20}`)
	defer func() { _ = db.Close() }()

	testhelpers.MockDBTime("2019-05-30 11:00:00")
	defer testhelpers.RestoreDBTime()

	var newAttemptID int64
	var err error
	assert.NoError(t, database.NewDataStore(db).InTransaction(func(store *database.DataStore) error {
		newAttemptID, err = store.Attempts().CreateNew(10, 20, 100)
		return err
	}))
	assert.Equal(t, int64(0), newAttemptID)
	var result resultType
	expectedTime := database.Time(time.Date(2019, 5, 30, 11, 0, 0, 0, time.UTC))
	assert.NoError(t, database.NewDataStore(db).Results().
		Where("attempt_id = ?", newAttemptID).
		Where("participant_id = ?", 10).
		Select("participant_id, attempt_id, item_id, started_at, latest_activity_at").Take(&result).Error())
	assert.Equal(t, resultType{
		ParticipantID:    10,
		AttemptID:        0,
		ItemID:           20,
		StartedAt:        &expectedTime,
		LatestActivityAt: expectedTime,
	}, result)
}

func TestAttemptStore_CreateNew_CreatesNewAttemptWhenResultAlreadyExists(t *testing.T) {
	db := testhelpers.SetupDBWithFixtureString(`
		groups:
			- {id: 10}
			- {id: 100}
		users:
			- {group_id: 100}
		items: [{id: 20, default_language_tag: fr}, {id: 30, default_language_tag: fr}]
		attempts:
			- {id: 0, participant_id: 10}
			- {id: 0, participant_id: 20}
		results:
			- {attempt_id: 0, participant_id: 10, item_id: 20}
			- {attempt_id: 0, participant_id: 10, item_id: 30}
			- {attempt_id: 0, participant_id: 20, item_id: 20}`)
	defer func() { _ = db.Close() }()

	testhelpers.MockDBTime("2019-05-30 11:00:00")
	defer testhelpers.RestoreDBTime()

	var newAttemptID int64
	var err error
	assert.NoError(t, database.NewDataStore(db).InTransaction(func(store *database.DataStore) error {
		newAttemptID, err = store.Attempts().CreateNew(10, 20, 100)
		return err
	}))
	assert.Equal(t, int64(1), newAttemptID)
	var result resultType
	expectedTime := database.Time(time.Date(2019, 5, 30, 11, 0, 0, 0, time.UTC))
	assert.NoError(t, database.NewDataStore(db).Results().
		Where("attempt_id = ?", newAttemptID).
		Where("participant_id = ?", 10).
		Select("participant_id, attempt_id, item_id, started_at, latest_activity_at").Take(&result).Error())
	assert.Equal(t, resultType{
		ParticipantID:    10,
		AttemptID:        1,
		ItemID:           20,
		StartedAt:        &expectedTime,
		LatestActivityAt: expectedTime,
	}, result)
	var attempt attemptType
	assert.NoError(t, database.NewDataStore(db).Attempts().ByID(newAttemptID).
		Where("participant_id = ?", 10).
		Select("participant_id, id, creator_id, parent_attempt_id, root_item_id, created_at").Take(&attempt).Error())
	assert.Equal(t, attemptType{
		ParticipantID:   10,
		ID:              1,
		ParentAttemptID: ptrInt64(0),
		RootItemID:      ptrInt64(20),
		CreatorID:       100,
		CreatedAt:       &expectedTime,
	}, attempt)
}

func TestAttemptStore_GetAttemptParticipantIDIfUserHasAccess(t *testing.T) {
	tests := []struct {
		name                  string
		fixture               string
		attemptID             int64
		itemID                int64
		userID                int64
		expectedFound         bool
		expectedParticipantID int64
	}{
		{
			name: "okay (full access)",
			fixture: `
				attempts: [{id: 1, participant_id: 111}]
				results: [{attempt_id: 1, participant_id: 111, item_id: 50}]`,
			attemptID:             1,
			userID:                111,
			expectedFound:         true,
			itemID:                50,
			expectedParticipantID: 111,
		},
		{
			name: "okay (content access)",
			fixture: `
				attempts: [{id: 1, participant_id: 101}]
				results: [{attempt_id: 1, participant_id: 101, item_id: 50}]`,
			attemptID:             1,
			userID:                101,
			expectedFound:         true,
			itemID:                50,
			expectedParticipantID: 101,
		},
		{
			name:      "okay (as a team member)",
			userID:    101,
			attemptID: 2,
			fixture: `
				attempts: [{id: 2, participant_id: 102}]
				results: [{attempt_id: 2, participant_id: 102, item_id: 60}]`,
			expectedFound:         true,
			itemID:                60,
			expectedParticipantID: 102,
		},
		{
			name: "user not found",
			fixture: `
				attempts: [{id: 1, participant_id: 121}]
				results: [{attempt_id: 1, participant_id: 121, item_id: 50}]`,
			userID:        404,
			attemptID:     1,
			itemID:        50,
			expectedFound: false,
		},
		{
			name:      "user doesn't have access to the item",
			userID:    121,
			attemptID: 1,
			itemID:    50,
			fixture: `
				attempts: [{id: 1, participant_id: 121}]
				results: [{attempt_id: 1, participant_id: 121, item_id: 50}]`,
			expectedFound: false,
		},
		{
			name:          "no attempts",
			userID:        101,
			attemptID:     100,
			itemID:        50,
			fixture:       ``,
			expectedFound: false,
		},
		{
			name:      "wrong item",
			userID:    101,
			attemptID: 1,
			itemID:    51,
			fixture: `
				attempts: [{id: 1, participant_id: 101}]
				results: [{attempt_id: 1, participant_id: 101, item_id: 51}]`,
			expectedFound: false,
		},
		{
			name:      "user is not a member of the team",
			userID:    101,
			attemptID: 1,
			itemID:    60,
			fixture: `
				attempts: [{id: 1, participant_id: 103}]
				results: [{attempt_id: 1, participant_id: 103, item_id: 60}]`,
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
					- {ancestor_group_id: 101, child_group_id: 101}
					- {ancestor_group_id: 102, child_group_id: 101}
					- {ancestor_group_id: 102, child_group_id: 102}
					- {ancestor_group_id: 111, child_group_id: 111}
					- {ancestor_group_id: 121, child_group_id: 121}
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
			found, participantID, err := store.Attempts().GetAttemptParticipantIDIfUserHasAccess(test.attemptID, test.itemID, user)
			assert.NoError(t, err)
			assert.Equal(t, test.expectedFound, found)
			assert.Equal(t, test.expectedParticipantID, participantID)
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
