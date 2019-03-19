package database

import (
	"database/sql"
	"runtime"
	"strings"
)

// UserItemStore implements database operations on `users_items`
type UserItemStore struct {
	*DataStore
}

// ComputeAllUserItems recomputes fields of users_items
func (s *UserItemStore) ComputeAllUserItems() (err error) {
	defer func() {
		if p := recover(); p != nil {
			switch e := p.(type) {
			case runtime.Error:
				panic(e)
			default:
				err = p.(error)
			}
		}
	}()

	// Use a lock so that we don't execute the listener multiple times in parallel
	var getLockResult int64
	mustNotBeError(s.db.Raw("SELECT GET_LOCK('listener_computeAllUserItems', 1)").Row().Scan(&getLockResult))
	if getLockResult != 1 {
		return nil
	}

	/*
		$db->exec("LOCK TABLES
		users_items as ancestors WRITE,
		users_items as descendants WRITE,
		history_users_items WRITE,
		items_ancestors READ,
		history_items_ancestors READ;
		");
	*/
	// We mark as 'todo' all ancestors of objects marked as 'todo'
	mustNotBeError(s.db.Exec(
		`UPDATE users_items AS ancestors
			JOIN items_ancestors
			ON (ancestors.idItem = items_ancestors.idItemAncestor AND
				items_ancestors.idItemAncestor != items_ancestors.idItemChild)
			JOIN users_items AS descendants
			ON (descendants.idItem = items_ancestors.idItemChild AND 
				descendants.idUser = ancestors.idUser)
			SET ancestors.sAncestorsComputationState = 'todo'
			WHERE descendants.sAncestorsComputationState = 'todo'`).Error)

	//$db->exec("UNLOCK TABLES;");
	hasChanges := true
	groupsItemsChanged := false

	var markAsProcessingStatement, updateActiveAttemptStatement, markAsDoneStatement,
		updateStatement, insertUnlocksStatement *sql.Stmt

	for hasChanges {
		// We mark as "processing" all objects that were marked as 'todo' and that have no children not marked as 'done'
		if markAsProcessingStatement == nil {
			const markAsProcessingQuery = `
				UPDATE ` + "`users_items`" + ` AS ` + "`parent`" + `
				JOIN (
					SELECT * FROM (
						SELECT ` + "`parent`.`ID`" + ` FROM ` + "`users_items`" + ` AS ` + "`parent`" + `
						WHERE ` + "`sAncestorsComputationState`" + ` = 'todo'
							AND NOT EXISTS (
								SELECT ` + "`items_items`.`idItemChild`" + `
								FROM ` + "`items_items`" + `
								JOIN ` + "`users_items`" + ` AS ` + "`children`" + `
								ON (` + "`children`.`idItem` = `items_items`.`idItemChild`" + `)
								WHERE ` + "`items_items`.`idItemParent` = `parent`.`idItem`" + ` AND
									` + "`children`.`sAncestorsComputationState`" + ` <> 'done' AND
									` + "`children`.`idUser` = `parent`.`idUser`" + `
							)
							ORDER BY parent.ID
							` + //FOR UPDATE
				`) AS tmp2
				) AS tmp
				SET sAncestorsComputationState = 'processing'
				WHERE tmp.ID = parent.ID
`
			markAsProcessingStatement, err = s.db.CommonDB().Prepare(markAsProcessingQuery)
			mustNotBeError(err)
			defer func() { mustNotBeError(markAsProcessingStatement.Close()) }()
		}
		_, err = markAsProcessingStatement.Exec()
		mustNotBeError(err)

		/** For every object marked as 'processing', we compute all the characteristics based on the children:
		* sLastActivityDate as the max of children's
		* nbTasksWithHelp, nbTasksTried, nbTaskSolved as the sum of children's field
		* nbChildrenValidated as the sum of children with bValidated == 1
		* bValidated, depending on the items_items.sCategory and items.sValidationType
		 */
		if updateActiveAttemptStatement == nil {
			const updateActiveAttemptQuery = `
				UPDATE users_items
				JOIN groups_attempts ON groups_attempts.ID = users_items.idAttemptActive
				SET users_items.sHintsRequested = groups_attempts.sHintsRequested
				WHERE users_items.sAncestorsComputationState = 'processing'`
			updateActiveAttemptStatement, err = s.db.CommonDB().Prepare(updateActiveAttemptQuery)
			mustNotBeError(err)
			defer func() { mustNotBeError(updateActiveAttemptStatement.Close()) }()
		}
		_, err = updateActiveAttemptStatement.Exec()
		mustNotBeError(err)

		// query only user_items with children
		const selectNewUsersItemsQuery = `
			SELECT DISTINCT
			` + "`users_items`.`ID`, `users_items`.`idUser`, `users_items`.`idItem`, `items`.`sValidationType`" + `
			FROM users_items
			JOIN items_items ON items_items.idItemParent = users_items.idItem
			JOIN items ON ` + "`items`.`ID` = `users_items`.`idItem`" + `
			WHERE ` + "`users_items`.`sAncestorsComputationState`" + ` = 'processing'`
		//ORDER BY ID`
		//FOR UPDATE`

		var rows []map[string]interface{}
		mustNotBeError(s.Raw(selectNewUsersItemsQuery).ScanIntoSliceOfMaps(&rows).Error())

		for _, row := range rows {
			if updateStatement == nil {
				const updateQuery = `
					UPDATE users_items
					JOIN
						(SELECT MAX(children.sLastActivityDate) AS sLastActivityDate,
							SUM(children.nbTasksTried) AS nbTasksTried, Sum(children.nbTasksWithHelp) AS nbTasksWithHelp,
							SUM(children.nbTasksSolved) AS nbTasksSolved, SUM(bValidated) AS nbChildrenValidated
						FROM users_items AS children 
						JOIN items_items ON items_items.idItemChild = children.idItem
						WHERE children.idUser = ? AND items_items.idItemParent = ? ` + // ?=idUser, ?=idItem
					//`FOR UPDATE`
					`) AS children_data
					JOIN
						(SELECT
							SUM(IF(task_children.ID IS NOT NULL AND task_children.bValidated, 1, 0)) AS nbChildrenValidated,
							SUM(IF(task_children.ID IS NOT NULL AND task_children.bValidated, 0, 1)) AS nbChildrenNonValidated,
							SUM(
								IF(items_items.sCategory = 'Validation' AND
									(task_children.ID IS NULL OR task_children.bValidated = 0), 1, 0)
							) AS nbChildrenCategory,
							MAX(task_children.sValidationDate) AS maxValidationDate,
							MAX(IF(items_items.sCategory = 'Validation', task_children.sValidationDate, NULL)) AS maxValidationDateCategories
						FROM items_items
						LEFT JOIN users_items AS task_children
						ON items_items.idItemChild = task_children.idItem AND task_children.idUser = ? ` + // ?=idUser
					` JOIN items ON items.ID = items_items.idItemChild
						WHERE items_items.idItemParent = ? AND items.sType != 'Course' AND items.bNoScore = 0 ` + // ?=idItem
					//`FOR UPDATE`
					` ) AS task_children_data
					SET users_items.sLastActivityDate = children_data.sLastActivityDate,
						users_items.nbTasksTried = children_data.nbTasksTried,
						users_items.nbTasksWithHelp = children_data.nbTasksWithHelp,
						users_items.nbTasksSolved = children_data.nbTasksSolved,
						users_items.nbChildrenValidated = children_data.nbChildrenValidated,
						users_items.bValidated =
							IF(
								users_items.bValidated = 1,
								1, ` + // users_items.bValidated = 1
					`			IF(
									STRCMP(?, 'Categories'), ` + // ?=sValidationType
					`				IF(
										STRCMP(?, 'All'), ` + // ?=sValidationType
					`					IF(
											STRCMP(?, 'AllButOne'), ` + // ?=sValidationType
					`						IF(
												STRCMP(?, 'One'), ` + // ?=sValidationType
					`							0, ` + // @sValidationType not in('Categories', 'All', 'AllButOne', 'One')
					`							IF(
													task_children_data.nbChildrenValidated > 0,
													1, ` + // @sValidationType == 'One' && task_children_data.nbChildrenValidated > 0
					`								0` + //   @sValidationType == 'One' && task_children_data.nbChildrenValidated <= 0
					`							)
											),
											IF(
												task_children_data.nbChildrenNonValidated < 2,
												1, ` + // @sValidationType == 'AllButOne' && task_children_data.nbChildrenNonValidated < 2
					`							0` + //   @sValidationType == 'AllButOne' && task_children_data.nbChildrenNonValidated >= 2
					`						)
										),
										IF(
											task_children_data.nbChildrenNonValidated = 0,
											1, ` + // @sValidationType == 'All' && task_children_data.nbChildrenNonValidated == 0
					`						0` + //   @sValidationType == 'All' && task_children_data.nbChildrenNonValidated != 0
					`					)
									),
									IF(
										task_children_data.nbChildrenCategory = 0,
										1, ` + // @sValidationType == 'Categories' && task_children_data.nbChildrenCategory == 0
					`					0` + //   @sValidationType == 'Categories' && task_children_data.nbChildrenCategory != 0
					`				)
								)
							),
						users_items.sValidationDate =
							IFNULL(
								users_items.sValidationDate,
								IF(
									STRCMP(?, 'Categories'), ` + // ?=sValidationType
					// 			users_items.sValidationDate IS NULL && @sValidationType != 'Categories'
					`				task_children_data.maxValidationDate, ` +
					// 			users_items.sValidationDate IS NULL && @sValidationType == 'Categories'
					`				task_children_data.maxValidationDateCategories
								)
							)
					WHERE users_items.ID = ?` // ?=ID
				updateStatement, err = s.db.CommonDB().Prepare(updateQuery)
				mustNotBeError(err)
				defer func() { mustNotBeError(updateStatement.Close()) }()
			}
			_, err = updateStatement.Exec(row["idUser"], row["idItem"], row["idUser"], row["idItem"],
				row["sValidationType"], row["sValidationType"], row["sValidationType"], row["sValidationType"],
				row["sValidationType"], row["ID"])
			mustNotBeError(err)
		}

		// Unlock items depending on bKeyObtained
		const selectUnlocksQuery = `
			SELECT users.idGroupSelf AS idGroup, items.idItemUnlocked as idsItems
			FROM users_items
			JOIN items ON users_items.idItem = items.ID
			JOIN users ON users_items.idUser = users.ID
			WHERE users_items.sAncestorsComputationState = 'processing' AND
				users_items.bKeyObtained = 1 AND items.idItemUnlocked IS NOT NULL`
		//ORDER BY users_items.ID`
		//FOR UPDATE`

		var unlocks []map[string]interface{}
		mustNotBeError(s.Raw(selectUnlocksQuery).ScanIntoSliceOfMaps(&unlocks).Error())

		if insertUnlocksStatement == nil {
			const insertUnlocksQuery = `
				INSERT INTO groups_items
				(idGroup, idItem, sPartialAccessDate, sCachedPartialAccessDate, bCachedPartialAccess)
				VALUES(?, ?, NOW(), NOW(), 1)
				ON DUPLICATE KEY UPDATE
				sPartialAccessDate = NOW(), sCachedPartialAccessDate = NOW(), bCachedPartialAccess = 1`
			insertUnlocksStatement, err = s.db.CommonDB().Prepare(insertUnlocksQuery)
			mustNotBeError(err)
			defer func() { mustNotBeError(insertUnlocksStatement.Close()) }()
		}
		for _, unlock := range unlocks {
			groupsItemsChanged = true
			idsItems := strings.Split(unlock["idsItems"].(string), ",")
			for _, idItem := range idsItems {
				_, err = insertUnlocksStatement.Exec(unlock["idGroup"], idItem)
				mustNotBeError(err)
			}
		}

		// Objects marked as 'processing' are now marked as 'done'
		if markAsDoneStatement == nil {
			const markAsDoneQuery = "UPDATE `users_items` SET `sAncestorsComputationState` = 'done' WHERE `sAncestorsComputationState` = 'processing'"
			markAsDoneStatement, err = s.db.CommonDB().Prepare(markAsDoneQuery)
			mustNotBeError(err)
			defer func() { mustNotBeError(markAsDoneStatement.Close()) }()
		}
		var result sql.Result
		result, err = markAsDoneStatement.Exec()
		mustNotBeError(err)
		var rowsAffected int64
		rowsAffected, err = result.RowsAffected()
		mustNotBeError(err)
		hasChanges = rowsAffected > 0
	}

	// Release the lock
	mustNotBeError(s.db.Raw("SELECT RELEASE_LOCK('listener_computeAllUserItems')").Row().Scan(&getLockResult))

	// If items have been unlocked, need to recompute access
	if groupsItemsChanged {
		_ = groupsItemsChanged // stub
		//Listeners::groupsItemsAfter($db);
	}
	return nil
}

func mustNotBeError(err error) {
	if err != nil {
		panic(err)
	}
}
