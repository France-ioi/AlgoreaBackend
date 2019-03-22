package database

import (
	"database/sql"
	"fmt"
	"runtime"
	"strconv"
	"strings"
)

// UserItemStore implements database operations on `users_items`
type UserItemStore struct {
	*DataStore
}

type groupItemPair struct {
	idGroup int64
	idItem  int64
}

// ComputeAllUserItems recomputes fields of users_items
func (s *UserItemStore) ComputeAllUserItems() (err error) {
	defer func() {
		if p := recover(); p != nil {
			switch e := p.(type) {
			case runtime.Error, *strconv.NumError:
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

	var markAsProcessingStatement, updateStatement *sql.Stmt
	groupItemsToUnlock := make(map[groupItemPair]bool)

	for hasChanges {
		// We mark as "processing" all objects that were marked as 'todo' and that have no children not marked as 'done'
		// This way we prevent infinite looping as we never process items that are ancestors of themselves
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
				`	) AS tmp2
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

		s.collectItemsToUnlock(groupItemsToUnlock)

		/** For every object marked as 'processing', we compute all the characteristics based on the children:
		* sLastActivityDate as the max of children's
		* nbTasksWithHelp, nbTasksTried, nbTaskSolved as the sum of children's field
		* nbChildrenValidated as the sum of children with bValidated == 1
		* bValidated, depending on the items_items.sCategory and items.sValidationType
		 */

		if updateStatement == nil {
			const updateQuery = `
					UPDATE users_items
					LEFT JOIN
						(SELECT MAX(children.sLastActivityDate) AS sLastActivityDate,
							SUM(children.nbTasksTried) AS nbTasksTried, SUM(children.nbTasksWithHelp) AS nbTasksWithHelp,
							SUM(children.nbTasksSolved) AS nbTasksSolved, SUM(bValidated) AS nbChildrenValidated,
							children.idUser AS idUser, items_items.idItemParent AS idItem
						FROM users_items AS children 
						JOIN items_items ON items_items.idItemChild = children.idItem
						GROUP BY children.idUser, items_items.idItemParent ` +
				//`FOR UPDATE`
				` ) AS children_data
					ON users_items.idUser = children_data.idUser AND children_data.idItem = users_items.idItem
					LEFT JOIN task_children_data_view AS task_children_data
						ON task_children_data.idUserItem = users_items.ID
					JOIN items ON users_items.idItem = items.ID
					LEFT JOIN items_items ON items_items.idItemParent = users_items.idItem
					LEFT JOIN groups_attempts ON groups_attempts.ID = users_items.idAttemptActive
					SET
						users_items.sLastActivityDate = IF(task_children_data.idUserItem IS NOT NULL AND items_items.ID IS NOT NULL, children_data.sLastActivityDate, users_items.sLastActivityDate),
						users_items.nbTasksTried = IF(task_children_data.idUserItem IS NOT NULL AND items_items.ID IS NOT NULL, children_data.nbTasksTried, users_items.nbTasksTried),
						users_items.nbTasksWithHelp = IF(task_children_data.idUserItem IS NOT NULL AND items_items.ID IS NOT NULL, children_data.nbTasksWithHelp, users_items.nbTasksWithHelp),
						users_items.nbTasksSolved = IF(task_children_data.idUserItem IS NOT NULL AND items_items.ID IS NOT NULL, children_data.nbTasksSolved, users_items.nbTasksSolved),
						users_items.nbChildrenValidated = IF(task_children_data.idUserItem IS NOT NULL AND items_items.ID IS NOT NULL, children_data.nbChildrenValidated, users_items.nbChildrenValidated),
						users_items.bValidated = IF(task_children_data.idUserItem IS NOT NULL AND items_items.ID IS NOT NULL, CASE
							WHEN users_items.bValidated = 1 THEN
								1
							WHEN STRCMP(items.sValidationType, 'Categories') = 0 THEN
								task_children_data.nbChildrenCategory = 0
							WHEN STRCMP(items.sValidationType, 'All') = 0 THEN
								task_children_data.nbChildrenNonValidated = 0
							WHEN STRCMP(items.sValidationType, 'AllButOne') = 0 THEN
								task_children_data.nbChildrenNonValidated < 2
							WHEN STRCMP(items.sValidationType, 'One') = 0 THEN
								task_children_data.nbChildrenValidated > 0
							ELSE
								0
							END, users_items.bValidated),
						users_items.sValidationDate = IF(task_children_data.idUserItem IS NOT NULL AND items_items.ID IS NOT NULL,
							IFNULL(
								users_items.sValidationDate,
								IF(
									STRCMP(items.sValidationType, 'Categories'), ` +
				//				users_items.sValidationDate IS NULL && @sValidationType != 'Categories'
				`					task_children_data.maxValidationDate, ` +
				//				users_items.sValidationDate IS NULL && @sValidationType == 'Categories'
				`					task_children_data.maxValidationDateCategories
								)
							), users_items.sValidationDate),
						users_items.sHintsRequested = IF(groups_attempts.ID IS NOT NULL, groups_attempts.sHintsRequested, users_items.sHintsRequested),
						users_items.sAncestorsComputationState = 'done'
					WHERE users_items.sAncestorsComputationState = 'processing'`
			updateStatement, err = s.db.CommonDB().Prepare(updateQuery)
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

	groupsUnlocked := s.unlockGroupItems(groupItemsToUnlock)

	// Release the lock
	mustNotBeError(s.db.Raw("SELECT RELEASE_LOCK('listener_computeAllUserItems')").Row().Scan(&getLockResult))

	// If items have been unlocked, need to recompute access
	if groupsUnlocked > 0 {
		_ = groupsUnlocked // stub
		//Listeners::groupsItemsAfter($db);
	}
	return nil
}

func (s *UserItemStore) collectItemsToUnlock(groupItemsToUnlock map[groupItemPair]bool) {
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
	var err error
	var unlocksResult []struct {
		IDGroup  int64  `gorm:"column:idGroup"`
		ItemsIds string `gorm:"column:idsItems"`
	}
	mustNotBeError(s.Raw(selectUnlocksQuery).Scan(&unlocksResult).Error())
	for _, unlock := range unlocksResult {
		idsItems := strings.Split(unlock.ItemsIds, ",")
		for _, idItem := range idsItems {
			var idItemInt64 int64
			if idItemInt64, err = strconv.ParseInt(idItem, 10, 64); err != nil {
				panic(err)
			}
			groupItemsToUnlock[groupItemPair{idGroup: unlock.IDGroup, idItem: idItemInt64}] = true
		}
	}
}

func (s *UserItemStore) unlockGroupItems(groupItemsToUnlock map[groupItemPair]bool) int64 {
	if len(groupItemsToUnlock) > 0 {
		query := `
						INSERT INTO groups_items
						(idGroup, idItem, sPartialAccessDate, sCachedPartialAccessDate, bCachedPartialAccess)
						VALUES `
		rowsData := make([]string, 0, len(groupItemsToUnlock))
		for item := range groupItemsToUnlock {
			rowsData = append(rowsData, fmt.Sprintf("(%d, %d, NOW(), NOW(), 1)", item.idGroup, item.idItem))
		}

		query += strings.Join(rowsData, ", ") +
			"ON DUPLICATE KEY UPDATE sPartialAccessDate = NOW(), sCachedPartialAccessDate = NOW(), bCachedPartialAccess = 1"
		result := s.db.Exec(query)
		mustNotBeError(result.Error)
		return result.RowsAffected
	}
	return 0
}

func mustNotBeError(err error) {
	if err != nil {
		panic(err)
	}
}
