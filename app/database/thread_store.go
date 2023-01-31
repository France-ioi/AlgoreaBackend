package database

// ThreadStore implements database operations on threads
type ThreadStore struct {
	*DataStore
}

// UpdateHelperGroupID updates all occurrences of a certain helper_group_id to a new value
func (s *ThreadStore) UpdateHelperGroupID(oldHelperGroupID int64, newHelperGroupID int64) (err error) {
	s.mustBeInTransaction()
	defer recoverPanics(&err)

	err = s.Threads().
		Where("helper_group_id = ?", oldHelperGroupID).
		UpdateColumn("helper_group_id", newHelperGroupID).
		Error()
	if err != nil {
		return err
	}

	return nil
}
