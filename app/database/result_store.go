package database

// ResultStore implements database operations on `results`
type ResultStore struct {
	*DataStore
}

func (s *ResultStore) ExistsForUserTeam(user *User, participantTeamID, attemptID, itemID int64) (bool, error) {
	return s.Groups().ByID(participantTeamID).Where("type = 'Team'").
		Joins("JOIN groups_groups_active ON groups_groups_active.parent_group_id = groups.id").
		Joins(`
				JOIN results ON results.participant_id = groups.id AND
					results.attempt_id = ? AND results.item_id = ?`, attemptID, itemID).
		Where("groups_groups_active.child_group_id = ?", user.GroupID).HasRows()
}

func (s *ResultStore) ByID(participantID, attemptID, itemID int64) *DB {
	return s.Where("participant_id = ? AND attempt_id = ? AND item_id = ?", participantID, attemptID, itemID)
}
