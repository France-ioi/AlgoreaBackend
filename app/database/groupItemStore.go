package database

import t "github.com/France-ioi/AlgoreaBackend/app/types"

// GroupItemStore implements database operations on `groups_items`
type GroupItemStore struct {
	db *DB
}

type GroupItem struct {
	ID               t.Int64 `db:"ID"`
	GroupID          t.Int64 `db:"idGroup"`
	ItemID           t.Int64 `db:"idItem"`
	FullAccessDate   string  `db:"sFullAccessDate"`   // should be a datetime
	CachedFullAccess bool    `db:"bCachedFullAccess"` // when the db does not know the default, they will get the go type default
	OwnerAccess      bool    `db:"bOwnerAccess"`      // when the db does not know the default, they will get the go type default
	CreatedUserID    int64   `db:"idUserCreated"`     // when the db does not know the default, they will get the go type default
	Version          int64   `db:"iVersion"`          // when the db does not know the default, they will get the go type default
}

func (s *GroupItemStore) createRaw(tx Tx, entry *GroupItem) (int64, error) {
	entry.FullAccessDate = "2018-01-01 00:00:00" // dummy
	entry.GroupID = *t.NewInt64(6)               // dummy
	if !entry.ID.Set {
		entry.ID = *t.NewInt64(generateID())
	}
	err := tx.insert("groups_items", entry)
	return entry.ID.Value, err
}
