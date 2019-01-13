package database

import (
	"fmt"

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

func (s *ItemItemStore) tableName() string {
	return "items_items"
}

// All creates a composable query without filtering
func (s *ItemItemStore) All() DB {
	return s.table(s.tableName())
}

// Insert does a INSERT query in the given table with data that may contain types.* types
func (s *ItemItemStore) Insert(data *ItemItem) error {
	return s.insert(s.tableName(), data)
}

// ChildrenOf .
func (s *ItemItemStore) ChildrenOf(parentID int64) ([]*ItemItem, error) {
	var ii []*ItemItem
	err := s.db.Table(s.tableName()).Where("idItemParent=?", parentID).Scan(&ii).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get children of item '%d': %v", parentID, err)
	}
	return ii, nil
}
