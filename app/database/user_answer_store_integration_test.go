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
	db := testhelpers.SetupDBWithFixtureString(`
		groups: [{id: 121}]
		users: [{group_id: 121}]`)
	defer func() { _ = db.Close() }()

	userAnswerStore := database.NewDataStore(db).UserAnswers()
	tests := []struct {
		name        string
		userGroupID int64
		itemID      int64
		attemptID   int64
		answer      string
	}{
		{name: "with attemptID", userGroupID: 121, itemID: 34, attemptID: 56, answer: "my answer"},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			newID, err := userAnswerStore.SubmitNewAnswer(test.userGroupID, test.itemID, test.attemptID, test.answer)

			assert.NoError(t, err)
			assert.NotZero(t, newID)

			type userAnswer struct {
				UserGroupID    int64
				ItemID         int64
				AttemptID      *int64
				Type           string
				Answer         string
				SubmittedAtSet bool
				Validated      bool
			}
			var insertedAnswer userAnswer
			assert.NoError(t,
				userAnswerStore.ByID(newID).
					Select("user_group_id, item_id, attempt_id, type, answer, "+
						"validated, ABS(TIMESTAMPDIFF(SECOND, submitted_at, NOW())) < 3 AS submitted_at_set").
					Scan(&insertedAnswer).Error())
			assert.Equal(t, userAnswer{
				UserGroupID:    test.userGroupID,
				ItemID:         test.itemID,
				AttemptID:      &test.attemptID,
				Type:           "Submission",
				Answer:         test.answer,
				SubmittedAtSet: true,
				Validated:      false,
			}, insertedAnswer)
		})
	}
}

