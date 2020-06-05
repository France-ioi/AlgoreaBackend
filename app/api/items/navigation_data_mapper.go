package items

import (
	"github.com/jinzhu/gorm"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/service"
)

// rawNavigationItem represents one row of a navigation subtree returned from the DB
type rawNavigationItem struct {
	// items
	ID                    int64
	Type                  string
	RequiresExplicitEntry bool
	EntryParticipantType  string
	NoScore               bool

	// title (from items_strings) in the user’s default language or (if not available) default language of the item
	Title       *string
	LanguageTag *string

	// results
	AttemptID        *int64
	ScoreComputed    float32
	Validated        bool
	StartedAt        *database.Time
	LatestActivityAt *database.Time
	EndedAt          *database.Time

	// attempts
	AttemptAllowsSubmissionsUntil database.Time

	// max from results of the current participant
	BestScore float32

	// items_items
	ParentItemID       int64
	HasVisibleChildren bool

	CanViewGeneratedValue      int
	CanGrantViewGeneratedValue int
	CanWatchGeneratedValue     int
	CanEditGeneratedValue      int
	IsOwnerGenerated           bool

	CanWatchForGroupResults  bool
	WatchedGroupCanView      int
	WatchedGroupAvgScore     float32
	WatchedGroupAllValidated bool
}

