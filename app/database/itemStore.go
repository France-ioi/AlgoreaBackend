package database

import (
	"fmt"

	"github.com/France-ioi/AlgoreaBackend/app/logging"
	"github.com/France-ioi/AlgoreaBackend/app/types"
)

// ItemStore implements database operations on items
type ItemStore struct {
	*DataStore
}

type ItemAccessDetails struct {
	// MAX(groups_items.bCachedFullAccess)
	FullAccess    bool  `sql:"column:fullAccess" json:"full_access"`
	// MAX(groups_items.bCachedPartialAccess)
	PartialAccess bool  `sql:"column:partialAccess" json:"partial_access"`
	// MAX(groups_items.bCachedGrayAccess)
	GrayedAccess  bool  `sql:"column:grayedAccess" json:"grayed_access"`
}

type itemAccessDetailsWithID struct {
	ItemID        int64 `sql:"column:idItem"`
	ItemAccessDetails
}

// Item matches the content the `items` table
type Item struct {
	ID                types.Int64  `sql:"column:ID"`
	Type              types.String `sql:"column:sType"`
	DefaultLanguageID types.Int64  `sql:"column:idDefaultLanguage"`
	TeamsEditable     types.Bool   `sql:"column:bTeamsEditable"`
	NoScore           types.Bool   `sql:"column:bNoScore"`
	Version           int64        `sql:"column:iVersion"` // use Go default in DB (to be fixed)
}

type NavigationItemCommonFields struct {
	// items
	ID                		int64  `sql:"column:ID" json:"item_id"`
	Type              		string `sql:"column:sType" json:"type"`
	TransparentFolder 		bool	 `sql:"colums:bTransparentFolder" json:"transparent_folder"`
	// whether items.idItemUnlocked is empty
	HasUnlockedItems  		bool   `sql:"hasUnlockedItems" json:"has_unlocked_items"`

	// title (from items_strings) in the user’s default language or (if not available) default language of the item
	Title         				string `sql:"column:sTitle" json:"title"`

	// from users_items for current user
	UserScore 						float32	`sql:"column:iScore" json:"user_score,omitempty"`
	UserValidated 				bool	  `sql:"column:bValidated" json:"user_validated,omitempty"`
	UserFinished					bool	  `sql:"column:bFinished" json:"user_finished,omitempty"`
	KeyObtained 					bool 	  `sql:"column:bKeyObtained" json:"key_obtained,omitempty"`
	SubmissionsAttempts   int64   `sql:"column:nbSubmissionsAttempts" json:"submissions_attempts,omitempty"`
	StartDate             string  `sql:"column:sStartDate" json:"start_date,omitempty"` // iso8601 str
	ValidationDate        string  `sql:"column:sValidationDate" json:"validation_date,omitempty"` // iso8601 str
	FinishDate            string  `sql:"column:sFinishDate" json:"finish_date,omitempty"` // iso8601 str
}

type NavigationItemChild struct {
	*NavigationItemCommonFields

	Order 						int64 `sql:"column:iChildOrder" json:"order"`
	AccessRestricted  bool  `sql:"column:bAccessRestricted" json:"access_restricted"`
}


// TreeItem represents the content of `items` table filled with some additional information
// from `items_strings` and `items_items`.
type TreeItem struct {
	ID            types.Int64  `sql:"column:ID"`
	Type          types.String `sql:"column:sType"`
	TeamsEditable bool         `sql:"column:bTeamsEditable"` // use Go default in DB (to be fixed)
	NoScore       bool         `sql:"column:bNoScore"`       // use Go default in DB (to be fixed)
	Version       int64        `sql:"column:iVersion"`       // use Go default in DB (to be fixed)
	Title         types.String `sql:"column:sTitle"`         // from items_strings
	Order         types.Int64  `sql:"column:iChildOrder"`    // from items_items
	ParentID      int64        `sql:"column:idItemParent"`
	TreeLevel     int64        `sql:"column:treeLevel"` // information if direct child of root
}

func (s *ItemStore) tableName() string {
	return "items"
}


func (s *ItemStore) GetRawNavigationData(rootID, userID, userLanguageID, defaultLanguageID int64) (*[]NavigationItemChild, error){
	var result []NavigationItemChild
	if err := s.Raw(
		"SELECT union_table.ID, union_table.sType, union_table.bTransparentFolder, " +
			"COALESCE(union_table.idItemUnlocked, '')<>'' as hasUnlockedItems, " +
			"COALESCE(ustrings.sTitle, dstrings.sTitle) AS sTitle, " +
			"users_items.iScore AS iScore, users_items.bValidated AS bValidated, " +
			"users_items.bFinished AS bFinished, users_items.bKeyObtained AS bKeyObtained, " +
			"users_items.nbSubmissionsAttempts AS nbSubmissionsAttempts, " +
			"users_items.sStartDate AS sStartDate, users_items.sValidationDate AS sValidationDate, " +
			"users_items.sFinishDate AS sFinishDate, " +
			"union_table.iChildOrder AS iChildOrder, " +
			"union_table.idItemParent AS idItemParent " +
			"FROM " +
			"(SELECT items.*, NULL AS idItemParent, NULL AS iChildOrder FROM items WHERE items.ID=? UNION " +
			"(SELECT items.*, idItemParent, iChildOrder FROM items, items_items " +
			" WHERE items.ID=idItemChild AND idItemParent=?" +
			" ORDER BY items_items.iChildOrder) UNION" +
			"(SELECT items.*, ii2.idItemParent, ii2.iChildOrder FROM items, items_items ii1 " +
			" JOIN items_items ii2 ON ii1.idItemChild = ii2.idItemParent " +
			" WHERE items.ID=ii2.idItemChild AND ii1.idItemParent=?" +
			" ORDER BY ii2.idItemParent, ii2.iChildOrder)) union_table " +
			"LEFT JOIN items_strings ustrings ON ustrings.idItem=union_table.ID AND ustrings.idLanguage=? " +
			"LEFT JOIN items_strings dstrings ON dstrings.idItem=union_table.ID AND dstrings.idLanguage=? " +
			"LEFT JOIN users_items ON users_items.idItem=union_table.ID AND users_items.idUser=?",
			rootID, rootID, rootID, userLanguageID, defaultLanguageID, userID).Scan(&result).Error(); err != nil {
				return nil, err
	}
	return &result, nil
}

