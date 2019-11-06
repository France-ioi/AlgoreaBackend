package database

import "database/sql"

// computeAllAccess recomputes fields of permissions_generated.
//
// It starts from group-item pairs marked with propagate_access = 'self' in `permissions_propagate`.
//
// 1. can_view_generated, can_grant_view_generated, can_watch_generated, can_edit_generated, is_owner_generated are updated.
//
// 3. Then the loop repeats from step 1 for all children (from items_items) of the processed permissions_generated.
//
// Notes:
//  - The function may loop endlessly if items_items is a cyclic graph.
//  - Processed group-item pairs are removed from permissions_propagate.
//
func (s *PermissionGrantedStore) computeAllAccess() {
	s.mustBeInTransaction()

	// ------------------------------------------------------------------------------------
	// Here we declare and prepare DB statements that will be used by the function later on
	// ------------------------------------------------------------------------------------
	var stmtMarkChildrenOfChildrenAsSelf, stmtDeleteProcessedChildren, stmtUpdatePermissionsGenerated, stmtMarkSelfAsChildren *sql.Stmt
	var err error

	// marking group-item pairs whose parents are marked with propagate_access='children' as 'self'
	const queryMarkChildrenOfChildrenAsSelf = `
		INSERT INTO permissions_propagate (group_id, item_id, propagate_access)
		SELECT
			parents.group_id,
			items_items.child_item_id,
			'self' as propagate_access
		FROM items_items
		JOIN permissions_generated AS parents
			ON parents.item_id = items_items.parent_item_id
		JOIN permissions_propagate AS parents_propagate
			ON parents_propagate.group_id = parents.group_id AND parents_propagate.item_id = parents.item_id AND
				 parents_propagate.propagate_access = 'children'
		ON DUPLICATE KEY UPDATE propagate_access='self'`
	stmtMarkChildrenOfChildrenAsSelf, err = s.db.CommonDB().Prepare(queryMarkChildrenOfChildrenAsSelf)
	mustNotBeError(err)
	defer func() { mustNotBeError(stmtMarkChildrenOfChildrenAsSelf.Close()) }()

	// deleting 'children' groups_items_propagate
	const queryDeleteProcessedChildren = `DELETE FROM permissions_propagate WHERE propagate_access = 'children'`
	stmtDeleteProcessedChildren, err = s.db.CommonDB().Prepare(queryDeleteProcessedChildren)
	mustNotBeError(err)
	defer func() { mustNotBeError(stmtDeleteProcessedChildren.Close()) }()

	// computation for group-item pairs marked as 'self' in permissions_propagate (so normally all of them)
	const queryUpdatePermissionsGenerated = `
		INSERT INTO permissions_generated
			(group_id, item_id, can_view_generated, can_grant_view_generated, can_watch_generated, can_edit_generated, is_owner_generated)
		SELECT
			permissions_propagate.group_id,
			permissions_propagate.item_id,
			IF(granted.is_owner, 'solution', GREATEST(
				IFNULL(granted.can_view_value, 1),
				IFNULL(new_data.can_view_value, 1)
			)) AS can_view_generated,
			IF(granted.is_owner, 'transfer', GREATEST(
				IFNULL(granted.can_grant_view_value, 1),
				IFNULL(new_data.can_grant_view_value, 1)
			)) AS can_grant_view_generated,
			IF(granted.is_owner, 'transfer', GREATEST(
				IFNULL(granted.can_watch_value, 1),
				IFNULL(new_data.can_watch_value, 1)
			)) AS can_watch_generated,
			IF(granted.is_owner, 'transfer', GREATEST(
				IFNULL(granted.can_edit_value, 1),
				IFNULL(new_data.can_edit_value, 1)
			)) AS can_edit_generated,
			IFNULL(granted.is_owner, 0) AS is_owner_generated
		FROM permissions_propagate
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
				parent.group_id = permissions_propagate.group_id AND items_items.child_item_id = permissions_propagate.item_id
			GROUP BY permissions_propagate.group_id, permissions_propagate.item_id
		) AS new_data ON 1
		LEFT JOIN LATERAL (
			SELECT
				MAX(can_view_value) AS can_view_value,
				MAX(can_grant_view_value) AS can_grant_view_value,
				MAX(can_watch_value) AS can_watch_value,
				MAX(can_edit_value) AS can_edit_value,
				MAX(is_owner) AS is_owner
			FROM permissions_granted
			WHERE permissions_granted.group_id = permissions_propagate.group_id AND
				permissions_granted.item_id = permissions_propagate.item_id
			GROUP BY permissions_granted.group_id, permissions_granted.item_id
		) AS granted ON 1
		WHERE propagate_access = 'self'
		ON DUPLICATE KEY UPDATE
			can_view_generated = VALUES(can_view_generated),
			can_grant_view_generated = VALUES(can_grant_view_generated),
			can_watch_generated = VALUES(can_watch_generated),
			can_edit_generated = VALUES(can_edit_generated),
			is_owner_generated = VALUES(is_owner_generated)`
	stmtUpdatePermissionsGenerated, err = s.db.CommonDB().Prepare(queryUpdatePermissionsGenerated)
	mustNotBeError(err)
	defer func() { mustNotBeError(stmtUpdatePermissionsGenerated.Close()) }()

	// marking 'self' permissions_propagate as 'children'
	const queryMarkSelfAsChildren = `
		UPDATE permissions_propagate
		SET propagate_access = 'children'
		WHERE propagate_access = 'self'`
	stmtMarkSelfAsChildren, err = s.db.CommonDB().Prepare(queryMarkSelfAsChildren)
	mustNotBeError(err)
	defer func() { mustNotBeError(stmtMarkSelfAsChildren.Close()) }()

	// ------------------------------------------------------------------------------------
	// Here we execute the statements
	// ------------------------------------------------------------------------------------
	hasChanges := true
	for hasChanges {
		_, err = stmtMarkChildrenOfChildrenAsSelf.Exec()
		mustNotBeError(err)
		_, err = stmtDeleteProcessedChildren.Exec()
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
