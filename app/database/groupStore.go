package database

import (
	"github.com/jmoiron/sqlx"
)

// GroupStore implements database operations on groups
type GroupStore struct {
	db *sqlx.DB
}

// NewGroupStore returns a GroupStore
func NewGroupStore(db *sqlx.DB) *GroupStore {
	return &GroupStore{db}
}

// GetAll returns all groups
func (s *GroupStore) GetAll(dest interface{}) error {
	return s.db.Select(dest, "SELECT ID, sName FROM groups")
}
