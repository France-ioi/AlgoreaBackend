package database

import (
	"database/sql"
)

const groups = "groups"

// createNewAncestors inserts new rows into
// the objectName_ancestors table (items_ancestors or groups_ancestors)
// for all rows marked with ancestors_computation_state="todo" in objectName_propagate
// (items_propagate or groups_propagate) and their descendants.
//
// Note: rows in *_propagate tables with `ancestors_computation_state`="todo"
// are added in the database in SQL triggers:
// - after_update_groups_groups
// - before_insert_items_items/groups_groups
// - before_delete_items_items/groups_groups.
func (s *DataStore) createNewAncestors(objectName, singleObjectName string) { /* #nosec */
	s.mustBeInTransaction()

	// We mark as 'todo' all descendants of objects marked as 'todo'
	query := `
		INSERT INTO  ` + objectName + `_propagate (id, ancestors_computation_state)
		SELECT ` + QuoteName(objectName+"_ancestors") + ".child_" + singleObjectName + `_id, 'todo'
		FROM ` + QuoteName(objectName+"_ancestors") + `
		JOIN ` + QuoteName(objectName+"_propagate") + ` AS ancestors
			ON ancestors.id = ` + QuoteName(objectName+"_ancestors") + ".ancestor_" + singleObjectName + `_id
		WHERE ancestors.ancestors_computation_state = 'todo'
		FOR SHARE OF ` + QuoteName(objectName+"_ancestors") + `
		FOR UPDATE OF ancestors
		ON DUPLICATE KEY UPDATE ancestors_computation_state = 'todo'` /* #nosec */

	mustNotBeError(s.db.Exec(query).Error)

	createTemporaryTableQuery := "CREATE TEMPORARY TABLE " + objectName + "_propagate_processing (id BIGINT NOT NULL)"
	dropTemporaryTableQuery := "DROP TEMPORARY TABLE IF EXISTS " + objectName + "_propagate_processing"
	mustNotBeError(s.db.Exec(createTemporaryTableQuery).Error)
	defer func() {
		mustNotBeError(s.db.Exec(dropTemporaryTableQuery).Error)
	}()

	relationsTable := objectName + "_" + objectName

	additionalRelationCondition := "1"
	if objectName == groups {
		additionalRelationCondition = "groups_groups.is_team_membership = 0"
	}
	// Next queries will be executed in the loop

	// We mark as processing all objects that were marked as 'todo' and that have no parents not marked as 'done'.
	// This way we prevent infinite looping as we never process objects that are descendants of themselves

	/* #nosec */
	markAsProcessingQuery := `
		INSERT INTO ` + objectName + `_propagate_processing (id)
		SELECT id
		FROM ` + objectName + `_propagate AS children
		WHERE children.ancestors_computation_state = 'todo' AND
			NOT EXISTS (
				SELECT 1
				FROM ` + relationsTable + `
					JOIN ` + objectName + `_propagate
						ON ` + objectName + `_propagate.id = ` + relationsTable + `.parent_` + singleObjectName + `_id AND
							 ` + objectName + `_propagate.ancestors_computation_state = 'todo'
				WHERE ` + relationsTable + `.child_` + singleObjectName + `_id = children.id AND ` + additionalRelationCondition + `
				LIMIT 1
				FOR SHARE OF ` + relationsTable + `
				FOR UPDATE OF ` + objectName + `_propagate
			)
		FOR UPDATE OF children` // #nosec

	createTemporaryTable, err := s.db.CommonDB().Prepare(createTemporaryTableQuery)
	mustNotBeError(err)
	defer func() { mustNotBeError(createTemporaryTable.Close()) }()
	markAsProcessing, err := s.db.CommonDB().Prepare(markAsProcessingQuery)
	mustNotBeError(err)
	defer func() { mustNotBeError(markAsProcessing.Close()) }()

	expiresAtColumn := ""
	expiresAtValueJoin := ""
	ignore := "IGNORE"

	if objectName == groups {
		expiresAtColumn = ", expires_at"
		expiresAtValueJoin = ", MAX(LEAST(groups_ancestors_join.expires_at, groups_groups.expires_at)) AS max_expires_at"
		ignore = ""
	}

	// For every object marked as processing, we compute all its ancestors
	recomputeQueries := make([]string, 0, 3)
	recomputeQueries = append(recomputeQueries, `
		DELETE `+objectName+`_ancestors
		FROM `+objectName+`_ancestors
			JOIN `+objectName+`_propagate_processing
				ON `+objectName+`_propagate_processing.id = `+objectName+`_ancestors.child_`+singleObjectName+`_id`, `
		INSERT `+ignore+` INTO `+objectName+`_ancestors
		(
			ancestor_`+singleObjectName+`_id,
			child_`+singleObjectName+`_id`+`
			`+expiresAtColumn+`
		)
		SELECT
			`+objectName+`_ancestors_join.ancestor_`+singleObjectName+`_id,
			`+relationsTable+`.child_`+singleObjectName+`_id
			`+expiresAtValueJoin+`
		FROM `+relationsTable+` AS `+relationsTable+`
		JOIN `+objectName+`_ancestors AS `+objectName+`_ancestors_join ON (
			`+objectName+`_ancestors_join.child_`+singleObjectName+`_id = `+relationsTable+`.parent_`+singleObjectName+`_id
		)
		JOIN `+objectName+`_propagate_processing ON (
			`+relationsTable+`.child_`+singleObjectName+`_id = `+objectName+`_propagate_processing.id
		)
		WHERE `+additionalRelationCondition) // #nosec
	if objectName == groups {
		recomputeQueries[0] += `
			AND groups_ancestors.ancestor_group_id != groups_ancestors.child_group_id` // do not delete group ancestors with is_self=1
		recomputeQueries[1] += `
				AND NOW() < groups_groups.expires_at
				AND NOW() < groups_ancestors_join.expires_at
			GROUP BY groups_groups.child_group_id, groups_ancestors_join.ancestor_group_id
			HAVING NOW() < max_expires_at
			FOR UPDATE OF ` + objectName + `_propagate_processing
			FOR SHARE OF ` + objectName + `_ancestors_join
			FOR SHARE OF ` + relationsTable
	} else {
		recomputeQueries[1] += `
			FOR UPDATE OF ` + objectName + `_propagate_processing
			FOR SHARE OF ` + objectName + `_ancestors_join
			FOR SHARE OF ` + relationsTable
		recomputeQueries = append(recomputeQueries, `
			INSERT IGNORE INTO items_ancestors (ancestor_item_id, child_item_id)
			SELECT items_items.parent_item_id, items_items.child_item_id
			FROM items_items
			JOIN items_propagate_processing ON items_items.child_item_id = items_propagate_processing.id
			FOR UPDATE OF items_propagate_processing
			FOR SHARE OF items_items`) // #nosec
	}

	recomputeAncestors := make([]*sql.Stmt, len(recomputeQueries))
	for i := 0; i < len(recomputeQueries); i++ {
		recomputeAncestors[i], err = s.db.CommonDB().Prepare(recomputeQueries[i])
		mustNotBeError(err)

		defer func(i int) { mustNotBeError(recomputeAncestors[i].Close()) }(i)
	}

	// Objects marked as processing are now marked as 'done'
	markAsDoneQuery := `
		UPDATE ` + objectName + `_propagate
		JOIN ` + objectName + `_propagate_processing
			ON ` + objectName + `_propagate.id = ` + objectName + `_propagate_processing.id
		SET ancestors_computation_state = 'done'` // #nosec
	markAsDone, err := s.db.CommonDB().Prepare(markAsDoneQuery)
	mustNotBeError(err)
	defer func() { mustNotBeError(markAsDone.Close()) }()
	dropTemporaryTable, err := s.db.CommonDB().Prepare(dropTemporaryTableQuery)
	mustNotBeError(err)
	defer func() { mustNotBeError(dropTemporaryTable.Close()) }()

	for {
		_, err = markAsProcessing.Exec()
		mustNotBeError(err)
		for i := 0; i < len(recomputeAncestors); i++ {
			_, err = recomputeAncestors[i].Exec()
			mustNotBeError(err)
		}

		var result sql.Result
		result, err = markAsDone.Exec()
		mustNotBeError(err)
		var rowsAffected int64
		rowsAffected, err = result.RowsAffected()
		mustNotBeError(err)
		if rowsAffected == 0 {
			break
		}

		_, err = dropTemporaryTable.Exec()
		mustNotBeError(err)

		_, err = createTemporaryTable.Exec()
		mustNotBeError(err)
	}
}
