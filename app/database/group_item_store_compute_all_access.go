package database

import "database/sql"

// computeAllAccess recomputes fields of groups_items.
//
// It starts from groups_items marked with sPropagateAccess = 'self'.
//
// 1. sCachedFullAccessDate, sCachedPartialAccessDate, bCachedManagerAccess,
// sCachedAccessSolutionsDate, sCachedGrayedAccessDate, and sCachedAccessReason are updated.
//
// 2. bCachedFullAccess, bCachedPartialAccess, bCachedAccessSolutions, bCachedGrayedAccess
// are zeroed for rows where the calculation revealed access rights revocation.
//
// 3. Then the loop repeats from step 1 for all children (from items_items) of the processed group_items.
//
// Notes:
//  - Items having bCustomChapter=1 are always skipped.
//  - Processed groups_items are marked with sPropagateAccess = 'done'
//  - The function may loop endlessly if items_items is a cyclic graph
//
func (s *GroupItemStore) computeAllAccess() {
	s.mustBeInTransaction()

	var stmtInsertMissingPropagate, stmtUpdatePropagateAccess, stmtInsertMissingChildren, stmtMarkDoNotPropagate,
		stmtMarkExistingChildren, stmtMarkFinishedItems, stmtUpdateGroupItems, stmtMarkChildrenItems *sql.Stmt
	var err error

	// inserting missing children of groups_items into groups_items
	// for groups_items_propagate having sPropagateAccess = 'children'
	const queryInsertMissingChildren = `
		INSERT IGNORE INTO groups_items (idGroup, idItem, idUserCreated, sCachedAccessReason, sAccessReason)
		SELECT
			parents.idGroup AS idGroup,
			items_items.idItemChild AS idItem,
			parents.idUserCreated AS idUserCreated,
			NULL AS sCachedAccessReason,
			NULL AS sAccessReason
		FROM items_items
		JOIN groups_items AS parents
			ON parents.idItem = items_items.idItemParent
		JOIN groups_items_propagate AS parents_propagate
			ON parents.ID = parents_propagate.ID AND parents_propagate.sPropagateAccess = 'children'`
	stmtInsertMissingChildren, err = s.db.CommonDB().Prepare(queryInsertMissingChildren)
	mustNotBeError(err)
	defer func() { mustNotBeError(stmtInsertMissingChildren.Close()) }()

	// inserting missing (or set sPropagateAccess='self' to existing) groups_items_propagate
	// for groups_items having sPropagateAccess='self'
	const queryInsertMissingPropagate = `
		INSERT INTO groups_items_propagate (ID, sPropagateAccess)
		SELECT
			groups_items.ID,
			'self' as sPropagateAccess
		FROM groups_items
		WHERE sPropagateAccess='self'
		ON DUPLICATE KEY UPDATE sPropagateAccess='self'`
	stmtInsertMissingPropagate, err = s.db.CommonDB().Prepare(queryInsertMissingPropagate)
	mustNotBeError(err)
	defer func() { mustNotBeError(stmtInsertMissingPropagate.Close()) }()

	// Set groups_items as set up for propagation
	// (switch groups_items.sPropagateAccess from 'self' to 'done')
	const queryUpdatePropagateAccess = `
		UPDATE groups_items
		SET sPropagateAccess='done'
		WHERE sPropagateAccess='self'`
	stmtUpdatePropagateAccess, err = s.db.CommonDB().Prepare(queryUpdatePropagateAccess)
	mustNotBeError(err)
	defer func() { mustNotBeError(stmtUpdatePropagateAccess.Close()) }()

	// mark as 'done' groups_items_propagate that shouldn't propagate (having items.bCustomChapter=1)
	const queryMarkDoNotPropagate = `
		INSERT INTO groups_items_propagate (ID, sPropagateAccess)
		SELECT
			groups_items.ID AS ID,
			'done' as sPropagateAccess
		FROM groups_items
		JOIN items
			ON groups_items.idItem = items.ID AND items.bCustomChapter
		ON DUPLICATE KEY UPDATE sPropagateAccess='done'`
	stmtMarkDoNotPropagate, err = s.db.CommonDB().Prepare(queryMarkDoNotPropagate)
	mustNotBeError(err)
	defer func() { mustNotBeError(stmtMarkDoNotPropagate.Close()) }()

	// marking 'self' groups_items sons of groups_items in groups_items_propagate
	// whose parents are marked with groups_items_propagate.sPropagateAccess='children'
	const queryMarkExistingChildren = `
		INSERT INTO groups_items_propagate (ID, sPropagateAccess)
		SELECT
			children.ID AS ID,
			'self' as sPropagateAccess
		FROM items_items
		JOIN groups_items AS parents
			ON parents.idItem = items_items.idItemParent
		JOIN groups_items AS children
			ON children.idItem = items_items.idItemChild AND children.idGroup = parents.idGroup
		JOIN groups_items_propagate AS parents_propagate
			ON parents_propagate.ID = parents.ID AND parents_propagate.sPropagateAccess = 'children'
		ON DUPLICATE KEY UPDATE sPropagateAccess='self'`
	stmtMarkExistingChildren, err = s.db.CommonDB().Prepare(queryMarkExistingChildren)
	mustNotBeError(err)
	defer func() { mustNotBeError(stmtMarkExistingChildren.Close()) }()

	// marking 'children' groups_items_propagate as 'done'
	const queryMarkFinishedItems = `
		UPDATE groups_items_propagate
		SET sPropagateAccess = 'done'
		WHERE sPropagateAccess = 'children'`
	stmtMarkFinishedItems, err = s.db.CommonDB().Prepare(queryMarkFinishedItems)
	mustNotBeError(err)
	defer func() { mustNotBeError(stmtMarkFinishedItems.Close()) }()

	// computation for groups_items marked as 'self' in groups_items_propagate.
	const queryUpdateGroupItems = `
		UPDATE groups_items
		LEFT JOIN (
			SELECT
				child.ID,
				MIN(parent.sCachedFullAccessDate) AS sCachedFullAccessDate,
				MIN(IF(items_items.bAccessRestricted = 0, parent.sCachedPartialAccessDate, NULL)) AS sCachedPartialAccessDate,
				MAX(parent.bCachedManagerAccess) AS bCachedManagerAccess,
				MIN(IF(items_items.bAccessRestricted AND items_items.bAlwaysVisible, parent.sCachedPartialAccessDate, NULL)) AS sCachedGrayedAccessDate,
				MIN(parent.sCachedAccessSolutionsDate) AS sCachedAccessSolutionsDate,
				CONCAT('From ancestor group(s) ', GROUP_CONCAT(
					DISTINCT IF(parent.sAccessReason = '', NULL, parent.sAccessReason)
					ORDER BY parent_item.ID
					SEPARATOR ', '
				)) AS sAccessReasonAncestors
			FROM groups_items AS child
			JOIN items_items
				ON items_items.idItemChild = child.idItem
			LEFT JOIN groups_items_propagate
				ON groups_items_propagate.ID = child.ID
			JOIN groups_items AS parent
				ON parent.idItem = items_items.idItemParent AND parent.idGroup = child.idGroup
			JOIN items AS parent_item
				ON parent_item.ID = items_items.idItemParent
			WHERE
				(groups_items_propagate.sPropagateAccess = 'self' OR groups_items_propagate.ID IS NULL) AND
				(
					parent.sCachedFullAccessDate IS NOT NULL OR
					(parent.sCachedPartialAccessDate IS NOT NULL AND (items_items.bAccessRestricted = 0 OR items_items.bAlwaysVisible)) OR
					parent.sCachedAccessSolutionsDate IS NOT NULL OR
					parent.bCachedManagerAccess
				) AND
				parent_item.bCustomChapter = 0
			GROUP BY child.ID
		) AS new_data
			USING(ID)
		JOIN groups_items_propagate USING(ID)
		SET
			groups_items.sCachedFullAccessDate = LEAST(
				IFNULL(new_data.sCachedFullAccessDate, groups_items.sFullAccessDate),
				IFNULL(groups_items.sFullAccessDate, new_data.sCachedFullAccessDate)
			),
			groups_items.sCachedPartialAccessDate = LEAST(
				IFNULL(new_data.sCachedPartialAccessDate, groups_items.sPartialAccessDate),
				IFNULL(groups_items.sPartialAccessDate, new_data.sCachedPartialAccessDate)
			),
			groups_items.bCachedManagerAccess = GREATEST(
				IFNULL(new_data.bCachedManagerAccess, 0),
				groups_items.bManagerAccess
			),
			groups_items.sCachedAccessSolutionsDate = LEAST(
				IFNULL(new_data.sCachedAccessSolutionsDate, groups_items.sAccessSolutionsDate),
				IFNULL(groups_items.sAccessSolutionsDate, new_data.sCachedAccessSolutionsDate)
			),
			groups_items.sCachedGrayedAccessDate = new_data.sCachedGrayedAccessDate,
			groups_items.sCachedAccessReason = new_data.sAccessReasonAncestors
		WHERE groups_items_propagate.sPropagateAccess = 'self'`
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
		SET sPropagateAccess = 'children'
		WHERE sPropagateAccess = 'self'`
	stmtMarkChildrenItems, err = s.db.CommonDB().Prepare(queryMarkChildrenItems)
	mustNotBeError(err)
	defer func() { mustNotBeError(stmtMarkChildrenItems.Close()) }()

	hasChanges := true
	for hasChanges {
		_, err = stmtInsertMissingChildren.Exec()
		mustNotBeError(err)
		_, err = stmtInsertMissingPropagate.Exec()
		mustNotBeError(err)
		_, err = stmtUpdatePropagateAccess.Exec()
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
		"bCachedFullAccess":      "sCachedFullAccessDate",
		"bCachedPartialAccess":   "sCachedPartialAccessDate",
		"bCachedAccessSolutions": "sCachedAccessSolutionsDate",
		"bCachedGrayedAccess":    "sCachedGrayedAccessDate",
	}

	statements := make([]*sql.Stmt, 0, len(listFields))
	for bAccessField, sAccessDateField := range listFields {
		statement, err := s.db.CommonDB().Prepare(`
			UPDATE groups_items
			JOIN groups_items_propagate USING(ID)
			SET ` + bAccessField + ` = false
			WHERE ` + bAccessField + ` = true AND
				groups_items_propagate.sPropagateAccess = 'self' AND
				(` + sAccessDateField + ` IS NULL OR ` + sAccessDateField + ` > NOW())`) // #nosec
		mustNotBeError(err)
		statements = append(statements, statement)
	}
	return statements
}
