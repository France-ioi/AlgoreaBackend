package database

import (
	"github.com/jmoiron/sqlx"
)

// ItemStore implements database operations on items
type ItemStore struct {
	db *DB
}

// NewItemStore returns a GroupStore
func NewItemStore(db *DB) *ItemStore {
	return &ItemStore{db}
}

func (s *ItemStore) Create(
	itemID int,
	itemType string,
	languageID int,
	title string,
	parentID int,
	order int,
) error {

	// groupItemStore := NewGroupItemStore(s.db)
	// itemItemStore := NewItemItemStore(s.db)
	// itemStringStore := NewItemStringStore(s.db)

	return s.db.inTransaction(func(tx *sqlx.Tx) error {
		var err error
		if err = s.createRaw(tx, itemID, itemType); err != nil {
			return err
		}
		// if err = groupItemStore.createRaw(tx, ...); err != nil {
		// 	return err
		// }
		// if err = itemStringStore.createRaw(tx, itemID, languageID, title); err != nil {
		// 	return err
		// }
		// if err = itemItemStore.createRaw(tx, ...); err != nil {
		// 	return err
		// }
		return nil
	})
}

func (s *ItemStore) createRaw(tx *sqlx.Tx, itemID int, itemType string) error {
	_, err := tx.Exec("INSERT INTO items (ID, sType, bTeamsEditable, bNoScore, iVersion) VALUES (?, ?, ?, ?, ?)", itemID, itemType, false, false, 0)
	return err
}
