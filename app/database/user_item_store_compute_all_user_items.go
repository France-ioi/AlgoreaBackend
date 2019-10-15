package database

import (
	"database/sql"
	"strconv"
	"strings"
	"time"

	"github.com/France-ioi/AlgoreaBackend/app/logging"
)

type groupItemPair struct {
	groupID int64
	itemID  int64
}

const computeAllUserItemsLockName = "listener_computeAllUserItems"
const computeAllUserItemsLockTimeout = 10 * time.Second

// ComputeAllUserItems recomputes fields of groups_attempts
// For groups_attempts marked with ancestors_computation_state = 'todo':
// 1. We mark all their ancestors in groups_attempts as 'todo'
//  (we consider a row in groups_attempts as an ancestor if it has the same value in group_id and
//  its item_id is an ancestor of the original row's item_id).
// 2. We process all objects that were marked as 'todo' and that have no children not marked as 'done'.
//  Then, if an object has children, we update
//    latest_activity_at, tasks_tried, tasks_with_help, tasks_solved, children_validated, validated, validated_at.
//  This step is repeated until no records are updated.
// 3. We insert new groups_items for each processed row with key_obtained=1 according to corresponding items.unlocked_item_ids.
func (s *UserItemStore) ComputeAllUserItems() (err error) {
	s.mustBeInTransaction()
	defer recoverPanics(&err)

	var groupsUnlocked int64

	// Use a lock so that we don't execute the listener multiple times in parallel
	mustNotBeError(s.WithNamedLock(computeAllUserItemsLockName, computeAllUserItemsLockTimeout, func(ds *DataStore) error {
		userItemStore := ds.UserItems()

		// We mark as 'todo' all ancestors of objects marked as 'todo'
		mustNotBeError(userItemStore.db.Exec(
			`UPDATE groups_attempts AS ancestors
			JOIN items_ancestors ON (
				ancestors.item_id = items_ancestors.ancestor_item_id AND
				items_ancestors.ancestor_item_id != items_ancestors.child_item_id
			)
			JOIN groups_attempts AS descendants ON (
				descendants.item_id = items_ancestors.child_item_id AND
				descendants.group_id = ancestors.group_id
			)
			SET ancestors.ancestors_computation_state = 'todo'
			WHERE descendants.ancestors_computation_state = 'todo'`).Error)

		hasChanges := true

		var markAsProcessingStatement, updateStatement *sql.Stmt
		groupItemsToUnlock := make(map[groupItemPair]bool)

		for hasChanges {
			// We mark as "processing" all objects that were marked as 'todo' and that have no children not marked as 'done'
			// This way we prevent infinite looping as we never process items that are ancestors of themselves
			if markAsProcessingStatement == nil {
				const markAsProcessingQuery = `
					UPDATE groups_attempts AS parent
					JOIN (
						SELECT *
						FROM (
							SELECT inner_parent.id
							FROM groups_attempts AS inner_parent
							WHERE ancestors_computation_state = 'todo'
								AND NOT EXISTS (
									SELECT items_items.child_item_id
									FROM items_items
									JOIN groups_attempts AS children
										ON children.item_id = items_items.child_item_id
									WHERE items_items.parent_item_id = inner_parent.item_id AND
										children.ancestors_computation_state <> 'done' AND
										children.group_id = inner_parent.group_id
								)
							) AS tmp2
					) AS tmp
						USING(id)
					SET ancestors_computation_state = 'processing'`

				markAsProcessingStatement, err = userItemStore.db.CommonDB().Prepare(markAsProcessingQuery)
				mustNotBeError(err)
				defer func() { mustNotBeError(markAsProcessingStatement.Close()) }()
			}
			_, err = markAsProcessingStatement.Exec()
			mustNotBeError(err)

			userItemStore.collectItemsToUnlock(groupItemsToUnlock)

			// For every object marked as 'processing', we compute all the characteristics based on the children:
			//  - latest_activity_at as the max of children's
			//  - tasks_with_help, tasks_tried, nbTaskSolved as the sum of children's per-item maximums
			//  - children_validated as the number of children items with validated == 1
			//  - validated, depending on the items_items.category and items.validation_type
			if updateStatement == nil {
				const updateQuery = `
					UPDATE groups_attempts AS target_groups_attempts
					LEFT JOIN LATERAL (
						SELECT
							target_groups_attempts.id,
							MAX(aggregated_children_attempts.latest_activity_at) AS latest_activity_at,
							IFNULL(SUM(aggregated_children_attempts.tasks_tried), 0) AS tasks_tried,
							IFNULL(SUM(aggregated_children_attempts.tasks_with_help), 0) AS tasks_with_help,
							IFNULL(SUM(aggregated_children_attempts.tasks_solved), 0) AS tasks_solved,
							IFNULL(SUM(aggregated_children_attempts.validated), 0) AS children_validated,
							SUM(IFNULL(NOT aggregated_children_attempts.validated, 1)) AS children_non_validated,
							SUM(items_items_with_scores.category = 'Validation' AND IFNULL(NOT aggregated_children_attempts.validated, 1))
								AS children_non_validated_categories,
							MAX(aggregated_children_attempts.validated_at) AS max_validated_at,
							MAX(IF(items_items_with_scores.category = 'Validation', aggregated_children_attempts.validated_at, NULL))
								AS max_validated_at_categories
						FROM items_items AS items_items_with_scores
						LEFT JOIN LATERAL (
							SELECT
								MAX(validated) AS validated,
								MIN(validated_at) AS validated_at,
								MAX(latest_activity_at) AS latest_activity_at,
								MAX(tasks_tried) AS tasks_tried,
								MAX(tasks_with_help) AS tasks_with_help,
								MAX(tasks_solved) AS tasks_solved
							FROM groups_attempts AS children_attempts
							WHERE children_attempts.group_id = target_groups_attempts.group_id AND
								children_attempts.item_id = items_items_with_scores.child_item_id
							GROUP BY children_attempts.group_id, children_attempts.item_id
						) AS aggregated_children_attempts ON 1
						JOIN items ON(
							items.id = items_items_with_scores.child_item_id
						)
						WHERE items_items_with_scores.parent_item_id = target_groups_attempts.item_id AND NOT items.no_score
						GROUP BY items_items_with_scores.parent_item_id
					) AS children_stats ON 1
					JOIN items
						ON target_groups_attempts.item_id = items.id
					SET
						target_groups_attempts.latest_activity_at = IF(children_stats.id IS NOT NULL,
							children_stats.latest_activity_at, target_groups_attempts.latest_activity_at),
						target_groups_attempts.tasks_tried = IF(children_stats.id IS NOT NULL,
							children_stats.tasks_tried, target_groups_attempts.tasks_tried),
						target_groups_attempts.tasks_with_help = IF(children_stats.id IS NOT NULL,
							children_stats.tasks_with_help, target_groups_attempts.tasks_with_help),
						target_groups_attempts.tasks_solved = IF(children_stats.id IS NOT NULL,
							children_stats.tasks_solved, target_groups_attempts.tasks_solved),
						target_groups_attempts.children_validated = IF(children_stats.id IS NOT NULL,
							children_stats.children_validated, target_groups_attempts.children_validated),
						target_groups_attempts.validated = IF(children_stats.id IS NOT NULL,
							CASE
								WHEN target_groups_attempts.validated = 1 THEN 1
								WHEN items.validation_type = 'Categories' THEN children_stats.children_non_validated_categories = 0
								WHEN items.validation_type = 'All' THEN children_stats.children_non_validated = 0
								WHEN items.validation_type = 'AllButOne' THEN children_stats.children_non_validated < 2
								WHEN items.validation_type = 'One' THEN children_stats.children_validated > 0
								ELSE 0
							END, target_groups_attempts.validated),
						target_groups_attempts.validated_at = IF(children_stats.id IS NOT NULL,
							IFNULL(
								target_groups_attempts.validated_at,
								IF(items.validation_type = 'Categories',
									children_stats.max_validated_at_categories, children_stats.max_validated_at)
							), target_groups_attempts.validated_at),
						target_groups_attempts.ancestors_computation_state = 'done'
					WHERE target_groups_attempts.ancestors_computation_state = 'processing'`
				updateStatement, err = userItemStore.db.CommonDB().Prepare(updateQuery)
				mustNotBeError(err)
				defer func() { mustNotBeError(updateStatement.Close()) }()
			}

			var result sql.Result
			result, err = updateStatement.Exec()
			mustNotBeError(err)
			var rowsAffected int64
			rowsAffected, err = result.RowsAffected()
			mustNotBeError(err)
			hasChanges = rowsAffected > 0
		}

		groupsUnlocked = userItemStore.unlockGroupItems(groupItemsToUnlock)
		return nil
	}))

	// If items have been unlocked, need to recompute access
	if groupsUnlocked > 0 {
		return s.GroupItems().After()
	}
	return nil
}

