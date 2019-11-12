package database

import "github.com/jinzhu/gorm"

// GroupAttemptStore implements database operations on `groups_attempts`
type GroupAttemptStore struct {
	*DataStore
}

// CreateNew creates inserts a new row into groups_attempts with group_id=groupID, item_id=itemID.
// It also sets order, started_at, latest_activity_at
func (s *GroupAttemptStore) CreateNew(groupID, itemID int64) (newID int64, err error) {
	s.mustBeInTransaction()
	recoverPanics(&err)

	mustNotBeError(s.Exec(
		"SET @maxIOrder = IFNULL((SELECT MAX(`order`) FROM `groups_attempts` WHERE `group_id` = ? AND item_id = ? FOR UPDATE), 0)",
		groupID, itemID).Error())

	mustNotBeError(s.DB.retryOnDuplicatePrimaryKeyError(func(db *DB) error {
		store := NewDataStore(db)
		newID = store.NewID()
		return store.db.Exec(`
			INSERT INTO groups_attempts (id, group_id, item_id, `+"`order`"+`, started_at, latest_activity_at)
			VALUES (?, ?, ?, @maxIOrder+1, NOW(), NOW())`,
			newID, groupID, itemID).Error
	}))
	return newID, nil
}

// GetAttemptItemIDIfUserHasAccess returns groups_attempts.item_id if:
//  1) the user has at least 'content' access to this item
//  2) the user is a member of groups_attempts.group_id  (if items.has_attempts = 1)
//  3) the user's group_id = groups_attempts.group_id (if items.has_attempts = 0)
func (s *GroupAttemptStore) GetAttemptItemIDIfUserHasAccess(attemptID int64, user *User) (found bool, itemID int64, err error) {
	recoverPanics(&err)
	mustNotBeError(err)
	usersGroupsQuery := s.GroupGroups().WhereUserIsMember(user).Select("parent_group_id")
	err = s.Items().WhereUserHasViewPermissionOnItems(user, "content").
		Joins("JOIN groups_attempts ON groups_attempts.item_id = items.id AND groups_attempts.id = ?", attemptID).
		Where("IF(items.has_attempts, groups_attempts.group_id IN ?, groups_attempts.group_id = ?)",
			usersGroupsQuery.SubQuery(), user.GroupID).
		PluckFirst("items.id", &itemID).Error()
	if gorm.IsRecordNotFoundError(err) {
		return false, 0, nil
	}
	mustNotBeError(err)
	return true, itemID, nil
}

// VisibleAndByItemID returns a composable query for getting groups_attempts with the following access rights
// restrictions:
// 1) the user should have at least 'content' access rights to the groups_attempts.item_id item,
// 2) the user is able to see answers related to his group's attempts, so:
//   (a) if items.has_attempts = 1, then the user should be a member of the groups_attempts.group_id team
//   (b) if items.has_attempts = 0, then groups_attempts.group_id should be equal to the user's self group
func (s *GroupAttemptStore) VisibleAndByItemID(user *User, itemID int64) *DB {
	usersGroupsQuery := s.GroupGroups().WhereUserIsMember(user).Select("parent_group_id")
	// the user should have at least 'content' access to the item
	itemsQuery := s.Items().ByID(itemID).WhereUserHasViewPermissionOnItems(user, "content")

	return s.
		// the user should have at least 'content' access to the users_answers.item_id
		Joins("JOIN ? AS items ON items.id = groups_attempts.item_id", itemsQuery.SubQuery()).
		// if items.has_attempts = 1, then groups_attempts.group_id should be one of the authorized user's groups,
		// otherwise groups_attempts.group_id should be equal to the user's self group
		Where("IF(items.has_attempts, groups_attempts.group_id IN ?, groups_attempts.group_id = ?)",
			usersGroupsQuery.SubQuery(), user.GroupID)
}
