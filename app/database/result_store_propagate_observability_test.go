package database

import (
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/France-ioi/AlgoreaBackend/v2/app/logging"
	"github.com/France-ioi/AlgoreaBackend/v2/app/loggingtest"
	"github.com/France-ioi/AlgoreaBackend/v2/testhelpers/testoutput"
)

func resetPropagationObservabilityStateForTest(t *testing.T) {
	t.Helper()
	resetPropagationObservabilityState()
	t.Cleanup(resetPropagationObservabilityState)
}

func TestFormatPropagationCounterDeltas(t *testing.T) {
	before := propagationStatusSnapshot{
		innodbRowsRead: 100, handlerReadKey: 50, innodbBufferPoolReadRequests: 200, ok: true,
	}
	after := propagationStatusSnapshot{
		innodbRowsRead: 130, handlerReadKey: 55, innodbBufferPoolReadRequests: 250, ok: true,
	}
	assert.Equal(t,
		", Innodb_rows_read=30, Handler_read_key=5, Innodb_buffer_pool_read_requests=50",
		formatPropagationCounterDeltas(before, after))
	assert.Empty(t, formatPropagationCounterDeltas(propagationStatusSnapshot{}, after))
	assert.Empty(t, formatPropagationCounterDeltas(before, propagationStatusSnapshot{}))

	// Negative deltas (e.g. after server restart) are clamped to zero.
	wrapped := propagationStatusSnapshot{
		innodbRowsRead: 10, handlerReadKey: 1, innodbBufferPoolReadRequests: 5, ok: true,
	}
	assert.Equal(t,
		", Innodb_rows_read=0, Handler_read_key=0, Innodb_buffer_pool_read_requests=0",
		formatPropagationCounterDeltas(before, wrapped))
}

func TestTruncateUTF8(t *testing.T) {
	assert.Equal(t, "abc", truncateUTF8("abc", 10))
	// "é" is 2 bytes; cutting at 1 must not leave a partial rune.
	assert.Equal(t, "…", truncateUTF8("éxxx", 1))
	assert.Equal(t, "a…", truncateUTF8("aéxxx", 2))
}

