package database

import (
	"errors"
	"fmt"
	"strings"

	"github.com/jinzhu/gorm"
)

// ItemStore implements database operations on items
type ItemStore struct {
	*DataStore
}

// Visible returns a view of the visible items for the given participant
func (s *ItemStore) Visible(groupID int64) *DB {
	return s.WhereItemsAreVisible(groupID)
}

// VisibleByID returns a view of the visible item identified by itemID, for the given participant
func (s *ItemStore) VisibleByID(groupID, itemID int64) *DB {
	return s.Visible(groupID).Where("items.id = ?", itemID)
}

// VisibleChildrenOfID returns a view of the visible children of item identified by itemID, for the given participant
func (s *ItemStore) VisibleChildrenOfID(groupID, itemID int64) *DB {
	return s.
		Visible(groupID).
		Joins("JOIN ? ii ON items.id=child_item_id", s.ItemItems().SubQuery()).
		Where("ii.parent_item_id = ?", itemID)
}

// VisibleGrandChildrenOfID returns a view of the visible grand-children of item identified by itemID, for the given participant
func (s *ItemStore) VisibleGrandChildrenOfID(groupID, itemID int64) *DB {
	return s.
		// visible items are the leaves (potential grandChildren)
		Visible(groupID).
		// get their parents' IDs (ii1)
		Joins("JOIN ? ii1 ON items.id = ii1.child_item_id", s.ItemItems().SubQuery()).
		// get their grand parents' IDs (ii2)
		Joins("JOIN ? ii2 ON ii2.child_item_id = ii1.parent_item_id", s.ItemItems().SubQuery()).
		Where("ii2.parent_item_id = ?", itemID)
}

// IsValidParticipationHierarchyForParentAttempt checks if the given list of item ids is a valid participation hierarchy
// for the given `parentAttemptID` which means all the following statements are true:
//  * the first item in `ids` is a root items (items.is_root) or a root activity (groups.root_activity_id) of a group
//    the `groupID` is a descendant of,
//  * `ids` is an ordered list of parent-child items,
//  * the `groupID` group has at least 'content' access on each of the items in `ids`,
//  * the `groupID` group has a started, allowing submission, not ended result for each item but the last,
//    with `parentAttemptID` (or its parent attempt each time we reach a root of an attempt) as the attempt,
//  * if `ids` consists of only one item, the `parentAttemptID` is zero.
func (s *ItemStore) IsValidParticipationHierarchyForParentAttempt(
	ids []int64, groupID, parentAttemptID int64, requireContentAccessToTheLastItem, withWriteLock bool) (bool, error) {
	if len(ids) == 0 || len(ids) == 1 && parentAttemptID != 0 {
		return false, nil
	}

	return s.participationHierarchyForParentAttempt(
		ids, groupID, parentAttemptID, true, requireContentAccessToTheLastItem, "1", withWriteLock).HasRows()
}

func (s *ItemStore) participationHierarchyForParentAttempt(
	ids []int64, groupID, parentAttemptID int64, requireAttemptsToBeActive, requireContentAccessToTheLastItem bool,
	columnsList string, withWriteLock bool) *DB {
	subQuery := s.itemAttemptChainWithoutAttemptForTail(
		ids, groupID, requireAttemptsToBeActive, requireContentAccessToTheLastItem, withWriteLock)

	if len(ids) > 1 {
		subQuery = subQuery.
			Where(fmt.Sprintf("attempts%d.id = ?", len(ids)-2), parentAttemptID)
	}

	subQuery = subQuery.Select(columnsList)
	visibleItems := s.Visible(groupID).
		Select("items.id, items.is_root, items.allows_multiple_attempts, visible.can_view_generated_value")

	return s.Raw("WITH visible_items AS ? ?", visibleItems.SubQuery(), subQuery.QueryExpr())
}

