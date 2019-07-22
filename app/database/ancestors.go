package database

import "database/sql"

const groups = "groups"

// createNewAncestors inserts new rows into
// the objectName_ancestors table (items_ancestors or groups_ancestor)
// for all rows marked with sAncestorsComputationState="todo" in objectName_propagate
// (items_propagate or groups_propagate) and their descendants
func (s *DataStore) createNewAncestors(objectName, upObjectName string) { /* #nosec */
	// We mark as 'todo' all descendants of objects marked as 'todo'
	query := `
		INSERT INTO  ` + objectName + `_propagate (ID, sAncestorsComputationState)
		SELECT descendants.ID, 'todo'
		FROM ` + objectName + ` as descendants
		JOIN ` + objectName + `_ancestors
			ON descendants.ID = ` + objectName + `_ancestors.id` + upObjectName + `Child
		JOIN ` + objectName + `_propagate AS ancestors
			ON ancestors.ID = ` + objectName + `_ancestors.id` + upObjectName + `Ancestor
		WHERE ancestors.sAncestorsComputationState = 'todo'
		ON DUPLICATE KEY UPDATE sAncestorsComputationState = 'todo'`

	mustNotBeError(s.db.Exec(query).Error)
	hasChanges := true

	groupsAcceptedCondition := ""
	if objectName == groups {
		groupsAcceptedCondition = ` AND (
			groups_groups.sType IN('invitationAccepted', 'requestAccepted','direct')
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
							JOIN ` + objectName + `_propagate AS parents
								ON parents.ID = ` + relationsTable + `.id` + upObjectName + `Parent
							WHERE parents.sAncestorsComputationState <> 'done'` + groupsAcceptedCondition + `
						) AS notready
					)`
	*/

	/* #nosec */
	query = `
		UPDATE ` + objectName + `_propagate AS children
		LEFT JOIN ` + relationsTable + `
			ON ` + relationsTable + `.id` + upObjectName + `Child = children.ID ` + groupsAcceptedCondition + `
		LEFT JOIN ` + objectName + `_propagate AS parents
			ON parents.ID = ` + relationsTable + `.id` + upObjectName + `Parent AND parents.sAncestorsComputationState <> 'done'
		SET children.sAncestorsComputationState='processing'
		WHERE children.sAncestorsComputationState = 'todo' AND parents.ID IS NULL`
	markAsProcessing, err := s.db.CommonDB().Prepare(query)
	mustNotBeError(err)
	defer func() { mustNotBeError(markAsProcessing.Close()) }()

	bIsSelfColumn := ""
	bIsSelfValue := ""
	if objectName == groups {
		bIsSelfColumn = ", bIsSelf"
		bIsSelfValue = ", '0' AS bIsSelf"
	}

	// For every object marked as 'processing', we compute all its ancestors
	insertQueries := make([]string, 0, 4)
	insertQueries = append(insertQueries, `
		INSERT IGNORE INTO `+objectName+`_ancestors
		(
			id`+upObjectName+`Ancestor,
			id`+upObjectName+`Child`+`
			`+bIsSelfColumn+`
		)
		SELECT
			`+relationsTable+`.id`+upObjectName+`Parent,
			`+relationsTable+`.id`+upObjectName+`Child
			`+bIsSelfValue+`
		FROM `+relationsTable+`
		JOIN `+objectName+`_propagate
		ON (
			`+relationsTable+`.id`+upObjectName+`Child = `+objectName+`_propagate.ID
		)
		WHERE
			`+objectName+`_propagate.sAncestorsComputationState = 'processing'`+groupsAcceptedCondition, `
		INSERT IGNORE INTO `+objectName+`_ancestors
		(
			id`+upObjectName+`Ancestor,
			id`+upObjectName+`Child`+`
			`+bIsSelfColumn+`
		)
		SELECT
			`+relationsTable+`.id`+upObjectName+`Parent,
			`+relationsTable+`.id`+upObjectName+`Child
			`+bIsSelfValue+`
		FROM `+relationsTable+`
		JOIN `+objectName+`_propagate
		ON (
			`+relationsTable+`.id`+upObjectName+`Parent = `+objectName+`_propagate.ID
		)
		WHERE
			`+objectName+`_propagate.sAncestorsComputationState = 'processing'`+groupsAcceptedCondition, `
		INSERT IGNORE INTO `+objectName+`_ancestors
		(
			id`+upObjectName+`Ancestor,
			id`+upObjectName+`Child`+`
			`+bIsSelfColumn+`
		)
		SELECT
			`+objectName+`_ancestors.id`+upObjectName+`Ancestor,
			`+relationsTable+`_join.id`+upObjectName+`Child
			`+bIsSelfValue+`
		FROM `+objectName+`_ancestors
		JOIN `+relationsTable+` AS `+relationsTable+`_join ON (
			`+relationsTable+`_join.id`+upObjectName+`Parent = `+objectName+`_ancestors.id`+upObjectName+`Child
		)
		JOIN `+objectName+`_propagate ON (
			`+relationsTable+`_join.id`+upObjectName+`Child = `+objectName+`_propagate.ID
		)
		WHERE 
			`+objectName+`_propagate.sAncestorsComputationState = 'processing'`) // #nosec
	if objectName == groups {
		insertQueries[2] += `
			AND (
				groups_groups_join.sType IN('invitationAccepted', 'requestAccepted', 'direct')
			)`
		insertQueries = append(insertQueries, `
			INSERT IGNORE INTO `+objectName+`_ancestors
			(
				id`+upObjectName+`Ancestor,
				id`+upObjectName+`Child`+`
				`+bIsSelfColumn+`
			)
			SELECT
				groups_propagate.ID AS idGroupAncestor,
				groups_propagate.ID AS idGroupChild,
				'1' AS bIsSelf
			FROM groups_propagate
			WHERE groups_propagate.sAncestorsComputationState = 'processing'`) // #nosec
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
		SET sAncestorsComputationState = 'done'
		WHERE sAncestorsComputationState = 'processing'` // #nosec
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