func TestBeginPropagationChunkCounters_Disabled(t *testing.T) {
	restore := SetPropagationLogChunkCountersForTests(false)
	defer restore()
	resetPropagationObservabilityStateForTest(t)

	ctx, _, _ := logging.NewContextWithNewMockLogger()
	db, mock := NewDBMock(ctx)
	defer func() { _ = db.Close() }()

	assert.Nil(t, beginPropagationChunkCounters(NewDataStore(db)))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestBeginPropagationChunkCounters_RequiresSameSession(t *testing.T) {
	restore := SetPropagationLogChunkCountersForTests(true)
	defer restore()
	resetPropagationObservabilityStateForTest(t)

	ctx, _, _ := logging.NewContextWithNewMockLogger()
	db, mock := NewDBMock(ctx)
	defer func() { _ = db.Close() }()

	// Pooled store: fail-safe nil without querying (Handler_read_key would be cross-session).
	assert.Nil(t, beginPropagationChunkCounters(NewDataStore(db)))
	require.NoError(t, mock.ExpectationsWereMet())

	mock.ExpectBegin()
	expectStatusCountersRow(mock, "100", "10", "1000")
	mock.ExpectCommit()
	require.NoError(t, NewDataStore(db).InTransaction(func(store *DataStore) error {
		before := beginPropagationChunkCounters(store)
		require.NotNil(t, before)
		assert.True(t, before.ok)
		return nil
	}))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestBeginPropagationChunkCounters_FixedConnection(t *testing.T) {
	restore := SetPropagationLogChunkCountersForTests(true)
	defer restore()
	resetPropagationObservabilityStateForTest(t)

	ctx, _, _ := logging.NewContextWithNewMockLogger()
	db, mock := NewDBMock(ctx)
	defer func() { _ = db.Close() }()

	mock.ExpectQuery(propagationStatusCountersQueryRegexp()).
		WillReturnRows(mock.NewRows([]string{"c1", "c2", "c3"}).AddRow("1", "2", "3"))
	require.NoError(t, db.WithFixedConnection(func(fixed *DB) error {
		before := beginPropagationChunkCounters(NewDataStore(fixed))
		require.NotNil(t, before)
		return nil
	}))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReadPropagationStatusSnapshot_SuccessAndDeltasViaLog(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	restoreCounters := SetPropagationLogChunkCountersForTests(true)
	defer restoreCounters()
	restoreThreshold := SetResultsPropagationSlowChunkThresholdForTests(time.Hour)
	defer restoreThreshold()
	resetPropagationObservabilityStateForTest(t)

	ctx, logger, logHook := logging.NewContextWithNewMockLogger()
	logger.ForceDebugLevel()
	db, mock := NewDBMock(ctx)
	defer func() { _ = db.Close() }()

	mock.ExpectBegin()
	expectStatusCountersRow(mock, "100", "10", "1000")
	expectStatusCountersRow(mock, "150", "12", "1100")
	mock.ExpectCommit()

	require.NoError(t, NewDataStore(db).InTransaction(func(store *DataStore) error {
		before := beginPropagationChunkCounters(store)
		require.NotNil(t, before)
		require.True(t, before.ok)
		logPropagationStepDurationf(store, time.Millisecond, before, "chunk done: %d", 1)
		return nil
	}))

	logs := (&loggingtest.Hook{Hook: logHook}).GetAllLogs()
	assert.Contains(t, logs, "Innodb_rows_read=50")
	assert.Contains(t, logs, "Handler_read_key=2")
	assert.Contains(t, logs, "Innodb_buffer_pool_read_requests=100")
	assert.NotContains(t, logs, "performance_schema")
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestLogPropagationStepDurationf_SkipsAfterSnapshotWhenDebugDisabled(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	restoreCounters := SetPropagationLogChunkCountersForTests(true)
	defer restoreCounters()
	restoreThreshold := SetResultsPropagationSlowChunkThresholdForTests(time.Hour)
	defer restoreThreshold()
	resetPropagationObservabilityStateForTest(t)

	ctx, logger, logHook := logging.NewContextWithNewMockLogger()
	_ = logger // default mock level is Info — Debug disabled
	db, mock := NewDBMock(ctx)
	defer func() { _ = db.Close() }()
	store := NewDataStore(db)

	before := &propagationStatusSnapshot{
		innodbRowsRead: 1, handlerReadKey: 1, innodbBufferPoolReadRequests: 1, ok: true,
	}
	// Fast chunk + info level: no after-snapshot query, no log line.
	logPropagationStepDurationf(store, time.Millisecond, before, "chunk")
	assert.Empty(t, (&loggingtest.Hook{Hook: logHook}).GetAllLogs())
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestBeginPropagationChunkCounters_DegradedReturnsNilAndSkipsAfter(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	restoreCounters := SetPropagationLogChunkCountersForTests(true)
	defer restoreCounters()
	restoreThreshold := SetResultsPropagationSlowChunkThresholdForTests(time.Hour)
	defer restoreThreshold()
	resetPropagationObservabilityStateForTest(t)

	ctx, logger, logHook := logging.NewContextWithNewMockLogger()
	logger.ForceDebugLevel()
	db, mock := NewDBMock(ctx)
	defer func() { _ = db.Close() }()

	mock.ExpectBegin()
	mock.ExpectQuery(propagationStatusCountersQueryRegexp()).
		WillReturnRows(mock.NewRows([]string{"c1", "c2", "c3"}).AddRow(nil, nil, nil))
	mock.ExpectCommit()

	require.NoError(t, NewDataStore(db).InTransaction(func(store *DataStore) error {
		before := beginPropagationChunkCounters(store)
		assert.Nil(t, before)
		assert.True(t, propagationStatusCountersUnavailable.Load())
		assert.Nil(t, beginPropagationChunkCounters(store))
		logPropagationStepDurationf(store, time.Millisecond, before, "chunk done: %d", 1)
		return nil
	}))

	logs := (&loggingtest.Hook{Hook: logHook}).GetAllLogs()
	assert.Contains(t, logs, "chunk done: 1")
	assert.NotContains(t, logs, "Innodb_rows_read=")
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReadPropagationStatusSnapshot_QueryError(t *testing.T) {
	resetPropagationObservabilityStateForTest(t)
	ctx, _, _ := logging.NewContextWithNewMockLogger()
	db, mock := NewDBMock(ctx)
	defer func() { _ = db.Close() }()

	mock.ExpectQuery(propagationStatusCountersQueryRegexp()).
		WillReturnError(errors.New("performance_schema disabled"))
	snap, err := readPropagationStatusSnapshot(NewDataStore(db))
	assert.False(t, snap.ok)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestBeginPropagationChunkCounters_TransientErrorUsesCooldown(t *testing.T) {
	restore := SetPropagationLogChunkCountersForTests(true)
	defer restore()
	resetPropagationObservabilityStateForTest(t)

	ctx, _, _ := logging.NewContextWithNewMockLogger()
	db, mock := NewDBMock(ctx)
	defer func() { _ = db.Close() }()

	mock.ExpectBegin()
	mock.ExpectQuery(propagationStatusCountersQueryRegexp()).
		WillReturnError(errors.New("driver: bad connection"))
	mock.ExpectCommit()

	require.NoError(t, NewDataStore(db).InTransaction(func(store *DataStore) error {
		assert.Nil(t, beginPropagationChunkCounters(store))
		assert.False(t, propagationStatusCountersUnavailable.Load())
		assert.True(t, countersObservabilityBlocked())
		assert.Nil(t, beginPropagationChunkCounters(store)) // cooldown: no second query
		return nil
	}))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReadPropagationStatusSnapshot_PartialRow(t *testing.T) {
	resetPropagationObservabilityStateForTest(t)
	ctx, _, _ := logging.NewContextWithNewMockLogger()
	db, mock := NewDBMock(ctx)
	defer func() { _ = db.Close() }()

	mock.ExpectQuery(propagationStatusCountersQueryRegexp()).
		WillReturnRows(mock.NewRows([]string{"c1", "c2", "c3"}).AddRow("100", nil, "1000"))
	snap, err := readPropagationStatusSnapshot(NewDataStore(db))
	require.NoError(t, err)
	assert.False(t, snap.ok)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReadPropagationStatusSnapshot_BadValue(t *testing.T) {
	resetPropagationObservabilityStateForTest(t)
	ctx, _, _ := logging.NewContextWithNewMockLogger()
	db, mock := NewDBMock(ctx)
	defer func() { _ = db.Close() }()

	mock.ExpectQuery(propagationStatusCountersQueryRegexp()).
		WillReturnRows(mock.NewRows([]string{"c1", "c2", "c3"}).AddRow("not-a-number", "1", "2"))
	snap, err := readPropagationStatusSnapshot(NewDataStore(db))
	require.NoError(t, err)
	assert.False(t, snap.ok)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestLogPropagationStepDurationf_SlowChunkDumpsInnodbTrx(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	restoreCounters := SetPropagationLogChunkCountersForTests(false)
	defer restoreCounters()
	restoreThreshold := SetResultsPropagationSlowChunkThresholdForTests(time.Nanosecond)
	defer restoreThreshold()
	resetPropagationObservabilityStateForTest(t)

	ctx, _, logHook := logging.NewContextWithNewMockLogger()
	db, mock := NewDBMock(ctx)
	defer func() { _ = db.Close() }()
	store := NewDataStore(db)

	// Percent signs in trx_query must not corrupt the Warn message (Critical 1 regression).
	queryWithPercent := "SELECT * FROM t WHERE name LIKE '%foo%' AND x = 100%"
	longQuery := strings.Repeat("SELECT ", 50)
	started1 := time.Date(2026, 7, 24, 10, 0, 0, 0, time.UTC)
	started2 := time.Date(2026, 7, 24, 10, 1, 0, 0, time.UTC)
	mock.ExpectQuery(longRunningInnodbTrxQueryRegexp()).
		WillReturnRows(mock.NewRows([]string{
			"trx_id", "trx_state", "trx_started", "trx_mysql_thread_id", "trx_rows_locked", "trx_query",
		}).AddRow("123", "RUNNING", started1, "42", "7", queryWithPercent).
			AddRow("124", "LOCK WAIT", started2, "43", "1", longQuery))

	logPropagationStepDurationf(store, time.Second, nil, "slow chunk: %d", 9)

	logs := (&loggingtest.Hook{Hook: logHook}).GetAllStructuredLogs()
	assert.Contains(t, logs, "level=warning")
	assert.Contains(t, logs, "slow chunk: 9")
	assert.Contains(t, logs, "long-running transactions:")
	assert.Contains(t, logs, "trx_id=123")
	assert.Contains(t, logs, "thread=42")
	assert.Contains(t, logs, `LIKE '%foo%'`)
	assert.NotContains(t, logs, "%!f(MISSING)")
	assert.Contains(t, logs, "…")
	assert.Contains(t, logs, "trx_id=124")
	assert.NotContains(t, logs, "INNODB_TRX")
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestLogPropagationStepDurationf_SlowChunkInnodbTrxErrorDisablesPermanently(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	restoreCounters := SetPropagationLogChunkCountersForTests(false)
	defer restoreCounters()
	restoreThreshold := SetResultsPropagationSlowChunkThresholdForTests(time.Nanosecond)
	defer restoreThreshold()
	resetPropagationObservabilityStateForTest(t)

	ctx, _, logHook := logging.NewContextWithNewMockLogger()
	db, mock := NewDBMock(ctx)
	defer func() { _ = db.Close() }()
	store := NewDataStore(db)

	mock.ExpectQuery(longRunningInnodbTrxQueryRegexp()).
		WillReturnError(&mysql.MySQLError{Number: mysqlERSpecificAccessDenied, Message: "Access denied; you need the PROCESS privilege"})

	logPropagationStepDurationf(store, time.Second, nil, "slow chunk")
	logPropagationStepDurationf(store, time.Second, nil, "slow chunk again")

	logs := (&loggingtest.Hook{Hook: logHook}).GetAllStructuredLogs()
	assert.Contains(t, logs, "disabling INNODB_TRX dumps")
	assert.Equal(t, 1, strings.Count(logs, "disabling INNODB_TRX dumps"))
	assert.NotContains(t, logs, "long-running transactions:")
	for _, line := range strings.Split(logs, "\n") {
		if strings.Contains(line, "Access denied") {
			assert.Regexp(t, `level=warning\b`, line)
		}
	}
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestFormatLongRunningInnodbTrxDump_CapsRowsOldestFirst(t *testing.T) {
	resetPropagationObservabilityStateForTest(t)
	ctx, _, _ := logging.NewContextWithNewMockLogger()
	db, mock := NewDBMock(ctx)
	defer func() { _ = db.Close() }()

	rows := mock.NewRows([]string{
		"trx_id", "trx_state", "trx_started", "trx_mysql_thread_id", "trx_rows_locked", "trx_query",
	})
	base := time.Date(2026, 7, 24, 10, 0, 0, 0, time.UTC)
	for i := range longRunningInnodbTrxFetchLimit {
		rows.AddRow(i, "RUNNING", base.Add(time.Duration(i)*time.Second), i, 0, "SELECT 1")
	}
	mock.ExpectQuery(longRunningInnodbTrxQueryRegexp()).WillReturnRows(rows)

	dump := formatLongRunningInnodbTrxDump(NewDataStore(db))
	assert.Contains(t, dump, "(+1 or more)")
	assert.Equal(t, maxLongRunningInnodbTrxRowsInLog, strings.Count(dump, "trx_id="))
	assert.Contains(t, dump, "trx_id=0")
	assert.NotContains(t, dump, fmt.Sprintf("trx_id=%d", maxLongRunningInnodbTrxRowsInLog))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestFormatLongRunningInnodbTrxDump_RowsErrUsesCooldown(t *testing.T) {
	testoutput.SuppressIfPasses(t)
	resetPropagationObservabilityStateForTest(t)

	ctx, _, _ := logging.NewContextWithNewMockLogger()
	db, mock := NewDBMock(ctx)
	defer func() { _ = db.Close() }()

	started := time.Date(2026, 7, 24, 10, 0, 0, 0, time.UTC)
	rows := mock.NewRows([]string{
		"trx_id", "trx_state", "trx_started", "trx_mysql_thread_id", "trx_rows_locked", "trx_query",
	}).AddRow("1", "RUNNING", started, "1", "0", "SELECT 1")
	rows.RowError(0, errors.New("row iteration failed"))
	mock.ExpectQuery(longRunningInnodbTrxQueryRegexp()).WillReturnRows(rows)

	assert.Empty(t, formatLongRunningInnodbTrxDump(NewDataStore(db)))
	assert.False(t, innodbTrxDumpUnavailable.Load())
	assert.True(t, innodbTrxObservabilityBlocked())
	assert.Empty(t, formatLongRunningInnodbTrxDump(NewDataStore(db))) // cooldown: no second query
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestFormatLongRunningInnodbTrxDump_EmptyResult(t *testing.T) {
	resetPropagationObservabilityStateForTest(t)
	ctx, _, _ := logging.NewContextWithNewMockLogger()
	db, mock := NewDBMock(ctx)
	defer func() { _ = db.Close() }()

	mock.ExpectQuery(longRunningInnodbTrxQueryRegexp()).
		WillReturnRows(mock.NewRows([]string{
			"trx_id", "trx_state", "trx_started", "trx_mysql_thread_id", "trx_rows_locked", "trx_query",
		}))
	assert.Empty(t, formatLongRunningInnodbTrxDump(NewDataStore(db)))
	assert.False(t, innodbTrxDumpUnavailable.Load())
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestFormatLongRunningInnodbTrxDump_ScanErrorUsesCooldown(t *testing.T) {
	testoutput.SuppressIfPasses(t)
	resetPropagationObservabilityStateForTest(t)

	ctx, _, _ := logging.NewContextWithNewMockLogger()
	db, mock := NewDBMock(ctx)
	defer func() { _ = db.Close() }()

	mock.ExpectQuery(longRunningInnodbTrxQueryRegexp()).
		WillReturnRows(mock.NewRows([]string{
			"trx_id", "trx_state", "trx_started", "trx_mysql_thread_id", "trx_rows_locked", "trx_query",
		}).AddRow("bad", "RUNNING", "not-a-time", "1", "0", nil))

	assert.Empty(t, formatLongRunningInnodbTrxDump(NewDataStore(db)))
	assert.False(t, innodbTrxDumpUnavailable.Load())
	assert.True(t, innodbTrxObservabilityBlocked())
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestDisableInnodbTrxDump_OnlyOnce(t *testing.T) {
	testoutput.SuppressIfPasses(t)
	resetPropagationObservabilityStateForTest(t)

	ctx, _, logHook := logging.NewContextWithNewMockLogger()
	db, _ := NewDBMock(ctx)
	defer func() { _ = db.Close() }()
	store := NewDataStore(db)

	permanent := &mysql.MySQLError{Number: mysqlERSpecificAccessDenied, Message: "Access denied"}
	disableInnodbTrxDump(store, permanent)
	disableInnodbTrxDump(store, permanent)
	logs := (&loggingtest.Hook{Hook: logHook}).GetAllLogs()
	assert.Equal(t, 1, strings.Count(logs, "disabling INNODB_TRX dumps"))
}

func TestBeginPropagationChunkCounters_PermanentErrorLatches(t *testing.T) {
	restore := SetPropagationLogChunkCountersForTests(true)
	defer restore()
	resetPropagationObservabilityStateForTest(t)

	ctx, _, _ := logging.NewContextWithNewMockLogger()
	db, mock := NewDBMock(ctx)
	defer func() { _ = db.Close() }()

	mock.ExpectBegin()
	mock.ExpectQuery(propagationStatusCountersQueryRegexp()).
		WillReturnError(&mysql.MySQLError{Number: mysqlERNoSuchTable, Message: "Table 'performance_schema.global_status' doesn't exist"})
	mock.ExpectCommit()

	require.NoError(t, NewDataStore(db).InTransaction(func(store *DataStore) error {
		assert.Nil(t, beginPropagationChunkCounters(store))
		assert.True(t, propagationStatusCountersUnavailable.Load())
		return nil
	}))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestLogPropagationStepDurationf_AfterSnapshotQueryError(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	restoreCounters := SetPropagationLogChunkCountersForTests(true)
	defer restoreCounters()
	restoreThreshold := SetResultsPropagationSlowChunkThresholdForTests(time.Hour)
	defer restoreThreshold()
	resetPropagationObservabilityStateForTest(t)

	ctx, logger, logHook := logging.NewContextWithNewMockLogger()
	logger.ForceDebugLevel()
	db, mock := NewDBMock(ctx)
	defer func() { _ = db.Close() }()
	store := NewDataStore(db)

	before := &propagationStatusSnapshot{
		innodbRowsRead: 1, handlerReadKey: 1, innodbBufferPoolReadRequests: 1, ok: true,
	}
	mock.ExpectQuery(propagationStatusCountersQueryRegexp()).
		WillReturnError(errors.New("driver: bad connection"))
	logPropagationStepDurationf(store, time.Millisecond, before, "chunk")
	assert.False(t, propagationStatusCountersUnavailable.Load())
	assert.True(t, countersObservabilityBlocked())
	assert.Contains(t, (&loggingtest.Hook{Hook: logHook}).GetAllLogs(), "chunk")
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestNullStringValue_Invalid(t *testing.T) {
	assert.Empty(t, nullStringValue(sql.NullString{}))
	assert.Equal(t, "x", nullStringValue(sql.NullString{String: "x", Valid: true}))
}

func TestLogPropagationStepDurationf_AfterSnapshotFailsDegrades(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	restoreCounters := SetPropagationLogChunkCountersForTests(true)
	defer restoreCounters()
	restoreThreshold := SetResultsPropagationSlowChunkThresholdForTests(time.Hour)
	defer restoreThreshold()
	resetPropagationObservabilityStateForTest(t)

	ctx, logger, logHook := logging.NewContextWithNewMockLogger()
	logger.ForceDebugLevel()
	db, mock := NewDBMock(ctx)
	defer func() { _ = db.Close() }()
	store := NewDataStore(db)

	before := &propagationStatusSnapshot{
		innodbRowsRead: 1, handlerReadKey: 1, innodbBufferPoolReadRequests: 1, ok: true,
	}
	mock.ExpectQuery(propagationStatusCountersQueryRegexp()).
		WillReturnRows(mock.NewRows([]string{"c1", "c2", "c3"}).AddRow(nil, nil, nil))
	logPropagationStepDurationf(store, time.Millisecond, before, "chunk")
	assert.True(t, propagationStatusCountersUnavailable.Load())
	assert.Contains(t, (&loggingtest.Hook{Hook: logHook}).GetAllLogs(), "chunk")
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestQueryWithoutLogging_ViaNamedLockConn(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	ctx, _, _ := logging.NewContextWithNewMockLogger()
	db, mock := NewDBMock(ctx)
	defer func() { _ = db.Close() }()

	mock.ExpectQuery(`(?s)^` + regexp.QuoteMeta("SELECT 1 AS ok") + `$`).
		WillReturnRows(mock.NewRows([]string{"ok"}).AddRow(1))
	mock.ExpectQuery(`(?s)^` + regexp.QuoteMeta("SELECT 2 AS ok") + `$`).
		WillReturnRows(mock.NewRows([]string{"ok"}).AddRow(2))

	//nolint:forcetypeassert // test reaches the conn wrapper used by named locks
	sqlDBW := db.db.CommonDB().(*sqlDBWrapper)
	connWrapper, err := sqlDBW.conn(ctx)
	require.NoError(t, err)
	defer func() { _ = connWrapper.close(nil) }()

	rows, err := connWrapper.queryRowsWithoutLogging("SELECT 1 AS ok")
	require.NoError(t, err)
	defer func() { _ = rows.Close() }()
	require.True(t, rows.Next())
	require.NoError(t, rows.Err())

	var got int
	require.NoError(t, connWrapper.queryRowWithoutLogging("SELECT 2 AS ok").Scan(&got))
	assert.Equal(t, 2, got)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestQueryRowsWithoutLogging_ViaTransaction(t *testing.T) {
	testoutput.SuppressIfPasses(t)
	resetPropagationObservabilityStateForTest(t)

	ctx, _, _ := logging.NewContextWithNewMockLogger()
	db, mock := NewDBMock(ctx)
	defer func() { _ = db.Close() }()

	mock.ExpectBegin()
	mock.ExpectQuery(longRunningInnodbTrxQueryRegexp()).
		WillReturnRows(mock.NewRows([]string{
			"trx_id", "trx_state", "trx_started", "trx_mysql_thread_id", "trx_rows_locked", "trx_query",
		}))
	mock.ExpectCommit()

	require.NoError(t, NewDataStore(db).InTransaction(func(store *DataStore) error {
		assert.Empty(t, formatLongRunningInnodbTrxDump(store))
		return nil
	}))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSetPropagationLogChunkCounters(t *testing.T) {
	restore := SetPropagationLogChunkCountersForTests(true)
	defer restore()

	propagationStatusCountersUnavailable.Store(true)
	innodbTrxDumpUnavailable.Store(true)

	SetPropagationLogChunkCounters(false)
	assert.False(t, getPropagationLogChunkCounters())
	assert.False(t, PropagationLogChunkCountersEnabled())
	assert.False(t, propagationStatusCountersUnavailable.Load())
	assert.False(t, innodbTrxDumpUnavailable.Load())

	SetPropagationLogChunkCounters(true)
	assert.True(t, getPropagationLogChunkCounters())
	assert.True(t, PropagationLogChunkCountersEnabled())
}

func TestLongRunningInnodbTrxFetchLimitMatchesQuery(t *testing.T) {
	assert.Equal(t, maxLongRunningInnodbTrxRowsInLog+1, longRunningInnodbTrxFetchLimit)
	assert.Contains(t, longRunningInnodbTrxQuery, "LIMIT 11")
}

func expectStatusCountersRow(mock sqlmock.Sqlmock, innodbRowsRead, handlerReadKey, bufferPool string) {
	mock.ExpectQuery(propagationStatusCountersQueryRegexp()).
		WillReturnRows(mock.NewRows([]string{"c1", "c2", "c3"}).
			AddRow(innodbRowsRead, handlerReadKey, bufferPool))
}

func propagationStatusCountersQueryRegexp() string {
	return "(?s)^" + regexp.QuoteMeta(strings.TrimSpace(propagationStatusCountersQuery)) + "$"
}

func longRunningInnodbTrxQueryRegexp() string {
	return "(?s)^" + regexp.QuoteMeta(strings.TrimSpace(longRunningInnodbTrxQuery)) + "$"
}
