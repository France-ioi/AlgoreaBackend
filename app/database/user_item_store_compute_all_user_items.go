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

// ComputeAllUserItems recomputes fields of users_items
// For users_items marked with ancestors_computation_state = 'todo':
// 1. We mark all their ancestors in users_items as 'todo'
//  (we consider a row in users_items as an ancestor if it has the same value in user_id and
//  its item_id is an ancestor of the original row's item_id).
// 2. We process all objects that were marked as 'todo' and that have no children not marked as 'done'.
//  Then, if an object has children, we update
//    last_activity_date, tasks_tried, tasks_with_help, tasks_solved, children_validated, validated, validation_date.
//  This step is repeated until no records are updated.
// 3. We insert new groups_items for each processed row with key_obtained=1 according to corresponding items.item_unlocked_id.
func (s *UserItemStore) ComputeAllUserItems() (err error) {
	s.mustBeInTransaction()
	defer recoverPanics(&err)

	var groupsUnlocked int64

	// Use a lock so that we don't execute the listener multiple times in parallel
	mustNotBeError(s.WithNamedLock(computeAllUserItemsLockName, computeAllUserItemsLockTimeout, func(ds *DataStore) error {
		userItemStore := ds.UserItems()

		// We mark as 'todo' all ancestors of objects marked as 'todo'
		mustNotBeError(userItemStore.db.Exec(
			`UPDATE users_items AS ancestors
			JOIN items_ancestors ON (
				ancestors.item_id = items_ancestors.item_ancestor_id AND
				items_ancestors.item_ancestor_id != items_ancestors.item_child_id
			)
			JOIN users_items AS descendants ON (
				descendants.item_id = items_ancestors.item_child_id AND
				descendants.user_id = ancestors.user_id
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
					UPDATE users_items AS parent
					JOIN (
						SELECT *
						FROM (
							SELECT inner_parent.id
							FROM users_items AS inner_parent
							WHERE ancestors_computation_state = 'todo'
								AND NOT EXISTS (
									SELECT items_items.item_child_id
									FROM items_items
									JOIN users_items AS children
										ON children.item_id = items_items.item_child_id
									WHERE items_items.item_parent_id = inner_parent.item_id AND
										children.ancestors_computation_state <> 'done' AND
										children.user_id = inner_parent.user_id
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
			//  - last_activity_date as the max of children's
			//  - tasks_with_help, tasks_tried, nbTaskSolved as the sum of children's field
			//  - children_validated as the sum of children with validated == 1
			//  - validated, depending on the items_items.category and items.validation_type
			if updateStatement == nil {
				const updateQuery = `
					UPDATE users_items
					LEFT JOIN (
						SELECT
							MAX(children.last_activity_date) AS last_activity_date,
							SUM(children.tasks_tried) AS tasks_tried,
							SUM(children.tasks_with_help) AS tasks_with_help,
							SUM(children.tasks_solved) AS tasks_solved,
							SUM(validated) AS children_validated,
							children.user_id AS user_id,
							items_items.item_parent_id AS item_id
						FROM users_items AS children 
						JOIN items_items ON items_items.item_child_id = children.item_id
						GROUP BY children.user_id, items_items.item_parent_id
					) AS children_data
						USING(user_id, item_id)
					LEFT JOIN task_children_data_view AS task_children_data
						ON task_children_data.user_item_id = users_items.id
					JOIN items
						ON users_items.item_id = items.id
					LEFT JOIN items_items
						ON items_items.item_parent_id = users_items.item_id
					SET
						users_items.last_activity_date = IF(task_children_data.user_item_id IS NOT NULL AND
							children_data.user_id IS NOT NULL AND items_items.id IS NOT NULL,
							children_data.last_activity_date, users_items.last_activity_date),
						users_items.tasks_tried = IF(task_children_data.user_item_id IS NOT NULL AND
							children_data.user_id IS NOT NULL AND items_items.id IS NOT NULL,
							children_data.tasks_tried, users_items.tasks_tried),
						users_items.tasks_with_help = IF(task_children_data.user_item_id IS NOT NULL AND
							children_data.user_id IS NOT NULL AND items_items.id IS NOT NULL,
							children_data.tasks_with_help, users_items.tasks_with_help),
						users_items.tasks_solved = IF(task_children_data.user_item_id IS NOT NULL AND
							children_data.user_id IS NOT NULL AND items_items.id IS NOT NULL,
							children_data.tasks_solved, users_items.tasks_solved),
						users_items.children_validated = IF(task_children_data.user_item_id IS NOT NULL AND
							children_data.user_id IS NOT NULL AND items_items.id IS NOT NULL,
							children_data.children_validated, users_items.children_validated),
						users_items.validated = IF(task_children_data.user_item_id IS NOT NULL AND items_items.id IS NOT NULL,
							CASE
								WHEN users_items.validated = 1 THEN 1
								WHEN items.validation_type = 'Categories' THEN task_children_data.children_category = 0
								WHEN items.validation_type = 'All' THEN task_children_data.children_non_validated = 0
								WHEN items.validation_type = 'AllButOne' THEN task_children_data.children_non_validated < 2
								WHEN items.validation_type = 'One' THEN task_children_data.children_validated > 0
								ELSE 0
							END, users_items.validated),
						users_items.validation_date = IF(task_children_data.user_item_id IS NOT NULL AND items_items.id IS NOT NULL,
							IFNULL(
								users_items.validation_date,
								IF(items.validation_type = 'Categories',
									task_children_data.max_validation_date_categories, task_children_data.max_validation_date)
							), users_items.validation_date),
						users_items.ancestors_computation_state = 'done'
					WHERE users_items.ancestors_computation_state = 'processing'`
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
			users.group_self_id AS group_id,
			items.item_unlocked_id as items_ids
		FROM users_items
		JOIN items ON users_items.item_id = items.id
		JOIN users ON users_items.user_id = users.id
		WHERE users_items.ancestors_computation_state = 'processing' AND
			users_items.key_obtained AND items.item_unlocked_id IS NOT NULL`
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
					"items.id":               unlock.ItemID,
					"items.item_unlocked_id": unlock.ItemsIDs,
					"error":                  err,
				}).Warn("cannot parse items.item_unlocked_id")
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
			(group_id, item_id, partial_access_date, cached_partial_access_date, cached_partial_access, user_created_id)
		VALUES (?, ?, NOW(), NOW(), 1, -1)` // Note: user_created_id is incorrect here, but it is required
	values := make([]interface{}, 0, len(groupItemsToUnlock)*2)
	valuesTemplate := ", (?, ?, NOW(), NOW(), 1, -1)"
	for item := range groupItemsToUnlock {
		values = append(values, item.groupID, item.itemID)
	}

	query += strings.Repeat(valuesTemplate, len(groupItemsToUnlock)-1) +
		" ON DUPLICATE KEY UPDATE partial_access_date = NOW(), cached_partial_access_date = NOW(), cached_partial_access = 1"
	result := s.db.Exec(query, values...)
	mustNotBeError(result.Error)
	return result.RowsAffected
}
