package database

import (
	"github.com/jinzhu/gorm"
)

// ThreadStore implements database operations on threads.
type ThreadStore struct {
	*DataStore
}

// UpdateHelperGroupID updates all occurrences of a certain helper_group_id to a new value.
func (s *ThreadStore) UpdateHelperGroupID(oldHelperGroupID, newHelperGroupID int64) {
	var err error

	s.mustBeInTransaction()
	defer recoverPanics(&err)

	err = s.Threads().
		Where("helper_group_id = ?", oldHelperGroupID).
		UpdateColumn("helper_group_id", newHelperGroupID).
		Error()
	mustNotBeError(err)
}

// GetThreadQuery returns a query to get a thread's information.
func (s *ThreadStore) GetThreadQuery(participantID, itemID int64) *DB {
	return s.
		Where("threads.participant_id = ?", participantID).
		Where("threads.item_id = ?", itemID).
		Limit(1)
}

// GetThreadStatus retrieves a thread's status.
func (s *ThreadStore) GetThreadStatus(participantID, itemID int64) string {
	var status string

	err := s.
		GetThreadQuery(participantID, itemID).
		Select("threads.status AS status").
		PluckFirst("status", &status).
		Error()
	if gorm.IsRecordNotFoundError(err) {
		return "not_started"
	}
	mustNotBeError(err)

	return status
}

// GetThreadInfo retrieves a thread's information in an interface.
func (s *ThreadStore) GetThreadInfo(participantID, itemID int64, out interface{}) error {
	return s.
		GetThreadQuery(participantID, itemID).
		Take(out).
		Error()
}

// UserCanWrite checks write permission from a user to a thread.
func (s *ThreadStore) UserCanWrite(user *User, participantID, itemID int64) bool {
	// In order to write in a thread, the thread needs to be opened and the user needs to either:
	// (1) be the participant of the thread
	// (2) have can_watch>=answer permission on the item AND can_watch_members on the participant
	// (3) be part of the group the participant has requested help to AND either have can_watch>=answer on the item
	//     OR have validated the item.

	threadStatus := s.GetThreadStatus(participantID, itemID)
	if IsThreadClosedStatus(threadStatus) {
		return false
	}

	if user.GroupID == participantID {
		return true
	}

	userCanWatchAnswer := user.CanWatchItemAnswer(s.DataStore, itemID)
	userCanWatchMembersOnParticipant := user.CanWatchGroupMembers(s.DataStore, participantID)
	if userCanWatchAnswer && userCanWatchMembersOnParticipant {
		return true
	}

	isMemberOfHelperGroup, err := s.
		GetThreadQuery(participantID, itemID).
		Joins(`JOIN groups_ancestors_active ON groups_ancestors_active.ancestor_group_id = threads.helper_group_id
			AND groups_ancestors_active.child_group_id = ?`, user.GroupID).
		HasRows()
	mustNotBeError(err)

	hasValidatedItem := user.HasValidatedItem(s.DataStore, itemID)

	return isMemberOfHelperGroup && (userCanWatchAnswer || hasValidatedItem)
}

// UserCanChangeStatus checks whether a user can change the status of a thread
//   - The participant of a thread can always switch the thread from open to any another other status.
//     He can only switch it from non-open to an open status if he is allowed to request help on this item
//   - A user who has can_watch>=answer on the item AND can_watch_members on the participant:
//     can always switch a thread to any open status (i.e. he can always open it but not close it)
//   - A user who can write on the thread can switch from an open status to another open status.
func (s *ThreadStore) UserCanChangeStatus(user *User, oldStatus, newStatus string, participantID, itemID int64) bool {
	if oldStatus == "" && newStatus == "" {
		return false
	}
	if oldStatus == newStatus {
		return true
	}

	wasOpen := IsThreadOpenStatus(oldStatus)
	willBeOpen := IsThreadOpenStatus(newStatus)

	if user.GroupID == participantID {
		// * the participant of a thread can always switch the thread from open to any another other status.
		// * he can only switch it from not-open to an open status if he is allowed to request help on this item.
		// -> "allowed request help" have been checked before calling this method, therefore, the user can always
		//     change the status in this situation.
		return true
	} else if willBeOpen {
		// a user who has can_watch>=answer on the item AND can_watch_members on the participant:
		// can always switch a thread to any open status (i.e. he can always open it but not close it)
		currentUserCanWatch := user.CanWatchItemAnswer(s.DataStore, itemID)
		userCanWatchMembersOnParticipant := user.CanWatchGroupMembers(s.DataStore, participantID)

		if currentUserCanWatch && userCanWatchMembersOnParticipant {
			return true
		} else if wasOpen {
			// a user who can write on the thread can switch from an open status to another open status
			return s.UserCanWrite(user, participantID, itemID)
		}
	}

	return false
}