func (s *UserItemStore) collectItemsToUnlock(groupItemsToUnlock map[groupItemPair]bool) {
	// Unlock items depending on key_obtained
	const selectUnlocksQuery = `
		SELECT
			items.id AS item_id,
			groups.id AS group_id,
			items.unlocked_item_ids as items_ids
		FROM groups_attempts
		JOIN items ON groups_attempts.item_id = items.id
		JOIN ` + "`groups`" + ` ON groups_attempts.group_id = groups.id
		WHERE groups_attempts.ancestors_computation_state = 'processing' AND
			groups_attempts.key_obtained AND items.unlocked_item_ids IS NOT NULL`
	var err error
	var unlocksResult []struct {
		ItemID   int64
		GroupID  int64
		ItemsIDs string
	}
	mustNotBeError(s.Raw(selectUnlocksQuery).Scan(&unlocksResult).Error())
	for _, unlock := range unlocksResult {
		idsItems := strings.Split(unlock.ItemsIDs, ",")
		for _, itemID := range idsItems {
			var itemIDInt64 int64
			if itemIDInt64, err = strconv.ParseInt(itemID, 10, 64); err != nil {
				logging.SharedLogger.WithFields(map[string]interface{}{
					"items.id":                unlock.ItemID,
					"items.unlocked_item_ids": unlock.ItemsIDs,
					"error":                   err,
				}).Warn("cannot parse items.unlocked_item_ids")
			} else {
				groupItemsToUnlock[groupItemPair{groupID: unlock.GroupID, itemID: itemIDInt64}] = true
			}
		}
	}
}

func (s *UserItemStore) unlockGroupItems(groupItemsToUnlock map[groupItemPair]bool) int64 {
	if len(groupItemsToUnlock) == 0 {
		return 0
	}
	query := `
		INSERT INTO groups_items
			(group_id, item_id, partial_access_since, cached_partial_access_since, cached_partial_access, creator_user_id)
		VALUES (?, ?, NOW(), NOW(), 1, -1)` // Note: creator_user_id is incorrect here, but it is required
	values := make([]interface{}, 0, len(groupItemsToUnlock)*2)
	valuesTemplate := ", (?, ?, NOW(), NOW(), 1, -1)"
	for item := range groupItemsToUnlock {
		values = append(values, item.groupID, item.itemID)
	}

	query += strings.Repeat(valuesTemplate, len(groupItemsToUnlock)-1) +
		" ON DUPLICATE KEY UPDATE partial_access_since = NOW(), cached_partial_access_since = NOW(), cached_partial_access = 1"
	result := s.db.Exec(query, values...)
	mustNotBeError(result.Error)
	return result.RowsAffected
}
