package database

import (
	"github.com/jinzhu/gorm"
)

// AnswerStore implements database operations on `answers`
type AnswerStore struct {
	*DataStore
}

// WithUsers creates a composable query for getting answers joined with users (via author_id)
func (s *AnswerStore) WithUsers() *AnswerStore {
	return &AnswerStore{
		NewDataStoreWithTable(
			s.Joins("JOIN users ON users.group_id = answers.author_id"), s.tableName,
		),
	}
}

// WithGroupAttempts creates a composable query for getting answers joined with groups_attempts
func (s *AnswerStore) WithGroupAttempts() *AnswerStore {
	return &AnswerStore{
		NewDataStoreWithTable(
			s.Joins("JOIN groups_attempts ON groups_attempts.id = answers.attempt_id"), s.tableName,
		),
	}
}

// WithItems joins `items` through `groups_attempts`
func (s *AnswerStore) WithItems() *AnswerStore {
	return &AnswerStore{
		NewDataStoreWithTable(
			s.WithGroupAttempts().Joins("JOIN items ON items.id = groups_attempts.item_id"), s.tableName,
		),
	}
}

// SubmitNewAnswer inserts a new row with type='Submission', submitted_at=NOW()
// into the `answers` table.
func (s *AnswerStore) SubmitNewAnswer(authorID, attemptID int64, answer string) (int64, error) {
	var answerID int64
	err := s.retryOnDuplicatePrimaryKeyError(func(db *DB) error {
		store := NewDataStore(db)
		answerID = store.NewID()
		return db.db.Exec(`
				INSERT INTO answers (id, author_id, attempt_id, answer, submitted_at)
				VALUES (?, ?, ?, ?, NOW())`,
			answerID, authorID, attemptID, answer).Error
	})
	return answerID, err
}

// GetOrCreateCurrentAnswer returns an id of the current answer for given authorID & attemptID
// or inserts a new row with type='Current' and submitted_at=NOW() into the `answers` table.
func (s *AnswerStore) GetOrCreateCurrentAnswer(authorID, attemptID int64) (answerID int64, err error) {
	s.mustBeInTransaction()
	recoverPanics(&err)

	err = s.WithWriteLock().
		Joins("JOIN groups_attempts ON groups_attempts.id = answers.attempt_id").
		Where("answers.author_id = ?", authorID).
		Where("answers.type = 'Current'").
		Where("attempt_id = ?", attemptID).
		PluckFirst("answers.id", &answerID).Error()
	if gorm.IsRecordNotFoundError(err) {
		err = s.retryOnDuplicatePrimaryKeyError(func(db *DB) error {
			store := NewDataStore(db)
			answerID = store.NewID()
			return db.Exec(`
				INSERT INTO answers (id, author_id, attempt_id, type, submitted_at)
				VALUES (?, ?, ?, 'Current', NOW())`,
				answerID, authorID, attemptID).Error()
		})
	}
	mustNotBeError(err)
	return answerID, err
}

// Visible returns a composable query for getting answers with the following access rights
// restrictions:
// 1) the user should have at least 'content' access rights to the answers.item_id item,
// 2) the user is able to see answers related to his group's attempts, so:
//   (a) if items.has_attempts = 1, then the user should be a member of the groups_attempts.group_id team
//   (b) if items.has_attempts = 0, then groups_attempts.group_id should be equal to the user's self group
func (s *AnswerStore) Visible(user *User) *DB {
	usersGroupsQuery := s.GroupGroups().WhereUserIsMember(user).Select("parent_group_id")
	// the user should have at least 'content' access to the item
	itemsQuery := s.Items().WhereUserHasViewPermissionOnItems(user, "content")

	return s.
		// the user should have at least 'content' access to the answers.item_id
		Joins("JOIN groups_attempts ON groups_attempts.id = answers.attempt_id").
		Joins("JOIN ? AS items ON items.id = groups_attempts.item_id", itemsQuery.SubQuery()).
		// if items.has_attempts = 1, then groups_attempts.group_id should be one of the authorized user's groups,
		// otherwise groups_attempts.group_id should be equal to the user's self group
		Where("IF(items.has_attempts, groups_attempts.group_id IN ?, groups_attempts.group_id = ?)",
			usersGroupsQuery.SubQuery(), user.GroupID)
}
