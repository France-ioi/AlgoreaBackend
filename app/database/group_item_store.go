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
	ID             types.Int64    `sql:"column:ID"`
	GroupID        types.Int64    `sql:"column:idGroup"`
	ItemID         types.Int64    `sql:"column:idItem"`
	CreatorUserID  types.Int64    `sql:"column:idUserCreated"`
	FullAccessDate types.Datetime `sql:"column:sFullAccessDate"`
	OwnerAccess    types.Bool     `sql:"column:bOwnerAccess"`
	ManagerAccess  types.Bool     `sql:"column:bManagerAccess"`
	// computed fields
	CachedFullAccessDate types.Datetime `sql:"column:sCachedFullAccessDate"`
	CachedFullAccess     types.Bool     `sql:"column:bCachedFullAccess"`

	Version int64 `sql:"column:iVersion"` // use Go default in DB (to be fixed)
}

// MatchingUserAncestors returns a composable query of group items matching groups of which the user is member
func (s *GroupItemStore) MatchingUserAncestors(user *User) *DB {
	db := s.GroupAncestors().UserAncestors(user)
	if db.Error() != nil {
		return db
	}
	userAncestors := db.SubQuery()
	return s.Joins("JOIN ? AS ancestors ON groups_items.idGroup = ancestors.idGroupAncestor", userAncestors)
}

func (s *GroupItemStore) after() {
	s.computeAllAccess()
	s.grantCachedAccessWhereNeeded()
}

func (s *GroupItemStore) removePartialAccess(groupID, itemID int64) {
	mustNotBeError(s.Where("idItem = ? AND idGroup = ? AND bManagerAccess = 0", itemID, groupID).
		UpdateColumn(map[string]interface{}{
			"sPartialAccessDate":       nil,
			"sCachedPartialAccessDate": nil,
			"bCachedPartialAccess":     0,
		}).Error())
}
