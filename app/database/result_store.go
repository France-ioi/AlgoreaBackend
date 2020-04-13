package database

// ResultStore implements database operations on `results`
type ResultStore struct {
	*DataStore
}

// ExistsForUserTeam checks if the given participantTeamID is a team of the given user and
// there is at least one result for this team linked to the given attempt & item
func (s *ResultStore) ExistsForUserTeam(user *User, participantTeamID, attemptID, itemID int64) (bool, error) {
	return s.Groups().ByID(participantTeamID).Where("type = 'Team'").
		Joins("JOIN groups_groups_active ON groups_groups_active.parent_group_id = groups.id").
		Joins(`
				JOIN results ON results.participant_id = groups.id AND
					results.attempt_id = ? AND results.item_id = ?`, attemptID, itemID).
		Where("groups_groups_active.child_group_id = ?", user.GroupID).HasRows()
}

// ByID returns a composable query for getting a result row by the primary key (participant_id, attemptID, itemID)
func (s *ResultStore) ByID(participantID, attemptID, itemID int64) *DB {
	return s.Where("results.participant_id = ? AND results.attempt_id = ? AND results.item_id = ?", participantID, attemptID, itemID)
}

// HintsInfo contains information on requested hints and their cached count
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
		WithWriteLock().Select("hints_requested, hints_cached").
		Joins("JOIN attempts ON attempts.participant_id = results.participant_id AND attempts.id = results.attempt_id").
		Where("NOW() < attempts.allows_submissions_until").
		Scan(&hintsInfo).Error())
	return &hintsInfo, nil
}
