package database

import "database/sql"

// computeAllAccess recomputes fields of groups_items.
//
// It starts from groups_items marked with propagate_access = 'self'.
//
// 1. cached_full_access_date, cached_partial_access_date, cached_manager_access,
// cached_access_solutions_date, cached_grayed_access_date, and cached_access_reason are updated.
//
// 2. cached_full_access, cached_partial_access, cached_access_solutions, cached_grayed_access
// are zeroed for rows where the calculation revealed access rights revocation.
//
// 3. Then the loop repeats from step 1 for all children (from items_items) of the processed group_items.
//
// Notes:
//  - Items having custom_chapter=1 are always skipped.
//  - Processed groups_items are marked with propagate_access = 'done'
//  - The function may loop endlessly if items_items is a cyclic graph
//
func (s *GroupItemStore) computeAllAccess() {
	s.mustBeInTransaction()

	var stmtCreateTemporaryTable, stmtDropTemporaryTable, stmtMarkDoNotPropagate,
		stmtMarkExistingChildren, stmtMarkFinishedItems, stmtUpdateGroupItems, stmtMarkChildrenItems *sql.Stmt
	var err error

	// We cannot JOIN groups_items_propagate directly in the INSERT query
	// because a trigger adds new rows into groups_items_propagate.
	const queryDropTemporaryTable = "DROP TEMPORARY TABLE IF EXISTS parents_propagate"
	stmtDropTemporaryTable, err = s.db.CommonDB().Prepare(queryDropTemporaryTable)
	mustNotBeError(err)
	defer func() { mustNotBeError(stmtDropTemporaryTable.Close()) }()

	const queryCreateTemporaryTable = `
		CREATE TEMPORARY TABLE parents_propagate
			SELECT id FROM groups_items_propagate WHERE propagate_access = 'children'`
	stmtCreateTemporaryTable, err = s.db.CommonDB().Prepare(queryCreateTemporaryTable)
	mustNotBeError(err)
	defer func() { mustNotBeError(stmtCreateTemporaryTable.Close()) }()

	// inserting missing children of groups_items into groups_items
	// for groups_items_propagate having propagate_access = 'children'
	const queryInsertMissingChildren = `
		INSERT IGNORE INTO groups_items (group_id, item_id, creator_user_id, cached_access_reason, access_reason)
		SELECT
			parents.group_id AS group_id,
			items_items.child_item_id AS item_id,
			parents.creator_user_id AS creator_user_id,
			NULL AS cached_access_reason,
			NULL AS access_reason
		FROM items_items
		JOIN groups_items AS parents
			ON parents.item_id = items_items.parent_item_id
		JOIN parents_propagate ON parents_propagate.id = parents.id`

	// mark as 'done' groups_items_propagate that shouldn't propagate (having items.custom_chapter=1)
	const queryMarkDoNotPropagate = `
		INSERT INTO groups_items_propagate (id, propagate_access)
		SELECT
			groups_items.id AS id,
			'done' as propagate_access
		FROM groups_items
		JOIN items
			ON groups_items.item_id = items.id AND items.custom_chapter
		ON DUPLICATE KEY UPDATE propagate_access='done'`
	stmtMarkDoNotPropagate, err = s.db.CommonDB().Prepare(queryMarkDoNotPropagate)
	mustNotBeError(err)
	defer func() { mustNotBeError(stmtMarkDoNotPropagate.Close()) }()

	// marking 'self' groups_items sons of groups_items in groups_items_propagate
	// whose parents are marked with groups_items_propagate.propagate_access='children'
	const queryMarkExistingChildren = `
		INSERT INTO groups_items_propagate (id, propagate_access)
		SELECT
			children.id AS id,
			'self' as propagate_access
		FROM items_items
		JOIN groups_items AS parents
			ON parents.item_id = items_items.parent_item_id
		JOIN groups_items AS children
			ON children.item_id = items_items.child_item_id AND children.group_id = parents.group_id
		JOIN groups_items_propagate AS parents_propagate
			ON parents_propagate.id = parents.id AND parents_propagate.propagate_access = 'children'
		ON DUPLICATE KEY UPDATE propagate_access='self'`
	stmtMarkExistingChildren, err = s.db.CommonDB().Prepare(queryMarkExistingChildren)
	mustNotBeError(err)
	defer func() { mustNotBeError(stmtMarkExistingChildren.Close()) }()

	// marking 'children' groups_items_propagate as 'done'
	const queryMarkFinishedItems = `
		UPDATE groups_items_propagate
		SET propagate_access = 'done'
		WHERE propagate_access = 'children'`
	stmtMarkFinishedItems, err = s.db.CommonDB().Prepare(queryMarkFinishedItems)
	mustNotBeError(err)
	defer func() { mustNotBeError(stmtMarkFinishedItems.Close()) }()

	// computation for groups_items marked as 'self' in groups_items_propagate.
	const queryUpdateGroupItems = `
		UPDATE groups_items
		LEFT JOIN (
			SELECT
				child.id,
				MIN(parent.cached_full_access_date) AS cached_full_access_date,
				MIN(IF(items_items.partial_access_propagation = 'AsPartial', parent.cached_partial_access_date, NULL)) AS cached_partial_access_date,
				MAX(parent.cached_manager_access) AS cached_manager_access,
				MIN(IF(items_items.partial_access_propagation = 'AsGrayed', parent.cached_partial_access_date, NULL))
					AS cached_grayed_access_date,
				MIN(parent.cached_access_solutions_date) AS cached_access_solutions_date,
				CONCAT('From ancestor group(s) ', GROUP_CONCAT(
					DISTINCT IF(parent.access_reason = '', NULL, parent.access_reason)
					ORDER BY parent_item.id
					SEPARATOR ', '
				)) AS access_reason_ancestors
			FROM groups_items AS child
			JOIN items_items
				ON items_items.child_item_id = child.item_id
			LEFT JOIN groups_items_propagate
				ON groups_items_propagate.id = child.id
			JOIN groups_items AS parent
				ON parent.item_id = items_items.parent_item_id AND parent.group_id = child.group_id
			JOIN items AS parent_item
				ON parent_item.id = items_items.parent_item_id
			WHERE
				(groups_items_propagate.propagate_access = 'self' OR groups_items_propagate.id IS NULL) AND
				(
					parent.cached_full_access_date IS NOT NULL OR
					(parent.cached_partial_access_date IS NOT NULL AND items_items.partial_access_propagation != 'None') OR
					parent.cached_access_solutions_date IS NOT NULL OR
					parent.cached_manager_access
				) AND
				parent_item.custom_chapter = 0
			GROUP BY child.id
		) AS new_data
			USING(id)
		JOIN groups_items_propagate USING(id)
		SET
			groups_items.cached_full_access_date = LEAST(
				IFNULL(new_data.cached_full_access_date, groups_items.full_access_date),
				IFNULL(groups_items.full_access_date, new_data.cached_full_access_date)
			),
			groups_items.cached_partial_access_date = LEAST(
				IFNULL(new_data.cached_partial_access_date, groups_items.partial_access_date),
				IFNULL(groups_items.partial_access_date, new_data.cached_partial_access_date)
			),
			groups_items.cached_manager_access = GREATEST(
				IFNULL(new_data.cached_manager_access, 0),
				groups_items.manager_access
			),
			groups_items.cached_access_solutions_date = LEAST(
				IFNULL(new_data.cached_access_solutions_date, groups_items.access_solutions_date),
				IFNULL(groups_items.access_solutions_date, new_data.cached_access_solutions_date)
			),
			groups_items.cached_grayed_access_date = new_data.cached_grayed_access_date,
			groups_items.cached_access_reason = new_data.access_reason_ancestors
		WHERE groups_items_propagate.propagate_access = 'self'`
	stmtUpdateGroupItems, err = s.db.CommonDB().Prepare(queryUpdateGroupItems)
	mustNotBeError(err)
	defer func() { mustNotBeError(stmtUpdateGroupItems.Close()) }()

	revokeCachedAccessStatements := s.prepareStatementsForRevokingCachedAccessWhereNeeded()
	defer func() {
		for _, statement := range revokeCachedAccessStatements {
			mustNotBeError(statement.Close())
		}
	}()

	// marking 'self' groups_items_propagate as 'children'
	const queryMarkChildrenItems = `
		UPDATE groups_items_propagate
		SET propagate_access = 'children'
		WHERE propagate_access = 'self'`
	stmtMarkChildrenItems, err = s.db.CommonDB().Prepare(queryMarkChildrenItems)
	mustNotBeError(err)
	defer func() { mustNotBeError(stmtMarkChildrenItems.Close()) }()

	hasChanges := true
	for hasChanges {
		_, err = stmtDropTemporaryTable.Exec()
		mustNotBeError(err)
		_, err = stmtCreateTemporaryTable.Exec()
		mustNotBeError(err)
		mustNotBeError(s.Exec(queryInsertMissingChildren).Error())
		_, err = stmtDropTemporaryTable.Exec()
		mustNotBeError(err)
		_, err = stmtMarkDoNotPropagate.Exec()
		mustNotBeError(err)
		_, err = stmtMarkExistingChildren.Exec()
		mustNotBeError(err)
		_, err = stmtMarkFinishedItems.Exec()
		mustNotBeError(err)
		_, err = stmtUpdateGroupItems.Exec()
		mustNotBeError(err)

		for _, statement := range revokeCachedAccessStatements {
			_, err = statement.Exec()
			mustNotBeError(err)
		}

		var result sql.Result
		result, err = stmtMarkChildrenItems.Exec()
		mustNotBeError(err)
		var rowsAffected int64
		rowsAffected, err = result.RowsAffected()
		mustNotBeError(err)
		hasChanges = rowsAffected > 0
	}
}

func (s *GroupItemStore) prepareStatementsForRevokingCachedAccessWhereNeeded() []*sql.Stmt {
	listFields := map[string]string{
		"cached_full_access":      "cached_full_access_date",
		"cached_partial_access":   "cached_partial_access_date",
		"cached_access_solutions": "cached_access_solutions_date",
		"cached_grayed_access":    "cached_grayed_access_date",
	}

	statements := make([]*sql.Stmt, 0, len(listFields))
	for accessField, accessDateField := range listFields {
		statement, err := s.db.CommonDB().Prepare(`
			UPDATE groups_items
			JOIN groups_items_propagate USING(id)
			SET ` + accessField + ` = false
			WHERE ` + accessField + ` = true AND
				groups_items_propagate.propagate_access = 'self' AND
				(` + accessDateField + ` IS NULL OR ` + accessDateField + ` > NOW())`) // #nosec
		mustNotBeError(err)
		statements = append(statements, statement)
	}
	return statements
}
