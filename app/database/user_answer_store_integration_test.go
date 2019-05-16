// +build !unit

package database_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/testhelpers"
)

func TestUserAnswerStore_SubmitNewAnswer(t *testing.T) {
	db := testhelpers.SetupDBWithFixture()

	userAnswerStore := database.NewDataStore(db).UserAnswers()
	tests := []struct {
		name      string
		userID    int64
		itemID    int64
		attemptID *int64
		answer    string
	}{
		{name: "with attemptID", userID: 12, itemID: 34, attemptID: ptrInt64(56), answer: "my answer"},
		{name: "without attemptID", userID: 34, itemID: 56, attemptID: nil, answer: "another answer"},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			newID, err := userAnswerStore.SubmitNewAnswer(test.userID, test.itemID, test.attemptID, test.answer)

			assert.NoError(t, err)
			assert.NotZero(t, newID)

			type userAnswer struct {
				UserID            int64  `gorm:"column:idUser"`
				ItemID            int64  `gorm:"column:idItem"`
				AttemptID         *int64 `gorm:"column:idAttempt"`
				Type              string `gorm:"column:sType"`
				Answer            string `gorm:"column:sAnswer"`
				SubmissionDateSet bool   `gorm:"column:submissionDateSet"`
				Validated         bool   `gorm:"column:bValidated"`
			}
			var insertedAnswer userAnswer
			assert.NoError(t,
				userAnswerStore.Select("idUser, idItem, idAttempt, sType, sAnswer, ABS(NOW() - sSubmissionDate) < 3 AS submissionDateSet").
					Where("ID = ?", newID).Scan(&insertedAnswer).Error())
			assert.Equal(t, userAnswer{
				UserID:            test.userID,
				ItemID:            test.itemID,
				AttemptID:         test.attemptID,
				Type:              "Submission",
				Answer:            test.answer,
				SubmissionDateSet: true,
				Validated:         false,
			}, insertedAnswer)
		})
	}
}

func ptrInt64(i int64) *int64 { return &i }
