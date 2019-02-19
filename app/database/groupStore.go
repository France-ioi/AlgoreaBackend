package database

// GroupStore implements database operations on groups
type GroupStore struct {
	*DataStore
}

// ByID returns a composable query of groups filtered by groupID
func (s *GroupStore) ByID(groupID int64) *DB {
	return s.Where("groups.ID = ?", groupID)
}
