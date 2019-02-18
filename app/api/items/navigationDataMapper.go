package items

import (
	"github.com/France-ioi/AlgoreaBackend/app/auth"
	"github.com/France-ioi/AlgoreaBackend/app/database"
)

// rawNavigationItem represents one row of a navigation subtree returned from the DB
type rawNavigationItem struct {
	// items
	ID                int64  `sql:"column:ID"`
	Type              string `sql:"column:sType"`
	TransparentFolder bool   `sql:"column:bTransparentFolder"`
	// whether items.idItemUnlocked is empty
	HasUnlockedItems bool `sql:"column:hasUnlockedItems"`
	AccessRestricted bool `sql:"column:bAccessRestricted"`

	// title (from items_strings) in the user’s default language or (if not available) default language of the item
	Title string `sql:"column:sTitle"`

	// from users_items for current user
	UserScore               float32 `sql:"column:iScore"`
	UserValidated           bool    `sql:"column:bValidated"`
	UserFinished            bool    `sql:"column:bFinished"`
	UserKeyObtained         bool    `sql:"column:bKeyObtained"`
	UserSubmissionsAttempts int64   `sql:"column:nbSubmissionsAttempts"`
	UserStartDate           string  `sql:"column:sStartDate"`      // iso8601 str
	UserValidationDate      string  `sql:"column:sValidationDate"` // iso8601 str
	UserFinishDate          string  `sql:"column:sFinishDate"`     // iso8601 str

	// items_items
	IDItemParent int64 `sql:"column:idItemParent"`
	Order        int64 `sql:"column:iChildOrder"`

	*database.ItemAccessDetails
}

// getRawNavigationData reads a navigation subtree from the DB and returns an array of rawNavigationItem's
func getRawNavigationData(dataStore *database.DataStore, rootID int64, user *auth.User) (*[]rawNavigationItem, error) {
	var result []rawNavigationItem
	items := dataStore.Items()

	// This query can be simplified if we add a column for relation degrees into `items_ancestors`

	commonAttributes := "items.ID, items.sType, items.bTransparentFolder, items.idItemUnlocked, items.idDefaultLanguage, fullAccess, partialAccess, grayedAccess"
	itemQ := items.VisibleByID(user, rootID).Select(commonAttributes + ", NULL AS idItemParent, NULL AS idItemGrandparent, NULL AS iChildOrder, NULL AS bAccessRestricted")
	childrenQ := items.VisibleChildrenOfID(user, rootID).Select(commonAttributes + ",	idItemParent, NULL AS idItemGrandparent, iChildOrder, bAccessRestricted")
	gChildrenQ := items.VisibleGrandChildrenOfID(user, rootID).Select(commonAttributes + ", ii1.idItemParent, ii2.idItemParent AS idItemGrandparent, ii1.iChildOrder, ii1.bAccessRestricted")
	itemThreeGenQ := itemQ.Union(childrenQ.QueryExpr()).Union(gChildrenQ.QueryExpr())

	query := dataStore.Raw(`
		SELECT items.ID, items.sType, items.bTransparentFolder,
			COALESCE(items.idItemUnlocked, '')<>'' as hasUnlockedItems,
			COALESCE(user_strings.sTitle, default_strings.sTitle) AS sTitle,
			users_items.iScore AS iScore, users_items.bValidated AS bValidated,
			users_items.bFinished AS bFinished, users_items.bKeyObtained AS bKeyObtained,
			users_items.nbSubmissionsAttempts AS nbSubmissionsAttempts,
			users_items.sStartDate AS sStartDate, users_items.sValidationDate AS sValidationDate,
			users_items.sFinishDate AS sFinishDate,
			items.iChildOrder AS iChildOrder,
			items.bAccessRestricted,
			items.idItemParent AS idItemParent,
			items.fullAccess, items.partialAccess, items.grayedAccess
		FROM ? items`, itemThreeGenQ.SubQuery()).
		Joins("LEFT JOIN users_items ON users_items.idItem=items.ID AND users_items.idUser=?", user.UserID).
		Order("idItemGrandparent, idItemParent, iChildOrder")
	query = dataStore.Items().JoinStrings(user, query)

	if err := query.Scan(&result).Error(); err != nil {
		return nil, err
	}
	return &result, nil
}
