package database

import (
	"errors"
	"fmt"
	"time"

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
	return s.Visible(user).Where("items.ID = ?", itemID)
}

// VisibleChildrenOfID returns a view of the visible children of item identified by itemID, for the given user
func (s *ItemStore) VisibleChildrenOfID(user *User, itemID int64) *DB {
	return s.
		Visible(user).
		Joins("JOIN ? ii ON items.ID=idItemChild", s.ItemItems().SubQuery()).
		Where("ii.idItemParent = ?", itemID)
}

// VisibleGrandChildrenOfID returns a view of the visible grand-children of item identified by itemID, for the given user
func (s *ItemStore) VisibleGrandChildrenOfID(user *User, itemID int64) *DB {
	return s.
		// visible items are the leaves (potential grandChildren)
		Visible(user).
		// get their parents' IDs (ii1)
		Joins("JOIN ? ii1 ON items.ID = ii1.idItemChild", s.ItemItems().SubQuery()).
		// get their grand parents' IDs (ii2)
		Joins("JOIN ? ii2 ON ii2.idItemChild = ii1.idItemParent", s.ItemItems().SubQuery()).
		Where("ii2.idItemParent = ?", itemID)
}

// AccessRights returns a composable query for getting
// (idItem, fullAccess, partialAccess, grayedAccess, accessSolutions) for the given user
func (s *ItemStore) AccessRights(user *User) *DB {
	return s.GroupItems().MatchingUserAncestors(user).
		Select(
			"idItem, MIN(sCachedFullAccessDate) <= NOW() AS fullAccess, " +
				"MIN(sCachedPartialAccessDate) <= NOW() AS partialAccess, " +
				"MIN(sCachedGrayedAccessDate) <= NOW() AS grayedAccess, " +
				"MIN(sCachedAccessSolutionsDate) <= NOW() AS accessSolutions").
		Group("idItem")
}

// HasManagerAccess returns whether the user has manager access to all the given item_id's
// It is assumed that the `OwnerAccess` implies manager access
func (s *ItemStore) HasManagerAccess(user *User, itemIDs ...int64) (found bool, err error) {
	var count int64

	idsMap := make(map[int64]bool, len(itemIDs))
	for _, itemID := range itemIDs {
		idsMap[itemID] = true
	}
	err = s.GroupItems().MatchingUserAncestors(user).
		WithWriteLock().
		Where("idItem IN (?) AND (bCachedManagerAccess OR bOwnerAccess)", itemIDs).
		Select("COUNT(DISTINCT idItem)").Count(&count).Error()
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

	if err := checkAccess(itemIDs, accessDetails); err != nil {
		log.Infof("checkAccess %v %v", itemIDs, accessDetails)
		log.Infof("User access validation failed: %v", err)
		return false, nil
	}
	return true, nil
}

// GetAccessDetailsForIDs returns access details for given item IDs and the given user
func (s *ItemStore) GetAccessDetailsForIDs(user *User, itemIDs []int64) ([]ItemAccessDetailsWithID, error) {
	var accessDetails []ItemAccessDetailsWithID
	db := s.AccessRights(user).
		Where("groups_items.idItem IN (?)", itemIDs).
		Scan(&accessDetails)
	if err := db.Error(); err != nil {
		return nil, err
	}
	return accessDetails, nil
}

// checkAccess checks if the user has access to all items:
// - user has to have full access to all items
// OR
// - user has to have full access to all but last, and grayed access to that last item.
func checkAccess(itemIDs []int64, accDets []ItemAccessDetailsWithID) error {
	for i, id := range itemIDs {
		last := i == len(itemIDs)-1
		if err := checkAccessForID(id, last, accDets); err != nil {
			return err
		}
	}
	return nil
}

func checkAccessForID(id int64, last bool, accDets []ItemAccessDetailsWithID) error {
	for _, res := range accDets {
		if res.ItemID != id {
			continue
		}
		if res.FullAccess || res.PartialAccess {
			// OK, user has full access.
			return nil
		}
		if res.GrayedAccess && last {
			// OK, user has grayed access on the last item.
			return nil
		}
		return fmt.Errorf("not enough perm on item_id %d", id)
	}

	// no row matching this item_id
	return fmt.Errorf("not visible item_id %d", id)
}

