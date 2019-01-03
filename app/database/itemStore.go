package database

import (
	"fmt"
	"github.com/davecgh/go-spew/spew"

	"github.com/France-ioi/AlgoreaBackend/app/auth"
	"github.com/France-ioi/AlgoreaBackend/app/logging"
	t "github.com/France-ioi/AlgoreaBackend/app/types"
)

// ItemStore implements database operations on items
type ItemStore struct {
	*DataStore
}

type itemAccessDetails struct {
	ItemID        int64 `sql:"column:idItem"`
	FullAccess    bool  `sql:"column:fullAccess"`
	PartialAccess bool  `sql:"column:partialAccess"`
	GrayedAccess  bool  `sql:"column:grayedAccess"`
}

type itemAncestorDetails struct {
	ID int64 `sql:"column:ID"`
	Type string `sql:"column:sType"`
	IdItemChild int64 `sql:"column:idItemChild"`
}

const (
	ItemTypeRoot      = "Root"
	ItemTypeCategory = "Category"
	ItemTypeChapter  = "Chapter"
	ItemTypeTask     = "Task"
	ItemTypeCourse   = "Course"
)

// Item matches the content the `items` table
type Item struct {
	ID            t.Int64  `sql:"column:ID"`
	Type          t.String `sql:"column:sType"`
	TeamsEditable bool     `sql:"column:bTeamsEditable"` // use Go default in DB (to be fixed)
	NoScore       bool     `sql:"column:bNoScore"`       // use Go default in DB (to be fixed)
	Version       int64    `sql:"column:iVersion"`       // use Go default in DB (to be fixed)
}

func (s *ItemStore) tableName() string {
	return "items"
}

// Insert does a INSERT query in the given table with data that may contain types.* types
func (s *ItemStore) Insert(data *Item) error {
	return s.db.insert(s.tableName(), data)
}

// IsValidHierarchy gets an ordered set of item ids and returns whether they forms a valid item hierarchy path from a root
func (s *ItemStore) IsValidHierarchy(ids []int64) (bool, error) {
	var ancDets []itemAncestorDetails
	db := s.db.Table(s.tableName()).Select("items.ID as ID, items.sType as sType, items_items.idItemChild as idItemChild").
		Joins("LEFT JOIN items_items ON items_items.idItemParent = items.ID").
		Where("items.ID in (?)", ids).Scan(&ancDets)
	if db.Error != nil {
		return false, db.Error
	}
	spew.Dump(ids, ancDets)
	// Todo: validate array
	if err := checkHierarchy(ancDets); err != nil {
		return false, err
	}
	return true, nil
}

// ValidateUserAccess gets a set of item ids and returns whether the given user is authorized to see them all
func (s *ItemStore) ValidateUserAccess(user *auth.User, itemIDs []int64) (bool, error) {

	var accDets []itemAccessDetails
	db := s.GroupItems().MatchingUserAncestors(user).
		Select("idItem, MAX(bCachedFullAccess) AS fullAccess, MAX(bCachedPartialAccess) AS partialAccess, MAX(bCachedGrayedAccess) AS grayedAccess").
		Where("groups_items.idItem IN (?)", itemIDs).
		Group("idItem").Scan(&accDets)
	if db.Error != nil {
		return false, db.Error
	}

	if err := checkAccess(itemIDs, accDets); err != nil {
		logging.Logger.Infof("User access validation failed: %v", err)
		return false, nil
	}
	return true, nil
}

// checkAccess checks if the user has access to all items:
// - user has to have full access to all items
// OR
// - user has to have full access to all but last, and grayed access to that last item.
func checkAccess(itemIDs []int64, accDets []itemAccessDetails) error {
	for i, id := range itemIDs {
		last := (i == len(itemIDs)-1)
		if err := checkAccessForID(id, last, accDets); err != nil {
			return err
		}
	}
	return nil
}

func checkAccessForID(id int64, last bool, accDets []itemAccessDetails) error {
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

func checkHierarchy(ancDets []itemAncestorDetails) error {
	items := map[int64]itemAncestorDetails{}

	for _, item := range ancDets {
		items[item.ID] = item
	}

	// find root
	var root itemAncestorDetails
	for _, item := range items {
		if item.Type == string(ItemTypeRoot) {
			root = item
		}
	}

	if root.ID == 0 {
		return fmt.Errorf("Incorrect hierarchy on given item ids")
	}

	if !checkHierarchyItem(root, items, 1) {
		fmt.Errorf("Incorrect hierarchy on given item ids")
	}

	return nil
}

func checkHierarchyItem(item itemAncestorDetails, items map[int64]itemAncestorDetails, level int) bool {
	if item.IdItemChild == 0 {
		if len(items) > level {
			return false
		}
		return true
	}

	if child, ok := items[item.IdItemChild]; !ok {
		return false
	} else {
		return checkHierarchyItem(child, items, level + 1)
	}
}
