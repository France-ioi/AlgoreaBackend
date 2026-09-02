//go:build !unit

package database_test

import (
	"database/sql"
	"strconv"
	"strings"
	"testing"
	_ "unsafe"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/testhelpers"
	"github.com/France-ioi/AlgoreaBackend/v2/testhelpers/testoutput"
)

func Test_moveFromResultsPropagateToResultsPropagateInternal_InsideTransaction(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db := testhelpers.SetupDBWithFixtureString(testhelpers.CreateTestContext(), `
		groups: [{id: 101}]
		attempts: [{participant_id: 101, id: 1}]
		items: [{id: 12, default_language_tag: fr}]
		results: [{participant_id: 101, attempt_id: 1, item_id: 12}]
		results_propagate:
			- {participant_id: 101, attempt_id: 1, item_id: 12, state: to_be_propagated}
	`)
	defer func() { _ = db.Close() }()

	store := database.NewDataStore(db)
	require.NoError(t, store.InTransaction(func(store *database.DataStore) error {
		var isolationRow struct {
			Isolation string `gorm:"column:transaction_isolation"`
		}
		require.NoError(t, store.Raw(
			"SELECT @@transaction_isolation AS transaction_isolation",
		).Scan(&isolationRow).Error())
		assert.Equal(t, "REPEATABLE-READ", isolationRow.Isolation,
			"in-transaction drain must not change session isolation")

		assert.True(t, moveFromResultsPropagateToResultsPropagateInternal(store))
		return nil
	}))
	assertResultsMarkedAsChanged(t, store, "results_propagate", nil)
	assertResultsMarkedAsChanged(t, store, "results_propagate_internal", []resultPrimaryKeyAndState{
		{ResultPrimaryKey: ResultPrimaryKey{101, 1, 12}, State: "to_be_propagated"},
	})
}

func Test_moveFromResultsPropagateToResultsPropagateInternal_EmptyTable(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db := testhelpers.SetupDBWithFixtureString(testhelpers.CreateTestContext(), `
		groups: [{id: 101}]
	`)
	defer func() { _ = db.Close() }()

	assert.False(t, moveFromResultsPropagateToResultsPropagateInternal(database.NewDataStore(db)))
}

func Test_SetResultsPropagateDrainChunkSizeForTests_PanicsWhenLessThanOne(t *testing.T) {
	assert.PanicsWithValue(t, "resultsPropagateDrainChunkSize must be >= 1", func() {
		database.SetResultsPropagateDrainChunkSizeForTests(0)
	})
}

func Test_moveFromResultsPropagateToResultsPropagateInternal_WalksRowsNotIdSpan(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db := testhelpers.SetupDBWithFixtureString(testhelpers.CreateTestContext(), `
		groups: [{id: 101}]
		attempts: [{participant_id: 101, id: 1}]
		items:
			- {id: 12, default_language_tag: fr}
			- {id: 13, default_language_tag: fr}
		results:
			- {participant_id: 101, attempt_id: 1, item_id: 12}
			- {participant_id: 101, attempt_id: 1, item_id: 13}
	`)
	defer func() { _ = db.Close() }()

	store := database.NewDataStore(db)
	// Gap of 99 empty ids: an id-span walk (fromID += chunkSize) would open ~100 chunk
	// transactions; a row walk opens 2 non-empty + 1 trailing empty.
	require.NoError(t, store.Exec(`
		INSERT INTO results_propagate (id, participant_id, attempt_id, item_id, state) VALUES
			(1, 101, 1, 12, 'to_be_propagated'),
			(100, 101, 1, 13, 'to_be_propagated')`).Error())

	restoreChunk := database.SetResultsPropagateDrainChunkSizeForTests(1)
	defer restoreChunk()

	txnCount := 0
	restoreHook := database.SetResultsPropagateDrainInsideChunkHookForTests(
		func(store *database.DataStore, _, _ int64, phase database.ResultsPropagateDrainChunkPhase) {
			// Count each chunk transaction once (AfterInsert shares the txn with AfterDelete).
			if phase == database.ResultsPropagateDrainChunkAfterDelete ||
				phase == database.ResultsPropagateDrainChunkSelectEmpty {
				txnCount++
			}
			if phase == database.ResultsPropagateDrainChunkAfterInsert {
				assertDrainTransactionIsolationReadCommitted(t, store)
			}
		})
	defer restoreHook()

	assert.True(t, moveFromResultsPropagateToResultsPropagateInternal(store))
	assert.Equal(t, 3, txnCount,
		"row-walking drain should run 2 row chunks + 1 trailing empty, not one txn per id gap")
	assertResultsMarkedAsChanged(t, store, "results_propagate", nil)
	assertResultsMarkedAsChanged(t, store, "results_propagate_internal", []resultPrimaryKeyAndState{
		{ResultPrimaryKey: ResultPrimaryKey{101, 1, 12}, State: "to_be_propagated"},
		{ResultPrimaryKey: ResultPrimaryKey{101, 1, 13}, State: "to_be_propagated"},
	})
}

