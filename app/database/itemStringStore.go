package database

import (
	"github.com/jmoiron/sqlx"
)

// ItemStringStore implements database operations on `items_strings`
type ItemStringStore struct {
	db *DB
}

// NewItemStringStore returns a ItemStringStore
func NewItemStringStore(db *DB) *ItemStringStore {
	return &ItemStringStore{db}
}

func (s *ItemStringStore) createRaw(tx *sqlx.Tx, itemID int, languageID int, title string) error {
	// FIXME dummy ID and version
	_, err := tx.Exec("INSERT INTO items_strings (ID, idItem, idLanguage, sTitle, iVersion) VALUES (?, ?, ?, ?, ?)", 1, itemID, languageID, title, 0)
	return err
}
