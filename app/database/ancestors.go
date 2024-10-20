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
// - after_insert_items/groups
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

	hasChanges := true

	relationsTable := objectName + "_" + objectName

	var additionalJoin, additionalLocking string
	if objectName == groups {
		additionalJoin = " JOIN `groups` AS parent ON parent.id = groups_groups.parent_group_id AND parent.type != 'Team' "
		additionalLocking = " FOR SHARE OF parent "
	}
	// Next queries will be executed in the loop

	// We mark as "processing" all objects that were marked as 'todo' and that have no parents not marked as 'done'
	// This way we prevent infinite looping as we never process objects that are descendants of themselves

	/* #nosec */
	markAsProcessingQuery := `
		UPDATE ` + objectName + `_propagate AS children
		SET children.ancestors_computation_state='processing'
		WHERE children.ancestors_computation_state = 'todo' AND NOT EXISTS (
			SELECT 1 FROM (
				SELECT 1
				FROM ` + relationsTable + `
					JOIN ` + objectName + `_propagate
						ON ` + objectName + `_propagate.id = ` + relationsTable + `.parent_` + singleObjectName + `_id AND
							 ` + objectName + `_propagate.ancestors_computation_state <> 'done'
					` + additionalJoin + `
				WHERE ` + relationsTable + `.child_` + singleObjectName + `_id = children.id
				FOR SHARE OF ` + relationsTable + `
				FOR UPDATE OF ` + objectName + `_propagate
				` + additionalLocking + `
			) has_undone_parents
		)`

	markAsProcessing, err := s.db.CommonDB().Prepare(markAsProcessingQuery)
	mustNotBeError(err)
	defer func() { mustNotBeError(markAsProcessing.Close()) }()

	expiresAtColumn := ""
	expiresAtValueJoin := ""
	ignore := "IGNORE"

	if objectName == groups {
		expiresAtColumn = ", expires_at"
		expiresAtValueJoin = ", LEAST(groups_ancestors_join.expires_at, groups_groups.expires_at)"
		ignore = ""
	}

	// For every object marked as 'processing', we compute all its ancestors
	recomputeQueries := make([]string, 0, 3)
	recomputeQueries = append(recomputeQueries, `
		DELETE `+objectName+`_ancestors
		FROM `+objectName+`_ancestors
			JOIN `+objectName+`_propagate
				ON `+objectName+`_propagate.id = `+objectName+`_ancestors.child_`+singleObjectName+`_id
		WHERE `+objectName+`_propagate.ancestors_computation_state = 'processing'`, `
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
		`+additionalJoin+`
		JOIN `+objectName+`_ancestors AS `+objectName+`_ancestors_join ON (
			`+objectName+`_ancestors_join.child_`+singleObjectName+`_id = `+relationsTable+`.parent_`+singleObjectName+`_id
		)
		JOIN `+objectName+`_propagate ON (
			`+relationsTable+`.child_`+singleObjectName+`_id = `+objectName+`_propagate.id
		)
		WHERE
			`+objectName+`_propagate.ancestors_computation_state = 'processing'`) // #nosec
	if objectName == groups {
		recomputeQueries[0] += `
			AND groups_ancestors.ancestor_group_id != groups_ancestors.child_group_id` // do not delete group ancestors with is_self=1
		recomputeQueries[1] += `
				AND NOW() < groups_groups.expires_at AND
				NOW() < LEAST(groups_ancestors_join.expires_at, groups_groups.expires_at)
			FOR UPDATE OF ` + objectName + `_propagate
			FOR SHARE OF ` + objectName + `_ancestors_join
			FOR SHARE OF ` + relationsTable + `
			FOR SHARE OF parent
			ON DUPLICATE KEY UPDATE
				expires_at = GREATEST(groups_ancestors.expires_at, LEAST(groups_ancestors_join.expires_at, groups_groups.expires_at))`
	} else {
		recomputeQueries[1] += `
			FOR UPDATE OF ` + objectName + `_propagate
			FOR SHARE OF ` + objectName + `_ancestors_join
			FOR SHARE OF ` + relationsTable
		recomputeQueries = append(recomputeQueries, `
			INSERT IGNORE INTO items_ancestors (ancestor_item_id, child_item_id)
			SELECT items_items.parent_item_id, items_items.child_item_id
			FROM items_items
			JOIN items_propagate ON items_items.child_item_id = items_propagate.id
			WHERE items_propagate.ancestors_computation_state = 'processing'
			FOR UPDATE OF items_propagate
			FOR SHARE OF items_items`) // #nosec
	}

	recomputeAncestors := make([]*sql.Stmt, len(recomputeQueries))
	for i := 0; i < len(recomputeQueries); i++ {
		recomputeAncestors[i], err = s.db.CommonDB().Prepare(recomputeQueries[i])
		mustNotBeError(err)

		defer func(i int) { mustNotBeError(recomputeAncestors[i].Close()) }(i)
	}

	// Objects marked as 'processing' are now marked as 'done'
	markAsDoneQuery := `
		UPDATE ` + objectName + `_propagate
		SET ancestors_computation_state = 'done'
		WHERE ancestors_computation_state = 'processing'` // #nosec
	markAsDone, err := s.db.CommonDB().Prepare(markAsDoneQuery)
	mustNotBeError(err)
	defer func() { mustNotBeError(markAsDone.Close()) }()

	for hasChanges {
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
		hasChanges = rowsAffected > 0
	}
}
