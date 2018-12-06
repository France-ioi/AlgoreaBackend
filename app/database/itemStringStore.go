package database

import t "github.com/France-ioi/AlgoreaBackend/app/types"

// ItemStringStore implements database operations on `items_strings`
type ItemStringStore struct {
  *DataStore
}

// ItemString matches the content the `items_strings` table
type ItemString struct {
  ID         t.Int64  `db:"ID"`
  ItemID     t.Int64  `db:"idItem"`
  LanguageID t.Int64  `db:"idLanguage"`
  Title      t.String `db:"sTitle"`
  Version    int64    `db:"iVersion"` // use Go default in DB (to be fixed)
}

func (s *ItemStringStore) createRaw(entry *ItemString) (int64, error) {
  if !entry.ID.Set {
    entry.ID = *t.NewInt64(generateID())
  }
  err := s.db.insert("items_strings", entry)
  return entry.ID.Value, err
}