// GetOne returns a single element of the tree structure.
func (s *ItemStore) GetOne(id, languageID int64) (*TreeItem, error) {
	var it TreeItem

	if err := s.ByID(id).
		Select("items.*, items_strings.sTitle as sTitle").
		Joins("LEFT JOIN items_strings ON (items.ID=items_strings.idItem)").
		Where("items_strings.idLanguage=?", languageID).
		Take(&it).Error(); err != nil {
		return nil, fmt.Errorf("failed to get item '%d': %v", id, err)
	}
	return &it, nil
}

// GetChildrenOf returns all children of the given root item.
func (s *ItemStore) GetChildrenOf(rootID, languageID int64) ([]*TreeItem, error) {
	var itt []*TreeItem

	err := s.All().
		Joins("JOIN items_ancestors ON (items.ID=items_ancestors.idItemChild)").
		Joins("JOIN items_strings ON (items_ancestors.idItemChild=items_strings.idItem)").
		Joins("JOIN items_items ON (items_ancestors.idItemChild=items_items.idItemChild)").
		Where("items_ancestors.idItemAncestor=? AND items_strings.idLanguage=?", rootID, languageID).
		Select("items.*, sTitle, iChildOrder, idItemParent").
		Scan(&itt).Error()
	if err != nil {
		return nil, fmt.Errorf("failed to get tree children of '%d': %v", rootID, err)
	}
	return itt, nil
}

// Insert does a INSERT query in the given table with data that may contain types.* types
func (s *ItemStore) Insert(data *Item) error {
	return s.insert(s.tableName(), data)
}

// ByID returns a composable query of items filtered by itemID
func (s *ItemStore) ByID(itemID int64) DB {
	return s.All().Where("items.ID = ?", itemID)
}

// All creates a composable query without filtering
func (s *ItemStore) All() DB {
	return s.table(s.tableName())
}

// HasManagerAccess returns whether the user has manager access to all the given item_id's
// It is assumed that the `OwnerAccess` implies manager access
func (s *ItemStore) HasManagerAccess(user AuthUser, itemID int64) (found bool, allowed bool, err error) {

	var dbRes []struct {
		ItemID        int64 `sql:"column:idItem"`
		ManagerAccess bool  `sql:"column:bManagerAccess"`
		OwnerAccess   bool  `sql:"column:bOwnerAccess"`
	}

	db := s.GroupItems().MatchingUserAncestors(user).
		Select("idItem, bManagerAccess, bOwnerAccess").
		Where("idItem = ?", itemID).
		Scan(&dbRes)
	if db.Error() != nil {
		return false, false, db.Error()
	}
	if len(dbRes) != 1 {
		return false, false, nil
	}
	item := dbRes[0]
	return true, item.ManagerAccess || item.OwnerAccess, nil
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
func (s *ItemStore) ValidateUserAccess(user AuthUser, itemIDs []int64) (bool, error) {

	var accDets []itemAccessDetailsWithID
	db := s.GroupItems().MatchingUserAncestors(user).
		Select("idItem, MAX(bCachedFullAccess) AS fullAccess, MAX(bCachedPartialAccess) AS partialAccess, MAX(bCachedGrayedAccess) AS grayedAccess").
		Where("groups_items.idItem IN (?)", itemIDs).
		Group("idItem").Scan(&accDets)
	if db.Error() != nil {
		return false, db.Error()
	}

	if err := checkAccess(itemIDs, accDets); err != nil {
		logging.Logger.Infof("checkAccess %v %v", itemIDs, accDets)
		logging.Logger.Infof("User access validation failed: %v", err)
		return false, nil
	}
	return true, nil
}

// checkAccess checks if the user has access to all items:
// - user has to have full access to all items
// OR
// - user has to have full access to all but last, and grayed access to that last item.
func checkAccess(itemIDs []int64, accDets []itemAccessDetailsWithID) error {
	for i, id := range itemIDs {
		last := i == len(itemIDs)-1
		if err := checkAccessForID(id, last, accDets); err != nil {
			return err
		}
	}
	return nil
}

func checkAccessForID(id int64, last bool, accDets []itemAccessDetailsWithID) error {
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

	db := s.ItemItems().All()
	previousID := ids[0]
	for index, id := range ids {
		if index == 0 {
			continue
		}

		db = db.Or("idItemParent=? AND idItemChild=?", previousID, id)
		previousID = id
	}

	count := 0
	if err := db.Count(&count).Error(); err != nil {
		return false, err
	}

	if count != len(ids)-1 {
		return false, nil
	}

	return true, nil
}