// Regression for Error 1205: concurrent uncommitted API inserts into the inbox must not block the drain.
func Test_moveFromResultsPropagateToResultsPropagateInternal_NoLockWaitOnUncommittedInsert(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	ctx := testhelpers.CreateTestContext()
	db := testhelpers.SetupDBWithFixtureString(ctx, `
		groups: [{id: 101}]
		attempts: [{participant_id: 101, id: 1}]
		items: [{id: 12, default_language_tag: fr}]
		results: [{participant_id: 101, attempt_id: 1, item_id: 12}]
		results_propagate:
			- {participant_id: 101, attempt_id: 1, item_id: 12, state: to_be_propagated}
	`)
	defer func() { _ = db.Close() }()

	rawDBA := testhelpers.OpenRawDBConnection(ctx)
	defer func() { _ = rawDBA.Close() }()

	connA, err := rawDBA.Conn(ctx)
	require.NoError(t, err)
	defer func() { _ = connA.Close() }()

	txA, err := connA.BeginTx(ctx, nil)
	require.NoError(t, err)
	defer func() { _ = txA.Rollback() }()

	_, err = txA.ExecContext(ctx, `
		INSERT INTO results_propagate (participant_id, attempt_id, item_id, state)
		VALUES (101, 1, 12, 'to_be_recomputed')`)
	require.NoError(t, err)

	assert.NotPanics(t, func() {
		assert.True(t, moveFromResultsPropagateToResultsPropagateInternal(database.NewDataStore(db)))
	})
}

func Test_moveFromResultsPropagateToResultsPropagateInternal_LocksOnlyCurrentChunkRange(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	ctx := testhelpers.CreateTestContext()
	db := testhelpers.SetupDBWithFixtureString(ctx, `
		groups: [{id: 101}, {id: 102}]
		attempts: [{participant_id: 101, id: 1}, {participant_id: 102, id: 1}]
		items:
			- {id: 12, default_language_tag: fr}
			- {id: 13, default_language_tag: fr}
			- {id: 14, default_language_tag: fr}
		results:
			- {participant_id: 101, attempt_id: 1, item_id: 12}
			- {participant_id: 101, attempt_id: 1, item_id: 13}
			- {participant_id: 102, attempt_id: 1, item_id: 14}
		results_propagate:
			- {participant_id: 101, attempt_id: 1, item_id: 12, state: to_be_propagated}
			- {participant_id: 101, attempt_id: 1, item_id: 13, state: to_be_propagated}
			- {participant_id: 102, attempt_id: 1, item_id: 14, state: to_be_propagated}
	`)
	defer func() { _ = db.Close() }()

	rawProbe := testhelpers.OpenRawDBConnection(ctx)
	defer func() { _ = rawProbe.Close() }()
	probeConn, err := rawProbe.Conn(ctx)
	require.NoError(t, err)
	defer func() { _ = probeConn.Close() }()

	var performanceSchemaEnabled int
	require.NoError(t, probeConn.QueryRowContext(ctx, "SELECT @@performance_schema").Scan(&performanceSchemaEnabled))
	if performanceSchemaEnabled == 0 {
		t.Skip("performance_schema is disabled")
	}

	restoreChunk := database.SetResultsPropagateDrainChunkSizeForTests(1)
	defer restoreChunk()

	seenChunk := false
	probedDataLocks := false
	seenRecordLock := false
	restoreHook := database.SetResultsPropagateDrainInsideChunkHookForTests(
		func(_ *database.DataStore, fromID, toID int64, phase database.ResultsPropagateDrainChunkPhase) {
			// Probe after DELETE: that is when inbox record locks are still held until commit.
			// Before DELETE the RC path holds no locks on results_propagate at all.
			if phase != database.ResultsPropagateDrainChunkAfterDelete {
				return
			}
			seenChunk = true

			rows, err := probeConn.QueryContext(ctx, `
				SELECT LOCK_TYPE, LOCK_DATA
				FROM performance_schema.data_locks
				WHERE OBJECT_SCHEMA = DATABASE()
				  AND OBJECT_NAME = 'results_propagate'`)
			if err != nil {
				t.Logf("skipping data_locks assertion (need SELECT on performance_schema.data_locks): %v", err)
				return
			}
			defer func() { _ = rows.Close() }()
			probedDataLocks = true

			for rows.Next() {
				var lockType string
				var lockData sql.NullString
				require.NoError(t, rows.Scan(&lockType, &lockData))
				if lockType == "TABLE" {
					continue
				}
				require.Equal(t, "RECORD", lockType)
				if !lockData.Valid {
					continue
				}
				lockID, err := strconv.ParseInt(lockData.String, 10, 64)
				if err != nil {
					assert.Fail(t, "unexpected non-numeric results_propagate lock",
						"LOCK_DATA=%q (supremum/gap locks must not appear on the RC drain path)", lockData.String)
					continue
				}
				seenRecordLock = true
				assert.Greater(t, lockID, fromID, "lock on id=%d outside chunk (%d, %d]", lockID, fromID, toID)
				assert.LessOrEqual(t, lockID, toID, "lock on id=%d outside chunk (%d, %d]", lockID, fromID, toID)
			}
			require.NoError(t, rows.Err())
		})
	defer restoreHook()

	assert.True(t, moveFromResultsPropagateToResultsPropagateInternal(database.NewDataStore(db)))
	assert.True(t, seenChunk)
	if probedDataLocks {
		assert.True(t, seenRecordLock,
			"after DELETE the chunk transaction must hold at least one RECORD lock on results_propagate")
	}
}

