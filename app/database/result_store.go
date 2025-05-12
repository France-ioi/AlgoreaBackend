package database

import "github.com/France-ioi/AlgoreaBackend/v2/golang"

// ResultStore implements database operations on `results`.
type ResultStore struct {
	*DataStore
}

// ByID returns a composable query for getting a result row by the primary key (participant_id, attemptID, itemID).
func (s *ResultStore) ByID(participantID, attemptID, itemID int64) *DB {
	return s.Where("results.participant_id = ? AND results.attempt_id = ? AND results.item_id = ?", participantID, attemptID, itemID)
}

// HintsInfo contains information on requested hints and their cached count.
type HintsInfo struct {
	HintsRequested *string
	HintsCached    int32
}

// GetHintsInfoForActiveAttempt returns HintsInfo of the result
// identified by given participantID, attemptID, itemID and linked to an active attempt.
// If such a result doesn't exist, the gorm.ErrRecordNotFound error is returned.
func (s *ResultStore) GetHintsInfoForActiveAttempt(participantID, attemptID, itemID int64) (result *HintsInfo, err error) {
	s.mustBeInTransaction()
	defer recoverPanics(&err)

	var hintsInfo HintsInfo
	mustNotBeError(s.Results().
		ByID(participantID, attemptID, itemID).
		WithCustomWriteLocks(golang.NewSet("attempts"), golang.NewSet("results")).
		Select("hints_requested, hints_cached").
		Joins("JOIN attempts ON attempts.participant_id = results.participant_id AND attempts.id = results.attempt_id").
		Where("NOW() < attempts.allows_submissions_until").
		Scan(&hintsInfo).Error())
	return &hintsInfo, nil
}

// MarkAsToBePropagated marks a given result as 'to_be_propagated'.
func (s *ResultStore) MarkAsToBePropagated(participantID, attemptID, itemID int64, propagateNow bool) error {
	err := s.Exec(`
		INSERT IGNORE INTO `+s.resultsPropagateTableName()+
		` (`+golang.If(s.arePropagationsSync(), "connection_id, ")+`participant_id, attempt_id, item_id, state)
		VALUES(`+golang.If(s.arePropagationsSync(), "CONNECTION_ID(), ")+`?, ?, ?, 'to_be_propagated')`, participantID, attemptID, itemID).Error()
	if err == nil && propagateNow {
		s.ScheduleResultsPropagation()
	}
	return err
}

func (s *ResultStore) resultsPropagateTableName() string {
	return golang.IfElse(s.arePropagationsSync(), "results_propagate_sync_conn", "results_propagate")
}
