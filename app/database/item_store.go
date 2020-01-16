package database

import (
	"errors"
	"fmt"

	"github.com/jinzhu/gorm"

	log "github.com/France-ioi/AlgoreaBackend/app/logging"
)

// ItemStore implements database operations on items
type ItemStore struct {
	*DataStore
}

// Visible returns a view of the visible items for the given user
func (s *ItemStore) Visible(user *User) *DB {
	return s.WhereItemsAreVisible(user)
}

// VisibleByID returns a view of the visible item identified by itemID, for the given user
func (s *ItemStore) VisibleByID(user *User, itemID int64) *DB {
	return s.Visible(user).Where("items.id = ?", itemID)
}

// VisibleChildrenOfID returns a view of the visible children of item identified by itemID, for the given user
func (s *ItemStore) VisibleChildrenOfID(user *User, itemID int64) *DB {
	return s.
		Visible(user).
		Joins("JOIN ? ii ON items.id=child_item_id", s.ItemItems().SubQuery()).
		Where("ii.parent_item_id = ?", itemID)
}

// VisibleGrandChildrenOfID returns a view of the visible grand-children of item identified by itemID, for the given user
func (s *ItemStore) VisibleGrandChildrenOfID(user *User, itemID int64) *DB {
	return s.
		// visible items are the leaves (potential grandChildren)
		Visible(user).
		// get their parents' IDs (ii1)
		Joins("JOIN ? ii1 ON items.id = ii1.child_item_id", s.ItemItems().SubQuery()).
		// get their grand parents' IDs (ii2)
		Joins("JOIN ? ii2 ON ii2.child_item_id = ii1.parent_item_id", s.ItemItems().SubQuery()).
		Where("ii2.parent_item_id = ?", itemID)
}

// CanGrantViewContentOnAll returns whether the user can grant 'content' view right on all the listed items (can_grant_view >= content)
func (s *ItemStore) CanGrantViewContentOnAll(user *User, itemIDs ...int64) (hasAccess bool, err error) {
	var count int64
	if len(itemIDs) == 0 {
		return true, nil
	}

	idsMap := make(map[int64]bool, len(itemIDs))
	for _, itemID := range itemIDs {
		idsMap[itemID] = true
	}
	err = s.Permissions().MatchingUserAncestors(user).
		WithWriteLock().
		Where("item_id IN (?)", itemIDs).
		WherePermissionIsAtLeast("grant_view", "content").
		Select("COUNT(DISTINCT item_id)").Count(&count).Error()
	if err != nil {
		return false, err
	}
	return count == int64(len(idsMap)), nil
}

// AreAllVisible returns whether all the items are visible to the user
func (s *ItemStore) AreAllVisible(user *User, itemIDs ...int64) (hasAccess bool, err error) {
	var count int64
	if len(itemIDs) == 0 {
		return true, nil
	}

	idsMap := make(map[int64]bool, len(itemIDs))
	for _, itemID := range itemIDs {
		idsMap[itemID] = true
	}
	err = s.Permissions().MatchingUserAncestors(user).
		WithWriteLock().
		Where("item_id IN (?)", itemIDs).
		Where("can_view_generated != 'none'").
		Select("COUNT(DISTINCT item_id)").Count(&count).Error()
	if err != nil {
		return false, err
	}
	return count == int64(len(idsMap)), nil
}

// IsValidHierarchy gets an ordered set of item ids and returns whether they forms a valid item hierarchy path from a root
func (s *ItemStore) IsValidHierarchy(ids []int64) (bool, error) {
	if len(ids) == 0 {
		return false, nil
	}

	if valid, err := s.isRootItem(ids[0]); !valid || err != nil {
		return valid, err
	}

	if valid, err := s.isHierarchicalChain(ids); !valid || err != nil {
		return valid, err
	}

	return true, nil
}

