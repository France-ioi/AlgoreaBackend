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

// ItemAccessDetails represents access rights for an item
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

// RawNavigationItem represents one row of a navigation subtree returned from the DB
type RawNavigationItem struct {
	// items
	ID                		int64    `sql:"column:ID"`
	Type              		string   `sql:"column:sType"`
	TransparentFolder 		bool	   `sql:"column:bTransparentFolder"`
	// whether items.idItemUnlocked is empty
	HasUnlockedItems  		bool     `sql:"column:hasUnlockedItems"`
	AccessRestricted  		bool  	 `sql:"column:bAccessRestricted"`

	// title (from items_strings) in the user’s default language or (if not available) default language of the item
	Title         				string   `sql:"column:sTitle"`

	// from users_items for current user
	UserScore 						float32	 `sql:"column:iScore"`
	UserValidated 				bool	   `sql:"column:bValidated"`
	UserFinished					bool	   `sql:"column:bFinished"`
	KeyObtained 					bool 	   `sql:"column:bKeyObtained"`
	SubmissionsAttempts   int64    `sql:"column:nbSubmissionsAttempts"`
	StartDate             string   `sql:"column:sStartDate"` // iso8601 str
	ValidationDate        string   `sql:"column:sValidationDate"` // iso8601 str
	FinishDate            string   `sql:"column:sFinishDate"` // iso8601 str

	// items_items
	IDItemParent					int64    `sql:"column:idItemParent"`
	Order 						    int64 	 `sql:"column:iChildOrder"`
}

func (s *ItemStore) tableName() string {
	return "items"
}

// GetRawNavigationData reads a navigation subtree from the DB and returns an array of RawNavigationItem's
func (s *ItemStore) GetRawNavigationData(rootID, userID, userLanguageID, defaultLanguageID int64) (*[]RawNavigationItem, error){
	var result []RawNavigationItem

	languageSelectPart := "COALESCE(ustrings.sTitle, dstrings.sTitle) AS sTitle, "
	languageJoinPart := "LEFT JOIN items_strings ustrings ON ustrings.idItem=union_table.ID AND ustrings.idLanguage=? "
	params := []interface{}{rootID, rootID, rootID}

	if userLanguageID == 0 {
		languageSelectPart = "dstrings.sTitle AS sTitle, "
		languageJoinPart = ""
	} else {
		params = append(params, userLanguageID)
	}

	params = append(params, defaultLanguageID, userID)

	// This query can be simplified if we add a column for relation degrees into `items_ancestors`
	if err := s.Raw(
		"SELECT union_table.ID, union_table.sType, union_table.bTransparentFolder, " +
			"COALESCE(union_table.idItemUnlocked, '')<>'' as hasUnlockedItems, " +
			languageSelectPart +
			"users_items.iScore AS iScore, users_items.bValidated AS bValidated, " +
			"users_items.bFinished AS bFinished, users_items.bKeyObtained AS bKeyObtained, " +
			"users_items.nbSubmissionsAttempts AS nbSubmissionsAttempts, " +
			"users_items.sStartDate AS sStartDate, users_items.sValidationDate AS sValidationDate, " +
			"users_items.sFinishDate AS sFinishDate, " +
			"union_table.iChildOrder AS iChildOrder, " +
			"union_table.bAccessRestricted, " +
			"union_table.idItemParent AS idItemParent " +
			"FROM " +
			"(SELECT items.ID, items.sType, items.bTransparentFolder, items.idItemUnlocked, " +
			" NULL AS idItemParent, NULL AS iChildOrder, NULL AS bAccessRestricted " +
			" FROM items WHERE items.ID=? UNION " +
			"(SELECT items.ID, items.sType, items.bTransparentFolder, items.idItemUnlocked, " +
			" idItemParent, iChildOrder, bAccessRestricted FROM items " +
			" JOIN items_items ON items.ID=idItemChild " +
			" WHERE idItemParent=?" +
			" ORDER BY items_items.iChildOrder) UNION" +
			"(SELECT  items.ID, items.sType, items.bTransparentFolder, items.idItemUnlocked," +
			" ii2.idItemParent, ii2.iChildOrder, ii2.bAccessRestricted FROM items " +
			" JOIN items_items ii1 ON ii1.idItemParent=? " +
			" JOIN items_items ii2 ON ii1.idItemChild = ii2.idItemParent " +
			" WHERE items.ID=ii2.idItemChild " +
			" ORDER BY ii2.idItemParent, ii2.iChildOrder)) union_table " +
			languageJoinPart +
			"LEFT JOIN items_strings dstrings ON dstrings.idItem=union_table.ID AND dstrings.idLanguage=? " +
			"LEFT JOIN users_items ON users_items.idItem=union_table.ID AND users_items.idUser=?",
			params...).Scan(&result).Error(); err != nil {
				return nil, err
	}
	return &result, nil
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
	accessDetails, err := s.getAccessDetailsForIDs(user, itemIDs)
	if err != nil {
		logging.Logger.Infof("User access rights loading failed: %v", err)
		return false, err
	}

	if err := checkAccess(itemIDs, accessDetails); err != nil {
		logging.Logger.Infof("checkAccess %v %v", itemIDs, accessDetails)
		logging.Logger.Infof("User access validation failed: %v", err)
		return false, nil
	}
	return true, nil
}

// getAccessDetailsForIDs returns access details for given item IDs and the given user
func (s *ItemStore) getAccessDetailsForIDs(user AuthUser, itemIDs []int64) ([]itemAccessDetailsWithID, error) {
	var accessDetails []itemAccessDetailsWithID
	db := s.GroupItems().MatchingUserAncestors(user).
		Select("idItem, MAX(bCachedFullAccess) AS fullAccess, MAX(bCachedPartialAccess) AS partialAccess, MAX(bCachedGrayedAccess) AS grayedAccess").
		Where("groups_items.idItem IN (?)", itemIDs).
		Group("idItem").Scan(&accessDetails)
	if err := db.Error(); err != nil {
		return nil, err
	}
	return accessDetails, nil
}

// GetAccessDetailsMapForIDs returns access details for given item IDs and the given user as a map (item_id->details)
func (s *ItemStore) GetAccessDetailsMapForIDs(user AuthUser, itemIDs []int64) (map[int64]ItemAccessDetails, error) {
	accessDetails, err := s.getAccessDetailsForIDs(user, itemIDs)
	if err != nil {
		return nil, err
	}
	accessDetailsMap := make(map[int64]ItemAccessDetails, len(accessDetails))
	for _, row := range accessDetails {
		accessDetailsMap[row.ItemID] = ItemAccessDetails{
			FullAccess: row.FullAccess,
			PartialAccess: row.PartialAccess,
			GrayedAccess: row.GrayedAccess,
		}
	}
	return accessDetailsMap, nil
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
	// For now, we don’t have a unique key for the pair ('idItemParent' and 'idItemChild') and
	// theoritically it’s still possible to have multiple rows with the same pair
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
