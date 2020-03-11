package database

import (
	"database/sql"
	"time"
)

const computeAllAttemptsLockName = "listener_computeAllAttempts"
const computeAllAttemptsLockTimeout = 10 * time.Second

// ComputeAllAttempts recomputes fields of attempts
// For attempts marked with result_propagation_state = 'to_be_propagated'/'to_be_recomputed':
// 1. We mark all their ancestors in attempts as 'to_be_recomputed'
//  (we consider a row in attempts as an ancestor if it has the same value in group_id and
//  its item_id is an ancestor of the original row's item_id).
// 2. We process all objects that are marked as 'to_be_recomputed' and that have no children marked as 'to_be_recomputed'.
//  Then, if an object has children, we update
//    latest_activity_at, tasks_tried, tasks_with_help, validated_at.
//  This step is repeated until no records are updated.
// 3. We insert new permissions_granted for each unlocked item according to corresponding item_unlocking_rules.
func (s *AttemptStore) ComputeAllAttempts() (err error) {
	s.mustBeInTransaction()
	defer recoverPanics(&err)

	var groupsUnlocked int64

	// Use a lock so that we don't execute the listener multiple times in parallel
	mustNotBeError(s.WithNamedLock(computeAllAttemptsLockName, computeAllAttemptsLockTimeout, func(ds *DataStore) error {
		// We mark as 'to_be_recomputed' all ancestors of objects marked as 'to_be_recomputed'/'to_be_propagated'
		// (this query can take more than 50 seconds to run when executed for the first time after the db migration)
		mustNotBeError(ds.db.Exec(
			`UPDATE attempts AS ancestors
			JOIN items_ancestors ON ancestors.item_id = items_ancestors.ancestor_item_id
			JOIN attempts AS descendants ON (
				descendants.item_id = items_ancestors.child_item_id AND
				descendants.group_id = ancestors.group_id
			)
			SET ancestors.result_propagation_state = 'to_be_recomputed'
			WHERE ancestors.result_propagation_state != 'to_be_recomputed' AND
				(descendants.result_propagation_state = 'to_be_recomputed' OR
				 descendants.result_propagation_state = 'to_be_propagated')`).Error)

		// Insert missing attempts for chapters having descendants with attempts marked as 'to_be_recomputed'/'to_be_propagated'.
		// We only create attempts for chapters which are (or have ancestors which are) visible to the group that attempted
		// to solve the descendant items. Chapters with explicit entry (items.entry_participant_type IS NOT NULL) are skipped).
		// (this query can take more than 25 seconds when executed for the first time after the db migration)
		mustNotBeError(ds.RetryOnDuplicatePrimaryKeyError(func(retryStore *DataStore) error {
			return retryStore.Exec(`
				INSERT INTO attempts (id, group_id, item_id, latest_activity_at, ` + "`order`, " + `result_propagation_state)
				SELECT
					FLOOR(RAND() * 1000000000) + FLOOR(RAND() * 1000000000) * 1000000000,
					descendants.group_id, items_ancestors.ancestor_item_id, '1000-01-01 00:00:00', 1, 'to_be_recomputed'
				FROM attempts AS descendants
				JOIN items_ancestors ON items_ancestors.child_item_id = descendants.item_id
				JOIN items ON items.id = items_ancestors.ancestor_item_id AND items.entry_participant_type IS NULL
				LEFT JOIN attempts AS existing ON (
					existing.group_id = descendants.group_id AND
					existing.item_id = items_ancestors.ancestor_item_id
				)
				WHERE (
					descendants.result_propagation_state = 'to_be_recomputed' OR
					descendants.result_propagation_state = 'to_be_propagated'
				) AND existing.id IS NULL AND (
					EXISTS(
						SELECT 1 FROM permissions_generated
						JOIN groups_ancestors_active
							ON groups_ancestors_active.ancestor_group_id = permissions_generated.group_id
						WHERE
							permissions_generated.item_id = items_ancestors.ancestor_item_id AND
							permissions_generated.can_view_generated != 'none' AND
							groups_ancestors_active.child_group_id = descendants.group_id
						LIMIT 1
					) OR EXISTS(
						SELECT 1 FROM permissions_generated
						JOIN groups_ancestors_active
							ON groups_ancestors_active.ancestor_group_id = permissions_generated.group_id
						WHERE
							permissions_generated.item_id IN (
								SELECT grand_ancestors.ancestor_item_id
								FROM items_ancestors AS grand_ancestors
								WHERE grand_ancestors.child_item_id = items_ancestors.ancestor_item_id
							) AND permissions_generated.can_view_generated != 'none' AND
							groups_ancestors_active.child_group_id = descendants.group_id
						LIMIT 1
				))
				GROUP BY descendants.group_id, items_ancestors.ancestor_item_id
			`).Error()
		}))

		hasChanges := true

		var markAsProcessingStatement, updateStatement *sql.Stmt

		for hasChanges {
			// We mark as "processing" all objects that were marked as 'to_be_recomputed' and
			// that have no children marked as 'to_be_recomputed'.
			// This way we prevent infinite looping as we never process items that are ancestors of themselves
			if markAsProcessingStatement == nil {
				const markAsProcessingQuery = `
					UPDATE attempts AS parent
					JOIN (
						SELECT *
						FROM (
							SELECT inner_parent.id
							FROM attempts AS inner_parent
							WHERE result_propagation_state = 'to_be_recomputed'
								AND NOT EXISTS (
									SELECT items_items.child_item_id
									FROM items_items
									JOIN attempts AS children
										ON children.item_id = items_items.child_item_id
									WHERE items_items.parent_item_id = inner_parent.item_id AND
										children.result_propagation_state = 'to_be_recomputed' AND
										children.group_id = inner_parent.group_id
								)
							) AS tmp2
					) AS tmp
						USING(id)
					SET result_propagation_state = 'processing'`

				markAsProcessingStatement, err = ds.db.CommonDB().Prepare(markAsProcessingQuery)
				mustNotBeError(err)
				defer func() { mustNotBeError(markAsProcessingStatement.Close()) }()
			}
			_, err = markAsProcessingStatement.Exec()
			mustNotBeError(err)

			// For every object marked as 'processing', we compute all the characteristics based on the children:
			//  - latest_activity_at as the max of children's
			//  - tasks_with_help, tasks_tried as the sum of children's per-item maximums
			//  - children_validated as the number of children items with validated == 1
			//  - validated, depending on the items_items.category and items.validation_type
			//    (an item should have at least one validated child to become validated itself by the propagation)
			if updateStatement == nil {
				const updateQuery = `
					UPDATE attempts AS target_attempts
					LEFT JOIN LATERAL (
						SELECT
							target_attempts.id,
							MAX(aggregated_children_attempts.latest_activity_at) AS latest_activity_at,
							SUM(aggregated_children_attempts.tasks_tried) AS tasks_tried,
							SUM(aggregated_children_attempts.tasks_with_help) AS tasks_with_help,
							SUM(aggregated_children_attempts.validated) AS children_validated,
							SUM(IFNULL(NOT aggregated_children_attempts.validated, 1)) AS children_non_validated,
							SUM(items_items.category = 'Validation' AND IFNULL(NOT aggregated_children_attempts.validated, 1))
								AS children_non_validated_categories,
							MAX(aggregated_children_attempts.validated_at) AS max_validated_at,
							MAX(IF(items_items.category = 'Validation', aggregated_children_attempts.validated_at, NULL))
								AS max_validated_at_categories,
							SUM(IFNULL(aggregated_children_attempts.score_computed, 0) * items_items.score_weight) /
								COALESCE(NULLIF(SUM(items_items.score_weight), 0), 1) AS average_score
						FROM items_items ` +
					// We use LEFT JOIN LATERAL to aggregate attempts grouped by target_attempts.group_id & items_items.child_item_id.
					// The usual LEFT JOIN conditions in the ON clause would group attempts before joining which would produce
					// wrong results.
					`	LEFT JOIN LATERAL (
							SELECT
								MAX(validated) AS validated,
								MIN(validated_at) AS validated_at,
								MAX(latest_activity_at) AS latest_activity_at,
								MAX(tasks_tried) AS tasks_tried,
								MAX(tasks_with_help) AS tasks_with_help,
								MAX(score_computed) AS score_computed
							FROM attempts AS children_attempts
							WHERE children_attempts.group_id = target_attempts.group_id AND
								children_attempts.item_id = items_items.child_item_id
							GROUP BY children_attempts.group_id, children_attempts.item_id
						) AS aggregated_children_attempts ON 1
						JOIN items ON(
							items.id = items_items.child_item_id
						)
						WHERE items_items.parent_item_id = target_attempts.item_id AND NOT items.no_score
						GROUP BY items_items.parent_item_id
					) AS children_stats ON 1
					JOIN items
						ON target_attempts.item_id = items.id
					SET
						target_attempts.latest_activity_at = GREATEST(
							IFNULL(children_stats.latest_activity_at, '1000-01-01 00:00:00'),
							target_attempts.latest_activity_at
						),
						target_attempts.tasks_tried = IFNULL(children_stats.tasks_tried, 0),
						target_attempts.tasks_with_help = IFNULL(children_stats.tasks_with_help, 0),
						target_attempts.validated_at = CASE
							WHEN children_stats.id IS NULL THEN NULL
							WHEN items.validation_type = 'Categories' AND children_stats.children_non_validated_categories = 0
								THEN children_stats.max_validated_at_categories
							WHEN items.validation_type = 'All' AND children_stats.children_non_validated = 0
								THEN children_stats.max_validated_at
							WHEN items.validation_type = 'AllButOne' AND children_stats.children_non_validated < 2
								THEN children_stats.max_validated_at
							WHEN items.validation_type = 'One' AND children_stats.children_validated > 0
								THEN children_stats.max_validated_at
							ELSE NULL
						END,
						target_attempts.score_computed = IF(items.no_score OR children_stats.average_score IS NULL,
							0,
							LEAST(GREATEST(CASE target_attempts.score_edit_rule
								WHEN 'set' THEN target_attempts.score_edit_value
								WHEN 'diff' THEN children_stats.average_score + target_attempts.score_edit_value
								ELSE children_stats.average_score
							END, 0), 100)),
						target_attempts.result_propagation_state = 'to_be_propagated'
					WHERE target_attempts.result_propagation_state = 'processing'`
				updateStatement, err = ds.db.CommonDB().Prepare(updateQuery)
				mustNotBeError(err)
				defer func() { mustNotBeError(updateStatement.Close()) }()
			}

			result, err := updateStatement.Exec()
			mustNotBeError(err)

			rowsAffected, err := result.RowsAffected()
			mustNotBeError(err)
			hasChanges = rowsAffected > 0
		}

		result := ds.db.Exec(`
			INSERT INTO permissions_granted
				(group_id, item_id, source_group_id, origin, can_view, latest_update_on)
				SELECT
					groups.id AS group_id,
					item_unlocking_rules.unlocked_item_id AS item_id,
					groups.id,
					'item_unlocking',
					'content',
					NOW()
				FROM attempts
				JOIN item_unlocking_rules ON item_unlocking_rules.unlocking_item_id = attempts.item_id AND
					item_unlocking_rules.score <= attempts.score_computed
				JOIN ` + "`groups`" + ` ON attempts.group_id = groups.id
				WHERE attempts.result_propagation_state = 'to_be_propagated'
			ON DUPLICATE KEY UPDATE
				latest_update_on = IF(can_view = 'content', latest_update_on, NOW()),
				can_view = 'content'`)

		mustNotBeError(result.Error)
		groupsUnlocked += result.RowsAffected

		return ds.db.Exec(`
			UPDATE attempts SET result_propagation_state = 'done'
				WHERE result_propagation_state = 'to_be_propagated'`).Error
	}))

	// If items have been unlocked, need to recompute access
	if groupsUnlocked > 0 {
		// generate permissions_generated from permissions_granted
		mustNotBeError(s.PermissionsGranted().After())
		// we should compute attempts again as new permissions were set and
		// triggers on permissions_generated likely marked some attempts as 'to_be_propagated'
		return s.ComputeAllAttempts()
	}
	return nil
}
