package database

import (
	"errors"
	"fmt"
	"strings"

	"github.com/jinzhu/gorm"

	log "github.com/France-ioi/AlgoreaBackend/app/logging"
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

// IsValidParticipationHierarchyForParentAttempt checks if the given list of item ids is a valid participation hierarchy
// for the given `parentAttemptID` which means all the following statements are true:
//  * the first item in `ids` is a root items (items.is_root) or the item of a group (groups.activity_id)
//    the `groupID` is a descendant of,
//  * `ids` is an ordered list of parent-child items,
//  * the `groupID` group has at least 'content' access on each of the items in `ids`,
//  * the `groupID` group has a started, allowing submission, not ended result for each item but the last,
//    with `parentAttemptID` (or its parent attempt each time we reach a root of an attempt) as the attempt,
//  * if `ids` consists of only one item, the `parentAttemptID` is zero.
func (s *ItemStore) IsValidParticipationHierarchyForParentAttempt(
	ids []int64, groupID, parentAttemptID int64, requireContentAccessToTheLastItem bool) (bool, error) {
	if len(ids) == 0 || len(ids) == 1 && parentAttemptID != 0 {
		return false, nil
	}

	return s.participationHierarchyForParentAttempt(ids, groupID, parentAttemptID, true, requireContentAccessToTheLastItem, "1").HasRows()
}

func (s *ItemStore) participationHierarchyForParentAttempt(
	ids []int64, groupID, parentAttemptID int64, requireAttemptsToBeActive, requireContentAccessToTheLastItem bool,
	columnsList string) *DB {
	subQuery := s.itemAttemptChainWithoutAttemptForTail(ids, groupID, requireAttemptsToBeActive, requireContentAccessToTheLastItem, false)

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
		Select("groups.activity_id")

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
//  * the first item in `ids` is a root items (items.is_root) or the item of a group (groups.activity_id)
//    the `groupID` is a descendant of,
//  * `ids` is an ordered list of parent-child items,
//  * the `groupID` group has at least 'content' access on each of the items in `ids` except for the last one and
//    at least 'info' access on the last one,
//  * the `groupID` group has a started result for each item but the last,
//    with `parentAttemptID` (or its parent attempt each time we reach a root of an attempt) as the attempt,
//  * if `ids` consists of only one item, the `parentAttemptID` is zero.
func (s *ItemStore) BreadcrumbsHierarchyForParentAttempt(ids []int64, groupID, parentAttemptID int64) (
	attemptIDMap map[int64]int64, attemptNumberMap map[int64]int, err error) {
	if len(ids) == 0 || len(ids) == 1 && parentAttemptID != 0 {
		return nil, nil, nil
	}

	defer recoverPanics(&err)

	columnsList := columnsListForBreadcrumbsHierarchy(ids[:len(ids)-1])
	query := s.participationHierarchyForParentAttempt(
		ids, groupID, parentAttemptID, false, false, columnsList)
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
func (s *ItemStore) BreadcrumbsHierarchyForAttempt(ids []int64, groupID, attemptID int64) (
	attemptIDMap map[int64]int64, attemptNumberMap map[int64]int, err error) {
	if len(ids) == 0 {
		return nil, nil, nil
	}

	defer recoverPanics(&err)

	columnsList := columnsListForBreadcrumbsHierarchy(ids)
	query := s.participationHierarchyForAttempt(
		ids, groupID, attemptID, false, false, columnsList, false)
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

func (s *ItemStore) participationHierarchyForAttempt(
	ids []int64, groupID, attemptID int64, requireAttemptsToBeActive, requireContentAccessToTheLastItem bool,
	columnsList string, withWriteLock bool) *DB {
	lastItemIndex := len(ids) - 1
	subQuery := s.
		itemAttemptChainWithoutAttemptForTail(ids, groupID, requireAttemptsToBeActive, requireContentAccessToTheLastItem, withWriteLock).
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
	if requireAttemptsToBeActive {
		subQuery = subQuery.
			Where(fmt.Sprintf("attempts%d.ended_at IS NULL AND NOW() < attempts%d.allows_submissions_until",
				lastItemIndex, lastItemIndex))
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
