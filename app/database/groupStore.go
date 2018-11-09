package database

// GroupStore implements database operations on groups
type GroupStore struct {
	db *DB
}

// NewGroupStore returns a GroupStore
func NewGroupStore(db *DB) *GroupStore {
	return &GroupStore{db}
}

// GetAll returns all groups
func (s *GroupStore) GetAll(dest interface{}) error {
	return s.db.Select(dest, "SELECT ID, sName FROM groups")
}
