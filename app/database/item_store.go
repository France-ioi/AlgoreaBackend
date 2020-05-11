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

// IsValidParticipationHierarchy checks if the given list of item ids is a valid participation hierarchy which means
// all the following statements are true:
//  * the first item in `ids` is a root items (items.is_root) or the item of a group (groups.activity_id)
//    the `groupID` is a descendant of,
//  * `ids` is an ordered list of parent-child items,
//  * the `groupID` group has at least 'content' access on each of the items in `ids`,
//  * the `groupID` group has a started, allowing submission, not ended result for each item but the last,
//    with `attemptID` (or its parent attempt each time we reach a root of an attempt) as the attempt,
//  * if `ids` consists of only one item, the `attemptID` is zero.
func (s *ItemStore) IsValidParticipationHierarchy(
	ids []int64, groupID, attemptID int64, requireContentAccessToTheLastItem bool) (bool, error) {
	if len(ids) == 0 || len(ids) == 1 && attemptID != 0 {
		return false, nil
	}

	participantActivities := s.ActiveGroupAncestors().Where("child_group_id = ?", groupID).
		Joins("JOIN `groups` ON groups.id = groups_ancestors_active.ancestor_group_id").
		WithWriteLock().Select("groups.activity_id")

	subQuery := s.Table("visible_items as items0").Where("items0.id = ?", ids[0]).
		Where("items0.is_root OR items0.id IN ?", participantActivities.SubQuery())

	for i := 1; i < len(ids); i++ {
		subQuery = subQuery.Joins(fmt.Sprintf(`
				JOIN results AS results%d ON results%d.participant_id = ? AND
					results%d.item_id = items%d.id AND results%d.started_at IS NOT NULL`, i-1, i-1, i-1, i-1, i-1), groupID).
			Joins(fmt.Sprintf(`
				JOIN attempts AS attempts%d ON attempts%d.participant_id = results%d.participant_id AND
					attempts%d.id = results%d.attempt_id AND attempts%d.ended_at IS NULL AND
					NOW() < attempts%d.allows_submissions_until`, i-1, i-1, i-1, i-1, i-1, i-1, i-1)).
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
	}

	if requireContentAccessToTheLastItem {
		subQuery = subQuery.Where(fmt.Sprintf("items%d.can_view_generated_value >= ?", len(ids)-1),
			s.PermissionsGranted().ViewIndexByName("content"))
	}

	if len(ids) > 1 {
		subQuery = subQuery.
			Where(fmt.Sprintf("attempts%d.id = ?", len(ids)-2), attemptID)
	}

	subQuery = subQuery.Select("1").WithWriteLock()
	visibleItems := s.Visible(groupID).Select("items.id, items.is_root, visible.can_view_generated_value").WithWriteLock()

	return s.Raw("WITH visible_items AS ? ?", visibleItems.SubQuery(), subQuery.QueryExpr()).HasRows()
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
