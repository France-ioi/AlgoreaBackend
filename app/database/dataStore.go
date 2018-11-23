package database

import (
  "math/rand"
)

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
    &GroupStore{db},
    &ItemStore{db},
  }
}

func generateID() int64 {
  // gen a 63-bits number as we want unsigned number stored in a 64-bits signed DB attribute
  // !!! to be safe, the insertion should be be retried if the ID conflicts with an existing entry
  return rand.Int63()
}
