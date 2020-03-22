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
	return s.Where("participant_id = ? AND attempt_id = ? AND item_id = ?", participantID, attemptID, itemID)
}
