package database

import (
	"github.com/France-ioi/AlgoreaBackend/app/auth"
	"github.com/jinzhu/gorm"
)

// GroupAncestorStore implements database operations on `groups_ancestors` (which is a precomputed cache over groups_groups)
type GroupAncestorStore struct {
  *DataStore
}

// All creates a composable query without filtering
func (s *GroupAncestorStore) All() *gorm.DB {
  return s.db.Table("groups_ancestors")
}

// UserAncestors returns a composable query of ancestors of user's self group, i.e. groups of which he is a member
func (s *GroupAncestorStore) UserAncestors(user *auth.User) *gorm.DB {
  return s.All().Where("idGroupChild = ?", user.SelfGroupID())
}
