package database

import (
	"time"
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
		Where("threads.status != 'closed' OR (threads.status = 'closed' AND threads.latest_update_at > ?)", twoWeeksAgo).
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
	if err != nil {
		status = "not_started"
	}

	return status
}