func (s *ItemStore) isRootItem(id int64) (bool, error) {
	count := 0
	if err := s.ByID(id).Where("sType='Root'").Count(&count).Error(); err != nil {
		return false, err
	}
	if count == 0 {
		return false, nil
	}
	return true, nil
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

		db = db.Or("idItemParent=? AND idItemChild=?", previousID, id)
		previousID = id
	}

	count := 0
	// For now, we don’t have a unique key for the pair ('idItemParent' and 'idItemChild') and
	// theoretically it’s still possible to have multiple rows with the same pair
	// of 'idItemParent' and 'idItemChild'.
	// The “Group(...)” here resolves the issue.
	if err := db.Group("idItemParent, idItemChild").Count(&count).Error(); err != nil {
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
	err = s.Visible(user).Where("fullAccess > 0 OR partialAccess > 0").Where("ID = ?", itemID).
		PluckFirst("bReadOnly", &readOnly).Error()
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
	// to which the user has at least grayed access.
	// Note that while an answer is always related to a task,
	// tasks cannot be time-limited, only chapters can.
	// So, actually here we select time-limited chapters that are ancestors of the task.
	var contestItems []struct {
		ItemID     int64 `gorm:"column:idItem"`
		FullAccess bool  `gorm:"column:fullAccess"`
	}

	mustNotBeError(s.Visible(user).
		Select("items.ID AS idItem, fullAccess").
		Joins("JOIN items_ancestors ON items_ancestors.idItemAncestor = items.ID").
		Where("items_ancestors.idItemChild = ?", itemID).
		Where("items.sDuration IS NOT NULL").
		Group("items.ID").Scan(&contestItems).Error())

	// The item is not time-limited itself and it doesn't have time-limited ancestors the user has access to.
	// Or maybe the user doesn't have access to the item at all... We ignore this possibility here
	if len(contestItems) == 0 {
		return true, nil // The user can submit an answer
	}

	activeContestErr := errors.New("the contest has not started yet or has already finished")

	activeContest := s.getActiveContestInfoForUser(user)
	if activeContest == nil {
		return false, activeContestErr
	}

	if activeContest.IsOver() {
		if activeContest.TeamMode != nil {
			s.closeTeamContest(activeContest.ItemID, user)
		} else {
			s.closeContest(activeContest.ItemID, user)
		}
		return false, activeContestErr
	}

	for i := range contestItems {
		if contestItems[i].FullAccess || activeContest.ItemID == contestItems[i].ItemID {
			return true, nil
		}
	}

	return false, errors.New(
		"the exercise for which you wish to submit an answer is a part " +
			"of a different competition than the one in progress")
}

type activeContestInfo struct {
	ItemID   int64
	UserID   int64
	TeamMode *string

	Now               time.Time
	DurationInSeconds int32
	EndTime           time.Time
	StartTime         time.Time
}

func (contest *activeContestInfo) IsOver() bool {
	return contest.EndTime.Before(contest.Now) || contest.EndTime.Equal(contest.Now)
}