func TestUserAnswerStore_GetOrCreateCurrentAnswer(t *testing.T) {
	attemptID := int64(56)
	tests := []struct {
		name                    string
		userGroupID             int64
		itemID                  int64
		attemptID               *int64
		expectedCurrentAnswerID int64
	}{
		{name: "create new with attemptID", userGroupID: 121, itemID: 34, attemptID: &attemptID},
		{name: "create new without attemptID", userGroupID: 121, itemID: 35},
		{name: "return existing with attemptID", userGroupID: 121, itemID: 33, attemptID: &attemptID, expectedCurrentAnswerID: 2},
		{name: "return existing without attemptID", userGroupID: 121, itemID: 34, expectedCurrentAnswerID: 5},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			db := testhelpers.SetupDBWithFixtureString(`
				groups: [{id: 111}, {id: 121}]
				users: [{login: 111, group_id: 111}, {login: 121, group_id: 121}]
				users_answers:
					- {id: 1, user_group_id: 111, item_id: 34, attempt_id: 56, type: Current, submitted_at: 2018-03-22 08:44:55}
					- {id: 2, user_group_id: 121, item_id: 33, attempt_id: 56, type: Current, submitted_at: 2018-03-22 08:44:55}
					- {id: 3, user_group_id: 121, item_id: 34, attempt_id: 55, type: Current, submitted_at: 2018-03-22 08:44:55}
					- {id: 4, user_group_id: 121, item_id: 34, attempt_id: 56, type: Submission, submitted_at: 2018-03-22 08:44:55}
					- {id: 5, user_group_id: 121, item_id: 34, type: Current, submitted_at: 2018-03-22 08:44:55}
					- {id: 6, user_group_id: 121, item_id: 35, type: Submission, submitted_at: 2018-03-22 08:44:55}`)
			defer func() { _ = db.Close() }()

			dataStore := database.NewDataStore(db)
			var currentAnswerID int64
			assert.NoError(t, dataStore.InTransaction(func(store *database.DataStore) error {
				var err error
				currentAnswerID, err = store.UserAnswers().
					GetOrCreateCurrentAnswer(test.userGroupID, test.itemID, test.attemptID)
				return err
			}))

			assert.NotZero(t, currentAnswerID)
			if test.expectedCurrentAnswerID > 0 {
				assert.Equal(t, test.expectedCurrentAnswerID, currentAnswerID)
			} else {
				assert.True(t, currentAnswerID > int64(6))
				type userAnswer struct {
					UserGroupID    int64
					ItemID         int64
					AttemptID      *int64
					Type           string
					SubmittedAtSet bool
					Validated      bool
				}
				var insertedAnswer userAnswer
				assert.NoError(t,
					dataStore.UserAnswers().ByID(currentAnswerID).
						Select(`
							user_group_id, item_id, attempt_id, type, validated,
							ABS(TIMESTAMPDIFF(SECOND, submitted_at, NOW())) < 3 AS submitted_at_set`).
						Scan(&insertedAnswer).Error())
				assert.Equal(t, userAnswer{
					UserGroupID:    test.userGroupID,
					ItemID:         test.itemID,
					AttemptID:      test.attemptID,
					Type:           "Current",
					SubmittedAtSet: true,
					Validated:      false,
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
		userGroupID   int64
		expectedFound bool
	}{
		{
			name: "okay (full access)",
			fixture: `
				users_answers: [{id: 200, user_group_id: 111, item_id: 50, attempt_id: 100, submitted_at: 2018-03-22 08:44:55}]
				groups_attempts: [{id: 100, group_id: 111, item_id: 50, order: 0}]`,
			userAnswerID:  200,
			userGroupID:   111,
			expectedFound: true,
		},
		{
			name: "okay (partial access)",
			fixture: `
				users_answers: [{id: 200, user_group_id: 101, item_id: 50, attempt_id: 100, submitted_at: 2018-03-22 08:44:55}]
				groups_attempts: [{id: 100, group_id: 101, item_id: 50, order: 0}]`,
			userAnswerID:  200,
			userGroupID:   101,
			expectedFound: true,
		},
		{
			name:         "okay (has_attempts=1, groups_groups.type=requestAccepted)",
			userGroupID:  101,
			userAnswerID: 200,
			fixture: `
				users_answers:
					- {id: 200, user_group_id: 101, item_id: 60, attempt_id: 100, submitted_at: 2018-03-22 08:44:55}
				groups_attempts:
					- {id: 100, group_id: 102, item_id: 60, order: 0}`,
			expectedFound: true,
		},
		{
			name:         "okay (has_attempts=1, groups_groups.type=joinedByCode)",
			userGroupID:  101,
			userAnswerID: 200,
			fixture: `
				users_answers:
					- {id: 200, user_group_id: 101, item_id: 60, attempt_id: 100, submitted_at: 2018-03-22 08:44:55}
				groups_attempts:
					- {id: 100, group_id: 140, item_id: 60, order: 0}`,
			expectedFound: true,
		},
		{
			name:         "okay (has_attempts=1, groups_groups.type=invitationAccepted)",
			userGroupID:  101,
			userAnswerID: 200,
			fixture: `
				users_answers:
					- {id: 200, user_group_id: 101, item_id: 60, attempt_id: 100, submitted_at: 2018-03-22 08:44:55}
				groups_attempts:
					- {id: 100, group_id: 110, item_id: 60, order: 0}`,
			expectedFound: true,
		},
		{
			name: "user not found",
			fixture: `
				groups_attempts: [{id: 100, group_id: 121, item_id: 50, order: 0}]
				users_answers: [{id: 200, user_group_id: 101, item_id: 60, attempt_id: 100, submitted_at: 2018-03-22 08:44:55}]`,
			userGroupID:   404,
			userAnswerID:  100,
			expectedFound: false,
		},
		{
			name:         "user doesn't have access to the item",
			userGroupID:  121,
			userAnswerID: 100,
			fixture: `
				users_answers: [{id: 100, user_group_id: 121, item_id: 50, attempt_id: 200, submitted_at: 2018-03-22 08:44:55}]
				groups_attempts: [{id: 200, group_id: 121, item_id: 50, order: 0}]`,
			expectedFound: false,
		},
		{
			name:          "no groups_attempts",
			userGroupID:   101,
			userAnswerID:  100,
			fixture:       `users_answers: [{id: 100, user_group_id: 101, item_id: 50, submitted_at: 2018-03-22 08:44:55}]`,
			expectedFound: false,
		},
		{
			name:         "wrong item in groups_attempts",
			userGroupID:  101,
			userAnswerID: 100,
			fixture: `
				users_answers: [{id: 100, user_group_id: 101, item_id: 50, attempt_id: 200, submitted_at: 2018-03-22 08:44:55}]
				groups_attempts: [{id: 200, group_id: 101, item_id: 51, order: 0}]`,
			expectedFound: false,
		},
		{
			name:         "no users_answers",
			userGroupID:  101,
			userAnswerID: 100,
			fixture: `
				groups_attempts: [{id: 100, group_id: 101, item_id: 50, order: 0}]`,
			expectedFound: false,
		},
		{
			name:         "user is not a member of the team (invitationSent)",
			userGroupID:  101,
			userAnswerID: 100,
			fixture: `
				users_answers: [{id: 100, user_group_id: 101, item_id: 60, attempt_id: 200, submitted_at: 2018-03-22 08:44:55}]
				groups_attempts: [{id: 200, group_id: 103, item_id: 60, order: 0}]`,
			expectedFound: false,
		},
		{
			name:         "user is not a member of the team (requestSent)",
			userGroupID:  101,
			userAnswerID: 100,
			fixture: `
				users_answers: [{id: 100, user_group_id: 101, item_id: 60, attempt_id: 200, submitted_at: 2018-03-22 08:44:55}]
				groups_attempts: [{id: 200, group_id: 104, item_id: 60, order: 0}]`,
			expectedFound: false,
		},
		{
			name:         "user is not a member of the team (invitationRefused)",
			userGroupID:  101,
			userAnswerID: 100,
			fixture: `
				users_answers: [{id: 100, user_group_id: 101, item_id: 60, attempt_id: 200, submitted_at: 2018-03-22 08:44:55}]
				groups_attempts: [{id: 200, group_id: 105, item_id: 60, order: 0}]`,
			expectedFound: false,
		},
		{
			name:         "user is not a member of the team (requestRefused)",
			userGroupID:  101,
			userAnswerID: 100,
			fixture: `
				users_answers: [{id: 100, user_group_id: 101, item_id: 60, attempt_id: 200, submitted_at: 2018-03-22 08:44:55}]
				groups_attempts: [{id: 200, group_id: 106, item_id: 60, order: 0}]`,
			expectedFound: false,
		},
		{
			name:         "user is not a member of the team (removed)",
			userGroupID:  101,
			userAnswerID: 100,
			fixture: `
				users_answers: [{id: 100, user_group_id: 101, item_id: 60, attempt_id: 200, submitted_at: 2018-03-22 08:44:55}]
				groups_attempts: [{id: 200, group_id: 107, item_id: 60, order: 0}]`,
			expectedFound: false,
		},
		{
			name:         "user is not a member of the team (left)",
			userGroupID:  101,
			userAnswerID: 100,
			fixture: `
				users_answers: [{id: 100, user_group_id: 101, item_id: 60, attempt_id: 200, submitted_at: 2018-03-22 08:44:55}]
				groups_attempts: [{id: 200, group_id: 108, item_id: 60, order: 0}]`,
			expectedFound: false,
		},
		{
			name:         "user is a member of the team (direct)",
			userGroupID:  101,
			userAnswerID: 100,
			fixture: `
				users_answers: [{id: 100, user_group_id: 101, item_id: 60, attempt_id: 200, submitted_at: 2018-03-22 08:44:55}]
				groups_attempts: [{id: 200, group_id: 109, item_id: 60, order: 0}]`,
			expectedFound: true,
		},
		{
			name:         "groups_attempts.group_id is not user's self group",
			userGroupID:  101,
			userAnswerID: 100,
			fixture: `
				users_answers: [{id: 100, user_group_id: 101, item_id: 50, attempt_id: 200, submitted_at: 2018-03-22 08:44:55}]
				groups_attempts: [{id: 200, group_id: 102, item_id: 50, order: 0}]`,
			expectedFound: false,
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			db := testhelpers.SetupDBWithFixtureString(`
				groups: [{id: 101}, {id: 111}, {id: 121}]
				users:
					- {login: "john", group_id: 101}
					- {login: "jane", group_id: 111}
					- {login: "guest", group_id: 121}
				groups_groups:
					- {parent_group_id: 102, child_group_id: 101, type: requestAccepted}
					- {parent_group_id: 103, child_group_id: 101, type: invitationSent}
					- {parent_group_id: 104, child_group_id: 101, type: requestSent}
					- {parent_group_id: 105, child_group_id: 101, type: invitationRefused}
					- {parent_group_id: 106, child_group_id: 101, type: requestRefused}
					- {parent_group_id: 107, child_group_id: 101, type: removed}
					- {parent_group_id: 108, child_group_id: 101, type: left}
					- {parent_group_id: 109, child_group_id: 101, type: direct}
					- {parent_group_id: 110, child_group_id: 101, type: invitationAccepted}
					- {parent_group_id: 140, child_group_id: 101, type: joinedByCode}
				groups_ancestors:
					- {ancestor_group_id: 101, child_group_id: 101, is_self: 1}
					- {ancestor_group_id: 102, child_group_id: 101, is_self: 0}
					- {ancestor_group_id: 102, child_group_id: 102, is_self: 1}
					- {ancestor_group_id: 111, child_group_id: 111, is_self: 1}
					- {ancestor_group_id: 121, child_group_id: 121, is_self: 1}
					- {ancestor_group_id: 109, child_group_id: 101, is_self: 0}
					- {ancestor_group_id: 109, child_group_id: 109, is_self: 1}
					- {ancestor_group_id: 110, child_group_id: 101, is_self: 0}
					- {ancestor_group_id: 110, child_group_id: 110, is_self: 1}
					- {ancestor_group_id: 140, child_group_id: 101, is_self: 0}
					- {ancestor_group_id: 140, child_group_id: 140, is_self: 1}
				items:
					- {id: 10, has_attempts: 0}
					- {id: 50, has_attempts: 0}
					- {id: 60, has_attempts: 1}
				permissions_generated:
					- {group_id: 101, item_id: 50, can_view_generated: content}
					- {group_id: 101, item_id: 60, can_view_generated: content}
					- {group_id: 111, item_id: 50, can_view_generated: content_with_descendants}
					- {group_id: 121, item_id: 50, can_view_generated: info}`,
				test.fixture)
			defer func() { _ = db.Close() }()
			store := database.NewDataStore(db)
			user := &database.User{}
			assert.NoError(t, user.LoadByGroupID(store, test.userGroupID))
			var resultID int64
			err := store.UserAnswers().Visible(user).
				Where("users_answers.id = ?", test.userAnswerID).
				PluckFirst("users_answers.id", &resultID).Error()
			if test.expectedFound {
				assert.NoError(t, err)
				assert.Equal(t, test.userAnswerID, resultID)
			} else {
				assert.Equal(t, gorm.ErrRecordNotFound, err)
			}
		})
	}
}
