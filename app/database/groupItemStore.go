package database

import (
	"github.com/jmoiron/sqlx"
)

// GroupItemStore implements database operations on `groups_items`
type GroupItemStore struct {
	db *DB
}

// NewGroupItemStore returns a GroupItemStore
func NewGroupItemStore(db *DB) *GroupItemStore {
	return &GroupItemStore{db}
}

func (s *GroupItemStore) createRaw(tx *sqlx.Tx, itemID int) error {
	_, err := tx.Exec("INSERT INTO groups_items (ID, idGroup, idItem, sFullAccessDate, bCachedFullAccess, bOwnerAccess, idUserCreated, iVersion) VALUES (?, ?, ?, ?, ?, ?, ?, ?)", 1, 6, itemID, "2018-01-01 00:00:00", true, true, 9, 0)
	return err
}
