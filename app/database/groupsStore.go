package database

import (
	"github.com/jmoiron/sqlx"
)

// GroupsStore implements database operations on groups
type GroupsStore struct {
	db *sqlx.DB
}

// NewGroupStore returns a GroupStore
func NewGroupStore(db *sqlx.DB) *GroupsStore {
	return &GroupsStore{db}
}

// GetAll returns all groups
func (s *GroupsStore) GetAll(dest interface{}) error {
	return s.db.Select(dest, "SELECT ID, sName FROM groups")
}