// ValidateUserAccess gets a set of item ids and returns whether the given user is authorized to see them all
func (s *ItemStore) ValidateUserAccess(user *User, itemIDs []int64) (bool, error) {
	accessDetails, err := s.GetAccessDetailsForIDs(user, itemIDs)
	if err != nil {
		log.Infof("User access rights loading failed: %v", err)
		return false, err
	}

	if err := s.checkAccess(itemIDs, accessDetails); err != nil {
		log.Infof("checkAccess %v %v", itemIDs, accessDetails)
		log.Infof("User access validation failed: %v", err)
		return false, nil
	}
	return true, nil
}

// GetAccessDetailsForIDs returns access details for given item IDs and the given user
func (s *ItemStore) GetAccessDetailsForIDs(user *User, itemIDs []int64) ([]ItemAccessDetailsWithID, error) {
	var valuesWithIDs []struct {
		ItemID                int64
		CanViewGeneratedValue int
	}
	db := s.Permissions().WithViewPermissionForUser(user, "info").
		Where("item_id IN (?)", itemIDs).
		Scan(&valuesWithIDs)
	if err := db.Error(); err != nil {
		return nil, err
	}
	accessDetails := make([]ItemAccessDetailsWithID, len(valuesWithIDs))
	for i := range valuesWithIDs {
		accessDetails[i].ItemID = valuesWithIDs[i].ItemID
		accessDetails[i].CanView = s.PermissionsGranted().ViewNameByIndex(valuesWithIDs[i].CanViewGeneratedValue)
	}
	return accessDetails, nil
}

// checkAccess checks if the user has access to all items:
// - user has to have full access to all items
// OR
// - user has to have full access to all but last, and info access to that last item.
func (s *ItemStore) checkAccess(itemIDs []int64, accDets []ItemAccessDetailsWithID) error {
	for i, id := range itemIDs {
		last := i == len(itemIDs)-1
		if err := s.checkAccessForID(id, last, accDets); err != nil {
			return err
		}
	}
	return nil
}

func (s *ItemStore) checkAccessForID(id int64, last bool, accDets []ItemAccessDetailsWithID) error {
	for _, res := range accDets {
		if res.ItemID != id {
			continue
		}
		if res.CanView != "" && s.PermissionsGranted().ViewIndexByName(res.CanView) >= s.PermissionsGranted().ViewIndexByName("content") {
			// OK, user has full access.
			return nil
		}
		if res.CanView == canViewInfo && last {
			// OK, user has info access on the last item.
			return nil
		}
		return fmt.Errorf("not enough perm on item_id %d", id)
	}

	// no row matching this item_id
	return fmt.Errorf("not visible item_id %d", id)
}

func (s *ItemStore) isRootItem(id int64) (bool, error) {
	return s.ByID(id).Where("is_root").HasRows()
}

func (s *ItemStore) isHierarchicalChain(ids []int64) (bool, error) {
	if len(ids) == 0 {
		return false, nil
	}

	if len(ids) == 1 {
		return true, nil
	}

	db := s.ItemItems().DB
	previousID := ids[0]
	for index, id := range ids {
		if index == 0 {
			continue
		}

		db = db.Or("parent_item_id=? AND child_item_id=?", previousID, id)
		previousID = id
	}

	count := 0
	// There is a unique key for the pair ('parent_item_id' and 'child_item_id') so count() will work correctly
	if err := db.Count(&count).Error(); err != nil {
		return false, err
	}

	if count != len(ids)-1 {
		return false, nil
	}

	return true, nil
}

// CheckSubmissionRights checks if the user can submit an answer for the given item (task):
// 1. If the task is inside a time-limited chapter, the method checks that the task is a part of
//    the user's active contest (or the user has full access to one of the task's chapters)
// 2. The method also checks that the item (task) exists and is not read-only.
//
// Note: This method doesn't check if the user has access to the item.
// Note 2: This method may also close the user's active contest (or the user's active team contest).
func (s *ItemStore) CheckSubmissionRights(itemID int64, user *User) (hasAccess bool, reason, err error) {
	s.mustBeInTransaction() // because it may close a contest
	recoverPanics(&err)

	var readOnly bool
	err = s.Visible(user).WherePermissionIsAtLeast("view", "content").
		Where("id = ?", itemID).
		PluckFirst("read_only", &readOnly).Error()
	if gorm.IsRecordNotFoundError(err) {
		return false, errors.New("no access to the task item"), nil
	}
	mustNotBeError(err)

	if readOnly {
		return false, errors.New("item is read-only"), nil
	}

	hasRights, reason := s.checkSubmissionRightsForTimeLimitedContest(itemID, user)
	if !hasRights {
		return hasRights, reason, nil
	}

	return true, nil, nil
}