func Test_moveFromResultsPropagateToResultsPropagateInternal_SingleRangeDeleteCoversJoinedAndOrphans(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db := testhelpers.SetupDBWithFixtureString(testhelpers.CreateTestContext(), `
		groups: [{id: 101}]
		attempts: [{participant_id: 101, id: 1}]
		items:
			- {id: 12, default_language_tag: fr}
			- {id: 13, default_language_tag: fr}
		results:
			- {participant_id: 101, attempt_id: 1, item_id: 12}
		results_propagate:
			- {participant_id: 101, attempt_id: 1, item_id: 12, state: to_be_propagated}
			- {participant_id: 101, attempt_id: 1, item_id: 13, state: to_be_propagated}
	`)
	defer func() { _ = db.Close() }()

	store := database.NewDataStore(db)
	assert.True(t, moveFromResultsPropagateToResultsPropagateInternal(store))

	assertResultsMarkedAsChanged(t, store, "results_propagate", nil)
	assertResultsMarkedAsChanged(t, store, "results_propagate_internal", []resultPrimaryKeyAndState{
		{ResultPrimaryKey: ResultPrimaryKey{101, 1, 12}, State: "to_be_propagated"},
	})
}

func Test_moveFromResultsPropagateToResultsPropagateInternal_RowsAppendedMidDrainSurvive(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	ctx := testhelpers.CreateTestContext()
	db := testhelpers.SetupDBWithFixtureString(ctx, `
		groups: [{id: 101}]
		attempts: [{participant_id: 101, id: 1}]
		items:
			- {id: 12, default_language_tag: fr}
			- {id: 13, default_language_tag: fr}
			- {id: 14, default_language_tag: fr}
		results:
			- {participant_id: 101, attempt_id: 1, item_id: 12}
			- {participant_id: 101, attempt_id: 1, item_id: 13}
			- {participant_id: 101, attempt_id: 1, item_id: 14}
		results_propagate:
			- {participant_id: 101, attempt_id: 1, item_id: 12, state: to_be_propagated}
			- {participant_id: 101, attempt_id: 1, item_id: 13, state: to_be_propagated}
	`)
	defer func() { _ = db.Close() }()

	rawDB := testhelpers.OpenRawDBConnection(ctx)
	defer func() { _ = rawDB.Close() }()
	appendConn, err := rawDB.Conn(ctx)
	require.NoError(t, err)
	defer func() { _ = appendConn.Close() }()

	restoreChunk := database.SetResultsPropagateDrainChunkSizeForTests(1)
	defer restoreChunk()

	var survivingID int64
	restoreHook := database.SetResultsPropagateDrainInsideChunkHookForTests(
		func(_ *database.DataStore, _, _ int64, phase database.ResultsPropagateDrainChunkPhase) {
			if phase != database.ResultsPropagateDrainChunkAfterInsert || survivingID != 0 {
				return
			}
			result, err := appendConn.ExecContext(ctx, `
				INSERT INTO results_propagate (participant_id, attempt_id, item_id, state)
				VALUES (101, 1, 14, 'to_be_propagated')`)
			require.NoError(t, err)
			survivingID, err = result.LastInsertId()
			require.NoError(t, err)
		})
	defer restoreHook()

	store := database.NewDataStore(db)
	assert.True(t, moveFromResultsPropagateToResultsPropagateInternal(store))
	require.NotZero(t, survivingID)

	var countRow struct{ Count int }
	require.NoError(t, store.Raw(
		"SELECT COUNT(*) AS count FROM results_propagate WHERE id = ?", survivingID,
	).Scan(&countRow).Error())
	assert.Equal(t, 1, countRow.Count)

	assertResultsMarkedAsChanged(t, store, "results_propagate_internal", []resultPrimaryKeyAndState{
		{ResultPrimaryKey: ResultPrimaryKey{101, 1, 12}, State: "to_be_propagated"},
		{ResultPrimaryKey: ResultPrimaryKey{101, 1, 13}, State: "to_be_propagated"},
	})
}

