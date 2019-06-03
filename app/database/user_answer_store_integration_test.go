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
	defer func() { _ = db.Close() }()

	userAnswerStore := database.NewDataStore(db).UserAnswers()
	tests := []struct {
		name      string
		userID    int64
		itemID    int64
		attemptID int64
		answer    string
	}{
		{name: "with attemptID", userID: 12, itemID: 34, attemptID: 56, answer: "my answer"},
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
				userAnswerStore.ByID(newID).
					Select("idUser, idItem, idAttempt, sType, sAnswer, "+
						"bValidated, ABS(NOW() - sSubmissionDate) < 3 AS submissionDateSet").
					Scan(&insertedAnswer).Error())
			assert.Equal(t, userAnswer{
				UserID:            test.userID,
				ItemID:            test.itemID,
				AttemptID:         &test.attemptID,
				Type:              "Submission",
				Answer:            test.answer,
				SubmissionDateSet: true,
				Validated:         false,
			}, insertedAnswer)
		})
	}
}

func TestUserAnswerStore_GetOrCreateCurrentAnswer(t *testing.T) {
	attemptID := int64(56)
	tests := []struct {
		name                    string
		userID                  int64
		itemID                  int64
		attemptID               *int64
		expectedCurrentAnswerID int64
	}{
		{name: "create new with attemptID", userID: 12, itemID: 34, attemptID: &attemptID},
		{name: "create new without attemptID", userID: 12, itemID: 35},
		{name: "return existing with attemptID", userID: 12, itemID: 33, attemptID: &attemptID, expectedCurrentAnswerID: 2},
		{name: "return existing without attemptID", userID: 12, itemID: 34, expectedCurrentAnswerID: 5},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			db := testhelpers.SetupDBWithFixtureString(`
				users_answers:
					- {ID: 1, idUser: 11, idItem: 34, idAttempt: 56, sType: Current}
					- {ID: 2, idUser: 12, idItem: 33, idAttempt: 56, sType: Current}
					- {ID: 3, idUser: 12, idItem: 34, idAttempt: 55, sType: Current}
					- {ID: 4, idUser: 12, idItem: 34, idAttempt: 56, sType: Submission}
					- {ID: 5, idUser: 12, idItem: 34, sType: Current}
					- {ID: 6, idUser: 12, idItem: 35, sType: Submission}`)
			defer func() { _ = db.Close() }()

			dataStore := database.NewDataStore(db)
			var currentAnswerID int64
			assert.NoError(t, dataStore.InTransaction(func(store *database.DataStore) error {
				var err error
				currentAnswerID, err = store.UserAnswers().
					GetOrCreateCurrentAnswer(test.userID, test.itemID, test.attemptID)
				return err
			}))

			assert.NotZero(t, currentAnswerID)
			if test.expectedCurrentAnswerID > 0 {
				assert.Equal(t, test.expectedCurrentAnswerID, currentAnswerID)
			} else {
				assert.True(t, currentAnswerID > int64(6))
				type userAnswer struct {
					UserID            int64  `gorm:"column:idUser"`
					ItemID            int64  `gorm:"column:idItem"`
					AttemptID         *int64 `gorm:"column:idAttempt"`
					Type              string `gorm:"column:sType"`
					SubmissionDateSet bool   `gorm:"column:submissionDateSet"`
					Validated         bool   `gorm:"column:bValidated"`
				}
				var insertedAnswer userAnswer
				assert.NoError(t,
					dataStore.UserAnswers().ByID(currentAnswerID).
						Select("idUser, idItem, idAttempt, sType, bValidated, ABS(NOW() - sSubmissionDate) < 3 AS submissionDateSet").
						Scan(&insertedAnswer).Error())
				assert.Equal(t, userAnswer{
					UserID:            test.userID,
					ItemID:            test.itemID,
					AttemptID:         test.attemptID,
					Type:              "Current",
					SubmissionDateSet: true,
					Validated:         false,
				}, insertedAnswer)
			}
		})
	}
}
