package database

import "time"

// GroupPendingRequestStore implements database operations on `group_pending_requests`
// (which stores requests that require an action from a user).
type GroupPendingRequestStore struct {
	*DataStore
}

// InviteParticipants add pending requests of invitation to participants to a group.
func (s GroupPendingRequestStore) InviteParticipants(groupID int64, participantIDs []int64) {
	invitationMaps := make([]map[string]interface{}, 0, len(participantIDs))
	for _, participantID := range participantIDs {
		invitationMaps = append(invitationMaps, map[string]interface{}{
			"group_id":  groupID,
			"member_id": participantID,
			"type":      "invitation",
			"at":        time.Now(),
		})
	}

	err := s.InsertOrUpdateMaps(invitationMaps, []string{"type", "at"})
	mustNotBeError(err)
}
