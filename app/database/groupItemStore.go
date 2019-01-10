package database

import (
	"github.com/France-ioi/AlgoreaBackend/app/types"
)

// GroupItemStore implements database operations on `groups_items`
type GroupItemStore struct {
	*DataStore
}

// GroupItem matches the content the `groups_items` table
type GroupItem struct {
	ID               types.Int64    `sql:"column:ID"`
	GroupID          types.Int64    `sql:"column:idGroup"`
	ItemID           types.Int64    `sql:"column:idItem"`
	CreatorUserID    types.Int64    `sql:"column:idUserCreated"`
	FullAccessDate   types.Datetime `sql:"column:sFullAccessDate"`
	CachedFullAccess bool           `sql:"column:bCachedFullAccess"` // use Go default in DB (to be fixed)
	OwnerAccess      bool           `sql:"column:bOwnerAccess"`      // use Go default in DB (to be fixed)
	Version          int64          `sql:"column:iVersion"`          // use Go default in DB (to be fixed)
}

func (s *GroupItemStore) tableName() string {
	return "groups_items"
}

// Insert does a INSERT query in the given table with data that may contain types.* types
func (s *GroupItemStore) Insert(data *GroupItem) error {
	return s.insert(s.tableName(), data)
}

// All creates a composable query without filtering
func (s *GroupItemStore) All() DB {
	return s.table(s.tableName())
}

// MatchingUserAncestors returns a composable query of group items matching groups of which the user is member
func (s *GroupItemStore) MatchingUserAncestors(user AuthUser) DB {
	userAncestors := s.GroupAncestors().UserAncestors(user).SubQuery()
	return s.All().Joins("JOIN ? AS ancestors ON groups_items.idGroup = ancestors.idGroupAncestor", userAncestors)
}