func (s *ItemStore) itemAttemptChainWithoutAttemptForTail(ids []int64, groupID int64,
	requireAttemptsToBeActive, requireContentAccessToTheLastItem, withWriteLock bool) *DB {
	participantActivities := s.ActiveGroupAncestors().Where("child_group_id = ?", groupID).
		Joins("JOIN `groups` ON groups.id = groups_ancestors_active.ancestor_group_id").
		Select("groups.root_activity_id")

	if withWriteLock {
		participantActivities = participantActivities.WithWriteLock()
	}

	subQuery := s.Table("visible_items as items0").Where("items0.id = ?", ids[0]).
		Where("items0.is_root OR items0.id IN ?", participantActivities.SubQuery())

	for i := 1; i < len(ids); i++ {
		subQuery = subQuery.Joins(fmt.Sprintf(`
				JOIN results AS results%d ON results%d.participant_id = ? AND
					results%d.item_id = items%d.id AND results%d.started_at IS NOT NULL`, i-1, i-1, i-1, i-1, i-1), groupID).
			Joins(fmt.Sprintf(`
				JOIN attempts AS attempts%d ON attempts%d.participant_id = results%d.participant_id AND
					attempts%d.id = results%d.attempt_id`, i-1, i-1, i-1, i-1, i-1)).
			Joins(
				fmt.Sprintf(
					"JOIN items_items AS items_items%d ON items_items%d.parent_item_id = items%d.id AND items_items%d.child_item_id = ?",
					i, i, i-1, i), ids[i]).
			Joins(fmt.Sprintf("JOIN visible_items AS items%d ON items%d.id = items_items%d.child_item_id", i, i, i)).
			Where(fmt.Sprintf("items%d.can_view_generated_value >= ?", i-1),
				s.PermissionsGranted().ViewIndexByName("content"))

		if i != len(ids)-1 {
			subQuery = subQuery.Where(fmt.Sprintf(
				"IF(attempts%d.root_item_id = items%d.id, attempts%d.parent_attempt_id, attempts%d.id) = attempts%d.id",
				i, i, i, i, i-1))
		}

		if requireAttemptsToBeActive {
			subQuery = subQuery.Where(fmt.Sprintf("attempts%d.ended_at IS NULL AND NOW() < attempts%d.allows_submissions_until", i-1, i-1))
		}
	}

	if requireContentAccessToTheLastItem {
		subQuery = subQuery.Where(fmt.Sprintf("items%d.can_view_generated_value >= ?", len(ids)-1),
			s.PermissionsGranted().ViewIndexByName("content"))
	}

	return subQuery
}

// BreadcrumbsHierarchyForParentAttempt returns attempts ids and 'order' (for items allowing multiple attempts)
// for the given list of item ids (but the last item) if it is a valid participation hierarchy
// for the given `parentAttemptID` which means all the following statements are true:
//  * the first item in `ids` is a root items (items.is_root) or a root activity (groups.root_activity_id) of a group
//    the `groupID` is a descendant of,
//  * `ids` is an ordered list of parent-child items,
//  * the `groupID` group has at least 'content' access on each of the items in `ids` except for the last one and
//    at least 'info' access on the last one,
//  * the `groupID` group has a started result for each item but the last,
//    with `parentAttemptID` (or its parent attempt each time we reach a root of an attempt) as the attempt,
//  * if `ids` consists of only one item, the `parentAttemptID` is zero.
func (s *ItemStore) BreadcrumbsHierarchyForParentAttempt(ids []int64, groupID, parentAttemptID int64, withWriteLock bool) (
	attemptIDMap map[int64]int64, attemptNumberMap map[int64]int, err error) {
	if len(ids) == 0 || len(ids) == 1 && parentAttemptID != 0 {
		return nil, nil, nil
	}

	defer recoverPanics(&err)

	columnsList := columnsListForBreadcrumbsHierarchy(ids[:len(ids)-1])
	query := s.participationHierarchyForParentAttempt(
		ids, groupID, parentAttemptID, false, false, columnsList, withWriteLock)
	var data []map[string]interface{}
	mustNotBeError(query.Limit(1).ScanIntoSliceOfMaps(&data).Error())
	if len(data) == 0 {
		return nil, nil, nil
	}

	attemptIDMap, attemptNumberMap = resultsForBreadcrumbsHierarchy(ids[:len(ids)-1], data[0])
	return
}

// BreadcrumbsHierarchyForAttempt returns attempts ids and 'order' (for items allowing multiple attempts)
// for the given list of item ids if it is a valid participation hierarchy
// for the given `attemptID` which means all the following statements are true:
//  * the first item in `ids` is a root items (items.is_root) or the item of a group (groups.activity_id)
//    the `groupID` is a descendant of,
//  * `ids` is an ordered list of parent-child items,
//  * the `groupID` group has at least 'content' access on each of the items in `ids` except for the last one and
//    at least 'info' access on the last one,
//  * the `groupID` group has a started result for each item,
//    with `attemptID` (or its parent attempt each time we reach a root of an attempt) as the attempt.
func (s *ItemStore) BreadcrumbsHierarchyForAttempt(ids []int64, groupID, attemptID int64, withWriteLock bool) (
	attemptIDMap map[int64]int64, attemptNumberMap map[int64]int, err error) {
	if len(ids) == 0 {
		return nil, nil, nil
	}

	defer recoverPanics(&err)

	columnsList := columnsListForBreadcrumbsHierarchy(ids)
	query := s.breadcrumbsHierarchyForAttempt(
		ids, groupID, attemptID, false, columnsList, withWriteLock)
	var data []map[string]interface{}
	mustNotBeError(query.Limit(1).ScanIntoSliceOfMaps(&data).Error())
	if len(data) == 0 {
		return nil, nil, nil
	}

	attemptIDMap, attemptNumberMap = resultsForBreadcrumbsHierarchy(ids, data[0])
	return
}