// Late-committed inbox row whose id is below the high-water mark must not be deleted unless the
// INSERT also considered it. Under READ COMMITTED a fresh read view between INSERT and DELETE
// would otherwise drop it silently.
func Test_moveFromResultsPropagateToResultsPropagateInternal_LateCommitBelowHWMNotLost(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	ctx := testhelpers.CreateTestContext()
	db := testhelpers.SetupDBWithFixtureString(ctx, `
		groups: [{id: 101}]
		attempts: [{participant_id: 101, id: 1}]
		items:
			- {id: 12, default_language_tag: fr}
			- {id: 13, default_language_tag: fr}
		results:
			- {participant_id: 101, attempt_id: 1, item_id: 12}
			- {participant_id: 101, attempt_id: 1, item_id: 13}
	`)
	defer func() { _ = db.Close() }()

	rawDBA := testhelpers.OpenRawDBConnection(ctx)
	defer func() { _ = rawDBA.Close() }()
	connA, err := rawDBA.Conn(ctx)
	require.NoError(t, err)
	defer func() { _ = connA.Close() }()

	// A inserts first (lower auto-increment id) and holds the transaction open.
	txA, err := connA.BeginTx(ctx, nil)
	require.NoError(t, err)
	defer func() { _ = txA.Rollback() }()
	resultA, err := txA.ExecContext(ctx, `
		INSERT INTO results_propagate (participant_id, attempt_id, item_id, state)
		VALUES (101, 1, 12, 'to_be_propagated')`)
	require.NoError(t, err)
	lateID, err := resultA.LastInsertId()
	require.NoError(t, err)

	// Committed row gets a higher id; MAX(id) therefore sits above lateID.
	store := database.NewDataStore(db)
	require.NoError(t, store.Exec(`
		INSERT INTO results_propagate (participant_id, attempt_id, item_id, state)
		VALUES (101, 1, 13, 'to_be_propagated')`).Error())

	restoreChunk := database.SetResultsPropagateDrainChunkSizeForTests(1)
	defer restoreChunk()

	committed := false
	restoreHook := database.SetResultsPropagateDrainInsideChunkHookForTests(
		func(_ *database.DataStore, _, _ int64, phase database.ResultsPropagateDrainChunkPhase) {
			// Commit after INSERT so DELETE still uses the pre-commit id list.
			if phase != database.ResultsPropagateDrainChunkAfterInsert || committed {
				return
			}
			committed = true
			require.NoError(t, txA.Commit())
		})
	defer restoreHook()

	assert.True(t, moveFromResultsPropagateToResultsPropagateInternal(store))

	var propagateCount struct{ Count int }
	require.NoError(t, store.Raw(
		"SELECT COUNT(*) AS count FROM results_propagate WHERE id = ?", lateID,
	).Scan(&propagateCount).Error())
	var internalCount struct{ Count int }
	require.NoError(t, store.Raw(`
		SELECT COUNT(*) AS count FROM results_propagate_internal
		WHERE participant_id = 101 AND attempt_id = 1 AND item_id = 12`).Scan(&internalCount).Error())

	assert.GreaterOrEqual(t, propagateCount.Count+internalCount.Count, 1,
		"late-committed row id=%d must remain in results_propagate (%d) or results_propagate_internal (%d)",
		lateID, propagateCount.Count, internalCount.Count)
}

func assertDrainTransactionIsolationReadCommitted(t *testing.T, store *database.DataStore) {
	t.Helper()

	// SET TRANSACTION (used by database/sql BeginTx) does not change @@transaction_isolation;
	// the live transaction level is on innodb_trx (needs PROCESS).
	var isoRow struct {
		Isolation string `gorm:"column:trx_isolation_level"`
	}
	err := store.Raw(`
		SELECT trx_isolation_level
		FROM information_schema.innodb_trx
		WHERE trx_mysql_thread_id = CONNECTION_ID()`).Scan(&isoRow).Error()
	if err != nil {
		if strings.Contains(err.Error(), "Access denied") || strings.Contains(err.Error(), "1227") {
			t.Logf("skipping READ COMMITTED assertion (need PROCESS for innodb_trx): %v", err)
			return
		}
		require.NoError(t, err)
	}
	assert.Equal(t, "READ COMMITTED", isoRow.Isolation)
}
