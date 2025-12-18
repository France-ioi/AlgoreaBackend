package database

import (
	"fmt"
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
	s.computeAllAccessWithCustomTables(
		"permissions_granted", "permissions_generated",
		s.permissionsPropagateTableName(), "", "",
		"BIGINT(20)", s.arePropagationsSync(), false, nil)
}

func (s *PermissionGrantedStore) computeAllAccessWithCustomTables(
	permissionsGrantedTableName, permissionsGeneratedTableName, permissionsPropagateTableName,
	permissionsGenerated2TableName, permissionsPropagate2TableName,
	groupIDType string,
	permissionsPropagateTableContainsConnectionID, skipTransactions bool,
	narrowToAncestorsOfItemWithID *int64,
) {
	permissionsGrantedTableNameQuoted := QuoteName(permissionsGrantedTableName)
	permissionsGeneratedTableNameQuoted := QuoteName(permissionsGeneratedTableName)
	permissionsPropagateTableNameQuoted := QuoteName(permissionsPropagateTableName)
	permissionsGenerated2TableNameQuoted := golang.If(permissionsGenerated2TableName != "", QuoteName(permissionsGenerated2TableName))
	permissionsPropagate2TableNameQuoted := golang.If(permissionsPropagate2TableName != "", QuoteName(permissionsPropagate2TableName))

	ensureTransactionFunc := func(s *DataStore, txFunc func(store *DataStore) error) error { return s.EnsureTransaction(txFunc) }
	if skipTransactions {
		ensureTransactionFunc = func(s *DataStore, txFunc func(store *DataStore) error) error { return txFunc(s) }
	}

	// marking group-item pairs whose parents are marked with propagate_to = 'children' as 'self'
	// (if permissionsPropagate2TableName is given, it will be the destination)
	queryMarkChildrenOfChildrenAsSelf := `
		INSERT INTO ` +
		golang.IfElse(permissionsPropagate2TableNameQuoted != "", permissionsPropagate2TableNameQuoted, permissionsPropagateTableNameQuoted) +
		` (` + golang.If(permissionsPropagateTableContainsConnectionID, "connection_id, ") + `group_id, item_id, propagate_to)
		SELECT
			` + golang.If(permissionsPropagateTableContainsConnectionID, "CONNECTION_ID(), ") + `
			parents_propagate.group_id,
			items_items.child_item_id,
			'self' as propagate_to
		FROM items_items
		JOIN ` + permissionsPropagateTableNameQuoted + ` AS parents_propagate
			ON parents_propagate.item_id = items_items.parent_item_id
		WHERE parents_propagate.propagate_to = 'children'` +
		golang.LazyIf(narrowToAncestorsOfItemWithID != nil,
			func() string {
				return fmt.Sprintf(` AND (items_items.child_item_id = %d OR EXISTS (
						SELECT 1 FROM items_ancestors WHERE child_item_id = %d AND ancestor_item_id = items_items.child_item_id
					))`, *narrowToAncestorsOfItemWithID, *narrowToAncestorsOfItemWithID)
			}) + `
		ON DUPLICATE KEY UPDATE propagate_to='self'`

	// if permissionsPropagate2TableName is given, we insert into it and then move all the rows to permissionsPropagateTableName
	queryCopyFromPermissionsPropagate2Table := golang.If(permissionsPropagate2TableNameQuoted != "", `
		INSERT INTO `+permissionsPropagateTableNameQuoted+
		` (`+golang.If(permissionsPropagateTableContainsConnectionID, "connection_id, ")+`group_id, item_id, propagate_to)
		SELECT
			`+golang.If(permissionsPropagateTableContainsConnectionID, "connection_id, ")+`
			group_id, item_id, propagate_to
		FROM `+permissionsPropagate2TableName+`
		ON DUPLICATE KEY UPDATE propagate_to='self'`)

	queryDeleteFromPermissionsPropagate2Table := golang.If(permissionsPropagate2TableNameQuoted != "",
		"DELETE FROM "+permissionsPropagate2TableNameQuoted)

	// deleting 'children' permissions_propagate
	queryDeleteProcessedChildren := `DELETE FROM ` + permissionsPropagateTableNameQuoted + ` WHERE propagate_to = 'children'`

	const queryDropTemporaryTable = `DROP TEMPORARY TABLE IF EXISTS permissions_propagate_processing`
	// creating permissions_propagate_processing
	queryCreateTemporaryTable := `CREATE TEMPORARY TABLE permissions_propagate_processing ` +
		`(group_id ` + groupIDType + ` NOT NULL, item_id BIGINT(20) NOT NULL, PRIMARY KEY (group_id, item_id))`

	const queryDropTemporaryTableForPostponedPermissions = `DROP TEMPORARY TABLE IF EXISTS permissions_propagate_postponed`
	queryCreateTemporaryTableForPostponedPermissions := `CREATE TEMPORARY TABLE permissions_propagate_postponed ` +
		`(group_id ` + groupIDType + ` NOT NULL, item_id BIGINT(20) NOT NULL, PRIMARY KEY (group_id, item_id))`

	// prepare the list of descendant permissions to postpone processing (their ancestors permissions have not been processed yet)
	// (this is much faster than searching for not processed ancestors for each item separately on the next insert)
	queryMarkPostponedPermissions := `
		INSERT INTO permissions_propagate_postponed (group_id, item_id)
		SELECT group_id, child_item_id
		FROM ` + permissionsPropagateTableNameQuoted + ` AS permissions_propagate
		JOIN items_ancestors ON items_ancestors.ancestor_item_id = permissions_propagate.item_id
		ON DUPLICATE KEY UPDATE item_id = VALUES(item_id)`

	// marking 'self' permissions_propagate that are not postponed (i.e. are not descendants of other 'self' permissions_propagate)
	// for processing in permissions_propagate_processing
	queryInsertIntoPermissionsPropagateProcessing := `
		INSERT INTO permissions_propagate_processing (group_id, item_id)
		SELECT permissions_propagate.group_id, permissions_propagate.item_id
		FROM ` + permissionsPropagateTableNameQuoted + ` AS permissions_propagate
		WHERE propagate_to = 'self' AND NOT EXISTS (
			SELECT 1
			FROM permissions_propagate_postponed
			WHERE permissions_propagate_postponed.group_id = permissions_propagate.group_id AND
			      permissions_propagate_postponed.item_id = permissions_propagate.item_id
		)
		FOR SHARE`

	// computation for group-item pairs marked as 'self' in permissions_propagate (so all of them)
	// (if permissionsGenerated2TableName is given, it will be the destination)
	queryUpdatePermissionsGenerated := `
		INSERT INTO ` +
		golang.IfElse(permissionsGenerated2TableNameQuoted != "", permissionsGenerated2TableNameQuoted, permissionsGeneratedTableNameQuoted) + `
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
		LEFT JOIN ` + permissionsGrantedTableNameQuoted + ` AS permissions_granted USING (group_id, item_id)
		LEFT JOIN items_items ON items_items.child_item_id = permissions_propagate_processing.item_id
		LEFT JOIN ` + permissionsGeneratedTableNameQuoted + ` AS parent
		  ON parent.item_id = items_items.parent_item_id AND parent.group_id = permissions_propagate_processing.group_id
		GROUP BY permissions_propagate_processing.group_id, permissions_propagate_processing.item_id
		ON DUPLICATE KEY UPDATE
			can_view_generated = VALUES(can_view_generated),
			can_grant_view_generated = VALUES(can_grant_view_generated),
			can_watch_generated = VALUES(can_watch_generated),
			can_edit_generated = VALUES(can_edit_generated),
			is_owner_generated = VALUES(is_owner_generated)`

	// if permissionsGenerated2TableName is given, we insert into it and then move all the rows to permissionsGeneratedTableName
	queryCopyFromPermissionsGenerated2Table := golang.If(permissionsGenerated2TableNameQuoted != "", `
		INSERT INTO `+
		permissionsGeneratedTableNameQuoted+
		` (group_id, item_id, can_view_generated, can_grant_view_generated, can_watch_generated, can_edit_generated, is_owner_generated)
		SELECT group_id, item_id, can_view_generated, can_grant_view_generated, can_watch_generated, can_edit_generated, is_owner_generated
		FROM `+permissionsGenerated2TableNameQuoted+`
		ON DUPLICATE KEY UPDATE
			can_view_generated = VALUES(can_view_generated),
			can_grant_view_generated = VALUES(can_grant_view_generated),
			can_watch_generated = VALUES(can_watch_generated),
			can_edit_generated = VALUES(can_edit_generated),
			is_owner_generated = VALUES(is_owner_generated)`)

	queryDeleteFromPermissionsGenerated2Table := golang.If(permissionsGenerated2TableNameQuoted != "",
		"DELETE FROM "+permissionsGenerated2TableNameQuoted)

	// marking 'self' permissions_propagate (so all of them) as 'children'
	queryMarkSelfAsChildren := `
		UPDATE ` + permissionsPropagateTableNameQuoted + ` AS permissions_propagate
		JOIN permissions_propagate_processing
			ON permissions_propagate_processing.group_id = permissions_propagate.group_id AND
			   permissions_propagate_processing.item_id = permissions_propagate.item_id
		SET permissions_propagate.propagate_to = 'children'`

	// ------------------------------------------------------------------------------------
	// Here we execute the statements
	// ------------------------------------------------------------------------------------
	hasChanges := true
	for hasChanges {
		CallBeforePropagationStepHook(PropagationStepAccessMain)

		mustNotBeError(ensureTransactionFunc(s.DataStore, func(store *DataStore) error {
			initTransactionTime := time.Now()

			mustNotBeError(store.Exec(queryCreateTemporaryTable).Error())
			defer store.Exec(queryDropTemporaryTable)

			result := store.Exec(queryMarkChildrenOfChildrenAsSelf)
			mustNotBeError(result.Error())
			if permissionsPropagate2TableNameQuoted != "" && result.RowsAffected() > 0 {
				mustNotBeError(store.Exec(queryCopyFromPermissionsPropagate2Table).Error())
				mustNotBeError(store.Exec(queryDeleteFromPermissionsPropagate2Table).Error())
			}

			mustNotBeError(store.Exec(queryDeleteProcessedChildren).Error())

			mustNotBeError(store.Exec(queryCreateTemporaryTableForPostponedPermissions).Error())
			defer store.Exec(queryDropTemporaryTableForPostponedPermissions)

			mustNotBeError(store.Exec(queryMarkPostponedPermissions).Error())
			mustNotBeError(store.Exec(queryInsertIntoPermissionsPropagateProcessing).Error())

			result = store.Exec(queryUpdatePermissionsGenerated)
			mustNotBeError(result.Error())
			if permissionsGenerated2TableNameQuoted != "" && result.RowsAffected() > 0 {
				mustNotBeError(store.Exec(queryCopyFromPermissionsGenerated2Table).Error())
				mustNotBeError(store.Exec(queryDeleteFromPermissionsGenerated2Table).Error())
			}

			result = store.Exec(queryMarkSelfAsChildren)
			mustNotBeError(result.Error())
			rowsAffected := result.RowsAffected()

			logging.EntryFromContext(store.ctx()).
				Debugf("Duration of permissions propagation step: %d rows affected, took %v", rowsAffected, time.Since(initTransactionTime))

			hasChanges = rowsAffected > 0

			return nil
		}))
	}
}

