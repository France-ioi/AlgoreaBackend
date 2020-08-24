package database

import (
	"database/sql"
	"time"
)

const propagateLockName = "listener_propagate"
const propagateLockTimeout = 10 * time.Second

// Propagate recomputes fields of results
// For results marked with result_propagation_state = 'to_be_propagated'/'to_be_recomputed':
// 1. We mark all their ancestors in results as 'to_be_recomputed'
//  (we consider a row in results as an ancestor if
//    a) it has the same value in group_id
//    b) its item_id is an ancestor of the original row's item_id
//    c) its attempt_id is equal to the original row's attempt_id for original rows with root_item_id != item_id or
//       its attempt_id is equal to the original row's parent_attempt_id for original rows with root_item_id = item_id).
// 2. We process all objects that are marked as 'to_be_recomputed' and that have no children marked as 'to_be_recomputed'.
//  Then, if an object has children, we update
//    latest_activity_at, tasks_tried, tasks_with_help, validated_at.
//  This step is repeated until no records are updated.
// 3. We insert new permissions_granted for each unlocked item according to corresponding item_dependencies.
func (s *ResultStore) Propagate() (err error) {
	s.mustBeInTransaction()
	defer recoverPanics(&err)

	var groupsUnlocked int64

	// Use a lock so that we don't execute the listener multiple times in parallel
	mustNotBeError(s.WithNamedLock(propagateLockName, propagateLockTimeout, func(ds *DataStore) error {
		// We mark as 'to_be_recomputed' results of all ancestors of items marked as 'to_be_recomputed'/'to_be_propagated'
		// with appropriate attempt_id.
		// Also, we insert missing results for chapters having descendants with results marked as 'to_be_recomputed'/'to_be_propagated'.
		// We only create results for chapters which are (or have ancestors which are) visible to the group that attempted
		// to solve the descendant items. Chapters requiring explicit entry or placed outside of the scope
		// of the attempts's root item are skipped).
		// (This query can take more than 30 seconds to run when executed for the first time after the db migration)
		mustNotBeError(ds.Exec(`
			INSERT INTO results (participant_id, attempt_id, item_id, latest_activity_at, result_propagation_state)
			WITH RECURSIVE results_to_insert (participant_id, attempt_id, item_id, result_exists, result_propagation_state) AS (
					SELECT results.participant_id,
								 IF(attempts.root_item_id = results.item_id, attempts.parent_attempt_id, results.attempt_id) AS attempt_id,
								 items_items.parent_item_id AS item_id,
								 existing.participant_id IS NOT NULL AS result_exists,
								 existing.result_propagation_state
					FROM results
					JOIN attempts ON attempts.participant_id = results.participant_id AND attempts.id = results.attempt_id
					JOIN items_items ON items_items.child_item_id = results.item_id
					JOIN items ON items.id = items_items.parent_item_id
					LEFT JOIN results AS existing
						ON existing.participant_id = results.participant_id AND
							 existing.attempt_id = IF(attempts.root_item_id = results.item_id, attempts.parent_attempt_id, results.attempt_id) AND
							 existing.item_id = items_items.parent_item_id
					WHERE NOT (items.requires_explicit_entry AND existing.participant_id IS NULL) AND
								(existing.result_propagation_state IS NULL OR existing.result_propagation_state != 'to_be_propagated') AND
								(results.result_propagation_state = 'to_be_recomputed' OR results.result_propagation_state = 'to_be_propagated')
				UNION
					SELECT results_to_insert.participant_id,
								 IF(attempts.root_item_id = results_to_insert.item_id, attempts.parent_attempt_id, results_to_insert.attempt_id) AS attempt_id,
								 items_items.parent_item_id AS item_id,
								 existing.participant_id IS NOT NULL AS result_exists,
								 existing.result_propagation_state
					FROM results_to_insert
					JOIN attempts ON attempts.participant_id = results_to_insert.participant_id AND attempts.id = results_to_insert.attempt_id
					JOIN items_items ON items_items.child_item_id = results_to_insert.item_id
					JOIN items ON items.id = items_items.parent_item_id
					LEFT JOIN results AS existing
						ON existing.participant_id = results_to_insert.participant_id AND
							 existing.attempt_id =
						     IF(attempts.root_item_id = results_to_insert.item_id, attempts.parent_attempt_id, results_to_insert.attempt_id) AND
							 existing.item_id = items_items.parent_item_id
					WHERE NOT (items.requires_explicit_entry AND existing.participant_id IS NULL) AND
								(existing.result_propagation_state IS NULL OR existing.result_propagation_state != 'to_be_propagated')
			)
			SELECT
				results_to_insert.participant_id, results_to_insert.attempt_id, results_to_insert.item_id, '1000-01-01 00:00:00', 'to_be_recomputed'
			FROM results_to_insert
			JOIN attempts ON attempts.participant_id = results_to_insert.participant_id AND attempts.id = results_to_insert.attempt_id
			LEFT JOIN items_ancestors AS root_item_descendant
				ON root_item_descendant.ancestor_item_id = attempts.root_item_id AND root_item_descendant.child_item_id = results_to_insert.item_id
			WHERE result_exists OR ((
				EXISTS(
					SELECT 1 FROM permissions_generated
					JOIN groups_ancestors_active
						ON groups_ancestors_active.ancestor_group_id = permissions_generated.group_id
					WHERE
						permissions_generated.item_id = results_to_insert.item_id AND
						permissions_generated.can_view_generated != 'none' AND
						groups_ancestors_active.child_group_id = results_to_insert.participant_id
					LIMIT 1
				) OR EXISTS(
					SELECT 1 FROM permissions_generated
					JOIN groups_ancestors_active
						ON groups_ancestors_active.ancestor_group_id = permissions_generated.group_id
					WHERE
						permissions_generated.item_id IN (
							SELECT grand_ancestors.ancestor_item_id
							FROM items_ancestors AS grand_ancestors
							WHERE grand_ancestors.child_item_id = results_to_insert.item_id
						) AND permissions_generated.can_view_generated != 'none' AND
						groups_ancestors_active.child_group_id = results_to_insert.participant_id
					LIMIT 1
			)) AND (
				attempts.root_item_id IS NULL OR attempts.root_item_id = results_to_insert.item_id OR
				root_item_descendant.ancestor_item_id IS NOT NULL))
			GROUP BY results_to_insert.participant_id, results_to_insert.attempt_id, results_to_insert.item_id
			ON DUPLICATE KEY UPDATE result_propagation_state = 'to_be_recomputed'
		`).Error())

		hasChanges := true

		var markAsProcessingStatement, updateStatement *sql.Stmt

		for hasChanges {
			// We mark as "processing" all objects that were marked as 'to_be_recomputed' and
			// that have no children (within the attempt or child attempts) marked as 'to_be_recomputed'.
			// This way we prevent infinite looping as we never process items that are ancestors of themselves
			if markAsProcessingStatement == nil {
				const markAsProcessingQuery = `
					UPDATE results AS parent
					JOIN (
						SELECT *
						FROM (
							SELECT inner_parent.participant_id, inner_parent.attempt_id, inner_parent.item_id
							FROM results AS inner_parent
							WHERE result_propagation_state = 'to_be_recomputed' AND
								NOT EXISTS (
									SELECT items_items.child_item_id
									FROM items_items
									JOIN results AS children
										ON children.item_id = items_items.child_item_id
									WHERE items_items.parent_item_id = inner_parent.item_id AND
										children.participant_id = inner_parent.participant_id AND
										children.attempt_id = inner_parent.attempt_id AND
										children.result_propagation_state = 'to_be_recomputed'
								) AND NOT EXISTS (
									SELECT items_items.child_item_id
									FROM items_items
									JOIN attempts
										ON attempts.root_item_id = items_items.child_item_id
									JOIN results AS children
										ON children.item_id = items_items.child_item_id AND
										   children.attempt_id = attempts.id
									WHERE items_items.parent_item_id = inner_parent.item_id AND
										attempts.participant_id = inner_parent.participant_id AND
										attempts.parent_attempt_id = inner_parent.attempt_id AND
										children.participant_id = inner_parent.participant_id AND
										children.result_propagation_state = 'to_be_recomputed'
								)
							) AS tmp2
					) AS tmp
						USING(participant_id, attempt_id, item_id)
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
					UPDATE results AS target_results
					LEFT JOIN LATERAL (
						SELECT
							target_results.participant_id,
							MAX(aggregated_children_results.latest_activity_at) AS latest_activity_at,
							SUM(aggregated_children_results.tasks_tried) AS tasks_tried,
							SUM(aggregated_children_results.tasks_with_help) AS tasks_with_help,
							SUM(aggregated_children_results.validated) AS children_validated,
							SUM(IFNULL(NOT aggregated_children_results.validated, 1)) AS children_non_validated,
							SUM(items_items.category = 'Validation' AND IFNULL(NOT aggregated_children_results.validated, 1))
								AS children_non_validated_categories,
							MAX(aggregated_children_results.validated_at) AS max_validated_at,
							MAX(IF(items_items.category = 'Validation', aggregated_children_results.validated_at, NULL))
								AS max_validated_at_categories,
							SUM(IFNULL(aggregated_children_results.score_computed, 0) * items_items.score_weight) /
								COALESCE(NULLIF(SUM(items_items.score_weight), 0), 1) AS average_score
						FROM items_items ` +
					// We use LEFT JOIN LATERAL to aggregate results grouped by target_results.participant_id & items_items.child_item_id.
					// The usual LEFT JOIN conditions in the ON clause would group results before joining which would produce
					// wrong results.
					`	LEFT JOIN LATERAL (
							SELECT
								MAX(validated) AS validated,
								MIN(validated_at) AS validated_at,
								MAX(latest_activity_at) AS latest_activity_at,
								MAX(tasks_tried) AS tasks_tried,
								MAX(tasks_with_help) AS tasks_with_help,
								MAX(score_computed) AS score_computed
							FROM results AS children_results
							JOIN attempts
								ON attempts.participant_id = children_results.participant_id AND
								   attempts.id = children_results.attempt_id
							WHERE children_results.participant_id = target_results.participant_id AND
								children_results.item_id = items_items.child_item_id AND
							  (children_results.attempt_id = target_results.attempt_id OR
							    (attempts.root_item_id = items_items.child_item_id AND
									 attempts.parent_attempt_id = target_results.attempt_id))
							GROUP BY children_results.participant_id, children_results.item_id
						) AS aggregated_children_results ON 1
						JOIN items ON(
							items.id = items_items.child_item_id
						)
						WHERE items_items.parent_item_id = target_results.item_id AND NOT items.no_score
						GROUP BY items_items.parent_item_id
					) AS children_stats ON 1
					JOIN items
						ON target_results.item_id = items.id
					SET
						target_results.latest_activity_at = GREATEST(
							IFNULL(children_stats.latest_activity_at, '1000-01-01 00:00:00'),
							target_results.latest_activity_at
						),
						target_results.tasks_tried = IFNULL(children_stats.tasks_tried, 0),
						target_results.tasks_with_help = IFNULL(children_stats.tasks_with_help, 0),
						target_results.validated_at = CASE
							WHEN children_stats.participant_id IS NULL THEN NULL
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
						target_results.score_computed = IF(items.no_score OR children_stats.average_score IS NULL,
							0,
							LEAST(GREATEST(CASE target_results.score_edit_rule
								WHEN 'set' THEN target_results.score_edit_value
								WHEN 'diff' THEN children_stats.average_score + target_results.score_edit_value
								ELSE children_stats.average_score
							END, 0), 100)),
						target_results.result_propagation_state = 'to_be_propagated'
					WHERE target_results.result_propagation_state = 'processing'`
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
				(group_id, item_id, source_group_id, origin, can_view, latest_update_at)
				SELECT
					groups.id AS group_id,
					item_dependencies.dependent_item_id AS item_id,
					groups.id,
					'item_unlocking',
					'content',
					NOW()
				FROM results
				JOIN item_dependencies ON item_dependencies.item_id = results.item_id AND
					item_dependencies.score <= results.score_computed AND item_dependencies.grant_content_view
				JOIN ` + "`groups`" + ` ON groups.id = results.participant_id
				WHERE results.result_propagation_state = 'to_be_propagated'
			ON DUPLICATE KEY UPDATE
				latest_update_at = IF(can_view = 'content', latest_update_at, NOW()),
				can_view = 'content'`)

		mustNotBeError(result.Error)
		groupsUnlocked += result.RowsAffected

		return ds.db.Exec(`
			UPDATE results SET result_propagation_state = 'done'
				WHERE result_propagation_state = 'to_be_propagated'`).Error
	}))

	// If items have been unlocked, need to recompute access
	if groupsUnlocked > 0 {
		// generate permissions_generated from permissions_granted
		mustNotBeError(s.PermissionsGranted().After())
		// we should compute attempts again as new permissions were set and
		// triggers on permissions_generated likely marked some attempts as 'to_be_propagated'
		return s.Propagate()
	}
	return nil
}
