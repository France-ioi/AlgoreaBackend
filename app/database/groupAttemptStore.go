package database

// GroupAttemptStore implements database operations on `groups_attempts`
type GroupAttemptStore struct {
	*DataStore
}

// ByAttemptID returns a composable query for filtering by idAttempt
func (s *GroupAttemptStore) ByAttemptID(attemptID int64) *DB {
	return s.Where("ID = ?", attemptID)
}