// ComputeAllAccess allows to call computeAllAccess() from outside.
//
// Note: The method propagates permissions synchronously. It does not use propagations scheduling.
// Callers probably want to call this method inside a transaction and mark the transaction with DataStore.SetPropagationsModeToSync()
// to ensure it will not process permissions that are marked for propagation by other transactions.
func (s *PermissionGrantedStore) ComputeAllAccess() (err error) {
	defer recoverPanics(&err)

	s.computeAllAccess()
	return nil
}

// CreateTemporaryTablesForPermissionsExplanation creates temporary tables
// permissions_granted_exp, permissions_generated_exp, permissions_generated_exp2,
// permissions_propagate_exp, permissions_propagate_exp2 for explaining permission computations.
// As we want to know the effect of each granted permission separately, `group_id` in all the tables is constructed as
// "{group_id}|{item_id}|{source_group_id}|{origin}" of granted permissions we propagate.
//
// It should be called either inside a transaction or with a fixed MySQL connection.
func (s *PermissionGrantedStore) CreateTemporaryTablesForPermissionsExplanation() (cleanupFunc func(), err error) {
	s.mustBeFixed()
	defer recoverPanics(&err)

	var cleanupFuncsToCall []func()
	cleanupFunc = func() {
		for i := len(cleanupFuncsToCall) - 1; i >= 0; i-- {
			cleanupFuncsToCall[i]()
		}
	}

	// group_id in these temporary tables is constructed as {group_id}|{item_id}|{source_group_id}|{origin}
	// of granted permissions we want to propagate
	mustNotBeError(s.Exec(`
		CREATE TEMPORARY TABLE permissions_granted_exp (
			` + "`group_id`" + ` CHAR(79) NOT NULL,
			` + "`item_id`" + ` BIGINT NOT NULL,
			` + "`source_group_id`" + ` BIGINT NOT NULL,
			` + "`origin`" + ` ENUM('group_membership','item_unlocking','self','other') NOT NULL,
			` + "`can_view`" + ` ENUM('none','info','content','content_with_descendants','solution') NOT NULL DEFAULT 'none',
			` + "`can_grant_view`" + `
				ENUM('none','enter','content','content_with_descendants','solution','solution_with_grant') NOT NULL DEFAULT 'none',
			` + "`can_watch`" + ` ENUM('none','result','answer','answer_with_grant') NOT NULL DEFAULT 'none',
			` + "`can_edit`" + ` ENUM('none','children','all','all_with_grant') NOT NULL DEFAULT 'none',
			` + "`is_owner`" + ` TINYINT(1) NOT NULL DEFAULT '0',
			` + "`can_view_value`" + ` TINYINT UNSIGNED GENERATED ALWAYS AS ((` + "`can_view`" + ` + 0)) VIRTUAL NOT NULL,
			` + "`can_grant_view_value`" + ` TINYINT UNSIGNED GENERATED ALWAYS AS ((` + "`can_grant_view`" + ` + 0)) VIRTUAL NOT NULL,
			` + "`can_watch_value`" + ` TINYINT UNSIGNED GENERATED ALWAYS AS ((` + "`can_watch`" + ` + 0)) VIRTUAL NOT NULL,
			` + "`can_edit_value`" + ` TINYINT UNSIGNED GENERATED ALWAYS AS ((` + "`can_edit`" + ` + 0)) VIRTUAL NOT NULL,
			PRIMARY KEY (` + "`group_id`,`item_id`,`source_group_id`,`origin`" + `),
			KEY ` + "`group_id_item_id` (`group_id`,`item_id`)" + `
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci`).Error())
	cleanupFuncsToCall = append(cleanupFuncsToCall, func() { s.Exec("DROP TEMPORARY TABLE permissions_granted_exp") })

	mustNotBeError(s.Exec(`
		CREATE TEMPORARY TABLE permissions_generated_exp (
			` + "`group_id`" + ` CHAR(79) NOT NULL,
			` + "`item_id`" + ` BIGINT NOT NULL,
			` + "`can_view_generated`" + ` ENUM('none','info','content','content_with_descendants','solution') NOT NULL DEFAULT 'none',
			` + "`can_grant_view_generated`" + `
				ENUM('none','enter','content','content_with_descendants','solution','solution_with_grant') NOT NULL DEFAULT 'none',
			` + "`can_watch_generated`" + ` ENUM('none','result','answer','answer_with_grant') NOT NULL DEFAULT 'none',
			` + "`can_edit_generated`" + ` ENUM('none','children','all','all_with_grant') NOT NULL DEFAULT 'none',
			` + "`is_owner_generated`" + ` TINYINT(1) NOT NULL DEFAULT '0',
			` + "`can_view_generated_value`" + ` TINYINT UNSIGNED GENERATED ALWAYS AS ((` + "`can_view_generated`" + ` + 0)) VIRTUAL NOT NULL,
			` + "`can_grant_view_generated_value`" + `
				TINYINT UNSIGNED GENERATED ALWAYS AS ((` + "`can_grant_view_generated`" + ` + 0)) VIRTUAL NOT NULL,
			` + "`can_watch_generated_value`" + ` TINYINT UNSIGNED GENERATED ALWAYS AS ((` + "`can_watch_generated`" + ` + 0)) VIRTUAL NOT NULL,
			` + "`can_edit_generated_value`" + ` TINYINT UNSIGNED GENERATED ALWAYS AS ((` + "`can_edit_generated`" + ` + 0)) VIRTUAL NOT NULL,
			` + "`hashed_group_id`" + ` CHAR(64) GENERATED ALWAYS AS ((SHA2(group_id, 256))) STORED NOT NULL,
			` + "`permissions_granted_group_id`" + `
				BIGINT GENERATED ALWAYS AS (SUBSTRING_INDEX(permissions_generated_exp.group_id, '|', 1)) VIRTUAL NOT NULL,
			` + "`permissions_granted_item_id`" + `
				BIGINT GENERATED ALWAYS AS (SUBSTRING_INDEX(SUBSTRING_INDEX(permissions_generated_exp.group_id, '|', 2), '|', -1)) VIRTUAL NOT NULL,
			` + "`permissions_granted_source_group_id`" + `
				BIGINT GENERATED ALWAYS AS (SUBSTRING_INDEX(SUBSTRING_INDEX(permissions_generated_exp.group_id, '|', 3), '|', -1)) VIRTUAL NOT NULL,
			` + "`permissions_granted_origin`" + `
				ENUM('group_membership','item_unlocking','self','other')
					GENERATED ALWAYS AS (SUBSTRING_INDEX(permissions_generated_exp.group_id, '|', -1)) VIRTUAL NOT NULL,
			PRIMARY KEY (` + "`group_id`,`item_id`" + `),
			KEY ` + "`hashed_group_id` (`hashed_group_id`, `item_id`)," + `
			KEY ` + "`item_id`" + `
				(
					` + "`item_id`, `permissions_granted_group_id`, `permissions_granted_item_id`, " + `
					` + "`permissions_granted_source_group_id`, `permissions_granted_origin`" + `
				)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci`).Error())
	cleanupFuncsToCall = append(cleanupFuncsToCall, func() { s.Exec("DROP TEMPORARY TABLE permissions_generated_exp") })

	mustNotBeError(s.Exec(`
		CREATE TEMPORARY TABLE permissions_generated_exp2 LIKE permissions_generated_exp`).Error())
	cleanupFuncsToCall = append(cleanupFuncsToCall, func() { s.Exec("DROP TEMPORARY TABLE permissions_generated_exp2") })

	mustNotBeError(s.Exec(`
		CREATE TEMPORARY TABLE permissions_propagate_exp (
			` + "`group_id`" + ` CHAR(79) NOT NULL NOT NULL,
			` + "`item_id`" + ` BIGINT NOT NULL,
			` + "`propagate_to`" + ` ENUM('self','children') NOT NULL,
			PRIMARY KEY ` + "(`group_id`,`item_id`)" + `,
			KEY ` + "`propagate_to_group_id_item_id` (`propagate_to`,`group_id`,`item_id`)" + `
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci`).Error())
	cleanupFuncsToCall = append(cleanupFuncsToCall, func() { s.Exec("DROP TEMPORARY TABLE permissions_propagate_exp") })

	mustNotBeError(s.Exec(`
		CREATE TEMPORARY TABLE permissions_propagate_exp2 LIKE permissions_propagate_exp`).Error())
	cleanupFuncsToCall = append(cleanupFuncsToCall, func() { s.Exec("DROP TEMPORARY TABLE permissions_propagate_exp2") })

	return cleanupFunc, nil
}

