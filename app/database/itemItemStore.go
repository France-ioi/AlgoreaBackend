package database

import t "github.com/France-ioi/AlgoreaBackend/app/types"

// ItemItemStore implements database operations on `items_items`
type ItemItemStore struct {
	db *DB
}

type ItemItem struct {
	ID           t.Int64 `db:"ID"`
	ParentItemID t.Int64 `db:"idItemParent"`
	ChildItemID  t.Int64 `db:"idItemChild"`
	Order        t.Int64 `db:"iChildOrder"`
	Difficulty   int64   `db:"iDifficulty"` // when the db does not know the default, they will get the go type default
	Version      int64   `db:"iVersion"`    // when the db does not know the default, they will get the go type default
}

func (s *ItemItemStore) createRaw(tx Tx, entry *ItemItem) (int64, error) {
	entry.ParentItemID = *t.NewInt64(4) // dummy

	if !entry.ID.Set {
		entry.ID = *t.NewInt64(generateID())
	}
	err := tx.insert("items_items", entry)
	return entry.ID.Value, err
}
