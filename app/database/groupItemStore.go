package database

import (
	"github.com/jinzhu/gorm"

	"github.com/France-ioi/AlgoreaBackend/app/auth"
	t "github.com/France-ioi/AlgoreaBackend/app/types"
)

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

// All creates a composable query without filtering
func (s *GroupItemStore) All() *gorm.DB {
	return s.db.Table("groups_items")
}

// MatchingUserAncestors returns a composable query of group items matching groups of which the user is member
func (s *GroupItemStore) MatchingUserAncestors(user *auth.User) *gorm.DB {
	userAncestors := s.GroupAncestors().UserAncestors(user).SubQuery()
	return s.All().Joins("JOIN ? AS ancestors ON groups_items.idGroup = ancestors.idGroupAncestor", userAncestors)
}
