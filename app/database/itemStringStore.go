package database

import (
	"fmt"

	"github.com/France-ioi/AlgoreaBackend/app/types"
)

// ItemStringStore implements database operations on `items_strings`
type ItemStringStore struct {
	*DataStore
}

// ItemString matches the content the `items_strings` table
type ItemString struct {
	ID          types.Int64  `sql:"column:ID"`
	ItemID      types.Int64  `sql:"column:idItem"`
	LanguageID  types.Int64  `sql:"column:idLanguage"`
	Title       types.String `sql:"column:sTitle"`
	ImageURL    types.String `sql:"column:sImageUrl"`
	Subtitle    types.String `sql:"column:sSubtitle"`
	Description types.String `sql:"column:sDescription"`
	Version     int64        `sql:"column:iVersion"` // use Go default in DB (to be fixed)
}

func (s *ItemStringStore) tableName() string {
	return "items_strings"
}

func (s *ItemStringStore) GetByItemID(id int64) (*ItemString, error) {
	var it ItemString
	if err := s.db.Table(s.tableName()).Where("idItem=?", id).First(&it).Error; err != nil {
		return nil, fmt.Errorf("failed to get item_string of item '%d': %v", id, err)
	}
	return &it, nil
}

func (s *ItemStringStore) GetByItemIDs(ids []int64) ([]*ItemString, error) {
	var itt []*ItemString
	if err := s.db.Table(s.tableName()).Where("idItem IN (?)", ids).Scan(&itt).Error; err != nil {
		return nil, fmt.Errorf("failed to get item_string of items %v: %v", ids, err)
	}
	return itt, nil
}

// Insert does a INSERT query in the given table with data that may contain types.* types
func (s *ItemStringStore) Insert(data *ItemString) error {
	return s.insert(s.tableName(), data)
}

// All creates a composable query without filtering
func (s *ItemStringStore) All() DB {
	return s.table(s.tableName())
}
