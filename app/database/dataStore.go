package database

// DataStore gather all stores for database operations on business data
type DataStore struct {
	db     *DB
	Groups *GroupStore
	Items  *ItemStore
}

// NewDataStore returns a DataStore
func NewDataStore(db *DB) *DataStore {
	return &DataStore{
		db,
		NewGroupStore(db),
		NewItemStore(db),
	}
}
