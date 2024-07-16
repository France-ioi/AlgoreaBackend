//go:build !unit

package database_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/testhelpers"
)

func TestAnswerStore_SubmitNewAnswer(t *testing.T) {
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
			newID, err := answerStore.SubmitNewAnswer(test.authorID, test.participantID, test.attemptID, test.itemID, test.answer)

			assert.NoError(t, err)
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
			assert.NoError(t,
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
