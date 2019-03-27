package database

func (s *GroupItemStore) computeAllAccess() {
	// Lock all tables during computation to avoid issues
	/*
		$queryLockTables = "LOCK TABLES
			groups_items WRITE,
			groups_items AS parents READ,
			groups_items AS children READ,
			groups_items AS parent READ,
			groups_items AS child READ,
			groups_items AS new_data READ,
			history_groups_items WRITE,
			groups_items_propagate WRITE,
			groups_items_propagate AS parents_propagate READ,
			items READ,
			items_items READ;";
		$queryUnlockTables = "UNLOCK TABLES;";
	*/

	// inserting missing groups_items_propagate
	const queryInsertMissingPropagate = `
		INSERT INTO groups_items_propagate (ID, sPropagateAccess)
		SELECT
			groups_items.ID,
			'self' as sPropagateAccess
		FROM groups_items
		WHERE sPropagateAccess='self'
		ON DUPLICATE KEY UPDATE sPropagateAccess='self'`

	// Set groups_items as set up for propagation
	const queryUpdatePropagateAccess = "UPDATE `groups_items` SET `sPropagateAccess`='done' WHERE `sPropagateAccess`='self'"

	// inserting missing children of groups_items marked as 'children'
	const queryInsertMissingChildren = `
		INSERT IGNORE INTO groups_items (idGroup, idItem, idUserCreated, sCachedAccessReason, sAccessReason)
		SELECT
			parents.idGroup AS idGroup,
			items_items.idItemChild AS idItem,
			parents.idUserCreated AS idUserCreated,
			'' AS sCachedAccessReason,
			'' AS sAccessReason
		FROM items_items
		JOIN groups_items AS parents
			ON parents.idItem = items_items.idItemParent
		JOIN groups_items_propagate AS parents_propagate
			ON parents.ID = parents_propagate.ID AND parents_propagate.sPropagateAccess = 'children'`

	// mark as 'done' items that shouldn't propagate
	const queryMarkDoNotPropagate = `
		INSERT IGNORE INTO groups_items_propagate (ID, sPropagateAccess)
		SELECT
			groups_items.ID AS ID,
			'done' as sPropagateAccess
		FROM groups_items
		JOIN items
			ON groups_items.idItem = items.ID AND items.bCustomChapter
		ON DUPLICATE KEY UPDATE sPropagateAccess='done'`

	// marking 'self' groups_items sons of groups_items marked as 'children'
	const queryMarkExistingChildren = `
		INSERT IGNORE INTO groups_items_propagate (ID, sPropagateAccess)
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

	// marking 'children' groups_items as 'done'
	const queryMarkFinishedItems = `
		UPDATE groups_items_propagate
		SET sPropagateAccess = 'done'
		WHERE sPropagateAccess = 'children'`

	// computation for groups_items marked as 'self'.
	// It also marks 'self' groups_items as 'children'
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
				CONCAT('From ancestor group(s) ', GROUP_CONCAT(parent.sAccessReason, ', ')) AS sAccessReasonAncestors
			FROM groups_items AS child
			JOIN items_items
				ON items_items.idItemChild = child.idItem
			LEFT JOIN groups_items_propagate
				ON groups_items_propagate.ID = child.ID
			JOIN groups_items AS parent
				ON parent.idItem = items_items.idItemParent AND parent.idGroup = child.idGroup
			WHERE
				(groups_items_propagate.sPropagateAccess = 'self' OR groups_items_propagate.ID IS NULL) AND
				(
					parent.sCachedFullAccessDate IS NOT NULL OR
					parent.sCachedPartialAccessDate IS NOT NULL OR
					parent.sCachedAccessSolutionsDate IS NOT NULL OR
					parent.sFullAccessDate IS NOT NULL OR
					parent.sPartialAccessDate IS NOT NULL OR
					parent.sAccessSolutionsDate IS NOT NULL OR
					parent.bManagerAccess
				)
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

	// marking 'self' groups_items as 'children'
	const queryMarkChildrenItems = `
		UPDATE groups_items_propagate
		SET sPropagateAccess = 'children'
		WHERE sPropagateAccess = 'self'`

	hasChanges := true
	for hasChanges {
		//mustNotBeError(s.db.Exec(queryLockTables).Error)
		mustNotBeError(s.db.Exec(queryInsertMissingChildren).Error)
		mustNotBeError(s.db.Exec(queryInsertMissingPropagate).Error)
		mustNotBeError(s.db.Exec(queryUpdatePropagateAccess).Error)
		mustNotBeError(s.db.Exec(queryMarkDoNotPropagate).Error)
		mustNotBeError(s.db.Exec(queryMarkExistingChildren).Error)
		mustNotBeError(s.db.Exec(queryMarkFinishedItems).Error)
		mustNotBeError(s.db.Exec(queryUpdateGroupItems).Error)
		result := s.db.Exec(queryMarkChildrenItems)
		mustNotBeError(result.Error)
		hasChanges = result.RowsAffected > 0
		//mustNotBeError(s.db.Exec(queryUnlockTables).Error)
	}

	/*
		// commented out in the PHP code
		// remove default groups_items (veeeery slow)
		// TODO :: maybe move to some cleaning cron
		const queryDeleteDefaultGI = "delete from `groups_items` where " +
			"    `sCachedAccessSolutionsDate` is null " +
			"and `sCachedPartialAccessDate` is null " +
			"and `sCachedFullAccessDate` is null " +
			"and `sCachedGrayedAccessDate` is null " +
			"and `sCachedAccessReason` = '' " +
			"and `sFullAccessDate` is null " +
			"and `sPartialAccessDate` is null " +
			"and `sAccessSolutionsDate` is null " +
			"and `bCachedManagerAccess` = 0 " +
			"and `bManagerAccess` = 0 " +
			"and `bOwnerAccess` = 0 " +
			"and `sAccessReason` = ''"
		//mustNotBeError(s.db.Exec(queryDeleteDefaultGI).Error)
	*/
}