// getRawNavigationData reads a navigation subtree from the DB and returns an array of rawNavigationItem's
func getRawNavigationData(dataStore *database.DataStore, rootID, groupID, attemptID int64,
	user *database.User, watchedGroupID int64, watchedGroupIDSet bool) []rawNavigationItem {
	var result []rawNavigationItem
	items := dataStore.Items()

	// This query can be simplified if we add a column for relation degrees into `items_ancestors`

	commonAttributes := `
		items.id, items.type, items.default_language_tag,
		can_view_generated_value, can_grant_view_generated_value, can_watch_generated_value, can_edit_generated_value, is_owner_generated`
	itemsQuery := items.VisibleByID(groupID, rootID).Select(
		commonAttributes+`, results.attempt_id, 0 AS requires_explicit_entry, NULL AS parent_item_id, NULL AS child_order,
			NULL AS score_computed, NULL AS validated, NULL AS started_at, NULL AS latest_activity_at,
			NULL AS allows_submissions_until, NULL AS ended_at, NULL AS entry_participant_type,
			0 AS no_score, 0 AS has_visible_children,
			NULL AS watched_group_can_view, 0 AS can_watch_for_group_results, 0 AS watched_group_avg_score, 0 AS watched_group_all_validated`).
		Joins(`
			JOIN results ON results.participant_id = ? AND results.attempt_id = ? AND
				results.item_id = items.id AND results.started_at IS NOT NULL`, groupID, attemptID)
	service.MustNotBeError(itemsQuery.Error())
	watchedGroupCanViewQuery := interface{}(gorm.Expr("NULL"))
	watchedGroupParticipantsQuery := interface{}(gorm.Expr("(SELECT NULL AS id)"))
	watchedGroupAvgScoreQuery := interface{}(gorm.Expr("(SELECT NULL AS avg_score, NULL AS all_validated)"))
	if watchedGroupIDSet {
		watchedGroupCanViewQuery = dataStore.Permissions().
			Joins("JOIN groups_ancestors_active ON groups_ancestors_active.ancestor_group_id = permissions.group_id").
			Where("groups_ancestors_active.child_group_id = ?", watchedGroupID).
			Where("permissions.item_id = items.id").
			Select("IFNULL(MAX(IFNULL(can_view_generated_value, 1)), 1)").SubQuery()
		watchedGroupParticipantsQuery = dataStore.ActiveGroupAncestors().
			Select("participant.id").
			Joins(`
				JOIN `+"`groups`"+` AS participant
					ON participant.id = groups_ancestors_active.child_group_id AND participant.type IN ('User', 'Team')`).
			Where("groups_ancestors_active.ancestor_group_id = ?", watchedGroupID).SubQuery()
		watchedGroupAvgScoreQuery = dataStore.Raw(
			"SELECT IFNULL(AVG(score), 0) AS avg_score, COUNT(*) > 0 AND COUNT(*) = SUM(validated) AS all_validated FROM ? AS stats",
			dataStore.Table("watched_group_participants").
				Joins(`
					LEFT JOIN (
						SELECT participant_id, score_computed, validated FROM results
						WHERE results.item_id = items.id
					) AS results ON results.participant_id = watched_group_participants.id`).
				Select("MAX(IFNULL(results.score_computed, 0)) AS score, MAX(IFNULL(results.validated, 0)) AS validated").
				Group("watched_group_participants.id").SubQuery()).SubQuery()
	}

	hasVisibleChildrenQuery := dataStore.Permissions().VisibleToGroup(groupID).
		Joins("JOIN items_items ON items_items.child_item_id = permissions.item_id").
		Where("items_items.parent_item_id = items.id").
		Select("1").Limit(1).SubQuery()

	canWatchResultEnumIndex := dataStore.PermissionsGranted().WatchIndexByName("result")
	childrenWithoutResultsQuery := dataStore.Raw("WITH watched_group_participants AS ? ?",
		watchedGroupParticipantsQuery,
		items.VisibleChildrenOfID(groupID, rootID).Select(
			commonAttributes+`,	items.requires_explicit_entry, parent_item_id, child_order,
					items.entry_participant_type, items.no_score,
					IFNULL(?, 0) AS has_visible_children, ? AS watched_group_can_view,
					can_watch_generated_value >= ? AS can_watch_for_group_results,
					IF(can_watch_generated_value >= ?, watched_group_stats.avg_score, 0) AS watched_group_avg_score,
					IF(can_watch_generated_value >= ?, watched_group_stats.all_validated, 0) AS watched_group_all_validated`,
			hasVisibleChildrenQuery, watchedGroupCanViewQuery, canWatchResultEnumIndex, canWatchResultEnumIndex, canWatchResultEnumIndex).
			Joins("JOIN LATERAL ? AS watched_group_stats", watchedGroupAvgScoreQuery).
			SubQuery()).SubQuery()
	// nolint:gosec
	childrenQuery :=
		dataStore.Raw(`
			SELECT `+commonAttributes+`, results.attempt_id, requires_explicit_entry, parent_item_id, child_order,
				results.score_computed, results.validated, results.started_at, results.latest_activity_at,
				attempts.allows_submissions_until, attempts.ended_at, items.entry_participant_type, items.no_score,
				has_visible_children, watched_group_can_view, can_watch_for_group_results, watched_group_avg_score, watched_group_all_validated
			FROM ? AS items
			LEFT JOIN results ON results.participant_id = ? AND results.item_id = items.id
			LEFT JOIN attempts ON attempts.participant_id = results.participant_id AND attempts.id = results.attempt_id
			WHERE attempts.id IS NULL OR IF(attempts.root_item_id <=> results.item_id, attempts.parent_attempt_id, attempts.id) = ?`,
			childrenWithoutResultsQuery, groupID, attemptID)
	service.MustNotBeError(childrenQuery.Error())

	allItemsQuery := itemsQuery.UnionAll(childrenQuery.SubQuery())
	service.MustNotBeError(allItemsQuery.Error())

	query := dataStore.Raw(`
		SELECT items.id, items.type, items.requires_explicit_entry, items.entry_participant_type, items.no_score,
			COALESCE(user_strings.title, default_strings.title) AS title,
			IF(user_strings.title IS NOT NULL, user_strings.language_tag, default_strings.language_tag) AS language_tag,
			IF(items.parent_item_id IS NOT NULL,
				IFNULL(
					(SELECT MAX(results.score_computed) AS best_score
					FROM results
					WHERE results.item_id = items.id AND results.participant_id = ?), 0),
				NULL) AS best_score,
			items.can_grant_view_generated_value,
			items.can_watch_generated_value, items.can_edit_generated_value, items.is_owner_generated,
			items.parent_item_id AS parent_item_id,
			items.can_view_generated_value,
			items.attempt_id,
			items.score_computed, items.validated, items.started_at, items.latest_activity_at,
			items.allows_submissions_until AS attempt_allows_submissions_until,
			items.ended_at,
			items.has_visible_children,
			items.can_watch_for_group_results,
			items.watched_group_can_view, items.watched_group_avg_score, items.watched_group_all_validated
		FROM ? items`, groupID, allItemsQuery.SubQuery()).
		JoinsUserAndDefaultItemStrings(user).
		Order("parent_item_id, child_order, attempt_id")

	service.MustNotBeError(query.Scan(&result).Error())
	return result
}
