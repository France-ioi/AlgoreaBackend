package database

// GroupStore implements database operations on groups
type GroupStore struct {
	*DataStore
}

// All creates a composable query without filtering
func (s *GroupStore) All() DB {
	return s.table("groups")
}