func (s *ItemStore) checkSubmissionRightsForTimeLimitedContest(itemID int64, user *User) (bool, error) {
	// TODO: handle case where the item is both in a contest and in a non-contest chapter the user has access to

	// ItemID & FullAccess for time-limited ancestors of the item
	// to which the user has at least 'info' access.
	// Note that while an answer is always related to a task,
	// tasks cannot be time-limited, only chapters can.
	// So, actually here we select time-limited chapters that are ancestors of the task.
	var contestItems []struct {
		ItemID     int64
		FullAccess bool
	}

	mustNotBeError(s.Visible(user).
		Select("items.id AS item_id, can_view_generated_value >= ? AS full_access",
			s.PermissionsGranted().ViewIndexByName("content_with_descendants")).
		Joins("JOIN items_ancestors ON items_ancestors.ancestor_item_id = items.id").
		Where("items_ancestors.child_item_id = ?", itemID).
		Where("items.duration IS NOT NULL").
		Group("items.id").Scan(&contestItems).Error())

	// The item is not time-limited itself and it doesn't have time-limited ancestors the user has access to.
	// Or maybe the user doesn't have access to the item at all... We ignore this possibility here
	if len(contestItems) == 0 {
		return true, nil // The user can submit an answer
	}

	activeContestItemID := s.getActiveContestItemIDForUser(user)
	if activeContestItemID == nil {
		return false, errors.New("the contest has not started yet or has already finished")
	}

	for i := range contestItems {
		if contestItems[i].FullAccess || *activeContestItemID == contestItems[i].ItemID {
			return true, nil
		}
	}

	return false, errors.New(
		"the exercise for which you wish to submit an answer is a part " +
			"of a different competition than the one in progress")
}

func (s *ItemStore) getActiveContestItemIDForUser(user *User) *int64 {
	// Get id of the item if the user has already started it, but hasn't finished yet
	// Note: the current API doesn't allow users to have more than one active contest
	// Note: attempts rows with 'entered_at' should exist to make this function return the info
	var itemID int64

	err := s.
		Select("items.id AS item_id").
		Joins(`
			JOIN groups_ancestors_active
				ON groups_ancestors_active.child_group_id = ?`, user.GroupID).
		Joins(`
			JOIN groups_groups_active
				ON groups_groups_active.parent_group_id = items.contest_participants_group_id AND
					groups_groups_active.child_group_id = groups_ancestors_active.child_group_id`).
		Joins(`JOIN attempts AS contest_participations ON contest_participations.item_id = items.id AND
			contest_participations.group_id = groups_ancestors_active.ancestor_group_id AND
			contest_participations.entered_at IS NOT NULL`).
		Group("items.id").
		Order("MIN(contest_participations.entered_at) DESC").
		PluckFirst("items.id", &itemID).Error()

	if gorm.IsRecordNotFoundError(err) {
		return nil
	}
	mustNotBeError(err)

	return &itemID
}

// ContestManagedByUser returns a composable query
// for getting a contest with the given item id managed by the given user
func (s *ItemStore) ContestManagedByUser(contestItemID int64, user *User) *DB {
	return s.ByID(contestItemID).Where("items.duration IS NOT NULL").
		Joins("JOIN permissions_generated ON permissions_generated.item_id = items.id").
		Joins(`
			JOIN groups_ancestors_active ON groups_ancestors_active.ancestor_group_id = permissions_generated.group_id AND
				groups_ancestors_active.child_group_id = ?`, user.GroupID).
		Group("items.id").
		HavingMaxPermissionAtLeast("view", "content_with_descendants")
}
