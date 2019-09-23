package items

import (
	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/service"
)

// rawNavigationItem represents one row of a navigation subtree returned from the DB
type rawNavigationItem struct {
	// items
	ID                int64
	Type              string
	TransparentFolder bool
	// whether items.item_unlocked_ids is empty
	HasUnlockedItems         bool
	PartialAccessPropagation string

	// title (from items_strings) in the user’s default language or (if not available) default language of the item
	Title string

	// from users_items for current user
	UserScore               float32        `sql:"column:score"`
	UserValidated           bool           `sql:"column:validated"`
	UserFinished            bool           `sql:"column:finished"`
	UserKeyObtained         bool           `sql:"column:key_obtained"`
	UserSubmissionsAttempts int32          `sql:"column:submissions_attempts"`
	UserStartDate           *database.Time `sql:"column:start_date"`
	UserValidationDate      *database.Time `sql:"column:validation_date"`
	UserFinishDate          *database.Time `sql:"column:finish_date"`

	// items_items
	ParentItemID int64
	Order        int32 `sql:"column:child_order"`

	*database.ItemAccessDetails

	ItemGrandparentID *int64
}

// getRawNavigationData reads a navigation subtree from the DB and returns an array of rawNavigationItem's
func getRawNavigationData(dataStore *database.DataStore, rootID int64, user *database.User) ([]rawNavigationItem, error) {
	var result []rawNavigationItem
	items := dataStore.Items()

	// This query can be simplified if we add a column for relation degrees into `items_ancestors`

	commonAttributes := "items.id, items.type, items.transparent_folder, items.unlocked_item_ids, items.default_language_id, " +
		"full_access, partial_access, grayed_access"
	itemQ := items.VisibleByID(user, rootID).Select(
		commonAttributes + ", NULL AS parent_item_id, NULL AS item_grandparent_id, NULL AS child_order, NULL AS partial_access_propagation")
	service.MustNotBeError(itemQ.Error())
	childrenQ := items.VisibleChildrenOfID(user, rootID).Select(
		commonAttributes + ",	parent_item_id, NULL AS item_grandparent_id, child_order, partial_access_propagation")
	service.MustNotBeError(childrenQ.Error())
	gChildrenQ := items.VisibleGrandChildrenOfID(user, rootID).Select(
		commonAttributes + ", ii1.parent_item_id, ii2.parent_item_id AS item_grandparent_id, ii1.child_order, ii1.partial_access_propagation")

	service.MustNotBeError(gChildrenQ.Error())
	itemThreeGenQ := itemQ.Union(childrenQ.QueryExpr()).Union(gChildrenQ.QueryExpr())
	service.MustNotBeError(itemThreeGenQ.Error())

	query := dataStore.Raw(`
		SELECT items.id, items.type, items.transparent_folder,
			COALESCE(items.unlocked_item_ids, '')<>'' as has_unlocked_items,
			COALESCE(user_strings.title, default_strings.title) AS title,
			users_items.score AS score, users_items.validated AS validated,
			users_items.finished AS finished, users_items.key_obtained AS key_obtained,
			users_items.submissions_attempts AS submissions_attempts,
			users_items.start_date AS start_date, users_items.validation_date AS validation_date,
			users_items.finish_date AS finish_date,
			items.child_order AS child_order,
			items.partial_access_propagation,
			items.parent_item_id AS parent_item_id,
			items.item_grandparent_id AS item_grandparent_id,
			items.full_access, items.partial_access, items.grayed_access
		FROM ? items`, itemThreeGenQ.SubQuery()).
		JoinsUserAndDefaultItemStrings(user).
		Joins("LEFT JOIN users_items ON users_items.item_id=items.id AND users_items.user_id=?", user.ID).
		Order("item_grandparent_id, parent_item_id, child_order")

	if err := query.Scan(&result).Error(); err != nil {
		return nil, err
	}
	return result, nil
}
