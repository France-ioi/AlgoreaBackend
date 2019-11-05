package database

import "database/sql"

// computeAllAccess recomputes fields of permissions_generated.
//
// It starts from permissions_generated marked with propagate_access = 'self'.
//
// 1. can_view_generated, can_grant_view_generated, can_watch_generated, can_edit_generated, is_owner_generated are updated.
//
// 3. Then the loop repeats from step 1 for all children (from items_items) of the processed permissions_generated.
//
// Notes:
//  - The function may loop endlessly if items_items is a cyclic graph.
//
func (s *PermissionGrantedStore) computeAllAccess() {
	s.mustBeInTransaction()

	// ------------------------------------------------------------------------------------
	// Here we declare and prepare DB statements that will be used by the function later on
	// ------------------------------------------------------------------------------------
	var stmtCreateTemporaryTable, stmtDropTemporaryTable,
		stmtMarkChildrenOfChildrenAsSelf, stmtMarkProcessedChildrenAsDone, stmtUpdatePermissionsGenerated, stmtMarkSelfAsChildren *sql.Stmt
	var err error

	// We cannot JOIN permissions_generated directly in the INSERT query
	// because a trigger adds new rows into permissions_generated.
	const queryDropTemporaryTable = "DROP TEMPORARY TABLE IF EXISTS parents_propagate"
	stmtDropTemporaryTable, err = s.db.CommonDB().Prepare(queryDropTemporaryTable)
	mustNotBeError(err)
	defer func() { mustNotBeError(stmtDropTemporaryTable.Close()) }()

	const queryCreateTemporaryTable = `
		CREATE TEMPORARY TABLE parents_propagate
			SELECT group_id, item_id FROM permissions_generated WHERE propagate_access = 'children'`
	stmtCreateTemporaryTable, err = s.db.CommonDB().Prepare(queryCreateTemporaryTable)
	mustNotBeError(err)
	defer func() { mustNotBeError(stmtCreateTemporaryTable.Close()) }()

	// inserting missing children of permissions_generated into permissions_generated
	// for permissions_generated having propagate_access = 'children'
	const queryInsertMissingChildren = `
		INSERT IGNORE INTO permissions_generated (group_id, item_id)
		SELECT
			parents.group_id AS group_id,
			items_items.child_item_id AS item_id
		FROM items_items
		JOIN permissions_generated AS parents
			ON parents.item_id = items_items.parent_item_id
		JOIN parents_propagate ON parents_propagate.group_id = parents.group_id AND parents_propagate.item_id = parents.item_id`

	// marking permissions_generated whose parents are marked with propagate_access='children' as 'self'
	const queryMarkChildrenOfChildrenAsSelf = `
		INSERT INTO permissions_generated (group_id, item_id, propagate_access)
		SELECT
			children.group_id,
			children.item_id,
			'self' as propagate_access
		FROM items_items
		JOIN permissions_generated AS parents
			ON parents.item_id = items_items.parent_item_id
		JOIN permissions_generated AS children
			ON children.item_id = items_items.child_item_id AND children.group_id = parents.group_id
		WHERE parents.propagate_access = 'children'
		ON DUPLICATE KEY UPDATE propagate_access='self'`
	stmtMarkChildrenOfChildrenAsSelf, err = s.db.CommonDB().Prepare(queryMarkChildrenOfChildrenAsSelf)
	mustNotBeError(err)
	defer func() { mustNotBeError(stmtMarkChildrenOfChildrenAsSelf.Close()) }()

	// mark 'children' in permissions_generated as 'done'
	const queryMarkProcessedChildrenAsDone = `UPDATE permissions_generated SET propagate_access = 'done' WHERE propagate_access = 'children'`
	stmtMarkProcessedChildrenAsDone, err = s.db.CommonDB().Prepare(queryMarkProcessedChildrenAsDone)
	mustNotBeError(err)
	defer func() { mustNotBeError(stmtMarkProcessedChildrenAsDone.Close()) }()

	// computation for permissions_generated marked as 'self'
	const queryUpdatePermissionsGenerated = `
		UPDATE permissions_generated
		LEFT JOIN LATERAL (
			SELECT
				MAX(LEAST(
					IF(LEAST(items_items.descendants_and_solution_view_propagation_value+2, parent.can_view_generated_value) = 3,
						items_items.content_view_propagation_value,
						IF(parent.can_view_generated = 'info', 1, parent.can_view_generated_value)),
					items_items.descendants_and_solution_view_propagation_value+2)) AS can_view_value,
				MAX(IF(items_items.grant_view_propagation, LEAST(parent.can_grant_view_generated_value, 4), 1)) AS can_grant_view_value,
				MAX(IF(items_items.watch_propagation, LEAST(parent.can_watch_generated_value, 3), 1)) AS can_watch_value,
				MAX(IF(items_items.edit_propagation, LEAST(parent.can_edit_generated_value, 3), 1)) AS can_edit_value
			FROM items_items
			JOIN permissions_generated AS parent
				ON parent.item_id = items_items.parent_item_id 
			JOIN items AS parent_item
				ON parent_item.id = items_items.parent_item_id
			WHERE
				parent.group_id = permissions_generated.group_id AND items_items.child_item_id = permissions_generated.item_id
			GROUP BY permissions_generated.group_id, permissions_generated.item_id
		) AS new_data ON 1
		LEFT JOIN LATERAL (
			SELECT
				MAX(can_view_value) AS can_view_value,
				MAX(can_grant_view_value) AS can_grant_view_value,
				MAX(can_watch_value) AS can_watch_value,
				MAX(can_edit_value) AS can_edit_value,
				MAX(is_owner) AS is_owner
			FROM permissions_granted
			WHERE permissions_granted.group_id = permissions_generated.group_id AND
				permissions_granted.item_id = permissions_generated.item_id
			GROUP BY permissions_granted.group_id, permissions_granted.item_id
		) AS granted ON 1
		SET
			permissions_generated.can_view_generated = IF(granted.is_owner, 'solution', GREATEST(
				IFNULL(granted.can_view_value, 1),
				IFNULL(new_data.can_view_value, 1)
			)),
			permissions_generated.can_grant_view_generated = IF(granted.is_owner, 'transfer', GREATEST(
				IFNULL(granted.can_grant_view_value, 1),
				IFNULL(new_data.can_grant_view_value, 1)
			)),
			permissions_generated.can_watch_generated = IF(granted.is_owner, 'transfer', GREATEST(
				IFNULL(granted.can_watch_value, 1),
				IFNULL(new_data.can_watch_value, 1)
			)),
			permissions_generated.can_edit_generated = IF(granted.is_owner, 'transfer', GREATEST(
				IFNULL(granted.can_edit_value, 1),
				IFNULL(new_data.can_edit_value, 1)
			)),
			permissions_generated.is_owner_generated = IFNULL(granted.is_owner, 0)
		WHERE permissions_generated.propagate_access = 'self'`
	stmtUpdatePermissionsGenerated, err = s.db.CommonDB().Prepare(queryUpdatePermissionsGenerated)
	mustNotBeError(err)
	defer func() { mustNotBeError(stmtUpdatePermissionsGenerated.Close()) }()

	// marking 'self' permissions_generated as 'children'
	const queryMarkSelfAsChildren = `
		UPDATE permissions_generated
		SET propagate_access = 'children'
		WHERE propagate_access = 'self'`
	stmtMarkSelfAsChildren, err = s.db.CommonDB().Prepare(queryMarkSelfAsChildren)
	mustNotBeError(err)
	defer func() { mustNotBeError(stmtMarkSelfAsChildren.Close()) }()

	// ------------------------------------------------------------------------------------
	// Here we execute the statements
	// ------------------------------------------------------------------------------------
	_, err = stmtDropTemporaryTable.Exec()
	mustNotBeError(err)

	hasChanges := true
	for hasChanges {
		_, err = stmtCreateTemporaryTable.Exec()
		mustNotBeError(err)
		mustNotBeError(s.Exec(queryInsertMissingChildren).Error())
		_, err = stmtDropTemporaryTable.Exec()
		mustNotBeError(err)
		_, err = stmtMarkChildrenOfChildrenAsSelf.Exec()
		mustNotBeError(err)
		_, err = stmtMarkProcessedChildrenAsDone.Exec()
		mustNotBeError(err)
		_, err = stmtUpdatePermissionsGenerated.Exec()
		mustNotBeError(err)

		var result sql.Result
		result, err = stmtMarkSelfAsChildren.Exec()
		mustNotBeError(err)
		var rowsAffected int64
		rowsAffected, err = result.RowsAffected()
		mustNotBeError(err)
		hasChanges = rowsAffected > 0
	}
}
