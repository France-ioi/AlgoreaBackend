package database

import (
	"github.com/jinzhu/gorm"
)

// GroupStore implements database operations on groups
type GroupStore struct {
  *DataStore
}

// All creates a composable query without filtering
func (s *GroupStore) All() *gorm.DB {
  return s.db.Table("groups")
}
