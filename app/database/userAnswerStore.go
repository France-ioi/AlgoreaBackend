package database

// UserAnswerStore implements database operations on `users_answers`
type UserAnswerStore struct {
	*DataStore
}

// All creates a composable query without filtering
func (s *UserAnswerStore) All() DB {
	return s.table("users_answers")
}

// WithUsers creates a composable query for getting answers joined with users
func (s *UserAnswerStore) WithUsers() DB {
	return s.All().Joins("JOIN users ON users.ID = users_answers.idUser")
}
