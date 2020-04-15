// +build !unit

package database_test

import (
	"testing"

	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/testhelpers"
)

func TestResultStore_ExistsForUserTeam(t *testing.T) {
	db := testhelpers.SetupDBWithFixtureString(`
		groups: [{id: 1, type: User}, {id: 100, type: Team}, {id: 101, type: Class},
		         {id: 102, type: Team}, {id: 103, type: Team}, {id: 104, type: Team}]
		groups_groups:
			- {parent_group_id: 100, child_group_id: 1}
			- {parent_group_id: 101, child_group_id: 1}
			- {parent_group_id: 102, child_group_id: 1, expires_at: 2019-05-30 11:00:00}
			- {parent_group_id: 103, child_group_id: 1}
			- {parent_group_id: 104, child_group_id: 1}
		results:
			- {participant_id: 100, attempt_id: 200, item_id: 300}
			- {participant_id: 101, attempt_id: 200, item_id: 300}
			- {participant_id: 102, attempt_id: 200, item_id: 300}
			- {participant_id: 103, attempt_id: 200, item_id: 301}
			- {participant_id: 104, attempt_id: 201, item_id: 300}`)
	for _, test := range []struct {
		name              string
		userGroupID       int64
		participantTeamID int64
		attemptID         int64
		itemID            int64
		expectedResult    bool
	}{
		{name: "okay", userGroupID: 1, participantTeamID: 100, attemptID: 200, itemID: 300, expectedResult: true},
		{name: "no such member", userGroupID: 2, participantTeamID: 100, attemptID: 200, itemID: 300},
		{name: "participantTeamID is not a team", userGroupID: 1, participantTeamID: 101, attemptID: 200, itemID: 300},
		{name: "expired membership", userGroupID: 1, participantTeamID: 102, attemptID: 200, itemID: 300},
		{name: "item_id of results doesn't match", userGroupID: 1, participantTeamID: 103, attemptID: 200, itemID: 300},
		{name: "attempt_id of results doesn't match", userGroupID: 1, participantTeamID: 104, attemptID: 200, itemID: 300},
	} {
		test := test
		t.Run(test.name, func(t *testing.T) {
			user := &database.User{GroupID: test.userGroupID}
			found, err := database.NewDataStore(db).Results().ExistsForUserTeam(user, test.participantTeamID, test.attemptID, test.itemID)
			assert.NoError(t, err)
			assert.Equal(t, test.expectedResult, found)
		})
	}
}

func TestResultStore_GetHintsInfoForActiveAttempt(t *testing.T) {
	db := testhelpers.SetupDBWithFixtureString(`
		attempts:
			- {participant_id: 11, id: 1, root_item_id: 112, allows_submissions_until: 3019-05-30 12:00:00}
			- {participant_id: 11, id: 2, root_item_id: 112}
			- {participant_id: 12, id: 2, root_item_id: 114, allows_submissions_until: 2019-05-30 12:00:00}
		results:
			- {participant_id: 11, attempt_id: 1, item_id: 12, started_at: 2019-05-30 11:00:00}
			- {participant_id: 11, attempt_id: 2, item_id: 12, hints_requested: '[0,1,"hint",null]', hints_cached: 4,
				started_at: 2019-07-30 11:00:00}
			- {participant_id: 12, attempt_id: 2, item_id: 14}`)
	defer func() { _ = db.Close() }()

	tests := []struct {
		name          string
		participantID int64
		attemptID     int64
		itemID        int64
		wantHintsInfo *database.HintsInfo
		wantError     error
	}{
		{name: "empty info", participantID: 11, attemptID: 1, itemID: 12, wantHintsInfo: &database.HintsInfo{}},
		{name: "with info", participantID: 11, attemptID: 2, itemID: 12,
			wantHintsInfo: &database.HintsInfo{
				HintsRequested: ptrString(`[0,1,"hint",null]`),
				HintsCached:    4,
			}},
		{name: "finished", participantID: 12, attemptID: 2, itemID: 14, wantHintsInfo: nil,
			wantError: gorm.ErrRecordNotFound},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			assert.NoError(t, database.NewDataStore(db).InTransaction(func(store *database.DataStore) error {
				hintsInfo, err := store.Results().GetHintsInfoForActiveAttempt(test.participantID, test.attemptID, test.itemID)
				assert.Equal(t, test.wantHintsInfo, hintsInfo)
				assert.Equal(t, test.wantError, err)
				return nil
			}))
		})
	}
}

func TestResultStore_Propagate(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{name: "basic", wantErr: false},
	}

	db := testhelpers.SetupDBWithFixture("results_propagation/main")
	defer func() { _ = db.Close() }()

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			err := database.NewDataStore(db).InTransaction(func(s *database.DataStore) error {
				return s.Results().Propagate()
			})
			if (err != nil) != tt.wantErr {
				t.Errorf("ResultStore.Propagate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestResultStore_Propagate_Concurrent(t *testing.T) {
	db := testhelpers.SetupDBWithFixture("results_propagation/main")
	defer func() { _ = db.Close() }()

	testhelpers.RunConcurrently(func() {
		s := database.NewDataStore(db)
		err := s.InTransaction(func(st *database.DataStore) error {
			return st.Results().Propagate()
		})
		assert.NoError(t, err)
	}, 30)
}
