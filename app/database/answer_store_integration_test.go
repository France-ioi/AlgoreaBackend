// +build !unit

package database_test

import (
	"testing"

	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/testhelpers"
)

func TestAnswerStore_SubmitNewAnswer(t *testing.T) {
	db := testhelpers.SetupDBWithFixtureString(`
		groups: [{id: 121}]
		users: [{group_id: 121}]
		groups_attempts: [{id: 56, group_id: 121, item_id: 34, order: 1}]`)
	defer func() { _ = db.Close() }()

	answerStore := database.NewDataStore(db).Answers()
	tests := []struct {
		name      string
		authorID  int64
		attemptID int64
		answer    string
	}{
		{name: "with attemptID", authorID: 121, attemptID: 56, answer: "my answer"},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			newID, err := answerStore.SubmitNewAnswer(test.authorID, test.attemptID, test.answer)

			assert.NoError(t, err)
			assert.NotZero(t, newID)

			type answer struct {
				AuthorID     int64
				AttemptID    int64
				Type         string
				Answer       string
				CreatedAtSet bool
			}
			var insertedAnswer answer
			assert.NoError(t,
				answerStore.ByID(newID).
					Select("author_id, attempt_id, type, answer, "+
						"ABS(TIMESTAMPDIFF(SECOND, created_at, NOW())) < 3 AS created_at_set").
					Scan(&insertedAnswer).Error())
			assert.Equal(t, answer{
				AuthorID:     test.authorID,
				AttemptID:    test.attemptID,
				Type:         "Submission",
				Answer:       test.answer,
				CreatedAtSet: true,
			}, insertedAnswer)
		})
	}
}

func TestAnswerStore_Visible(t *testing.T) {
	tests := []struct {
		name          string
		fixture       string
		answerID      int64
		userID        int64
		expectedFound bool
	}{
		{
			name: "okay (full access)",
			fixture: `
				groups_attempts: [{id: 100, group_id: 111, item_id: 50, order: 0}]
				answers: [{id: 200, author_id: 111, attempt_id: 100, created_at: 2018-03-22 08:44:55}]`,
			answerID:      200,
			userID:        111,
			expectedFound: true,
		},
		{
			name: "okay (content access)",
			fixture: `
				groups_attempts: [{id: 100, group_id: 101, item_id: 50, order: 0}]
				answers: [{id: 200, author_id: 101, attempt_id: 100, created_at: 2018-03-22 08:44:55}]`,
			answerID:      200,
			userID:        101,
			expectedFound: true,
		},
		{
			name:     "okay (a team member)",
			userID:   101,
			answerID: 200,
			fixture: `
				groups_attempts:
					- {id: 100, group_id: 102, item_id: 60, order: 0}
				answers:
					- {id: 200, author_id: 101, attempt_id: 100, created_at: 2018-03-22 08:44:55}`,
			expectedFound: true,
		},
		{
			name: "user not found",
			fixture: `
				groups_attempts: [{id: 100, group_id: 121, item_id: 50, order: 0}]
				answers: [{id: 200, author_id: 101, attempt_id: 100, created_at: 2018-03-22 08:44:55}]`,
			userID:        404,
			answerID:      100,
			expectedFound: false,
		},
		{
			name:     "user doesn't have access to the item",
			userID:   121,
			answerID: 100,
			fixture: `
				groups_attempts: [{id: 200, group_id: 121, item_id: 50, order: 0}]
				answers: [{id: 100, author_id: 121, attempt_id: 200, created_at: 2018-03-22 08:44:55}]`,
			expectedFound: false,
		},
		{
			name:     "wrong item in groups_attempts",
			userID:   101,
			answerID: 100,
			fixture: `
				groups_attempts: [{id: 200, group_id: 101, item_id: 51, order: 0}]
				answers: [{id: 100, author_id: 101, attempt_id: 200, created_at: 2018-03-22 08:44:55}]`,
			expectedFound: false,
		},
		{
			name:     "no answers",
			userID:   101,
			answerID: 100,
			fixture: `
				groups_attempts: [{id: 100, group_id: 101, item_id: 50, order: 0}]`,
			expectedFound: false,
		},
		{
			name:     "user is not a member of the team",
			userID:   101,
			answerID: 100,
			fixture: `
				groups_attempts: [{id: 200, group_id: 103, item_id: 60, order: 0}]
				answers: [{id: 100, author_id: 101, attempt_id: 200, created_at: 2018-03-22 08:44:55}]`,
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
					- {parent_group_id: 102, child_group_id: 101}
				groups_ancestors:
					- {ancestor_group_id: 101, child_group_id: 101, is_self: 1}
					- {ancestor_group_id: 102, child_group_id: 101, is_self: 0}
					- {ancestor_group_id: 102, child_group_id: 102, is_self: 1}
					- {ancestor_group_id: 111, child_group_id: 111, is_self: 1}
					- {ancestor_group_id: 121, child_group_id: 121, is_self: 1}
				items:
					- {id: 10}
					- {id: 50}
					- {id: 60}
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
			var resultID int64
			err := store.Answers().Visible(user).
				Where("answers.id = ?", test.answerID).
				PluckFirst("answers.id", &resultID).Error()
			if test.expectedFound {
				assert.NoError(t, err)
				assert.Equal(t, test.answerID, resultID)
			} else {
				assert.Equal(t, gorm.ErrRecordNotFound, err)
			}
		})
	}
}
