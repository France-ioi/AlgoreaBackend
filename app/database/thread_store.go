package database

import (
	"time"

	"github.com/jinzhu/gorm"
)

// ThreadStore implements database operations on threads
type ThreadStore struct {
	*DataStore
}

// UpdateHelperGroupID updates all occurrences of a certain helper_group_id to a new value
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

// CanRetrieveThread checks whether a user can retrieve a thread
func (s *ThreadStore) CanRetrieveThread(user *User, participantID, itemID int64) bool {
	// TODO: Try to make the permission checks one query with OR instead of using subqueries.

	// TODO: We need to update GORM for this and use https://gorm.io/docs/advanced_query.html#Group-Conditions
	// Update in progress by Dmitry: https://github.com/France-ioi/AlgoreaBackend/issues/769

	// we check the permissions first without joining the threads because we need to distinguish between an
	// access error and the non-existence of the thread, which should be reported as status=not_started

	// check if the current-user is the thread participant and allowed to "can_view >= content" the item
	currentUserParticipantCanViewContent, err := s.Permissions().MatchingUserAncestors(user).
		Where("? = ?", user.GroupID, participantID).
		Where("permissions.item_id = ?", itemID).
		WherePermissionIsAtLeast("view", "content").
		Select("1").
		Limit(1).
		HasRows()
	mustNotBeError(err)
	if currentUserParticipantCanViewContent {
		return true
	}

	// the current-user has the "can_watch >= answer" permission on the item
	currentUserCanWatch, err := s.Permissions().MatchingUserAncestors(user).
		Where("permissions.item_id = ?", itemID).
		WherePermissionIsAtLeast("watch", "answer").
		Select("1").
		Limit(1).
		HasRows()
	mustNotBeError(err)
	if currentUserCanWatch {
		return true
	}

	// the following rules all matches:
	// the current-user is descendant of the thread helper_group
	// the thread is either open (=waiting_for_participant or =waiting_for_trainer), or closed for less than 2 weeks
	// the current-user has validated the item

	now := time.Now()
	twoWeeksAgo := now.AddDate(0, 0, -14)
	currentUserCanHelp, err := s.Threads().
		Joins("JOIN results ON results.item_id = threads.item_id").
		Joins("JOIN groups_ancestors_active ON groups_ancestors_active.child_group_id = ?", user.GroupID).
		Where("threads.helper_group_id = groups_ancestors_active.ancestor_group_id").
		Where("threads.item_id = ?", itemID).
		Where("threads.status != 'closed' OR threads.latest_update_at > ?", twoWeeksAgo).
		Where("results.participant_id = ?", user.GroupID).
		Where("results.validated").
		Limit(1).
		HasRows()
	mustNotBeError(err)

	return currentUserCanHelp
}

// GetThreadStatus retrieves a thread's status
func (s *ThreadStore) GetThreadStatus(participantID, itemID int64) string {
	var status string

	err := s.Threads().
		Select("threads.status AS status").
		Where("threads.participant_id = ?", participantID).
		Where("threads.item_id = ?", itemID).
		Limit(1).
		PluckFirst("status", &status).
		Error()
	if gorm.IsRecordNotFoundError(err) {
		return "not_started"
	}
	mustNotBeError(err)

	return status
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
//   - The participant of a thread can always switch the thread from open to any another other status.
//     He can only switch it from non-open to an open status if he is allowed to request help on this item
//   - A user who has can_watch>=answer on the item AND can_watch_members on the participant:
//     can always switch a thread to any open status (i.e. he can always open it but not close it)
//   - A user who can write on the thread can switch from an open status to another open status.
func (s *ThreadStore) UserCanChangeStatus(user *User, oldStatus string, newStatus string,
	participantID, itemID, newHelperGroupID int64) bool {
	if oldStatus == "" && newStatus == "" {
		return false
	}
	if oldStatus == newStatus {
		return true
	}

	wasOpen := oldStatus == "waiting_for_trainer" || oldStatus == "waiting_for_participant"
	willBeOpen := newStatus == "waiting_for_trainer" || newStatus == "waiting_for_participant"

	// The participant of a thread can always switch the thread from open to any another other status.
	if user.GroupID == participantID {
		if wasOpen {
			return true
		}

		// He can only switch it from not-open to an open status if he is allowed to request help on this item (see “specific permission” above)
		if willBeOpen {
			canRequestHelpTo := s.CanRequestHelpTo(user, itemID, newHelperGroupID)

			return canRequestHelpTo
		}
	} else {
		// the current-user has the "can_watch >= answer" permission on the item
		currentUserCanWatch, err := s.Permissions().MatchingUserAncestors(user).
			Where("permissions.item_id = ?", itemID).
			WherePermissionIsAtLeast("watch", "answer").
			Select("1").
			Limit(1).
			HasRows()
		mustNotBeError(err)
		if !currentUserCanWatch {
			return false
		}

		userCanWatchMembersOnParticipant, err := s.ActiveGroupAncestors().ManagedByUser(user).
			Where("groups_ancestors_active.child_group_id = ?", participantID).
			Where("group_managers.can_watch_members").
			Select("1").
			Limit(1).
			HasRows()
		mustNotBeError(err)
		if !userCanWatchMembersOnParticipant {
			return false
		}
	}

	return true
}

// CanRequestHelpTo checks whether the user can request help on an item to a group.
func (s *ThreadStore) CanRequestHelpTo(user *User, itemID, newHelperGroupID int64) bool {
	// TODO: Implement this!
	return true
}
