package database

// WhereUsersAreDescendantsOfGroup joins `groups_ancestors_active` on ancestor_group_id=groupID & child_group_id=users.group_id.
func (conn *DB) WhereUsersAreDescendantsOfGroup(groupID int64) *DB {
	return conn.
		Joins("JOIN groups_ancestors_active ON groups_ancestors_active.child_group_id=users.group_id").
		Where("groups_ancestors_active.ancestor_group_id = ?", groupID)
}

// CheckIfTeamParticipationsConflictWithExistingUserMemberships returns true if the given team
// has at least one active participation conflicting with active participations of the given user's teams.
func (s *DataStore) CheckIfTeamParticipationsConflictWithExistingUserMemberships(
	teamID, userGroupID int64, withLock bool,
) (bool, error) {
	teamAttemptsQuery := s.Attempts().
		Where("attempts.participant_id = ?", teamID).Where("root_item_id IS NOT NULL").
		Group("root_item_id")
	if withLock {
		teamAttemptsQuery = teamAttemptsQuery.WithExclusiveWriteLock()
	}

	query := s.ActiveGroupGroups().Where("child_group_id = ?", userGroupID).
		Where("is_team_membership = 1").
		Joins("JOIN (?) AS team_attempts", teamAttemptsQuery. // all the team's attempts (not only active ones)
									Select("root_item_id AS item_id, MAX(NOW() < attempts.allows_submissions_until) AS is_active").QueryExpr()).
		Joins("JOIN items ON items.id = team_attempts.item_id").
		Joins("JOIN attempts ON attempts.participant_id = parent_group_id AND attempts.root_item_id = items.id").
		Where("parent_group_id != ?", teamID).
		Where("(team_attempts.is_active AND NOW() < attempts.allows_submissions_until) OR NOT items.allows_multiple_attempts")
	if withLock {
		query = query.WithExclusiveWriteLock()
	}
	return query.HasRows()
}
