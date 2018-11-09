package database

import "github.com/jmoiron/sqlx"

// ItemStringStore implements database operations on `items_strings`
type ItemStringStore struct {
	db *DB
}

// NewItemStringStore returns a ItemStringStore
func NewItemStringStore(db *DB) *ItemStringStore {
	return &ItemStringStore{db}
}

func (s *ItemStringStore) createRaw(tx *sqlx.Tx, itemID int, langageID int, title string) error {
	_, err := tx.Exec("INSERT INTO items_strings (ID, idItem, idLanguage, sTitle) VALUES ($1, $2, $3, $4)", 1, itemID, langageID, title)
	return err
}