// ComputePermissionsExplanation computes explanation of permissions inserted into `permissions_granted_exp`
// and marked to be propagated to 'self' in `permissions_propagate_exp`.
// This is done by running the permissions propagation narrowed to ancestors of the given item and the item itself (if itemID is not nil)
// using custom temporary tables `permissions_granted_exp`, `permissions_generated_exp`, and `permissions_propagate_exp`.
// As we want to know the effect of each granted permission separately, we insert all the granted permissions
// into `permissions_granted_exp` which has a tricky `group_id` column constructed as
// "{group_id}|{item_id}|{source_group_id}|{origin}" of granted permissions we want to propagate.
// This way, each granted permission will be propagated separately, without grouping permissions having
// the same (`group_id`, `item_id`) pair.
// To make this work, other temporary tables also have a `group_id` column constructed in the same way.
// The result of the computation is stored in `permissions_generated_exp`.
//
// As there is a bug in MySQL making it impossible to use a temporary table twice in the same query
// (see https://bugs.mysql.com/bug.php?id=10327), we will use additional temporary tables
// `permissions_generated_exp2` & `permissions_propagate_exp2` to get around the issue.
//
// All the temporary tables should be created beforehand by calling CreateTemporaryTablesForPermissionsExplanation().
//
// The `group_id` field in these temporary tables is constructed as {group_id}|{item_id}|{source_group_id}|{origin}
// of granted permissions we want to propagate.
//
// Note: The method propagates permissions synchronously. It does not use propagations scheduling.
// It's a good idea to call this method on a fixed MySQL connection outside of transactions
// to ensure it will not lock the database.
func (s *PermissionGrantedStore) ComputePermissionsExplanation(itemID *int64) (err error) {
	s.mustBeFixed()
	defer recoverPanics(&err)

	s.computeAllAccessWithCustomTables(
		"permissions_granted_exp", "permissions_generated_exp", "permissions_propagate_exp",
		"permissions_generated_exp2", "permissions_propagate_exp2",
		"CHAR(79)", false, true, itemID)
	return nil
}
