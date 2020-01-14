package database

import "github.com/jinzhu/gorm"

// AttemptStore implements database operations on `attempts`
type AttemptStore struct {
	*DataStore
}

// CreateNew creates inserts a new row into attempts with group_id=groupID, item_id=itemID.
// It also sets order, started_at, latest_activity_at
func (s *AttemptStore) CreateNew(groupID, itemID, creatorID int64) (newID int64, err error) {
	s.mustBeInTransaction()
	recoverPanics(&err)

	mustNotBeError(s.Exec(
		"SET @maxIOrder = IFNULL((SELECT MAX(`order`) FROM `attempts` WHERE `group_id` = ? AND item_id = ? FOR UPDATE), 0)",
		groupID, itemID).Error())

	mustNotBeError(s.DB.retryOnDuplicatePrimaryKeyError(func(db *DB) error {
		store := NewDataStore(db)
		newID = store.NewID()
		return store.db.Exec(`
			INSERT INTO attempts (id, group_id, item_id, creator_id, `+"`order`)"+`VALUES (?, ?, ?, ?, @maxIOrder+1)`,
			newID, groupID, itemID, creatorID).Error
	}))
	return newID, nil
}

// GetAttemptItemIDIfUserHasAccess returns attempts.item_id if:
//  1) the user has at least 'content' access to this item
//  2) the user is a member of attempts.group_id or the user's group_id = attempts.group_id
func (s *AttemptStore) GetAttemptItemIDIfUserHasAccess(attemptID int64, user *User) (found bool, itemID int64, err error) {
	recoverPanics(&err)
	mustNotBeError(err)
	usersGroupsQuery := s.GroupGroups().WhereUserIsMember(user).Select("parent_group_id")
	err = s.Items().WhereUserHasViewPermissionOnItems(user, "content").
		Joins("JOIN attempts ON attempts.item_id = items.id AND attempts.id = ?", attemptID).
		Where("attempts.group_id = ? OR attempts.group_id IN ?",
			user.GroupID, usersGroupsQuery.SubQuery()).
		PluckFirst("items.id", &itemID).Error()
	if gorm.IsRecordNotFoundError(err) {
		return false, 0, nil
	}
	mustNotBeError(err)
	return true, itemID, nil
}
