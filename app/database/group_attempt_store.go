package database

import "github.com/jinzhu/gorm"

// GroupAttemptStore implements database operations on `groups_attempts`
type GroupAttemptStore struct {
	*DataStore
}

// After is a "listener" that calls UserItemStore::PropagateAttempts() & UserItemStore::ComputeAllUserItems()
func (s *GroupAttemptStore) After() error {
	s.mustBeInTransaction()

	if err := s.UserItems().PropagateAttempts(); err != nil {
		return err
	}
	if err := s.UserItems().ComputeAllUserItems(); err != nil {
		return err
	}
	return nil
}

// CreateNew creates inserts a new row into groups_attempts with idGroup=groupID, idItem=itemID.
// It also sets iOrder, sStartDate, sLastActivityDate
func (s *GroupAttemptStore) CreateNew(groupID, itemID int64) (newID int64, err error) {
	s.mustBeInTransaction()
	recoverPanics(&err)

	mustNotBeError(s.Exec(
		"SET @maxIOrder = IFNULL((SELECT MAX(iOrder) FROM `groups_attempts` WHERE `idGroup` = ? AND idItem = ? FOR UPDATE), 0)",
		groupID, itemID).Error())

	mustNotBeError(s.DB.retryOnDuplicatePrimaryKeyError(func(db *DB) error {
		store := NewDataStore(db)
		newID = store.NewID()
		return store.db.Exec(`
			INSERT INTO groups_attempts (ID, idGroup, idItem, iOrder, sStartDate, sLastActivityDate)
			VALUES (?, ?, ?, @maxIOrder+1, NOW(), NOW())`,
			newID, groupID, itemID).Error
	}))
	return newID, nil
}

// GetAttemptItemIDIfUserHasAccess returns groups_attempts.idItem if:
//  1) the user has at least partial access to this item
//  2) the user is a member of groups_attempts.idGroup  (if items.bHasAttempts = 1)
//  3) the user's idGroupSelf = groups_attempts.idGroup (if items.bHasAttempts = 0)
func (s *GroupAttemptStore) GetAttemptItemIDIfUserHasAccess(attemptID int64, user *User) (found bool, itemID int64, err error) {
	recoverPanics(&err)
	selfGroupID, err := user.SelfGroupID()
	if err == ErrUserNotFound {
		return false, 0, nil
	}
	mustNotBeError(err)
	usersGroupsQuery := s.GroupGroups().WhereUserIsMember(user).Select("idGroupParent")
	err = s.Items().Visible(user).
		Joins("JOIN groups_attempts ON groups_attempts.idItem = items.ID AND groups_attempts.ID = ?", attemptID).
		Joins("JOIN users_items ON users_items.idItem = items.ID AND users_items.idUser = ?", user.UserID).
		Where("partialAccess > 0 OR fullAccess > 0").
		Where("IF(items.bHasAttempts, groups_attempts.idGroup IN ?, groups_attempts.idGroup = ?)",
			usersGroupsQuery.SubQuery(), selfGroupID).
		PluckFirst("items.ID", &itemID).Error()
	if gorm.IsRecordNotFoundError(err) {
		return false, 0, nil
	}
	mustNotBeError(err)
	return true, itemID, nil
}
