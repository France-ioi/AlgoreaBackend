package database

// ThreadStore implements database operations on threads.
type ThreadStore struct {
	*DataStore
}

// UpdateHelperGroupID updates all occurrences of a certain helper_group_id to a new value.
func (s *ThreadStore) UpdateHelperGroupID(oldHelperGroupID, newHelperGroupID int64) {
	var err error

	s.mustBeInTransaction()
	defer recoverPanics(&err)

	err = s.Threads().
		Where("helper_group_id = ?", oldHelperGroupID).
		UpdateColumn("helper_group_id", newHelperGroupID).
		Error()
	mustNotBeError(err)
}
