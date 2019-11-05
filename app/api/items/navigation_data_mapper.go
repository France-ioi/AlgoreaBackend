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
	HasUnlockedItems       bool
	ContentViewPropagation string

	// title (from items_strings) in the user’s default language or (if not available) default language of the item
	Title *string

	// from groups_attempts for the active attempt of the current user
	UserAttemptID           *int64         `sql:"column:attempt_id"`
	UserScore               float32        `sql:"column:score"`
	UserValidated           bool           `sql:"column:validated"`
	UserFinished            bool           `sql:"column:finished"`
	UserKeyObtained         bool           `sql:"column:key_obtained"`
	UserSubmissionsAttempts int32          `sql:"column:submissions_attempts"`
	UserStartedAt           *database.Time `sql:"column:started_at"`
	UserValidatedAt         *database.Time `sql:"column:validated_at"`
	UserFinishedAt          *database.Time `sql:"column:finished_at"`

	// items_items
	ParentItemID int64
	Order        int32 `sql:"column:child_order"`

	CanViewGeneratedValue int

	ItemGrandparentID *int64
}

// getRawNavigationData reads a navigation subtree from the DB and returns an array of rawNavigationItem's
func getRawNavigationData(dataStore *database.DataStore, rootID int64, user *database.User) ([]rawNavigationItem, error) {
	var result []rawNavigationItem
	items := dataStore.Items()

	// This query can be simplified if we add a column for relation degrees into `items_ancestors`

	commonAttributes := "items.id, items.type, items.transparent_folder, items.unlocked_item_ids, items.default_language_id, " +
		"can_view_generated_value"
	itemQ := items.VisibleByID(user, rootID).Select(
		commonAttributes + ", NULL AS parent_item_id, NULL AS item_grandparent_id, NULL AS child_order, NULL AS content_view_propagation")
	service.MustNotBeError(itemQ.Error())
	childrenQ := items.VisibleChildrenOfID(user, rootID).Select(
		commonAttributes + ",	parent_item_id, NULL AS item_grandparent_id, child_order, content_view_propagation")
	service.MustNotBeError(childrenQ.Error())
	gChildrenQ := items.VisibleGrandChildrenOfID(user, rootID).Select(
		commonAttributes + ", ii1.parent_item_id, ii2.parent_item_id AS item_grandparent_id, ii1.child_order, ii1.content_view_propagation")

	service.MustNotBeError(gChildrenQ.Error())
	itemThreeGenQ := itemQ.Union(childrenQ.QueryExpr()).Union(gChildrenQ.QueryExpr())
	service.MustNotBeError(itemThreeGenQ.Error())

	query := dataStore.Raw(`
		SELECT items.id, items.type, items.transparent_folder,
			COALESCE(items.unlocked_item_ids, '')<>'' as has_unlocked_items,
			COALESCE(user_strings.title, default_strings.title) AS title,
			groups_attempts.id AS attempt_id,
			groups_attempts.score AS score, groups_attempts.validated AS validated,
			groups_attempts.finished AS finished, groups_attempts.key_obtained AS key_obtained,
			groups_attempts.submissions_attempts AS submissions_attempts,
			groups_attempts.started_at AS started_at, groups_attempts.validated_at AS validated_at,
			groups_attempts.finished_at AS finished_at,
			items.child_order AS child_order,
			items.content_view_propagation,
			items.parent_item_id AS parent_item_id,
			items.item_grandparent_id AS item_grandparent_id,
			items.can_view_generated_value
		FROM ? items`, itemThreeGenQ.SubQuery()).
		JoinsUserAndDefaultItemStrings(user).
		Joins("LEFT JOIN users_items ON users_items.item_id=items.id AND users_items.user_group_id=?", user.GroupID).
		Joins("LEFT JOIN groups_attempts ON groups_attempts.id=users_items.active_attempt_id").
		Order("item_grandparent_id, parent_item_id, child_order")

	if err := query.Scan(&result).Error(); err != nil {
		return nil, err
	}
	return result, nil
}
