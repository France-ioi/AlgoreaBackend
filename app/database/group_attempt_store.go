package database

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
