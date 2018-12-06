package database

// GroupStore implements database operations on groups
type GroupStore struct {
  *DataStore
}

// GetAll returns all groups
func (s *GroupStore) GetAll(dest interface{}) error {
  return s.db.Raw("SELECT ID, sName FROM groups").Scan(dest).Error
}