func columnsListForBreadcrumbsHierarchy(ids []int64) string {
	columnsList := "1"
	if len(ids) > 0 {
		var columnsBuilder strings.Builder
		for idIndex := 0; idIndex < len(ids); idIndex++ {
			if idIndex != 0 {
				_, _ = columnsBuilder.WriteString(", ")
			}
			_, _ = columnsBuilder.WriteString(fmt.Sprintf(`
				IF(items%d.allows_multiple_attempts, (
					SELECT number FROM (
						SELECT results.attempt_id, ROW_NUMBER() OVER (ORDER BY started_at) AS number
						FROM results
						JOIN attempts ON attempts.participant_id = results.participant_id and attempts.id = results.attempt_id
						WHERE results.participant_id = attempts%d.participant_id AND
							results.item_id = items%d.id AND
							attempts.parent_attempt_id <=> attempts%d.parent_attempt_id
					) AS numbers WHERE numbers.attempt_id = attempts%d.id
				), NULL) AS number%d,
				attempts%d.id AS attempt%d`, idIndex, idIndex, idIndex, idIndex, idIndex, idIndex, idIndex, idIndex))
		}
		columnsList = columnsBuilder.String()
	}
	return columnsList
}

func resultsForBreadcrumbsHierarchy(ids []int64, data map[string]interface{}) (
	attemptIDMap map[int64]int64, attemptNumberMap map[int64]int) {
	attemptIDMap = make(map[int64]int64, len(ids))
	attemptNumberMap = make(map[int64]int, len(ids))
	for idIndex := 0; idIndex < len(ids); idIndex++ {
		attemptIDMap[ids[idIndex]] = data[fmt.Sprintf("attempt%d", idIndex)].(int64)
		numberKey := fmt.Sprintf("number%d", idIndex)
		if data[numberKey] != nil {
			attemptNumberMap[ids[idIndex]] = int(data[numberKey].(int64))
		}
	}
	return attemptIDMap, attemptNumberMap
}

func (s *ItemStore) breadcrumbsHierarchyForAttempt(
	ids []int64, groupID, attemptID int64, requireContentAccessToTheLastItem bool,
	columnsList string, withWriteLock bool) *DB {
	lastItemIndex := len(ids) - 1
	subQuery := s.
		itemAttemptChainWithoutAttemptForTail(ids, groupID, false, requireContentAccessToTheLastItem, withWriteLock).
		Where(fmt.Sprintf("attempts%d.id = ?", lastItemIndex), attemptID)
	subQuery = subQuery.
		Joins(fmt.Sprintf(`
				JOIN results AS results%d ON results%d.participant_id = ? AND
					results%d.item_id = items%d.id AND results%d.started_at IS NOT NULL`,
			lastItemIndex, lastItemIndex, lastItemIndex, lastItemIndex, lastItemIndex), groupID).
		Joins(fmt.Sprintf(`
				JOIN attempts AS attempts%d ON attempts%d.participant_id = results%d.participant_id AND
					attempts%d.id = results%d.attempt_id`, lastItemIndex, lastItemIndex, lastItemIndex, lastItemIndex, lastItemIndex))
	if len(ids) > 1 {
		subQuery = subQuery.Where(fmt.Sprintf(
			"IF(attempts%d.root_item_id = items%d.id, attempts%d.parent_attempt_id, attempts%d.id) = attempts%d.id",
			lastItemIndex, lastItemIndex, lastItemIndex, lastItemIndex, lastItemIndex-1))
	}

	subQuery = subQuery.Select(columnsList)
	visibleItems := s.Visible(groupID).
		Select("items.id, items.is_root, items.allows_multiple_attempts, visible.can_view_generated_value")

	if withWriteLock {
		subQuery = subQuery.WithWriteLock()
		visibleItems = visibleItems.WithWriteLock()
	}
	return s.Raw("WITH visible_items AS ? ?", visibleItems.SubQuery(), subQuery.QueryExpr())
}

// CheckSubmissionRights checks if the participant group can submit an answer for the given item (task),
// i.e. the item (task) exists and is not read-only and the participant has at least content:view permission on the item;
func (s *ItemStore) CheckSubmissionRights(participantID, itemID int64) (hasAccess bool, reason, err error) {
	s.mustBeInTransaction()
	recoverPanics(&err)

	var readOnly bool
	err = s.Visible(participantID).WherePermissionIsAtLeast("view", "content").
		Where("id = ?", itemID).
		WithWriteLock().
		PluckFirst("read_only", &readOnly).Error()
	if gorm.IsRecordNotFoundError(err) {
		return false, errors.New("no access to the task item"), nil
	}
	mustNotBeError(err)

	if readOnly {
		return false, errors.New("item is read-only"), nil
	}

	return true, nil, nil
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
