package database

import (
	"database/sql"
	"strconv"
	"strings"
	"time"

	"github.com/France-ioi/AlgoreaBackend/app/logging"
)

type groupItemPair struct {
	idGroup int64
	idItem  int64
}

const computeAllUserItemsLockName = "listener_computeAllUserItems"
const computeAllUserItemsLockTimeout = 10 * time.Second

// ComputeAllUserItems recomputes fields of users_items
// For users_items marked with sAncestorsComputationState = 'todo':
// 1. We mark all their ancestors in users_items as 'todo'
//  (we consider a row in users_items as an ancestor if it has the same value in idUser and
//  its idItem is an ancestor of the original row's idItem).
// 2. We process all objects that were marked as 'todo' and that have no children not marked as 'done'.
//  Then we copy sHintsRequested from related groups_attempts for them.
//  If an object has children, we update
//    sLastActivityDate, nbTasksTried, nbTasksWithHelp, nbTasksSolved, nbChildrenValidated, bValidated, sValidationDate.
//  This step is repeated until no records are updated.
// 3. We insert new groups_items for each processed row with bKeyObtained=1 according to corresponding items.idItemUnlocked.
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
				ancestors.idItem = items_ancestors.idItemAncestor AND
				items_ancestors.idItemAncestor != items_ancestors.idItemChild
			)
			JOIN users_items AS descendants ON (
				descendants.idItem = items_ancestors.idItemChild AND
				descendants.idUser = ancestors.idUser
			)
			SET ancestors.sAncestorsComputationState = 'todo'
			WHERE descendants.sAncestorsComputationState = 'todo'`).Error)

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
							SELECT inner_parent.ID
							FROM users_items AS inner_parent
							WHERE sAncestorsComputationState = 'todo'
								AND NOT EXISTS (
									SELECT items_items.idItemChild
									FROM items_items
									JOIN users_items AS children
										ON children.idItem = items_items.idItemChild
									WHERE items_items.idItemParent = inner_parent.idItem AND
										children.sAncestorsComputationState <> 'done' AND
										children.idUser = inner_parent.idUser
								)
							) AS tmp2
					) AS tmp
						USING(ID)
					SET sAncestorsComputationState = 'processing'`

				markAsProcessingStatement, err = userItemStore.db.CommonDB().Prepare(markAsProcessingQuery)
				mustNotBeError(err)
				defer func() { mustNotBeError(markAsProcessingStatement.Close()) }()
			}
			_, err = markAsProcessingStatement.Exec()
			mustNotBeError(err)

			userItemStore.collectItemsToUnlock(groupItemsToUnlock)

			// For every object marked as 'processing', we compute all the characteristics based on the children:
			//  - sLastActivityDate as the max of children's
			//  - nbTasksWithHelp, nbTasksTried, nbTaskSolved as the sum of children's field
			//  - nbChildrenValidated as the sum of children with bValidated == 1
			//  - bValidated, depending on the items_items.sCategory and items.sValidationType
			if updateStatement == nil {
				const updateQuery = `
					UPDATE users_items
					LEFT JOIN (
						SELECT
							MAX(children.sLastActivityDate) AS sLastActivityDate,
							SUM(children.nbTasksTried) AS nbTasksTried,
							SUM(children.nbTasksWithHelp) AS nbTasksWithHelp,
							SUM(children.nbTasksSolved) AS nbTasksSolved,
							SUM(bValidated) AS nbChildrenValidated,
							children.idUser AS idUser,
							items_items.idItemParent AS idItem
						FROM users_items AS children 
						JOIN items_items ON items_items.idItemChild = children.idItem
						GROUP BY children.idUser, items_items.idItemParent
					) AS children_data
						USING(idUser, idItem)
					LEFT JOIN task_children_data_view AS task_children_data
						ON task_children_data.idUserItem = users_items.ID
					JOIN items
						ON users_items.idItem = items.ID
					LEFT JOIN items_items
						ON items_items.idItemParent = users_items.idItem
					LEFT JOIN groups_attempts
						ON groups_attempts.ID = users_items.idAttemptActive
					SET
						users_items.sLastActivityDate = IF(task_children_data.idUserItem IS NOT NULL AND items_items.ID IS NOT NULL, children_data.sLastActivityDate, users_items.sLastActivityDate),
						users_items.nbTasksTried = IF(task_children_data.idUserItem IS NOT NULL AND items_items.ID IS NOT NULL, children_data.nbTasksTried, users_items.nbTasksTried),
						users_items.nbTasksWithHelp = IF(task_children_data.idUserItem IS NOT NULL AND items_items.ID IS NOT NULL, children_data.nbTasksWithHelp, users_items.nbTasksWithHelp),
						users_items.nbTasksSolved = IF(task_children_data.idUserItem IS NOT NULL AND items_items.ID IS NOT NULL, children_data.nbTasksSolved, users_items.nbTasksSolved),
						users_items.nbChildrenValidated = IF(task_children_data.idUserItem IS NOT NULL AND items_items.ID IS NOT NULL, children_data.nbChildrenValidated, users_items.nbChildrenValidated),
						users_items.bValidated = IF(task_children_data.idUserItem IS NOT NULL AND items_items.ID IS NOT NULL,
							CASE
								WHEN users_items.bValidated = 1 THEN 1
								WHEN items.sValidationType = 'Categories' THEN task_children_data.nbChildrenCategory = 0
								WHEN items.sValidationType = 'All' THEN task_children_data.nbChildrenNonValidated = 0
								WHEN items.sValidationType = 'AllButOne' THEN task_children_data.nbChildrenNonValidated < 2
								WHEN items.sValidationType = 'One' THEN task_children_data.nbChildrenValidated > 0
								ELSE 0
							END, users_items.bValidated),
						users_items.sValidationDate = IF(task_children_data.idUserItem IS NOT NULL AND items_items.ID IS NOT NULL,
							IFNULL(
								users_items.sValidationDate,
								IF(items.sValidationType = 'Categories', task_children_data.maxValidationDateCategories, task_children_data.maxValidationDate)
							), users_items.sValidationDate),
						users_items.sHintsRequested = IF(groups_attempts.ID IS NOT NULL, groups_attempts.sHintsRequested, users_items.sHintsRequested),
						users_items.sAncestorsComputationState = 'done'
					WHERE users_items.sAncestorsComputationState = 'processing'`
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
		s.GroupItems().after()
	}
	return nil
}

