package database

import (
  "math/rand"
)

// DataStore gather all stores for database operations on business data
type DataStore struct {
  db     *DB
}

// NewDataStore returns a DataStore
func NewDataStore(db *DB) *DataStore {
  return &DataStore{db}
}

// Users returns a UserStore
func (s *DataStore) Users() *UserStore {
  return &UserStore{s}
}

// Items returns a ItemStore
func (s *DataStore) Items() *ItemStore {
  return &ItemStore{s}
}

// Groups returns a GroupStore
func (s *DataStore) Groups() *GroupStore {
  return &GroupStore{s}
}

func generateID() int64 {
  // gen a 63-bits number as we want unsigned number stored in a 64-bits signed DB attribute
  // !!! to be safe, the insertion should be be retried if the ID conflicts with an existing entry
  return rand.Int63()
}