// Closes the time-limited contest if needed or returns time stats
func (s *ItemStore) getActiveContestInfoForUser(user *User) *activeContestInfo {
	// Get info for the item if the user has already started it, but hasn't finished yet
	// Note: the current API doesn't allow users to have more than one active contest
	// Note: both users_items & groups_items rows should exist to make this function return the info
	var results []struct {
		Now                     time.Time `gorm:"column:now"`
		DurationInSeconds       int32     `gorm:"column:duration"`
		ItemID                  int64     `gorm:"column:idItem"`
		AdditionalTimeInSeconds int32     `gorm:"column:additionalTime"`
		ContestStartDate        time.Time `gorm:"column:sContestStartDate"`
		TeamMode                *string   `gorm:"column:sTeamMode"`
	}
	mustNotBeError(s.
		Select(`
			NOW() AS now,
			TIME_TO_SEC(items.sDuration) AS duration,
			items.ID AS idItem,
			items.sTeamMode,
			IFNULL(SUM(TIME_TO_SEC(groups_items.sAdditionalTime)), 0) AS additionalTime,
			users_items.sContestStartDate AS sContestStartDate`).
		Joins("JOIN groups_items ON groups_items.idItem = items.ID").
		Joins(`
			JOIN groups_ancestors ON groups_ancestors.idGroupAncestor = groups_items.idGroup AND
				groups_ancestors.idGroupChild = ?`, user.SelfGroupID).
		Joins(`
			JOIN users_items ON users_items.idItem = items.ID AND users_items.idUser = ? AND
				users_items.sContestStartDate IS NOT NULL AND users_items.sFinishDate IS NULL`, user.ID).
		Group("items.ID").
		Order("users_items.sContestStartDate DESC").Scan(&results).Error())

	if len(results) == 0 {
		return nil
	}

	if len(results) > 1 {
		log.Warnf("User with ID = %d has %d (>1) active contests", user.ID, len(results))
	}

	totalDuration := results[0].DurationInSeconds + results[0].AdditionalTimeInSeconds
	endTime := results[0].ContestStartDate.Add(time.Duration(totalDuration) * time.Second)

	return &activeContestInfo{
		Now:               results[0].Now,
		DurationInSeconds: totalDuration,
		StartTime:         results[0].ContestStartDate,
		EndTime:           endTime,
		ItemID:            results[0].ItemID,
		UserID:            user.ID,
		TeamMode:          results[0].TeamMode,
	}
}

func (s *ItemStore) closeContest(itemID int64, user *User) {
	mustNotBeError(s.UserItems().
		Where("idItem = ? AND idUser = ?", itemID, user.ID).
		UpdateColumn("sFinishDate", Now()).Error())

	groupItemStore := s.GroupItems()

	// TODO: "remove partial access if other access were present" (what did he mean???)
	if user.SelfGroupID != nil {
		groupItemStore.removePartialAccess(*user.SelfGroupID, itemID)
		mustNotBeError(groupItemStore.db.Exec(`
		DELETE groups_items
		FROM groups_items
		JOIN items_ancestors ON
			items_ancestors.idItemChild = groups_items.idItem AND
			items_ancestors.idItemAncestor = ?
		WHERE groups_items.idGroup = ? AND
			(sCachedFullAccessDate IS NULL OR sCachedFullAccessDate > NOW()) AND
			bOwnerAccess = 0 AND bManagerAccess = 0`, itemID, *user.SelfGroupID).Error)
		// we do not need to call GroupItemStore.After() because we do not grant new access here
		groupItemStore.computeAllAccess()
	}
}

func (s *ItemStore) closeTeamContest(itemID int64, user *User) {
	var teamGroupID int64
	mustNotBeError(s.Groups().TeamGroupByTeamItemAndUser(itemID, user).PluckFirst("groups.ID", &teamGroupID).Error())

	// Set contest as finished
	/*
		// We would use this block if UPDATEs with JOINs were fixed in jinzhu/gorm
		mustNotBeError(s.UserItems().
			Joins("JOIN users ON users.ID = users_items.idUser").
			Joins(`JOIN groups_groups
				ON groups_groups.idGroupChild = users.idGroupSelf AND groups_groups.idGroupParent = ?`, teamGroupID).
			Where("users_items.idItem = ?", itemID).
			UpdateColumn("sFinishDate", Now()).Error())
	*/ // nolint:gocritic
	mustNotBeError(s.db.Exec(`
		UPDATE users_items
		JOIN users ON users.ID = users_items.idUser
		JOIN groups_groups
			ON groups_groups.idGroupChild = users.idGroupSelf AND
				groups_groups.sType IN ('invitationAccepted', 'requestAccepted') AND
				groups_groups.idGroupParent = ?
		SET sFinishDate = NOW()
		WHERE users_items.idItem = ?`, teamGroupID, itemID).Error)

	groupItemStore := s.GroupItems()
	// Remove access
	groupItemStore.removePartialAccess(teamGroupID, itemID)

	// we do not need to call GroupItemStore.After() because we do not grant new access here
	groupItemStore.computeAllAccess()
}
