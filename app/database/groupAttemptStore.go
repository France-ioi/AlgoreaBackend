package database

// GroupAttemptStore implements database operations on `groups_attempts`
type GroupAttemptStore struct {
	*DataStore
}

// ByID returns a composable query for filtering by groups_attempts.ID
func (s *GroupAttemptStore) ByID(attemptID int64) *DB {
	return s.Where("groups_attempts.ID = ?", attemptID)
}
