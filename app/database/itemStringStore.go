package database

import (
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

// InsertData does a INSERT query in the given table with data that may contain types.* types
func (s *ItemStringStore) InsertData(data *ItemString) error {
	return s.DB.Insert(s.tableName(), data)
}

// All creates a composable query without filtering
func (s *ItemStringStore) All() DB {
	return s.Table(s.tableName())
}
