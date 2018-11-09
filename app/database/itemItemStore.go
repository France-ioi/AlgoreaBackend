package database

import (
	"github.com/jmoiron/sqlx"
)

// ItemItemStore implements database operations on `items_items`
type ItemItemStore struct {
	db *DB
}

// NewItemItemStore returns a ItemItemStore
func NewItemItemStore(db *DB) *ItemItemStore {
	return &ItemItemStore{db}
}

func (s *ItemItemStore) createRaw(tx *sqlx.Tx, parentID int, childID int, order int) error {
	// FIXME dummy ID and version
	_, err := tx.Exec("INSERT INTO items_items (ID, idItemParent, idItemChild, iChildOrder, iDifficulty, iVersion) VALUES (?, ?, ?, ?, ?, ?)", 1, parentID, childID, order, 0, 0)
	return err
}
