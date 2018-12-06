package database

import t "github.com/France-ioi/AlgoreaBackend/app/types"

// ItemItemStore implements database operations on `items_items`
type ItemItemStore struct {
  *DataStore
}

// ItemItem matches the content the `items_items` table
type ItemItem struct {
  ID           t.Int64 `db:"ID"`
  ParentItemID t.Int64 `db:"idItemParent"`
  ChildItemID  t.Int64 `db:"idItemChild"`
  Order        t.Int64 `db:"iChildOrder"`
  Difficulty   int64   `db:"iDifficulty"` // use Go default in DB (to be fixed)
  Version      int64   `db:"iVersion"`    // use Go default in DB (to be fixed)
}

func (s *ItemItemStore) createRaw(entry *ItemItem) (int64, error) {
  entry.ParentItemID = *t.NewInt64(4) // dummy

  if !entry.ID.Set {
    entry.ID = *t.NewInt64(generateID())
  }
  err := s.db.insert("items_items", entry)
  return entry.ID.Value, err
}
