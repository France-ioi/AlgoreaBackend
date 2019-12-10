package database

import (
	"database/sql"
	"time"
)

const computeAllGroupAttemptsLockName = "listener_computeAllGroupAttempts"
const computeAllGroupAttemptsLockTimeout = 10 * time.Second

// ComputeAllGroupAttempts recomputes fields of groups_attempts
// For groups_attempts marked with ancestors_computation_state = 'todo':
// 1. We mark all their ancestors in groups_attempts as 'todo'
//  (we consider a row in groups_attempts as an ancestor if it has the same value in group_id and
//  its item_id is an ancestor of the original row's item_id).
// 2. We process all objects that were marked as 'todo' and that have no children not marked as 'done'.
//  Then, if an object has children, we update
//    latest_activity_at, tasks_tried, tasks_with_help, tasks_solved, children_validated, validated_at.
//  This step is repeated until no records are updated.
// 3. We insert new permissions_granted for each unlocked item according to corresponding item_unlocking_rules.
func (s *GroupAttemptStore) ComputeAllGroupAttempts() (err error) {
	s.mustBeInTransaction()
	defer recoverPanics(&err)

	var groupsUnlocked int64

	// Use a lock so that we don't execute the listener multiple times in parallel
	mustNotBeError(s.WithNamedLock(computeAllGroupAttemptsLockName, computeAllGroupAttemptsLockTimeout, func(ds *DataStore) error {
		// We mark as 'todo' all ancestors of objects marked as 'todo'
		// (this query can take more than 50 seconds to run when executed for the first time after the db migration)
		mustNotBeError(ds.db.Exec(
			`UPDATE groups_attempts AS ancestors
			JOIN items_ancestors ON (
				ancestors.item_id = items_ancestors.ancestor_item_id AND
				items_ancestors.ancestor_item_id != items_ancestors.child_item_id
			)
			JOIN groups_attempts AS descendants ON (
				descendants.item_id = items_ancestors.child_item_id AND
				descendants.group_id = ancestors.group_id
			)
			SET ancestors.ancestors_computation_state = 'todo'
			WHERE descendants.ancestors_computation_state = 'todo'`).Error)

		hasChanges := true

		var markAsProcessingStatement, updateStatement, unlockStatement, markAsDoneStatement *sql.Stmt

		for hasChanges {
			// We mark as "processing" all objects that were marked as 'todo' and that have no children not marked as 'done'
			// This way we prevent infinite looping as we never process items that are ancestors of themselves
			if markAsProcessingStatement == nil {
				const markAsProcessingQuery = `
					UPDATE groups_attempts AS parent
					JOIN (
						SELECT *
						FROM (
							SELECT inner_parent.id
							FROM groups_attempts AS inner_parent
							WHERE ancestors_computation_state = 'todo'
								AND NOT EXISTS (
									SELECT items_items.child_item_id
									FROM items_items
									JOIN groups_attempts AS children
										ON children.item_id = items_items.child_item_id
									WHERE items_items.parent_item_id = inner_parent.item_id AND
										children.ancestors_computation_state <> 'done' AND
										children.group_id = inner_parent.group_id
								)
							) AS tmp2
					) AS tmp
						USING(id)
					SET ancestors_computation_state = 'processing'`

				markAsProcessingStatement, err = ds.db.CommonDB().Prepare(markAsProcessingQuery)
				mustNotBeError(err)
				defer func() { mustNotBeError(markAsProcessingStatement.Close()) }()
			}
			_, err = markAsProcessingStatement.Exec()
			mustNotBeError(err)

			// For every object marked as 'processing', we compute all the characteristics based on the children:
			//  - latest_activity_at as the max of children's
			//  - tasks_with_help, tasks_tried, nbTaskSolved as the sum of children's per-item maximums
			//  - children_validated as the number of children items with validated == 1
			//  - validated, depending on the items_items.category and items.validation_type
			//    (an item should have at least one validated child to become validated itself by the propagation)
			if updateStatement == nil {
				const updateQuery = `
					UPDATE groups_attempts AS target_groups_attempts
					LEFT JOIN LATERAL (
						SELECT
							target_groups_attempts.id,
							MAX(aggregated_children_attempts.latest_activity_at) AS latest_activity_at,
							IFNULL(SUM(aggregated_children_attempts.tasks_tried), 0) AS tasks_tried,
							IFNULL(SUM(aggregated_children_attempts.tasks_with_help), 0) AS tasks_with_help,
							IFNULL(SUM(aggregated_children_attempts.tasks_solved), 0) AS tasks_solved,
							IFNULL(SUM(aggregated_children_attempts.validated), 0) AS children_validated,
							SUM(IFNULL(NOT aggregated_children_attempts.validated, 1)) AS children_non_validated,
							SUM(items_items.category = 'Validation' AND IFNULL(NOT aggregated_children_attempts.validated, 1))
								AS children_non_validated_categories,
							MAX(aggregated_children_attempts.validated_at) AS max_validated_at,
							MAX(IF(items_items.category = 'Validation', aggregated_children_attempts.validated_at, NULL))
								AS max_validated_at_categories,
							SUM(IFNULL(aggregated_children_attempts.score, 0) * items_items.score_weight) /
								COALESCE(NULLIF(IFNULL(SUM(items_items.score_weight), 0), 0), 1) AS average_score
						FROM items_items ` +
					// We use LEFT JOIN LATERAL to aggregate attempts grouped by target_groups_attempts.group_id & items_items.child_item_id.
					// The usual LEFT JOIN conditions in the ON clause would group attempts before joining which would produce
					// wrong results.
					`	LEFT JOIN LATERAL (
							SELECT
								MAX(validated) AS validated,
								MIN(validated_at) AS validated_at,
								MAX(latest_activity_at) AS latest_activity_at,
								MAX(tasks_tried) AS tasks_tried,
								MAX(tasks_with_help) AS tasks_with_help,
								MAX(tasks_solved) AS tasks_solved,
								MAX(score) AS score
							FROM groups_attempts AS children_attempts
							WHERE children_attempts.group_id = target_groups_attempts.group_id AND
								children_attempts.item_id = items_items.child_item_id
							GROUP BY children_attempts.group_id, children_attempts.item_id
						) AS aggregated_children_attempts ON 1
						JOIN items ON(
							items.id = items_items.child_item_id
						)
						WHERE items_items.parent_item_id = target_groups_attempts.item_id AND NOT items.no_score
						GROUP BY items_items.parent_item_id
					) AS children_stats ON 1
					JOIN items
						ON target_groups_attempts.item_id = items.id
					SET
						target_groups_attempts.latest_activity_at = IF(children_stats.id IS NOT NULL,
							children_stats.latest_activity_at, target_groups_attempts.latest_activity_at),
						target_groups_attempts.tasks_tried = IF(children_stats.id IS NOT NULL,
							children_stats.tasks_tried, target_groups_attempts.tasks_tried),
						target_groups_attempts.tasks_with_help = IF(children_stats.id IS NOT NULL,
							children_stats.tasks_with_help, target_groups_attempts.tasks_with_help),
						target_groups_attempts.tasks_solved = IF(children_stats.id IS NOT NULL,
							children_stats.tasks_solved, target_groups_attempts.tasks_solved),
						target_groups_attempts.children_validated = IF(children_stats.id IS NOT NULL,
							children_stats.children_validated, target_groups_attempts.children_validated),
						target_groups_attempts.validated_at = IFNULL(target_groups_attempts.validated_at, CASE
							WHEN children_stats.id IS NULL THEN NULL
							WHEN items.validation_type = 'Categories' AND children_stats.children_non_validated_categories = 0
							  THEN children_stats.max_validated_at_categories
							WHEN items.validation_type = 'All' AND children_stats.children_non_validated = 0 THEN children_stats.max_validated_at
							WHEN items.validation_type = 'AllButOne' AND children_stats.children_non_validated < 2 THEN children_stats.max_validated_at
							WHEN items.validation_type = 'One' AND children_stats.children_validated > 0 THEN children_stats.max_validated_at
							ELSE NULL
							END),
						target_groups_attempts.score = IF(items.no_score, 0, IF(children_stats.id IS NOT NULL,
							children_stats.average_score, target_groups_attempts.score))
					WHERE target_groups_attempts.ancestors_computation_state = 'processing'`
				updateStatement, err = ds.db.CommonDB().Prepare(updateQuery)
				mustNotBeError(err)
				defer func() { mustNotBeError(updateStatement.Close()) }()
			}

			_, err = updateStatement.Exec()
			mustNotBeError(err)

			if unlockStatement == nil {
				const unlockQuery = `
					INSERT INTO permissions_granted
						(group_id, item_id, giver_group_id, can_view, latest_update_on)
						SELECT
							groups.id AS group_id,
							item_unlocking_rules.unlocked_item_id AS item_id,
							-1,
							'content',
							NOW()
						FROM groups_attempts
						JOIN item_unlocking_rules ON item_unlocking_rules.unlocking_item_id = groups_attempts.item_id AND
							item_unlocking_rules.score <= groups_attempts.score
						JOIN ` + "`groups`" + ` ON groups_attempts.group_id = groups.id
						WHERE groups_attempts.ancestors_computation_state = 'processing'
					ON DUPLICATE KEY UPDATE
						latest_update_on = IF(can_view = 'content', latest_update_on, NOW()),
						can_view = GREATEST(can_view_value, 3 /* 'content' */)`
				unlockStatement, err = ds.db.CommonDB().Prepare(unlockQuery)
				mustNotBeError(err)
				defer func() { mustNotBeError(unlockStatement.Close()) }()
			}

			var result sql.Result
			result, err = unlockStatement.Exec()
			mustNotBeError(err)
			var rowsAffected int64
			rowsAffected, err = result.RowsAffected()
			mustNotBeError(err)
			groupsUnlocked += rowsAffected

			if markAsDoneStatement == nil {
				const markAsDoneQuery = `UPDATE groups_attempts SET ancestors_computation_state = 'done'
						WHERE ancestors_computation_state = 'processing'`
				markAsDoneStatement, err = ds.db.CommonDB().Prepare(markAsDoneQuery)
				mustNotBeError(err)
				defer func() { mustNotBeError(markAsDoneStatement.Close()) }()
			}

			result, err = markAsDoneStatement.Exec()
			rowsAffected, err = result.RowsAffected()
			mustNotBeError(err)
			hasChanges = rowsAffected > 0
			mustNotBeError(err)
		}

		return nil
	}))

	// If items have been unlocked, need to recompute access
	if groupsUnlocked > 0 {
		return s.PermissionsGranted().After()
	}
	return nil
}
