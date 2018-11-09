package database

import "github.com/jmoiron/sqlx"

// ItemItemStore implements database operations on `items_items`
type ItemItemStore struct {
	db *DB
}

// NewItemItemStore returns a ItemItemStore
func NewItemItemStore(db *DB) *ItemItemStore {
	return &ItemItemStore{db}
}

func (s *ItemItemStore) createRaw(tx *sqlx.Tx, itemID int, itemType string) error {
	_, err := tx.Exec("INSERT INTO items (ID, sType) VALUES ($1, $2)", itemID, itemType)
	return err
}
