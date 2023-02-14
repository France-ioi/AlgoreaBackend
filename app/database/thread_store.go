package database

import "errors"

// ThreadStore implements database operations on threads
type ThreadStore struct {
	*DataStore
}

// UpdateHelperGroupID updates all occurrences of a certain helper_group_id to a new value
func (s *ThreadStore) UpdateHelperGroupID(oldHelperGroupID, newHelperGroupID int64) (err error) {
	s.mustBeInTransaction()
	defer recoverPanics(&err)

	err = s.Threads().
		Where("helper_group_id = ?", oldHelperGroupID).
		UpdateColumn("helper_group_id", newHelperGroupID).
		Error()
	if err != nil {
		return err
	}

	return nil
}

// UserCanWrite checks write permission from a user to a thread
func (s *ThreadStore) UserCanWrite(user *User, participantID int64, itemID int64) (bool, error) {
	// In order to write in a thread, the thread needs to be opened and the user needs to either:
	// (1) be the participant of the thread
	// (2) have can_watch>=answer permission on the item AND can_watch_members on the participant
	// (3) be part of the group the participant has requested help to AND either have can_watch>=answer on the item
	//     OR have validated the item.

	userIsParticipant := user.GroupID == participantID

	userCanWatchAnswer, err := s.Permissions().MatchingUserAncestors(user).
		Where("permissions.item_id = ?", itemID).
		WherePermissionIsAtLeast("watch", "answer").
		Select("1").
		Limit(1).
		HasRows()
	if err != nil {
		return false, err
	}

	userCanWatchMembersOnParticipant, err := s.ActiveGroupAncestors().ManagedByUser(user).
		Where("groups_ancestors_active.child_group_id = ?", participantID).
		Where("group_managers.can_watch_members").
		Select("1").
		Limit(1).
		HasRows()
	if err != nil {
		return false, err
	}

	isMemberOfHelperGroup, err := s.Threads().
		Joins(`JOIN groups_ancestors_active ON groups_ancestors_active.ancestor_group_id = threads.helper_group_id
			AND groups_ancestors_active.child_group_id = ?`, user.GroupID).
		Where("threads.item_id = ?", itemID).
		Where("threads.participant_id = ?", participantID).
		Limit(1).
		HasRows()
	if err != nil {
		return false, err
	}

	hasValidatedItem, err := s.Threads().
		Joins("JOIN results ON results.item_id = threads.item_id").
		Where("threads.item_id = ?", itemID).
		Where("threads.participant_id = ?", participantID).
		Where("results.validated").
		Limit(1).
		HasRows()
	if err != nil {
		return false, err
	}

	return userIsParticipant ||
		(userCanWatchAnswer && userCanWatchMembersOnParticipant) ||
		(isMemberOfHelperGroup && (userCanWatchAnswer || hasValidatedItem)), nil
}

// UserCanChangeStatus checks whether a user can change the status of a thread
// - The participant of a thread can always switch the thread from open to any another other status.
//    He can only switch it from non-open to an open status if he is allowed to request help on this item
// - A user who has can_watch>=answer on the item AND can_watch_members on the participant:
//   can always switch a thread to any open status (i.e. he can always open it but not close it)
// - A user who can write on the thread can switch from an open status to another open status.
func (s *ThreadStore) UserCanChangeStatus(user *User, oldStatus string, newStatus string, participantID int64,
	itemID int64) (bool, error) {
	if oldStatus == "" {
		// TODO: Permissions to create a thread
		return false, errors.New("needs implementation")
	}
	if oldStatus == newStatus {
		return true, nil
	}

	wasOpen := oldStatus == "waiting_for_trainer" || oldStatus == "waiting_for_participant"
	willBeOpen := newStatus == "waiting_for_trainer" || newStatus == "waiting_for_participant"

	// The participant of a thread can always switch the thread from open to any another other status.
	if user.GroupID == participantID {
		if wasOpen {
			return true, nil
		}

		// He can only switch it from not-open to an open status if he is allowed to request help on this item (see “specific permission” above)
		if willBeOpen {
			// TODO: Check if allowed to request_help on this item, when forum permissions are merged
		}
	} else {

	}

	return true, nil
}
