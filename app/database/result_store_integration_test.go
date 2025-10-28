//go:build !unit

package database_test

import (
	"testing"
	_ "unsafe"

	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/golang"
	"github.com/France-ioi/AlgoreaBackend/v2/testhelpers"
	"github.com/France-ioi/AlgoreaBackend/v2/testhelpers/testoutput"
)

func TestResultStore_GetHintsInfoForActiveAttempt(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db := testhelpers.SetupDBWithFixtureString(testhelpers.CreateTestContext(), `
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
		{
			name: "with info", participantID: 11, attemptID: 2, itemID: 12,
			wantHintsInfo: &database.HintsInfo{
				HintsRequested: golang.Ptr(`[0,1,"hint",null]`),
				HintsCached:    4,
			},
		},
		{
			name: "finished", participantID: 12, attemptID: 2, itemID: 14, wantHintsInfo: nil,
			wantError: gorm.ErrRecordNotFound,
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			testoutput.SuppressIfPasses(t)

			assert.NoError(t, database.NewDataStore(db).InTransaction(func(store *database.DataStore) error {
				hintsInfo, err := store.Results().GetHintsInfoForActiveAttempt(test.participantID, test.attemptID, test.itemID)
				assert.Equal(t, test.wantHintsInfo, hintsInfo)
				assert.Equal(t, test.wantError, err)
				return nil
			}))
		})
	}
}

func Test_moveFromResultsPropagateToResultsPropagateInternal(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db := testhelpers.SetupDBWithFixtureString(testhelpers.CreateTestContext(), `
		groups: [{id: 101}, {id: 102}]
		attempts: [{participant_id: 11, id: 1}, {participant_id: 12, id: 2}]
		items:
			- {id: 12, default_language_tag: fr}
			- {id: 13, default_language_tag: fr}
			- {id: 14, default_language_tag: fr}
			- {id: 15, default_language_tag: fr}
			- {id: 404, default_language_tag: fr}
		results:
			- {participant_id: 101, attempt_id: 1, item_id: 12}
			- {participant_id: 102, attempt_id: 2, item_id: 12}
			- {participant_id: 101, attempt_id: 1, item_id: 13}
			- {participant_id: 102, attempt_id: 2, item_id: 13}
			- {participant_id: 101, attempt_id: 1, item_id: 14}
			- {participant_id: 102, attempt_id: 2, item_id: 14}
			- {participant_id: 101, attempt_id: 1, item_id: 15}
			- {participant_id: 102, attempt_id: 2, item_id: 15}
			- {participant_id: 102, attempt_id: 2, item_id: 16}
		results_propagate:
			- {participant_id: 101, attempt_id: 1, item_id: 12, state: 'to_be_propagated'}
			- {participant_id: 102, attempt_id: 2, item_id: 12, state: 'to_be_recomputed'}
			- {participant_id: 101, attempt_id: 1, item_id: 13, state: 'to_be_propagated'}
			- {participant_id: 102, attempt_id: 2, item_id: 13, state: 'to_be_recomputed'}
			- {participant_id: 101, attempt_id: 1, item_id: 14, state: 'to_be_propagated'}
			- {participant_id: 102, attempt_id: 2, item_id: 14, state: 'to_be_recomputed'}
			- {participant_id: 101, attempt_id: 1, item_id: 15, state: 'to_be_propagated'}
			- {participant_id: 102, attempt_id: 2, item_id: 15, state: 'to_be_recomputed'}
			- {participant_id: 101, attempt_id: 1, item_id: 404, state: 'to_be_propagated'}
			- {participant_id: 101, attempt_id: 404, item_id: 12, state: 'to_be_recomputed'}
			- {participant_id: 404, attempt_id: 1, item_id: 12, state: 'to_be_recomputed'}
		results_propagate_internal:
			- {participant_id: 101, attempt_id: 1, item_id: 13, state: 'to_be_propagated'}
			- {participant_id: 102, attempt_id: 2, item_id: 13, state: 'to_be_propagated'}
			- {participant_id: 101, attempt_id: 1, item_id: 14, state: 'to_be_recomputed'}
			- {participant_id: 102, attempt_id: 2, item_id: 14, state: 'to_be_recomputed'}
			- {participant_id: 101, attempt_id: 1, item_id: 15, state: 'propagating'}
			- {participant_id: 102, attempt_id: 2, item_id: 15, state: 'propagating'}
			- {participant_id: 102, attempt_id: 2, item_id: 16, state: 'to_be_propagated'}
	`)
	defer func() { _ = db.Close() }()

	store := database.NewDataStore(db)
	result := moveFromResultsPropagateToResultsPropagateInternal(store)
	assert.True(t, result)

	assertResultsMarkedAsChanged(t, store, "results_propagate", nil)
	assertResultsMarkedAsChanged(t, store, "results_propagate_internal", []resultPrimaryKeyAndState{
		{ResultPrimaryKey: ResultPrimaryKey{101, 1, 12}, State: "to_be_propagated"},
		{ResultPrimaryKey: ResultPrimaryKey{102, 2, 12}, State: "to_be_recomputed"},
		{ResultPrimaryKey: ResultPrimaryKey{101, 1, 13}, State: "to_be_propagated"},
		{ResultPrimaryKey: ResultPrimaryKey{102, 2, 13}, State: "to_be_recomputed"},
		{ResultPrimaryKey: ResultPrimaryKey{101, 1, 14}, State: "to_be_recomputed"},
		{ResultPrimaryKey: ResultPrimaryKey{102, 2, 14}, State: "to_be_recomputed"},
		{ResultPrimaryKey: ResultPrimaryKey{101, 1, 15}, State: "to_be_propagated"},
		{ResultPrimaryKey: ResultPrimaryKey{102, 2, 15}, State: "to_be_recomputed"},
		{ResultPrimaryKey: ResultPrimaryKey{102, 2, 16}, State: "to_be_propagated"},
	})
}

func Test_moveFromResultsPropagateToResultsPropagateInternal_NothingMoved(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db := testhelpers.SetupDBWithFixtureString(testhelpers.CreateTestContext(), `
		results_propagate:
			- {participant_id: 101, attempt_id: 1, item_id: 404, state: 'to_be_propagated'}
			- {participant_id: 101, attempt_id: 404, item_id: 12, state: 'to_be_recomputed'}
			- {participant_id: 404, attempt_id: 1, item_id: 12, state: 'to_be_recomputed'}
	`)
	defer func() { _ = db.Close() }()

	store := database.NewDataStore(db)
	result := moveFromResultsPropagateToResultsPropagateInternal(store)
	assert.False(t, result)

	assertResultsMarkedAsChanged(t, store, "results_propagate", nil)
	assertResultsMarkedAsChanged(t, store, "results_propagate_internal", nil)
}

func TestResultStore_Propagate(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	tests := []struct {
		name    string
		wantErr bool
	}{
		{name: "basic", wantErr: false},
	}

	db := testhelpers.SetupDBWithFixture(testhelpers.CreateTestContext(), "results_propagation/main")
	defer func() { _ = db.Close() }()

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			testoutput.SuppressIfPasses(t)

			err := database.NewDataStore(db).InTransaction(func(s *database.DataStore) error {
				return s.Results().Propagate()
			})
			if (err != nil) != tt.wantErr {
				t.Errorf("ResultStore.propagate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Works locally but fails twice for every run on CI, losing 20 minutes each time.
// Comment for now until the current emergency is over.
// func TestResultStore_Propagate_Concurrent(t *testing.T) {
//	db := testhelpers.SetupDBWithFixture(testhelpers.CreateTestContext(), "results_propagation/main")
//	defer func() { _ = db.Close() }()
//
//	testhelpers.RunConcurrently(func() {
//		s := database.NewDataStoreWithContext(db.GetContext(), db)
//		err := s.InTransaction(func(st *database.DataStore) error {
//			st.ScheduleResultsPropagation()
//			return nil
//		})
//		assert.NoError(t, err)
//	}, 30)
// }

//go:linkname moveFromResultsPropagateToResultsPropagateInternal github.com/France-ioi/AlgoreaBackend/v2/app/database.moveFromResultsPropagateToResultsPropagateInternal
func moveFromResultsPropagateToResultsPropagateInternal(store *database.DataStore) bool
