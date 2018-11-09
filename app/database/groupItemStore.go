package database

import "github.com/jmoiron/sqlx"

// GroupItemStore implements database operations on `groups_items`
type GroupItemStore struct {
	db *DB
}

// NewGroupItemStore returns a GroupItemStore
func NewGroupItemStore(db *DB) *GroupItemStore {
	return &GroupItemStore{db}
}

func (s *GroupItemStore) createRaw(tx *sqlx.Tx, itemID int, itemType string) error {
	_, err := tx.Exec("INSERT INTO items (ID, sType) VALUES ($1, $2)", itemID, itemType)
	return err
}
