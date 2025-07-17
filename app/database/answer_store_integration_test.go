//go:build !unit

package database_test

import (
	"reflect"
	"testing"

	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/golang"
	"github.com/France-ioi/AlgoreaBackend/v2/testhelpers"
	"github.com/France-ioi/AlgoreaBackend/v2/testhelpers/testoutput"
)

func TestAnswerStore_SubmitNewAnswer(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db := testhelpers.SetupDBWithFixtureString(`
		groups: [{id: 121}]
		users: [{group_id: 121}]
		attempts: [{id: 56, participant_id: 121}]
		results: [{participant_id: 121, attempt_id: 56, item_id: 456}]`)
	defer func() { _ = db.Close() }()

	answerStore := database.NewDataStore(db).Answers()
	tests := []struct {
		name          string
		authorID      int64
		participantID int64
		attemptID     int64
		itemID        int64
		answer        string
	}{
		{name: "with attemptID", authorID: 121, participantID: 121, attemptID: 56, itemID: 456, answer: "my answer"},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			testoutput.SuppressIfPasses(t)

			newID, err := answerStore.SubmitNewAnswer(test.authorID, test.participantID, test.attemptID, test.itemID, test.answer)

			require.NoError(t, err)
			assert.NotZero(t, newID)

			type answer struct {
				AuthorID      int64
				ParticipantID int64
				AttemptID     int64
				ItemID        int64
				Type          string
				Answer        string
				CreatedAtSet  bool
			}
			var insertedAnswer answer
			require.NoError(t,
				answerStore.ByID(newID).
					Select("author_id, participant_id, attempt_id, item_id, type, answer, "+
						"ABS(TIMESTAMPDIFF(SECOND, created_at, NOW())) < 3 AS created_at_set").
					Scan(&insertedAnswer).Error())
			assert.Equal(t, answer{
				AuthorID:      test.authorID,
				ParticipantID: test.participantID,
				AttemptID:     test.attemptID,
				ItemID:        test.itemID,
				Type:          "Submission",
				Answer:        test.answer,
				CreatedAtSet:  true,
			}, insertedAnswer)
		})
	}
}

func TestAnswerStore_CreateNewAnswer(t *testing.T) {
	testoutput.SuppressIfPasses(t)
	db := testhelpers.SetupDBWithFixtureString(`
		groups: [{id: 121}, {id: 122}]
		users: [{group_id: 122}]
		attempts: [{id: 56, participant_id: 121}]
		results: [{participant_id: 121, attempt_id: 56, item_id: 456}]`)
	defer func() { _ = db.Close() }()

	answerStore := database.NewDataStore(db).Answers()
	newID, err := answerStore.CreateNewAnswer(int64(122), int64(121), int64(56), int64(456), "Saved", "answer", golang.Ptr("State"))
	require.NoError(t, err)
	require.NotZero(t, newID)

	type answer struct {
		AuthorID      int64
		ParticipantID int64
		AttemptID     int64
		ItemID        int64
		Type          string
		Answer        string
		State         string
		CreatedAtSet  bool
	}
	var insertedAnswer answer
	require.NoError(t,
		answerStore.ByID(newID).
			Select("author_id, participant_id, attempt_id, item_id, type, answer, state, "+
				"ABS(TIMESTAMPDIFF(SECOND, created_at, NOW())) < 3 AS created_at_set").
			Scan(&insertedAnswer).Error())
	assert.Equal(t, answer{
		AuthorID:      122,
		ParticipantID: 121,
		AttemptID:     56,
		ItemID:        456,
		State:         "State",
		Type:          "Saved",
		Answer:        "answer",
		CreatedAtSet:  true,
	}, insertedAnswer)
}

func TestAnswerStore_CreateNewAnswer_Duplicate(t *testing.T) {
	testoutput.SuppressIfPasses(t)
	db := testhelpers.SetupDBWithFixtureString(`
		groups: [{id: 121}]
		users: [{group_id: 121}]
		attempts: [{id: 56, participant_id: 121}]
		results: [{participant_id: 121, attempt_id: 56, item_id: 456}]
		answers: [{id: 1, author_id: 121, participant_id: 121, attempt_id: 56, item_id: 456, type: "Submission",
		           answer: "my answer", created_at: "2023-01-01 00:00:00"}]`)
	defer func() { _ = db.Close() }()

	var nextID int64
	monkey.PatchInstanceMethod(reflect.TypeOf(&database.DataStore{}), "NewID", func(_ *database.DataStore) int64 {
		nextID++
		return nextID
	})
	defer monkey.UnpatchAll()

	answerStore := database.NewDataStore(db).Answers()
	newID, err := answerStore.CreateNewAnswer(int64(121), int64(121), int64(56), int64(456), "Saved", "my answer", nil)
	require.NoError(t, err)
	assert.Equal(t, int64(2), newID)
}
