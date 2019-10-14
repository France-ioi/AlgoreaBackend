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
			//  - tasks_with_help, tasks_tried, nbTaskSolved as the sum of children's field
			//  - children_validated as the sum of children with validated == 1
			//  - validated, depending on the items_items.category and items.validation_type
			if updateStatement == nil {
				const updateQuery = `
					WITH task_children_data_view AS (
						WITH best_scores AS (
							SELECT group_id, item_id, MAX(score) AS score, MAX(validated) AS validated,
								MIN(validated_at) AS validated_at
							FROM groups_attempts
							GROUP BY group_id, item_id
						)
						SELECT
							parent_groups_attempts.id,
							SUM(IF(task_children.group_id IS NOT NULL AND task_children.validated, 1, 0)) AS children_validated,
							SUM(IF(task_children.group_id IS NOT NULL AND task_children.validated, 0, 1)) AS children_non_validated,
							SUM(IF(items_items.category = 'Validation' AND
								(ISNULL(task_children.group_id) OR task_children.validated != 1), 1, 0)) AS children_category,
							MAX(task_children.validated_at) AS max_validated_at,
							MAX(IF(items_items.category = 'Validation', task_children.validated_at, NULL)) AS max_validated_at_categories
						FROM groups_attempts AS parent_groups_attempts
						JOIN items_items ON(
							parent_groups_attempts.item_id = items_items.parent_item_id
						)
						LEFT JOIN best_scores AS task_children ON(
							items_items.child_item_id = task_children.item_id AND
							task_children.group_id = parent_groups_attempts.group_id
						)
						JOIN items ON(
							items.ID = items_items.child_item_id
						)
						WHERE items.type <> 'Course' AND items.no_score = 0
						GROUP BY parent_groups_attempts.id
					)
					UPDATE groups_attempts
					LEFT JOIN (
						SELECT
							MAX(children.latest_activity_at) AS latest_activity_at,
							SUM(children.tasks_tried) AS tasks_tried,
							SUM(children.tasks_with_help) AS tasks_with_help,
							SUM(children.tasks_solved) AS tasks_solved,
							SUM(validated) AS children_validated,
							children.group_id AS group_id,
							items_items.parent_item_id AS item_id
						FROM groups_attempts AS children 
						JOIN items_items ON items_items.child_item_id = children.item_id
						GROUP BY children.group_id, items_items.parent_item_id
					) AS children_data
						USING(group_id, item_id)
					LEFT JOIN task_children_data_view AS task_children_data
						ON task_children_data.id = groups_attempts.id
					JOIN items
						ON groups_attempts.item_id = items.id
					LEFT JOIN items_items
						ON items_items.parent_item_id = groups_attempts.item_id
					SET
						groups_attempts.latest_activity_at = IF(task_children_data.id IS NOT NULL AND
							children_data.group_id IS NOT NULL AND items_items.id IS NOT NULL,
							children_data.latest_activity_at, groups_attempts.latest_activity_at),
						groups_attempts.tasks_tried = IF(task_children_data.id IS NOT NULL AND
							children_data.group_id IS NOT NULL AND items_items.id IS NOT NULL,
							children_data.tasks_tried, groups_attempts.tasks_tried),
						groups_attempts.tasks_with_help = IF(task_children_data.id IS NOT NULL AND
							children_data.group_id IS NOT NULL AND items_items.id IS NOT NULL,
							children_data.tasks_with_help, groups_attempts.tasks_with_help),
						groups_attempts.tasks_solved = IF(task_children_data.id IS NOT NULL AND
							children_data.group_id IS NOT NULL AND items_items.id IS NOT NULL,
							children_data.tasks_solved, groups_attempts.tasks_solved),
						groups_attempts.children_validated = IF(task_children_data.id IS NOT NULL AND
							children_data.group_id IS NOT NULL AND items_items.id IS NOT NULL,
							children_data.children_validated, groups_attempts.children_validated),
						groups_attempts.validated = IF(task_children_data.id IS NOT NULL AND items_items.id IS NOT NULL,
							CASE
								WHEN groups_attempts.validated = 1 THEN 1
								WHEN items.validation_type = 'Categories' THEN task_children_data.children_category = 0
								WHEN items.validation_type = 'All' THEN task_children_data.children_non_validated = 0
								WHEN items.validation_type = 'AllButOne' THEN task_children_data.children_non_validated < 2
								WHEN items.validation_type = 'One' THEN task_children_data.children_validated > 0
								ELSE 0
							END, groups_attempts.validated),
						groups_attempts.validated_at = IF(task_children_data.id IS NOT NULL AND items_items.id IS NOT NULL,
							IFNULL(
								groups_attempts.validated_at,
								IF(items.validation_type = 'Categories',
									task_children_data.max_validated_at_categories, task_children_data.max_validated_at)
							), groups_attempts.validated_at),
						groups_attempts.ancestors_computation_state = 'done'
					WHERE groups_attempts.ancestors_computation_state = 'processing'`
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
