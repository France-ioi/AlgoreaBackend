package database

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
