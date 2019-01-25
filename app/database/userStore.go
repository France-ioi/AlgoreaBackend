package database

// UserStore implements database operations on `users`
type UserStore struct {
	*DataStore
}

// ByID returns a composable query of users filtered by userID
func (s *UserStore) ByID(userID int64) DB {
	return s.All().Where("users.ID = ?", userID)
}

// All creates a composable query without filtering
func (s *UserStore) All() DB {
	return s.table("users")
}
