package database

import (
	"github.com/France-ioi/AlgoreaBackend/app/types"
)

// ItemItemStore implements database operations on `items_items`
type ItemItemStore struct {
	*DataStore
}

// ItemItem matches the content the `items_items` table
type ItemItem struct {
	ID           types.Int64 `sql:"column:ID"`
	ParentItemID types.Int64 `sql:"column:idItemParent"`
	ChildItemID  types.Int64 `sql:"column:idItemChild"`
	Order        types.Int64 `sql:"column:iChildOrder"`
	Difficulty   int64       `sql:"column:iDifficulty"` // use Go default in DB (to be fixed)
	Version      int64       `sql:"column:iVersion"`    // use Go default in DB (to be fixed)
}

// Insert does a INSERT query in the given table with data that may contain types.* types
func (s *ItemItemStore) Insert(data *ItemItem) error {
	return s.insert(s.tableName, data)
}

// ChildrenOf returns a composable query for selecting children of the given item
func (s *ItemItemStore) ChildrenOf(parentID int64) *ItemItemStore {
	return &ItemItemStore{
		NewDataStoreWithTable(s.Where("items_items.idItemParent=?", parentID), s.tableName),
	}
}