func (s *UserItemStore) collectItemsToUnlock(groupItemsToUnlock map[groupItemPair]bool) {
	// Unlock items depending on bKeyObtained
	const selectUnlocksQuery = `
		SELECT
			items.ID AS idItem,
			users.idGroupSelf AS idGroup,
			items.idItemUnlocked as idsItems
		FROM users_items
		JOIN items ON users_items.idItem = items.ID
		JOIN users ON users_items.idUser = users.ID
		WHERE users_items.sAncestorsComputationState = 'processing' AND
			users_items.bKeyObtained AND items.idItemUnlocked IS NOT NULL`
	var err error
	var unlocksResult []struct {
		IDItem   int64  `gorm:"column:idItem"`
		IDGroup  int64  `gorm:"column:idGroup"`
		ItemsIds string `gorm:"column:idsItems"`
	}
	mustNotBeError(s.Raw(selectUnlocksQuery).Scan(&unlocksResult).Error())
	for _, unlock := range unlocksResult {
		idsItems := strings.Split(unlock.ItemsIds, ",")
		for _, idItem := range idsItems {
			var idItemInt64 int64
			if idItemInt64, err = strconv.ParseInt(idItem, 10, 64); err != nil {
				logging.SharedLogger.WithFields(map[string]interface{}{
					"items.ID":             unlock.IDItem,
					"items.idItemUnlocked": unlock.ItemsIds,
					"error":                err,
				}).Warn("cannot parse items.idItemUnlocked")
			} else {
				groupItemsToUnlock[groupItemPair{idGroup: unlock.IDGroup, idItem: idItemInt64}] = true
			}
		}
	}
}

func (s *UserItemStore) unlockGroupItems(groupItemsToUnlock map[groupItemPair]bool) int64 {
	if len(groupItemsToUnlock) <= 0 {
		return 0
	}
	query := `
		INSERT INTO groups_items
			(idGroup, idItem, sPartialAccessDate, sCachedPartialAccessDate, bCachedPartialAccess)
		VALUES (?, ?, NOW(), NOW(), 1)`
	values := make([]interface{}, 0, len(groupItemsToUnlock)*2)
	valuesTemplate := ", (?, ?, NOW(), NOW(), 1)"
	for item := range groupItemsToUnlock {
		values = append(values, item.idGroup, item.idItem)
	}

	query += strings.Repeat(valuesTemplate, len(groupItemsToUnlock)-1) +
		" ON DUPLICATE KEY UPDATE sPartialAccessDate = NOW(), sCachedPartialAccessDate = NOW(), bCachedPartialAccess = 1"
	result := s.db.Exec(query, values...)
	mustNotBeError(result.Error)
	return result.RowsAffected
}
