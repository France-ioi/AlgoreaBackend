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
