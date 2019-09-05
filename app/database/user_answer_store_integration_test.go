// +build !unit

package database_test

import (
	"testing"

	"github.com/jinzhu/gorm"
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
						"bValidated, ABS(TIMESTAMPDIFF(SECOND, sSubmissionDate, NOW())) < 3 AS submissionDateSet").
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
						Select("idUser, idItem, idAttempt, sType, bValidated, ABS(TIMESTAMPDIFF(SECOND, sSubmissionDate, NOW())) < 3 AS submissionDateSet").
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

func TestUserAnswerStore_Visible(t *testing.T) {
	tests := []struct {
		name          string
		fixture       string
		userAnswerID  int64
		userID        int64
		expectedFound bool
	}{
		{
			name: "okay (full access)",
			fixture: `
				users_answers: [{ID: 200, idUser: 11, idItem: 50, idAttempt: 100}]
				groups_attempts: [{ID: 100, idGroup: 111, idItem: 50}]`,
			userAnswerID:  200,
			userID:        11,
			expectedFound: true,
		},
		{
			name: "okay (partial access)",
			fixture: `
				users_answers: [{ID: 200, idUser: 10, idItem: 50, idAttempt: 100}]
				groups_attempts: [{ID: 100, idGroup: 101, idItem: 50}]`,
			userAnswerID:  200,
			userID:        10,
			expectedFound: true,
		},
		{
			name:         "okay (bHasAttempts=1, groups_groups.sType=requestAccepted)",
			userID:       10,
			userAnswerID: 200,
			fixture: `
				users_answers:
					- {ID: 200, idUser: 10, idItem: 60, idAttempt: 100}
				groups_attempts:
					- {ID: 100, idGroup: 102, idItem: 60}`,
			expectedFound: true,
		},
		{
			name:         "okay (bHasAttempts=1, groups_groups.sType=joinedByCode)",
			userID:       10,
			userAnswerID: 200,
			fixture: `
				users_answers:
					- {ID: 200, idUser: 10, idItem: 60, idAttempt: 100}
				groups_attempts:
					- {ID: 100, idGroup: 140, idItem: 60}`,
			expectedFound: true,
		},
		{
			name:         "okay (bHasAttempts=1, groups_groups.sType=invitationAccepted)",
			userID:       10,
			userAnswerID: 200,
			fixture: `
				users_answers:
					- {ID: 200, idUser: 10, idItem: 60, idAttempt: 100}
				groups_attempts:
					- {ID: 100, idGroup: 110, idItem: 60}`,
			expectedFound: true,
		},
		{
			name: "user not found",
			fixture: `
				groups_attempts: [{ID: 100, idGroup: 121, idItem: 50}]
				users_answers: [{ID: 200, idUser: 10, idItem: 60, idAttempt: 100}]`,
			userID:        404,
			userAnswerID:  100,
			expectedFound: false,
		},
		{
			name:         "user doesn't have access to the item",
			userID:       12,
			userAnswerID: 100,
			fixture: `
				users_answers: [{ID: 100, idUser: 12, idItem: 50, idAttempt: 200}]
				groups_attempts: [{ID: 200, idGroup: 121, idItem: 50}]`,
			expectedFound: false,
		},
		{
			name:          "no groups_attempts",
			userID:        10,
			userAnswerID:  100,
			fixture:       `users_answers: [{ID: 100, idUser: 10, idItem: 50}]`,
			expectedFound: false,
		},
		{
			name:         "wrong item in groups_attempts",
			userID:       10,
			userAnswerID: 100,
			fixture: `
				users_answers: [{ID: 100, idUser: 10, idItem: 50, idAttempt: 200}]
				groups_attempts: [{ID: 200, idGroup: 101, idItem: 51}]`,
			expectedFound: false,
		},
		{
			name:         "no users_answers",
			userID:       10,
			userAnswerID: 100,
			fixture: `
				groups_attempts: [{ID: 100, idGroup: 101, idItem: 50}]`,
			expectedFound: false,
		},
		{
			name:         "user is not a member of the team (invitationSent)",
			userID:       10,
			userAnswerID: 100,
			fixture: `
				users_answers: [{ID: 100, idUser: 10, idItem: 60, idAttempt: 200}]
				groups_attempts: [ID: 200, idGroup: 103, idItem: 60]`,
			expectedFound: false,
		},
		{
			name:         "user is not a member of the team (requestSent)",
			userID:       10,
			userAnswerID: 100,
			fixture: `
				users_answers: [{ID: 100, idUser: 10, idItem: 60, idAttempt: 200}]
				groups_attempts: [ID: 200, idGroup: 104, idItem: 60]`,
			expectedFound: false,
		},
		{
			name:         "user is not a member of the team (invitationRefused)",
			userID:       10,
			userAnswerID: 100,
			fixture: `
				users_answers: [{ID: 100, idUser: 10, idItem: 60, idAttempt: 200}]
				groups_attempts: [ID: 200, idGroup: 105, idItem: 60]`,
			expectedFound: false,
		},
		{
			name:         "user is not a member of the team (requestRefused)",
			userID:       10,
			userAnswerID: 100,
			fixture: `
				users_answers: [{ID: 100, idUser: 10, idItem: 60, idAttempt: 200}]
				groups_attempts: [ID: 200, idGroup: 106, idItem: 60]`,
			expectedFound: false,
		},
		{
			name:         "user is not a member of the team (removed)",
			userID:       10,
			userAnswerID: 100,
			fixture: `
				users_answers: [{ID: 100, idUser: 10, idItem: 60, idAttempt: 200}]
				groups_attempts: [ID: 200, idGroup: 107, idItem: 60]`,
			expectedFound: false,
		},
		{
			name:         "user is not a member of the team (left)",
			userID:       10,
			userAnswerID: 100,
			fixture: `
				users_answers: [{ID: 100, idUser: 10, idItem: 60, idAttempt: 200}]
				groups_attempts: [ID: 200, idGroup: 108, idItem: 60]`,
			expectedFound: false,
		},
		{
			name:         "user is not a member of the team (direct)",
			userID:       10,
			userAnswerID: 100,
			fixture: `
				users_answers: [{ID: 100, idUser: 10, idItem: 60, idAttempt: 200}]
				groups_attempts: [ID: 200, idGroup: 109, idItem: 60]`,
			expectedFound: false,
		},
		{
			name:         "groups_attempts.idGroup is not user's self group",
			userID:       10,
			userAnswerID: 100,
			fixture: `
				users_answers: [{ID: 100, idUser: 10, idItem: 50, idAttempt: 200}]
				groups_attempts: [ID: 200, idGroup: 102, idItem: 50]`,
			expectedFound: false,
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			db := testhelpers.SetupDBWithFixtureString(`
				users:
					- {ID: 10, sLogin: "john", idGroupSelf: 101}
					- {ID: 11, sLogin: "jane", idGroupSelf: 111}
					- {ID: 12, sLogin: "guest", idGroupSelf: 121}
				groups_groups:
					- {idGroupParent: 102, idGroupChild: 101, sType: requestAccepted}
					- {idGroupParent: 103, idGroupChild: 101, sType: invitationSent}
					- {idGroupParent: 104, idGroupChild: 101, sType: requestSent}
					- {idGroupParent: 105, idGroupChild: 101, sType: invitationRefused}
					- {idGroupParent: 106, idGroupChild: 101, sType: requestRefused}
					- {idGroupParent: 107, idGroupChild: 101, sType: removed}
					- {idGroupParent: 108, idGroupChild: 101, sType: left}
					- {idGroupParent: 109, idGroupChild: 101, sType: direct}
					- {idGroupParent: 110, idGroupChild: 101, sType: invitationAccepted}
					- {idGroupParent: 140, idGroupChild: 101, sType: joinedByCode}
				groups_ancestors:
					- {idGroupAncestor: 101, idGroupChild: 101, bIsSelf: 1}
					- {idGroupAncestor: 102, idGroupChild: 101, bIsSelf: 0}
					- {idGroupAncestor: 102, idGroupChild: 102, bIsSelf: 1}
					- {idGroupAncestor: 111, idGroupChild: 111, bIsSelf: 1}
					- {idGroupAncestor: 121, idGroupChild: 121, bIsSelf: 1}
					- {idGroupAncestor: 109, idGroupChild: 101, bIsSelf: 0}
					- {idGroupAncestor: 109, idGroupChild: 109, bIsSelf: 1}
					- {idGroupAncestor: 110, idGroupChild: 101, bIsSelf: 0}
					- {idGroupAncestor: 110, idGroupChild: 110, bIsSelf: 1}
					- {idGroupAncestor: 140, idGroupChild: 101, bIsSelf: 0}
					- {idGroupAncestor: 140, idGroupChild: 140, bIsSelf: 1}
				items:
					- {ID: 10, bHasAttempts: 0}
					- {ID: 50, bHasAttempts: 0}
					- {ID: 60, bHasAttempts: 1}
				groups_items:
					- {idGroup: 101, idItem: 50, sCachedPartialAccessDate: "2017-05-29T06:38:38Z"}
					- {idGroup: 101, idItem: 60, sCachedPartialAccessDate: "2017-05-29T06:38:38Z"}
					- {idGroup: 111, idItem: 50, sCachedFullAccessDate: "2017-05-29T06:38:38Z"}
					- {idGroup: 121, idItem: 50, sCachedGrayedAccessDate: "2017-05-29T06:38:38Z"}`,
				test.fixture)
			defer func() { _ = db.Close() }()
			store := database.NewDataStore(db)
			user := &database.User{}
			assert.NoError(t, user.LoadByID(store, test.userID))
			var resultID int64
			err := store.UserAnswers().Visible(user).
				Where("users_answers.ID = ?", test.userAnswerID).
				PluckFirst("users_answers.ID", &resultID).Error()
			if test.expectedFound {
				assert.NoError(t, err)
				assert.Equal(t, test.userAnswerID, resultID)
			} else {
				assert.Equal(t, gorm.ErrRecordNotFound, err)
			}
		})
	}
}
