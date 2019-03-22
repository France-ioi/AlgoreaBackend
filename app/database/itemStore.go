package database

import (
	"fmt"

	log "github.com/France-ioi/AlgoreaBackend/app/logging"
	"github.com/France-ioi/AlgoreaBackend/app/types"
)

// ItemStore implements database operations on items
type ItemStore struct {
	*DataStore
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

func (s *ItemStore) tableName() string {
	return "items"
}

// Visible returns a view of the visible items for the given user
func (s *ItemStore) Visible(user AuthUser) *DB {
	return s.WhereItemsAreVisible(user)
}

// VisibleByID returns a view of the visible item identified by itemID, for the given user
func (s *ItemStore) VisibleByID(user AuthUser, itemID int64) *DB {
	return s.Visible(user).Where("items.ID = ?", itemID)
}

// VisibleChildrenOfID returns a view of the visible children of item identified by itemID, for the given user
func (s *ItemStore) VisibleChildrenOfID(user AuthUser, itemID int64) *DB {
	return s.
		Visible(user).
		Joins("JOIN ? ii ON items.ID=idItemChild", s.ItemItems().SubQuery()).
		Where("ii.idItemParent = ?", itemID)
}

// VisibleGrandChildrenOfID returns a view of the visible grand-children of item identified by itemID, for the given user
func (s *ItemStore) VisibleGrandChildrenOfID(user AuthUser, itemID int64) *DB {
	return s.
		Visible(user).                                                                       // visible items are the leaves (potential grandChildren)
		Joins("JOIN ? ii1 ON items.ID = ii1.idItemChild", s.ItemItems().SubQuery()).         // get their parents' IDs (ii1)
		Joins("JOIN ? ii2 ON ii2.idItemChild = ii1.idItemParent", s.ItemItems().SubQuery()). // get their grand parents' IDs (ii2)
		Where("ii2.idItemParent = ?", itemID)
}

// RawItem represents one row of the getItem service data returned from the DB
type RawItem struct {
	// items
	ID                     int64  `sql:"column:ID"`
	Type                   string `sql:"column:sType"`
	DisplayDetailsInParent bool   `sql:"column:bDisplayDetailsInParent"`
	ValidationType         string `sql:"column:sValidationType"`
	HasUnlockedItems       bool   `sql:"column:hasUnlockedItems"` // whether items.idItemUnlocked is empty
	ScoreMinUnlock         int64  `sql:"column:iScoreMinUnlock"`
	TeamMode               string `sql:"column:sTeamMode"`
	TeamsEditable          bool   `sql:"column:bTeamsEditable"`
	TeamMaxMembers         int64  `sql:"column:iTeamMaxMembers"`
	HasAttempts            bool   `sql:"column:bHasAttempts"`
	AccessOpenDate         string `sql:"column:sAccessOpenDate"` // iso8601 str
	Duration               string `sql:"column:sDuration"`
	EndContestDate         string `sql:"column:sEndContestDate"` // iso8601 str
	NoScore                bool   `sql:"column:bNoScore"`
	GroupCodeEnter         bool   `sql:"column:groupCodeEnter"`

	// root node only
	TitleBarVisible *bool   `sql:"column:bTitleBarVisible"`
	ReadOnly        *bool   `sql:"column:bReadOnly"`
	FullScreen      *string `sql:"column:sFullScreen"`
	ShowSource      *bool   `sql:"column:bShowSource"`
	ValidationMin   *int64  `sql:"column:iValidationMin"`
	ShowUserInfos   *bool   `sql:"column:bShowUserInfos"`
	ContestPhase    *string `sql:"column:sContestPhase"`
	URL             *string `sql:"column:sUrl"`          // only if not a chapter
	UsesAPI         *bool   `sql:"column:bUsesAPI"`      // only if not a chapter
	HintsAllowed    *bool   `sql:"column:bHintsAllowed"` // only if not a chapter

	// from items_strings: in the user’s default language or (if not available) default language of the item
	StringLanguageID  int64  `sql:"column:idLanguage"`
	StringTitle       string `sql:"column:sTitle"`
	StringImageURL    string `sql:"column:sImageUrl"`
	StringSubtitle    string `sql:"column:sSubtitle"`
	StringDescription string `sql:"column:sDescription"`
	StringEduComment  string `sql:"column:sEduComment"`

	// from users_items for current user
	UserActiveAttemptID     int64   `sql:"column:idAttemptActive"`
	UserScore               float32 `sql:"column:iScore"`
	UserSubmissionsAttempts int64   `sql:"column:nbSubmissionsAttempts"`
	UserValidated           bool    `sql:"column:bValidated"`
	UserFinished            bool    `sql:"column:bFinished"`
	UserKeyObtained         bool    `sql:"column:bKeyObtained"`
	UserHintsCached         int64   `sql:"column:nbHintsCached"`
	UserStartDate           string  `sql:"column:sStartDate"`        // iso8601 str
	UserValidationDate      string  `sql:"column:sValidationDate"`   // iso8601 str
	UserFinishDate          string  `sql:"column:sFinishDate"`       // iso8601 str
	UserContestStartDate    string  `sql:"column:sContestStartDate"` // iso8601 str
	UserState               *string `sql:"column:sState"`            // only if not a chapter
	UserAnswer              *string `sql:"column:sAnswer"`           // only if not a chapter

	// items_items
	Order            int64   `sql:"column:iChildOrder"`
	Category         *string `sql:"column:sCategory"`
	AlwaysVisible    *bool   `sql:"column:bAlwaysVisible"`
	AccessRestricted *bool   `sql:"column:bAccessRestricted"`

	*ItemAccessDetails
}

// GetRawItemData reads data needed by the getItem service from the DB and returns an array of RawItem's
func (s *ItemStore) GetRawItemData(rootID, userID, userLanguageID int64, user AuthUser) (*[]RawItem, error) {
	var result []RawItem

	commonColumns := `items.ID AS ID,
		items.sType,
		items.bDisplayDetailsInParent,
		items.sValidationType,
		items.idItemUnlocked,
		items.iScoreMinUnlock,
		items.sTeamMode,
		items.bTeamsEditable,
		items.iTeamMaxMembers,
		items.bHasAttempts,
		items.sAccessOpenDate,
		items.sDuration,
		items.sEndContestDate,
		items.bNoScore,
		items.idDefaultLanguage,
		items.groupCodeEnter, `

	rootItemQuery := s.ByID(rootID).Select(
		commonColumns + `items.bTitleBarVisible,
		items.bReadOnly,
		items.sFullScreen,
		items.bShowSource,
		items.iValidationMin,
		items.bShowUserInfos,
		items.sContestPhase,
		items.sUrl,
		IF(items.sType <> 'Chapter', items.bUsesAPI, NULL) AS bUsesAPI,
		IF(items.sType <> 'Chapter', items.bHintsAllowed, NULL) AS bHintsAllowed,
		NULL AS iChildOrder, NULL AS sCategory, NULL AS bAlwaysVisible, NULL AS bAccessRestricted`)

	childrenQuery := s.Select(
		commonColumns+`NULL AS bTitleBarVisible,
		NULL AS bReadOnly,
		NULL AS sFullScreen,
		NULL AS bShowSource,
		NULL AS iValidationMin,
		NULL AS bShowUserInfos,
		NULL AS sContestPhase,
		NULL AS sUrl,
		NULL AS bUsesAPI,
		NULL AS bHintsAllowed,
		iChildOrder, sCategory, bAlwaysVisible, bAccessRestricted`).
		Joins("JOIN items_items ON items.ID=idItemChild AND idItemParent=?", rootID)

	unionQuery := rootItemQuery.UnionAll(childrenQuery.QueryExpr())
	// This query can be simplified if we add a column for relation degrees into `items_ancestors`
	query := s.Raw(`
    SELECT
		  items.ID,
      items.sType,
		  items.bDisplayDetailsInParent,
      items.sValidationType,`+
		// idItemUnlocked is a comma-separated list of item IDs which will be unlocked if this item is validated
		// Here we consider both NULL and an empty string as FALSE
		` COALESCE(items.idItemUnlocked, '')<>'' as hasUnlockedItems,
			items.iScoreMinUnlock,
			items.sTeamMode,
			items.bTeamsEditable,
			items.iTeamMaxMembers,
			items.bHasAttempts,
			items.sAccessOpenDate,
			items.sDuration,
			items.sEndContestDate,
			items.bNoScore,
			items.groupCodeEnter,

			COALESCE(user_strings.idLanguage, default_strings.idLanguage) AS idLanguage,
			IF(user_strings.idLanguage IS NULL, default_strings.sTitle, user_strings.sTitle) AS sTitle,
			IF(user_strings.idLanguage IS NULL, default_strings.sImageUrl, user_strings.sImageUrl) AS sImageUrl,
			IF(user_strings.idLanguage IS NULL, default_strings.sSubtitle, user_strings.sSubtitle) AS sSubtitle,
			IF(user_strings.idLanguage IS NULL, default_strings.sDescription, user_strings.sDescription) AS sDescription,
			IF(user_strings.idLanguage IS NULL, default_strings.sEduComment, user_strings.sEduComment) AS sEduComment,

			users_items.idAttemptActive AS idAttemptActive,
			users_items.iScore AS iScore,
			users_items.nbSubmissionsAttempts AS nbSubmissionsAttempts,
			users_items.bValidated AS bValidated,
			users_items.bFinished AS bFinished,
			users_items.bKeyObtained AS bKeyObtained,
			users_items.nbHintsCached AS nbHintsCached,
			users_items.sStartDate AS sStartDate,
			users_items.sValidationDate AS sValidationDate,
			users_items.sFinishDate AS sFinishDate,
			users_items.sContestStartDate AS sContestStartDate,
			IF(items.sType <> 'Chapter', users_items.sState, NULL) as sState,
			users_items.sAnswer,

			items.iChildOrder AS iChildOrder,
			items.sCategory AS sCategory,
			items.bAlwaysVisible,
			items.bAccessRestricted, `+
		// inputItem only
		` items.bTitleBarVisible,
			items.bReadOnly,
			items.sFullScreen,
			items.bShowSource,
			items.iValidationMin,
			items.bShowUserInfos,
			items.sContestPhase,
			items.sUrl,
			items.bUsesAPI,
			items.bHintsAllowed,
			accessRights.fullAccess, accessRights.partialAccess, accessRights.grayedAccess, accessRights.accessSolutions
    FROM ? items `, unionQuery.SubQuery()).
		JoinsUserAndDefaultItemStrings(user).
		Joins("LEFT JOIN users_items ON users_items.idItem=items.ID AND users_items.idUser=?", userID).
		Joins("JOIN ? accessRights on accessRights.idItem=items.ID AND (fullAccess>0 OR partialAccess>0 OR grayedAccess>0)",
			s.AccessRights(user).SubQuery()).
		Order("iChildOrder")

	if err := query.Scan(&result).Error(); err != nil {
		return nil, err
	}
	return &result, nil
}

// AccessRights returns a composable query for getting
// (idItem, fullAccess, partialAccess, grayedAccess, accessSolutions) for the given user
func (s *ItemStore) AccessRights(user AuthUser) *DB {
	return s.GroupItems().MatchingUserAncestors(user).
		Select(
			"idItem, MIN(sCachedFullAccessDate) <= NOW() AS fullAccess, " +
				"MIN(sCachedPartialAccessDate) <= NOW() AS partialAccess, " +
				"MIN(sCachedGrayedAccessDate) <= NOW() AS grayedAccess, " +
				"MIN(sCachedAccessSolutionsDate) <= NOW() AS accessSolutions").
		Group("idItem")
}

// Insert does a INSERT query in the given table with data that may contain types.* types
func (s *ItemStore) Insert(data *Item) error {
	return s.insert(s.tableName(), data)
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
func (s *ItemStore) GetAccessDetailsForIDs(user AuthUser, itemIDs []int64) ([]ItemAccessDetailsWithID, error) {
	var accessDetails []ItemAccessDetailsWithID
	db := s.AccessRights(user).
		Where("groups_items.idItem IN (?)", itemIDs).
		Scan(&accessDetails)
	if err := db.Error(); err != nil {
		return nil, err
	}
	return accessDetails, nil
}

// GetAccessDetailsMapForIDs returns access details for given item IDs and the given user as a map (item_id->details)
func (s *ItemStore) GetAccessDetailsMapForIDs(user AuthUser, itemIDs []int64) (map[int64]ItemAccessDetails, error) {
	accessDetails, err := s.GetAccessDetailsForIDs(user, itemIDs)
	if err != nil {
		return nil, err
	}
	accessDetailsMap := make(map[int64]ItemAccessDetails, len(accessDetails))
	for _, row := range accessDetails {
		accessDetailsMap[row.ItemID] = ItemAccessDetails{
			FullAccess:      row.FullAccess,
			PartialAccess:   row.PartialAccess,
			GrayedAccess:    row.GrayedAccess,
			AccessSolutions: row.AccessSolutions,
		}
	}
	return accessDetailsMap, nil
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
