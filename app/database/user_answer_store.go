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
			s.Joins("JOIN users ON users.ID = users_answers.idUser"), s.tableName,
		),
	}
}

// WithGroupAttempts creates a composable query for getting answers joined with groups_attempts
func (s *UserAnswerStore) WithGroupAttempts() *UserAnswerStore {
	return &UserAnswerStore{
		NewDataStoreWithTable(
			s.Joins("JOIN groups_attempts ON groups_attempts.ID = users_answers.idAttempt"), s.tableName,
		),
	}
}

// WithItems joins `items`
func (s *UserAnswerStore) WithItems() *UserAnswerStore {
	return &UserAnswerStore{
		NewDataStoreWithTable(
			s.Joins("JOIN items ON items.ID = users_answers.idItem"), s.tableName,
		),
	}
}

// SubmitNewAnswer inserts a new row with sType='Submission', bValidated=0, sSubmissionDate=NOW()
// into the `users_answers` table.
func (s *UserAnswerStore) SubmitNewAnswer(userID, itemID, attemptID int64, answer string) (int64, error) {
	var userAnswerID int64
	err := s.retryOnDuplicatePrimaryKeyError(func(db *DB) error {
		store := NewDataStore(db)
		userAnswerID = store.NewID()
		return db.db.Exec(`
				INSERT INTO users_answers (ID, idUser, idItem, idAttempt, sAnswer, sSubmissionDate, bValidated)
				VALUES (?, ?, ?, ?, ?, NOW(), 0)`,
			userAnswerID, userID, itemID, attemptID, answer).Error
	})
	return userAnswerID, err
}

// GetOrCreateCurrentAnswer returns an ID of the current users_answers for given userID, itemID, attemptID
// or inserts a new row with sType='Current' and sSubmissionDate=NOW() into the `users_answers` table.
func (s *UserAnswerStore) GetOrCreateCurrentAnswer(userID, itemID int64, attemptID *int64) (userAnswerID int64, err error) {
	s.mustBeInTransaction()
	recoverPanics(&err)

	query := s.WithWriteLock().
		Where("idUser = ?", userID).
		Where("idItem = ?", itemID).
		Where("sType = 'Current'")
	if attemptID == nil {
		query = query.Where("idAttempt IS NULL")
	} else {
		query = query.Where("idAttempt = ?", *attemptID)
	}
	err = query.PluckFirst("ID", &userAnswerID).Error()
	if gorm.IsRecordNotFoundError(err) {
		err = s.retryOnDuplicatePrimaryKeyError(func(db *DB) error {
			store := NewDataStore(db)
			userAnswerID = store.NewID()
			return db.Exec(`
				INSERT INTO users_answers (ID, idUser, idItem, idAttempt, sType, sSubmissionDate)
				VALUES (?, ?, ?, ?, 'Current', NOW())`,
				userAnswerID, userID, itemID, attemptID).Error()
		})
	}
	mustNotBeError(err)
	return userAnswerID, err
}

// Visible returns a composable query for getting users_answers with the following access rights
// restrictions:
// 1) the user should have at least partial access rights to the users_answers.idItem item,
// 2) the user is able to see answers related to his group's attempts, so:
//   (a) if items.bHasAttempts = 1, then the user should be a member of the groups_attempts.idGroup team
//   (b) if items.bHasAttempts = 0, then groups_attempts.idGroup should be equal to the user's self group
func (s *UserAnswerStore) Visible(user *User) *DB {
	usersGroupsQuery := s.GroupGroups().WhereUserIsMember(user).Select("idGroupParent")
	// the user should have at least partial access to the item
	itemsQuery := s.Items().Visible(user).Where("partialAccess > 0 OR fullAccess > 0")

	return s.
		// the user should have at least partial access to the users_answers.idItem
		Joins("JOIN ? AS items ON items.ID = users_answers.idItem", itemsQuery.SubQuery()).
		Joins("JOIN groups_attempts ON groups_attempts.idItem = users_answers.idItem AND groups_attempts.ID = users_answers.idAttempt").
		// if items.bHasAttempts = 1, then groups_attempts.idGroup should be one of the authorized user's groups,
		// otherwise groups_attempts.idGroup should be equal to the user's self group
		Where("IF(items.bHasAttempts, groups_attempts.idGroup IN ?, groups_attempts.idGroup = ?)",
			usersGroupsQuery.SubQuery(), user.SelfGroupID)
}
