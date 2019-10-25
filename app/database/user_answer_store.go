package database

import (
	"github.com/jinzhu/gorm"
)

// UserAnswerStore implements database operations on `users_answers`
type UserAnswerStore struct {
	*DataStore
}

// WithUsers creates a composable query for getting answers joined with users
func (s *UserAnswerStore) WithUsers() *UserAnswerStore {
	return &UserAnswerStore{
		NewDataStoreWithTable(
			s.Joins("JOIN users ON users.group_id = users_answers.user_group_id"), s.tableName,
		),
	}
}

// WithGroupAttempts creates a composable query for getting answers joined with groups_attempts
func (s *UserAnswerStore) WithGroupAttempts() *UserAnswerStore {
	return &UserAnswerStore{
		NewDataStoreWithTable(
			s.Joins("JOIN groups_attempts ON groups_attempts.id = users_answers.attempt_id"), s.tableName,
		),
	}
}

// WithItems joins `items`
func (s *UserAnswerStore) WithItems() *UserAnswerStore {
	return &UserAnswerStore{
		NewDataStoreWithTable(
			s.Joins("JOIN items ON items.id = users_answers.item_id"), s.tableName,
		),
	}
}

// SubmitNewAnswer inserts a new row with type='Submission', validated=0, submitted_at=NOW()
// into the `users_answers` table.
func (s *UserAnswerStore) SubmitNewAnswer(userGroupID, itemID, attemptID int64, answer string) (int64, error) {
	var userAnswerID int64
	err := s.retryOnDuplicatePrimaryKeyError(func(db *DB) error {
		store := NewDataStore(db)
		userAnswerID = store.NewID()
		return db.db.Exec(`
				INSERT INTO users_answers (id, user_group_id, item_id, attempt_id, answer, submitted_at, validated)
				VALUES (?, ?, ?, ?, ?, NOW(), 0)`,
			userAnswerID, userGroupID, itemID, attemptID, answer).Error
	})
	return userAnswerID, err
}

// GetOrCreateCurrentAnswer returns an id of the current users_answers for given userID, itemID, attemptID
// or inserts a new row with type='Current' and submitted_at=NOW() into the `users_answers` table.
func (s *UserAnswerStore) GetOrCreateCurrentAnswer(userGroupID, itemID int64, attemptID *int64) (userAnswerID int64, err error) {
	s.mustBeInTransaction()
	recoverPanics(&err)

	query := s.WithWriteLock().
		Where("user_group_id = ?", userGroupID).
		Where("item_id = ?", itemID).
		Where("type = 'Current'")
	if attemptID == nil {
		query = query.Where("attempt_id IS NULL")
	} else {
		query = query.Where("attempt_id = ?", *attemptID)
	}
	err = query.PluckFirst("id", &userAnswerID).Error()
	if gorm.IsRecordNotFoundError(err) {
		err = s.retryOnDuplicatePrimaryKeyError(func(db *DB) error {
			store := NewDataStore(db)
			userAnswerID = store.NewID()
			return db.Exec(`
				INSERT INTO users_answers (id, user_group_id, item_id, attempt_id, type, submitted_at)
				VALUES (?, ?, ?, ?, 'Current', NOW())`,
				userAnswerID, userGroupID, itemID, attemptID).Error()
		})
	}
	mustNotBeError(err)
	return userAnswerID, err
}

// Visible returns a composable query for getting users_answers with the following access rights
// restrictions:
// 1) the user should have at least partial access rights to the users_answers.item_id item,
// 2) the user is able to see answers related to his group's attempts, so:
//   (a) if items.has_attempts = 1, then the user should be a member of the groups_attempts.group_id team
//   (b) if items.has_attempts = 0, then groups_attempts.group_id should be equal to the user's self group
func (s *UserAnswerStore) Visible(user *User) *DB {
	usersGroupsQuery := s.GroupGroups().WhereUserIsMember(user).Select("parent_group_id")
	// the user should have at least partial access to the item
	itemsQuery := s.Items().Visible(user).Where("partial_access > 0 OR full_access > 0")

	return s.
		// the user should have at least partial access to the users_answers.item_id
		Joins("JOIN ? AS items ON items.id = users_answers.item_id", itemsQuery.SubQuery()).
		Joins("JOIN groups_attempts ON groups_attempts.item_id = users_answers.item_id AND groups_attempts.id = users_answers.attempt_id").
		// if items.has_attempts = 1, then groups_attempts.group_id should be one of the authorized user's groups,
		// otherwise groups_attempts.group_id should be equal to the user's self group
		Where("IF(items.has_attempts, groups_attempts.group_id IN ?, groups_attempts.group_id = ?)",
			usersGroupsQuery.SubQuery(), user.GroupID)
}
