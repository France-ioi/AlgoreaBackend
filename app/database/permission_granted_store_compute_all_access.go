package database

import (
	"time"

	"github.com/France-ioi/AlgoreaBackend/v2/app/logging"
	"github.com/France-ioi/AlgoreaBackend/v2/golang"
)

// computeAllAccess recomputes fields of permissions_generated.
//
// It starts from group-item pairs marked with propagate_to = 'self' in `permissions_propagate`.
// Those are created by SQL triggers:
// - after_insert_permissions_granted
// - after_update_permissions_granted
// - after_delete_permissions_granted
// - after_insert_items_items
//
// 1. can_view_generated, can_grant_view_generated, can_watch_generated, can_edit_generated, is_owner_generated are updated.
//
// 3. Then the loop repeats from step 1 for all children (from items_items) of the processed permissions_generated.
//
// Notes:
//   - The function may loop endlessly if items_items is a cyclic graph.
//   - Processed group-item pairs are removed from permissions_propagate.
func (s *PermissionGrantedStore) computeAllAccess() {
	permissionsPropagateTableName := s.permissionsPropagateTableName()

	// marking group-item pairs whose parents are marked with propagate_to = 'children' as 'self'
	queryMarkChildrenOfChildrenAsSelf := `
		INSERT INTO ` + permissionsPropagateTableName +
		` (` + golang.If(s.arePropagationsSync(), "connection_id, ") + `group_id, item_id, propagate_to)
		SELECT
			` + golang.If(s.arePropagationsSync(), "CONNECTION_ID(), ") + `
			parents.group_id,
			items_items.child_item_id,
			'self' as propagate_to
		FROM items_items
		JOIN permissions_generated AS parents
			ON parents.item_id = items_items.parent_item_id
		JOIN ` + permissionsPropagateTableName + ` AS parents_propagate
			ON parents_propagate.group_id = parents.group_id AND parents_propagate.item_id = parents.item_id
		WHERE parents_propagate.propagate_to = 'children'
		GROUP BY parents.group_id, items_items.child_item_id
		ON DUPLICATE KEY UPDATE propagate_to='self'`

	// deleting 'children' permissions_propagate
	queryDeleteProcessedChildren := `DELETE FROM ` + permissionsPropagateTableName + ` WHERE propagate_to = 'children'`

	const queryDropTemporaryTable = `DROP TEMPORARY TABLE IF EXISTS permissions_propagate_processing`
	// creating permissions_propagate_processing
	const queryCreateTemporaryTable = `CREATE TEMPORARY TABLE permissions_propagate_processing ` +
		`(group_id BIGINT(20) NOT NULL, item_id BIGINT(20) NOT NULL, PRIMARY KEY (group_id, item_id))`

	// marking 'self' permissions_propagate that are not descendants of other 'self' permissions_propagate for processing
	// in permissions_propagate_processing
	queryInsertIntoPermissionsPropagateProcessing := `
		INSERT INTO permissions_propagate_processing (group_id, item_id)
		SELECT group_id, item_id
		FROM ` + permissionsPropagateTableName + `
		WHERE propagate_to = 'self' AND (
			SELECT 1
			FROM ` + permissionsPropagateTableName + ` AS ancestor_propagate
			JOIN items_ancestors
				ON items_ancestors.child_item_id = ` + permissionsPropagateTableName + `.item_id AND
				   items_ancestors.ancestor_item_id = ancestor_propagate.item_id
			WHERE ancestor_propagate.group_id = ` + permissionsPropagateTableName + `.group_id AND
			      ancestor_propagate.propagate_to = 'self'
			LIMIT 1
			FOR SHARE
		) IS NULL
		FOR SHARE`

	// computation for group-item pairs marked as 'self' in permissions_propagate (so all of them)
	const queryUpdatePermissionsGenerated = `
		INSERT INTO permissions_generated
			(group_id, item_id, can_view_generated, can_grant_view_generated, can_watch_generated, can_edit_generated, is_owner_generated)
		SELECT STRAIGHT_JOIN
			permissions_propagate_processing.group_id,
			permissions_propagate_processing.item_id,
			IF(MAX(permissions_granted.is_owner), 'solution', GREATEST(
				IFNULL(MAX(permissions_granted.can_view_value), 1),
				IFNULL(MAX(
					CASE
					WHEN parent.can_view_generated IS NULL OR parent.can_view_generated IN ('none', 'info') THEN 1 /* none */
					WHEN parent.can_view_generated = 'content' OR items_items.upper_view_levels_propagation = 'use_content_view_propagation' THEN
						CASE items_items.content_view_propagation
						WHEN 'as_info' THEN 2 /* info */
						WHEN 'as_content' THEN 3 /* content */
						ELSE 1 /* none */
						END
					WHEN items_items.upper_view_levels_propagation = 'as_content_with_descendants' THEN 4 /* content_with_descendants */
					ELSE parent.can_view_generated_value
					END), 1)
			)) AS can_view_generated,
			IF(MAX(permissions_granted.is_owner), 'solution_with_grant', GREATEST(
				IFNULL(MAX(permissions_granted.can_grant_view_value), 1),
				IFNULL(MAX(IF(items_items.grant_view_propagation, LEAST(parent.can_grant_view_generated_value, 5 /* solution */), 1)), 1)
			)) AS can_grant_view_generated,
			IF(MAX(permissions_granted.is_owner), 'answer_with_grant', GREATEST(
				IFNULL(MAX(permissions_granted.can_watch_value), 1),
				IFNULL(MAX(IF(items_items.watch_propagation, LEAST(parent.can_watch_generated_value, 3 /* answer */), 1)), 1)
			)) AS can_watch_generated,
			IF(MAX(permissions_granted.is_owner), 'all_with_grant', GREATEST(
				IFNULL(MAX(permissions_granted.can_edit_value), 1),
				IFNULL(MAX(IF(items_items.edit_propagation, LEAST(parent.can_edit_generated_value, 3 /* all */), 1)), 1)
			)) AS can_edit_generated,
			IFNULL(MAX(permissions_granted.is_owner), 0) AS is_owner_generated
		FROM permissions_propagate_processing
		LEFT JOIN permissions_granted USING (group_id, item_id)
		LEFT JOIN items_items ON items_items.child_item_id = permissions_propagate_processing.item_id
		LEFT JOIN permissions_generated AS parent
		  ON parent.item_id = items_items.parent_item_id AND parent.group_id = permissions_propagate_processing.group_id
		GROUP BY permissions_propagate_processing.group_id, permissions_propagate_processing.item_id
		ON DUPLICATE KEY UPDATE
			can_view_generated = VALUES(can_view_generated),
			can_grant_view_generated = VALUES(can_grant_view_generated),
			can_watch_generated = VALUES(can_watch_generated),
			can_edit_generated = VALUES(can_edit_generated),
			is_owner_generated = VALUES(is_owner_generated)`

	// marking 'self' permissions_propagate (so all of them) as 'children'
	queryMarkSelfAsChildren := `
		UPDATE ` + permissionsPropagateTableName + `
		JOIN permissions_propagate_processing
			ON permissions_propagate_processing.group_id = ` + permissionsPropagateTableName + `.group_id AND
			   permissions_propagate_processing.item_id = ` + permissionsPropagateTableName + `.item_id
		SET ` + permissionsPropagateTableName + `.propagate_to = 'children'`

	// ------------------------------------------------------------------------------------
	// Here we execute the statements
	// ------------------------------------------------------------------------------------
	hasChanges := true
	for hasChanges {
		CallBeforePropagationStepHook(PropagationStepAccessMain)

		mustNotBeError(s.EnsureTransaction(func(store *DataStore) error {
			initTransactionTime := time.Now()

			mustNotBeError(store.Exec(queryCreateTemporaryTable).Error())
			defer store.Exec(queryDropTemporaryTable)

			mustNotBeError(store.Exec(queryMarkChildrenOfChildrenAsSelf).Error())
			mustNotBeError(store.Exec(queryDeleteProcessedChildren).Error())
			mustNotBeError(store.Exec(queryInsertIntoPermissionsPropagateProcessing).Error())
			mustNotBeError(store.Exec(queryUpdatePermissionsGenerated).Error())

			result := store.Exec(queryMarkSelfAsChildren)
			mustNotBeError(result.Error())
			rowsAffected := result.RowsAffected()

			logging.SharedLogger.WithContext(store.ctx).
				Debugf("Duration of permissions propagation step: %d rows affected, took %v", rowsAffected, time.Since(initTransactionTime))

			hasChanges = rowsAffected > 0

			return nil
		}))
	}
}
