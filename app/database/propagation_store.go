package database

const PropagationID = 1

// PropagationStore implements database operations on `propagations`
// (used for the propagation system).
type PropagationStore struct {
	*DataStore
}

// ByID checks whether there is an async propagation scheduled.
func (s *PropagationStore) ByID(propagationID int64) *DB {
	return s.Where("propagation_id = ?", propagationID)
}

// AsyncPropagationScheduled checks whether there is an async propagation scheduled.
func (s *PropagationStore) AsyncPropagationScheduled() bool {
	var asyncPropagationScheduled bool
	err := s.
		ByID(PropagationID).
		Select("propagate").
		PluckFirst("propagate", &asyncPropagationScheduled).
		Error()
	mustNotBeError(err)

	return asyncPropagationScheduled
}

// ScheduleAsyncPropagation schedules an async propagation.
// AsyncPropagationScheduled() becomes true after this function is called.
func (s *PropagationStore) ScheduleAsyncPropagation() {
	// We need to execute an UPDATE statement, so the BEFORE UPDATE trigger runs.
	err := s.
		ByID(PropagationID).
		UpdateColumn(map[string]interface{}{
			"propagation_id": PropagationID,
		}).
		Error()
	mustNotBeError(err)
}
