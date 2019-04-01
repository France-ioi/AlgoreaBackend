package database

import "database/sql"

// createNewAncestors creates inserts new rows into
// the objectName_ancestors table (items_ancestors or groups_ancestor)
// for all rows marked with sAncestorsComputationState="todo" in objectName_propagate
// (items_propagate or groups_propagate) and their descendants
//
//
func (s *DataStore) createNewAncestors(objectName, upObjectName string) {
	// We mark as 'todo' all descendants of objects marked as 'todo'
	query := `
		INSERT INTO  ` + objectName + `_propagate (ID, sAncestorsComputationState)
		SELECT descendants.ID, 'todo'
		FROM ` + objectName + ` as descendants
		JOIN ` + objectName + `_ancestors
		ON (
			descendants.ID = ` + objectName + `_ancestors.id` + upObjectName + `Child
		)
		JOIN ` + objectName + `_propagate AS ancestors
		ON (
			ancestors.ID = ` + objectName + `_ancestors.id` + upObjectName + `Ancestor
		)
		WHERE ancestors.sAncestorsComputationState = 'todo'
		ON DUPLICATE KEY UPDATE sAncestorsComputationState = 'todo'`

	mustNotBeError(s.db.Exec(query).Error)
	hasChanges := true

	groupsAcceptedCondition := ""
	if objectName == "groups" {
		groupsAcceptedCondition = ` AND (
			groups_groups.sType = 'invitationAccepted' OR
			groups_groups.sType = 'requestAccepted' OR
			groups_groups.sType = 'direct'
		)`
	}

	relationsTable := objectName + "_" + objectName

	// Next queries will be executed in the loop

	// We mark as "processing" all objects that were marked as 'todo' and that have no parents not marked as 'done'
	// This way we prevent infinite looping as we never process objects that are descendants of themselves
	/*
		// TODO: this query is super slow (> 2.5s sometimes)
		query = `
				UPDATE ` + objectName + `_propagate AS children
				SET
					sAncestorsComputationState = 'processing'
				WHERE
					sAncestorsComputationState = 'todo' AND
					children.ID NOT IN (
						SELECT id` + upObjectName + `Child
						FROM (
							SELECT ` + relationsTable + `.id` + upObjectName + `Child
							FROM ` + relationsTable + `
							JOIN ` + objectName + `_propagate AS parents ON (
								parents.ID = ` + relationsTable + `.id` + upObjectName + `Parent
							)
							WHERE parents.sAncestorsComputationState <> 'done'` + groupsAcceptedCondition + `
						) AS notready
					)`
	*/

	query = `
		UPDATE ` + objectName + `_propagate AS children
		LEFT JOIN (
			SELECT
				` + relationsTable + `.id` + upObjectName + `Child,
				MIN(parents.sAncestorsComputationState) AS min_parents_state,
				MAX(parents.sAncestorsComputationState) AS max_parents_state
			FROM ` + relationsTable + `
			JOIN ` + objectName + `_propagate AS parents ON (
				parents.ID = ` + relationsTable + `.id` + upObjectName + `Parent
			) WHERE 1` + groupsAcceptedCondition + `
			GROUP BY id` + upObjectName + `Child
		) AS not_ready ON (
			not_ready.id` + upObjectName + `Child=children.ID AND
			(
				not_ready.min_parents_state <> 'done' OR
				not_ready.max_parents_state <>'done'
			)
		)
		SET sAncestorsComputationState = 'processing'
		WHERE
			sAncestorsComputationState = 'todo' AND
			not_ready.id` + upObjectName + `Child IS NULL`
	markAsProcessing, err := s.db.CommonDB().Prepare(query)
	mustNotBeError(err)
	defer func() { mustNotBeError(markAsProcessing.Close()) }()

	bIsSelfColumn := ""
	bIsSelfValue := ""
	if objectName == "groups" {
		bIsSelfColumn = ", bIsSelf"
		bIsSelfValue = ", '0' AS bIsSelf"
	}

	// For every object marked as 'processing', we compute all its ancestors
	query = `
		INSERT IGNORE INTO ` + objectName + `_ancestors
		(
			id` + upObjectName + `Ancestor,
			id` + upObjectName + `Child` + `
			` + bIsSelfColumn + `
		)
		SELECT
			` + relationsTable + `.id` + upObjectName + `Parent,
			` + relationsTable + `.id` + upObjectName + `Child
			` + bIsSelfValue + `
		FROM ` + relationsTable + `
		JOIN ` + objectName + `_propagate
		ON (
			` + relationsTable + `.id` + upObjectName + `Child = ` + objectName + `_propagate.ID OR
			` + relationsTable + `.id` + upObjectName + `Parent = ` + objectName + `_propagate.ID
		)
		WHERE
			` + objectName + `_propagate.sAncestorsComputationState = 'processing'` + groupsAcceptedCondition + `
		UNION ALL
		SELECT
			` + objectName + `_ancestors.id` + upObjectName + `Ancestor,
			` + relationsTable + `_join.id` + upObjectName + `Child
			` + bIsSelfValue + `
		FROM ` + objectName + `_ancestors
		JOIN ` + relationsTable + ` AS ` + relationsTable + `_join ON (
			` + relationsTable + `_join.id` + upObjectName + `Parent = ` + objectName + `_ancestors.id` + upObjectName + `Child
		)
		JOIN ` + objectName + `_propagate ON (
			` + relationsTable + `_join.id` + upObjectName + `Child = ` + objectName + `_propagate.ID
		)
		WHERE 
			` + objectName + `_propagate.sAncestorsComputationState = 'processing'`
	if objectName == "groups" {
		query += `
			AND (
				groups_groups_join.sType = 'invitationAccepted' OR
				groups_groups_join.sType = 'requestAccepted' OR
				groups_groups_join.sType = 'direct'
			)
		UNION ALL
		SELECT
			` + `groups_propagate.ID AS idGroupAncestor,
			` + `groups_propagate.ID AS idGroupChild,
			'1' AS bIsSelf
		FROM groups_propagate
		WHERE groups_propagate.sAncestorsComputationState = 'processing'`
	}

	insertAncestors, err := s.db.CommonDB().Prepare(query)
	mustNotBeError(err)
	defer func() { mustNotBeError(insertAncestors.Close()) }()

	// Objects marked as 'processing' are now marked as 'done'
	query = `
		UPDATE ` + objectName + `_propagate
		SET sAncestorsComputationState = 'done'
		WHERE sAncestorsComputationState = 'processing'`
	markAsDone, err := s.db.CommonDB().Prepare(query)
	mustNotBeError(err)
	defer func() { mustNotBeError(markAsDone.Close()) }()

	for hasChanges {
		_, err = markAsProcessing.Exec()
		mustNotBeError(err)
		_, err = insertAncestors.Exec()
		mustNotBeError(err)

		var result sql.Result
		result, err = markAsDone.Exec()
		mustNotBeError(err)
		var rowsAffected int64
		rowsAffected, err = result.RowsAffected()
		mustNotBeError(err)
		hasChanges = rowsAffected > 0
	}
}
