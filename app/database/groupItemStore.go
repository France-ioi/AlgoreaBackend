package database

import t "github.com/France-ioi/AlgoreaBackend/app/types"

// GroupItemStore implements database operations on `groups_items`
type GroupItemStore struct {
  *DataStore
}

// GroupItem matches the content the `groups_items` table
type GroupItem struct {
  ID               t.Int64 `db:"ID"`
  GroupID          t.Int64 `db:"idGroup"`
  ItemID           t.Int64 `db:"idItem"`
  FullAccessDate   string  `db:"sFullAccessDate"`   // should be a datetime
  CachedFullAccess bool    `db:"bCachedFullAccess"` // use Go default in DB (to be fixed)
  OwnerAccess      bool    `db:"bOwnerAccess"`      // use Go default in DB (to be fixed)
  CreatedUserID    int64   `db:"idUserCreated"`     // use Go default in DB (to be fixed)
  Version          int64   `db:"iVersion"`          // use Go default in DB (to be fixed)
}

func (s *GroupItemStore) createRaw(entry *GroupItem) (int64, error) {
  entry.FullAccessDate = "2018-01-01 00:00:00" // dummy
  entry.GroupID = *t.NewInt64(6)               // dummy
  if !entry.ID.Set {
    entry.ID = *t.NewInt64(generateID())
  }
  err := s.db.insert("groups_items", entry)
  return entry.ID.Value, err
}
