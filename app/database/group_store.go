package database

// GroupStore implements database operations on groups
type GroupStore struct {
	*DataStore
}

// ManagedBy returns a composable query for getting all the groups the user can manage.
//
// The `groups` in the result may be duplicated since
// there can be different paths to a managed group through the `group_managers` table and
// the group ancestry graph.
func (s *GroupStore) ManagedBy(user *User) *DB {
	return s.
		Joins(`
			JOIN groups_ancestors_active
				ON groups_ancestors_active.child_group_id = groups.id`).
		Joins(`
			JOIN group_managers
				ON group_managers.group_id = groups_ancestors_active.ancestor_group_id`).
		Joins(`
			JOIN groups_ancestors_active AS user_ancestors
				ON user_ancestors.ancestor_group_id = group_managers.manager_id AND
					user_ancestors.child_group_id = ?`, user.GroupID)
}

// TeamGroupForTeamItemAndUser returns a composable query for getting a team that
//  1) the given user is a member of
//  2) has an unexpired attempt with root_item_id = `itemID`.
// If more than one team is found (which should be impossible), the one with the smallest `groups.id` is returned.
func (s *GroupStore) TeamGroupForTeamItemAndUser(itemID int64, user *User) *DB {
	return s.
		Joins(`JOIN groups_groups_active
			ON groups_groups_active.parent_group_id = groups.id AND
				groups_groups_active.child_group_id = ?`, user.GroupID).
		Joins(`
			JOIN attempts ON attempts.participant_id = groups.id AND
				attempts.root_item_id = ? AND NOW() < attempts.allows_submissions_until`, itemID).
		Where("groups.type = 'Team'").
		Order("groups.id").
		Limit(1) // The current API doesn't allow users to join multiple teams working on the same item
}

// TeamGroupForUser returns a composable query for getting team group of the given user with given id
func (s *GroupStore) TeamGroupForUser(teamGroupID int64, user *User) *DB {
	return s.
		ByID(teamGroupID).
		Joins(`JOIN groups_groups_active
			ON groups_groups_active.parent_group_id = groups.id AND
				groups_groups_active.child_group_id = ?`, user.GroupID).
		Where("groups.type = 'Team'")
}

// CreateNew creates a new group with given name and type.
// It also runs GroupGroupStore.createNewAncestors().
func (s *GroupStore) CreateNew(name, groupType string) (groupID int64, err error) {
	s.mustBeInTransaction()
	defer recoverPanics(&err)
	mustNotBeError(s.RetryOnDuplicatePrimaryKeyError(func(retryStore *DataStore) error {
		groupID = retryStore.NewID()
		return retryStore.Groups().InsertMap(map[string]interface{}{
			"id":         groupID,
			"name":       name,
			"type":       groupType,
			"created_at": Now(),
		})
	}))
	if groupType == "Team" {
		mustNotBeError(s.Attempts().InsertMap(map[string]interface{}{
			"participant_id": groupID,
			"id":             0,
			"creator_id":     nil,
			"created_at":     Now(),
		}))
	}
	s.GroupGroups().createNewAncestors()
	return groupID, nil
}
