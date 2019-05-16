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
