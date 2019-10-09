package database

import (
	"database/sql"
)

const groups = "groups"

// createNewAncestors inserts new rows into
// the objectName_ancestors table (items_ancestors or groups_ancestor)
// for all rows marked with ancestors_computation_state="todo" in objectName_propagate
// (items_propagate or groups_propagate) and their descendants
func (s *DataStore) createNewAncestors(objectName, singleObjectName string) { /* #nosec */
	// We mark as 'todo' all descendants of objects marked as 'todo'
	query := `
		INSERT INTO  ` + objectName + `_propagate (id, ancestors_computation_state)
		SELECT descendants.id, 'todo'
		FROM ` + QuoteName(objectName) + ` AS descendants
		JOIN ` + QuoteName(objectName+"_ancestors") + `
			ON descendants.id = ` + QuoteName(objectName+"_ancestors") + ".child_" + singleObjectName + `_id
		JOIN ` + QuoteName(objectName+"_propagate") + ` AS ancestors
			ON ancestors.id = ` + QuoteName(objectName+"_ancestors") + ".ancestor_" + singleObjectName + `_id
		WHERE ancestors.ancestors_computation_state = 'todo'
		ON DUPLICATE KEY UPDATE ancestors_computation_state = 'todo'`

	mustNotBeError(s.db.Exec(query).Error)
	hasChanges := true

	groupsAcceptedCondition := ""
	if objectName == groups {
		groupsAcceptedCondition = " AND (groups_groups.type" + GroupRelationIsActiveCondition + ") AND NOW() < groups_groups.expires_at"
	}

	relationsTable := objectName + "_" + objectName

	// Next queries will be executed in the loop

	// We mark as "processing" all objects that were marked as 'todo' and that have no parents not marked as 'done'
	// This way we prevent infinite looping as we never process objects that are descendants of themselves

	/* #nosec */
	query = `
		UPDATE ` + objectName + `_propagate AS children
		LEFT JOIN ` + relationsTable + `
			ON ` + relationsTable + `.child_` + singleObjectName + `_id = children.id ` + groupsAcceptedCondition + `
		LEFT JOIN ` + objectName + `_propagate AS parents
			ON parents.id = ` + relationsTable + `.parent_` + singleObjectName + `_id AND parents.ancestors_computation_state <> 'done'
		SET children.ancestors_computation_state='processing'
		WHERE children.ancestors_computation_state = 'todo' AND parents.id IS NULL`
	markAsProcessing, err := s.db.CommonDB().Prepare(query)
	mustNotBeError(err)
	defer func() { mustNotBeError(markAsProcessing.Close()) }()

	isSelfColumn := ""
	isSelfValue := ""
	expiresAtColumn := ""
	expiresAtValue := ""
	expiresAtValueJoin := ""

	if objectName == groups {
		isSelfColumn = ", is_self"
		isSelfValue = ", '0' AS is_self"
		expiresAtColumn = ", expires_at"
		expiresAtValue = ", groups_groups.expires_at"
		expiresAtValueJoin = ", LEAST(groups_ancestors_select.expires_at, groups_groups_join.expires_at)"
	}

	// For every object marked as 'processing', we compute all its ancestors
	insertQueries := make([]string, 0, 4)
	insertQueries = append(insertQueries, `
		INSERT IGNORE INTO `+objectName+`_ancestors
		(
			ancestor_`+singleObjectName+`_id,
			child_`+singleObjectName+`_id`+`
			`+isSelfColumn+expiresAtColumn+`
		)
		SELECT
			`+relationsTable+`.parent_`+singleObjectName+`_id,
			`+relationsTable+`.child_`+singleObjectName+`_id
			`+isSelfValue+expiresAtValue+`
		FROM `+relationsTable+`
		JOIN `+objectName+`_propagate
		ON (
			`+relationsTable+`.child_`+singleObjectName+`_id = `+objectName+`_propagate.id
		)
		WHERE
			`+objectName+`_propagate.ancestors_computation_state = 'processing'`+groupsAcceptedCondition, `
		INSERT IGNORE INTO `+objectName+`_ancestors
		(
			ancestor_`+singleObjectName+`_id,
			child_`+singleObjectName+`_id`+`
			`+isSelfColumn+expiresAtColumn+`
		)
		SELECT
			`+relationsTable+`.parent_`+singleObjectName+`_id,
			`+relationsTable+`.child_`+singleObjectName+`_id
			`+isSelfValue+expiresAtValue+`
		FROM `+relationsTable+`
		JOIN `+objectName+`_propagate
		ON (
			`+relationsTable+`.parent_`+singleObjectName+`_id = `+objectName+`_propagate.id
		)
		WHERE
			`+objectName+`_propagate.ancestors_computation_state = 'processing'`+groupsAcceptedCondition, `
		INSERT IGNORE INTO `+objectName+`_ancestors
		(
			ancestor_`+singleObjectName+`_id,
			child_`+singleObjectName+`_id`+`
			`+isSelfColumn+expiresAtColumn+`
		)
		SELECT
			`+objectName+`_ancestors_select.ancestor_`+singleObjectName+`_id,
			`+relationsTable+`_join.child_`+singleObjectName+`_id
			`+isSelfValue+expiresAtValueJoin+`
		FROM `+objectName+`_ancestors AS `+objectName+`_ancestors_select
		JOIN `+relationsTable+` AS `+relationsTable+`_join ON (
			`+relationsTable+`_join.parent_`+singleObjectName+`_id = `+objectName+`_ancestors_select.child_`+singleObjectName+`_id
		)
		JOIN `+objectName+`_propagate ON (
			`+relationsTable+`_join.child_`+singleObjectName+`_id = `+objectName+`_propagate.id
		)
		WHERE
			`+objectName+`_propagate.ancestors_computation_state = 'processing'`) // #nosec
	if objectName == groups {
		insertQueries[2] += `
				AND (groups_groups_join.type` + GroupRelationIsActiveCondition + `) AND NOT groups_ancestors_select.is_self
			ON DUPLICATE KEY UPDATE
				expires_at = GREATEST(groups_ancestors.expires_at, LEAST(groups_ancestors_select.expires_at, groups_groups_join.expires_at))`
		insertQueries = append(insertQueries, `
			INSERT IGNORE INTO `+objectName+`_ancestors
			(
				ancestor_`+singleObjectName+`_id,
				child_`+singleObjectName+`_id`+`
				`+isSelfColumn+`
			)
			SELECT
				groups_propagate.id AS ancestor_group_id,
				groups_propagate.id AS child_group_id,
				'1' AS is_self
			FROM groups_propagate
			WHERE groups_propagate.ancestors_computation_state = 'processing'`) // #nosec
	}

	insertAncestors := make([]*sql.Stmt, len(insertQueries))
	for i := 0; i < len(insertQueries); i++ {
		insertAncestors[i], err = s.db.CommonDB().Prepare(insertQueries[i])
		mustNotBeError(err)
		defer func(i int) { mustNotBeError(insertAncestors[i].Close()) }(i)
	}

	// Objects marked as 'processing' are now marked as 'done'
	query = `
		UPDATE ` + objectName + `_propagate
		SET ancestors_computation_state = 'done'
		WHERE ancestors_computation_state = 'processing'` // #nosec
	markAsDone, err := s.db.CommonDB().Prepare(query)
	mustNotBeError(err)
	defer func() { mustNotBeError(markAsDone.Close()) }()

	for hasChanges {
		_, err = markAsProcessing.Exec()
		mustNotBeError(err)
		for i := 0; i < len(insertAncestors); i++ {
			_, err = insertAncestors[i].Exec()
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
