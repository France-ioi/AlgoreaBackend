package database

import "time"

// GroupMembershipChangeStore implements database operations on `group_membership_changes`
// (which stores the history of group membership changes).
type GroupMembershipChangeStore struct {
	*DataStore
}

// InsertEntries inserts multiple entries into the group_membership_changes table.
func (s GroupMembershipChangeStore) InsertEntries(initiatorID, groupID int64, memberIDs []int64, action string) {
	var groupMembershipChangesEntries []map[string]interface{}
	for _, memberID := range memberIDs {
		groupMembershipChangesEntries = append(groupMembershipChangesEntries, map[string]interface{}{
			"group_id":     groupID,
			"member_id":    memberID,
			"at":           time.Now(),
			"action":       action,
			"initiator_id": initiatorID,
		})
	}

	err := s.InsertMaps(groupMembershipChangesEntries)
	mustNotBeError(err)
}
