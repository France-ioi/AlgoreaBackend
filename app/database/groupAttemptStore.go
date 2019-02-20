package database

// GroupAttemptStore implements database operations on `groups_attempts`
type GroupAttemptStore struct {
	*DataStore
}

// All creates a composable query without filtering
func (s *GroupAttemptStore) All() DB {
	return s.Table("groups_attempts")
}

// ByAttemptID returns a composable query for filtering by idAttempt
func (s *GroupAttemptStore) ByAttemptID(attemptID int64) DB {
	return s.All().Where("ID = ?", attemptID)
}
