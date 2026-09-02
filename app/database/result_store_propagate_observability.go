package database

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unicode/utf8"

	"github.com/go-sql-driver/mysql"

	"github.com/France-ioi/AlgoreaBackend/v2/app/logging"
)

// Caps keep a flood of abandoned transactions from exploding a single Warn line.
const (
	maxLongRunningInnodbTrxRowsInLog = 10
	// Must match the LIMIT literal in longRunningInnodbTrxQuery (fetch one extra so we can
	// report that more rows exist without an unbounded transfer).
	longRunningInnodbTrxFetchLimit = 11
	maxInnodbTrxQueryLenInLog      = 200

	// MySQL ER_SPECIFIC_ACCESS_DENIED_ERROR / ER_NO_SUCH_TABLE — permanent for this process.
	mysqlERSpecificAccessDenied = 1227
	mysqlERNoSuchTable          = 1146
)

//nolint:gochecknoglobals // overridable via config / tests
var (
	propagationLogChunkCountersMu sync.RWMutex
	propagationLogChunkCounters   = true

	// Permanent degrade after empty/NULL counters or a permanent MySQL error (performance_schema
	// OFF, missing table privileges, etc.) so we do not pay a round-trip every chunk.
	propagationStatusCountersUnavailable atomic.Bool
	// Permanent degrade after a permanent INNODB_TRX failure (e.g. missing PROCESS privilege).
	innodbTrxDumpUnavailable atomic.Bool

	// Transient observability failures use a cooldown instead of a one-way latch so an incident
	// blip does not silence diagnostics for the rest of a long-lived API process.
	propagationStatusCountersCooldownUntilNano atomic.Int64
	innodbTrxDumpCooldownUntilNano             atomic.Int64
	// Default cooldown after a transient observability failure.
	propagationObservabilityCooldown = 5 * time.Minute
)

// SetPropagationLogChunkCounters enables or disables per-chunk performance_schema counter
// snapshots attached to propagation duration logs (config: propagation.logChunkCounters).
// Clearing latches here lets app.Reset revive diagnostics after a prior degrade.
func SetPropagationLogChunkCounters(enabled bool) {
	propagationLogChunkCountersMu.Lock()
	defer propagationLogChunkCountersMu.Unlock()
	propagationLogChunkCounters = enabled
	resetPropagationObservabilityState()
}

// PropagationLogChunkCountersEnabled reports whether per-chunk counter snapshots are enabled.
func PropagationLogChunkCountersEnabled() bool {
	return getPropagationLogChunkCounters()
}

func getPropagationLogChunkCounters() bool {
	propagationLogChunkCountersMu.RLock()
	defer propagationLogChunkCountersMu.RUnlock()
	return propagationLogChunkCounters
}

func resetPropagationObservabilityState() {
	propagationStatusCountersUnavailable.Store(false)
	innodbTrxDumpUnavailable.Store(false)
	propagationStatusCountersCooldownUntilNano.Store(0)
	innodbTrxDumpCooldownUntilNano.Store(0)
}

func countersObservabilityBlocked() bool {
	if propagationStatusCountersUnavailable.Load() {
		return true
	}
	until := propagationStatusCountersCooldownUntilNano.Load()
	return until > 0 && time.Now().UnixNano() < until
}

func innodbTrxObservabilityBlocked() bool {
	if innodbTrxDumpUnavailable.Load() {
		return true
	}
	until := innodbTrxDumpCooldownUntilNano.Load()
	return until > 0 && time.Now().UnixNano() < until
}

func isPermanentObservabilityError(err error) bool {
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) {
		switch mysqlErr.Number {
		case mysqlERSpecificAccessDenied, mysqlERNoSuchTable:
			return true
		}
	}
	// Plain errors from mocks / older drivers still carry the privilege text.
	msg := err.Error()
	return strings.Contains(msg, "Access denied") || strings.Contains(msg, "PROCESS privilege")
}

func noteCountersQueryError(err error) {
	if isPermanentObservabilityError(err) {
		propagationStatusCountersUnavailable.Store(true)
		return
	}
	propagationStatusCountersCooldownUntilNano.Store(
		time.Now().Add(propagationObservabilityCooldown).UnixNano())
}

// propagationStatusSnapshot holds status counters for one chunk boundary.
// Innodb_rows_read and Innodb_buffer_pool_read_requests come from performance_schema.global_status
// (server-wide: deltas include concurrent API traffic, so under load "high rows read" can be other
// sessions and must not alone refute a lock-wait diagnosis). Handler_read_key comes from
// session_status (this connection only) — before/after must share a transaction or fixed
// connection (enforced by beginPropagationChunkCounters).
// Interpretation when paired with chunk duration (the point of collecting these):
//   - high Innodb_rows_read + high duration → examining too many rows (fan-out), if traffic is quiet
//   - low Innodb_rows_read + high duration → blocked, not working (e.g. 2026-07-24)
//
// Prefer the INNODB_TRX dump on slow chunks as the uncontaminated "who is blocking" signal.
type propagationStatusSnapshot struct {
	innodbRowsRead               int64
	handlerReadKey               int64
	innodbBufferPoolReadRequests int64
	ok                           bool
}

// Single-row form so we can use queryRowWithoutLogging (no Info SQL log lines).
// Handler_read_key is session-scoped; InnoDB counters are global-only in MySQL 8.0.
const propagationStatusCountersQuery = `
	SELECT
		(SELECT VARIABLE_VALUE FROM performance_schema.global_status
			WHERE VARIABLE_NAME = 'Innodb_rows_read'),
		(SELECT VARIABLE_VALUE FROM performance_schema.session_status
			WHERE VARIABLE_NAME = 'Handler_read_key'),
		(SELECT VARIABLE_VALUE FROM performance_schema.global_status
			WHERE VARIABLE_NAME = 'Innodb_buffer_pool_read_requests')`

// Oldest long-running transactions first; LIMIT bounds transfer on a busy server.
// Requires GRANT PROCESS ON *.* for the application DB user.
const longRunningInnodbTrxQuery = `
	SELECT trx_id, trx_state, trx_started, trx_mysql_thread_id, trx_rows_locked, trx_query
	FROM information_schema.INNODB_TRX
	WHERE trx_started < NOW() - INTERVAL 30 SECOND
	ORDER BY trx_started
	LIMIT 11`

// beginPropagationChunkCounters snapshots status counters when the config flag is on.
// Returns nil when disabled, degraded/cooling down, not on a same-session-safe store, or the
// snapshot is not usable, so callers skip the after-snapshot (zero extra queries on the common
// performance_schema=OFF path). Same-session: requires isInTransaction() || isFixed() so
// session-scoped Handler_read_key before/after cannot hit different pool connections.
func beginPropagationChunkCounters(store *DataStore) *propagationStatusSnapshot {
	if !getPropagationLogChunkCounters() || countersObservabilityBlocked() {
		return nil
	}
	if !store.isInTransaction() && !store.isFixed() {
		return nil
	}
	snap, err := readPropagationStatusSnapshot(store)
	if err != nil {
		noteCountersQueryError(err)
		return nil
	}
	if !snap.ok {
		// Empty/NULL/unparseable: typical when performance_schema is OFF — permanent.
		propagationStatusCountersUnavailable.Store(true)
		return nil
	}
	return &snap
}

func readPropagationStatusSnapshot(store *DataStore) (propagationStatusSnapshot, error) {
	row := store.queryRowWithoutLogging(propagationStatusCountersQuery)
	var innodbRowsRead, handlerReadKey, bufferPool sql.NullString
	if err := row.Scan(&innodbRowsRead, &handlerReadKey, &bufferPool); err != nil {
		return propagationStatusSnapshot{}, err
	}
	if !innodbRowsRead.Valid || !handlerReadKey.Valid || !bufferPool.Valid {
		return propagationStatusSnapshot{}, nil
	}

	rowsRead, err1 := strconv.ParseInt(innodbRowsRead.String, 10, 64)
	readKey, err2 := strconv.ParseInt(handlerReadKey.String, 10, 64)
	poolReqs, err3 := strconv.ParseInt(bufferPool.String, 10, 64)
	if err1 != nil || err2 != nil || err3 != nil {
		// Unparseable values are an empty snapshot (permanent degrade), not a retryable query error.
		return propagationStatusSnapshot{}, nil //nolint:nilerr // deliberate: treat like empty/NULL
	}
	return propagationStatusSnapshot{
		innodbRowsRead:               rowsRead,
		handlerReadKey:               readKey,
		innodbBufferPoolReadRequests: poolReqs,
		ok:                           true,
	}, nil
}

func nonNegativeDelta(after, before int64) int64 {
	delta := after - before
	if delta < 0 {
		return 0
	}
	return delta
}

// formatPropagationCounterDeltas returns a suffix for the duration log, or "" when either
// snapshot is missing/degraded so the duration-only path remains usable without performance_schema.
func formatPropagationCounterDeltas(before, after propagationStatusSnapshot) string {
	if !before.ok || !after.ok {
		return ""
	}
	return fmt.Sprintf(
		", Innodb_rows_read=%d, Handler_read_key=%d, Innodb_buffer_pool_read_requests=%d",
		nonNegativeDelta(after.innodbRowsRead, before.innodbRowsRead),
		nonNegativeDelta(after.handlerReadKey, before.handlerReadKey),
		nonNegativeDelta(after.innodbBufferPoolReadRequests, before.innodbBufferPoolReadRequests),
	)
}

func formatLongRunningInnodbTrxDump(store *DataStore) string {
	if innodbTrxObservabilityBlocked() {
		return ""
	}

	rows, err := store.queryRowsWithoutLogging(longRunningInnodbTrxQuery)
	if err != nil {
		noteInnodbTrxFailure(store, err)
		return ""
	}
	defer func() { _ = rows.Close() }()

	parts := make([]string, 0, maxLongRunningInnodbTrxRowsInLog)
	var fetched int
	for rows.Next() {
		fetched++
		var (
			trxID, trxState, trxThread, trxRowsLocked sql.NullString
			trxStarted                                sql.NullTime
			trxQuery                                  sql.NullString
		)
		if scanErr := rows.Scan(&trxID, &trxState, &trxStarted, &trxThread, &trxRowsLocked, &trxQuery); scanErr != nil {
			noteInnodbTrxFailure(store, scanErr)
			return ""
		}
		if len(parts) < maxLongRunningInnodbTrxRowsInLog {
			parts = append(parts, formatInnodbTrxRowValues(trxID, trxState, trxStarted, trxThread, trxRowsLocked, trxQuery))
		}
	}
	if err := rows.Err(); err != nil {
		noteInnodbTrxFailure(store, err)
		return ""
	}
	if len(parts) == 0 {
		return ""
	}

	dump := strings.Join(parts, "; ")
	if fetched > maxLongRunningInnodbTrxRowsInLog {
		// LIMIT caps fetched at longRunningInnodbTrxFetchLimit, so we cannot report an exact remainder.
		dump += " (+1 or more)"
	}
	return dump
}

func noteInnodbTrxFailure(store *DataStore, err error) {
	if isPermanentObservabilityError(err) {
		disableInnodbTrxDumpPermanently(store, err)
		return
	}
	innodbTrxDumpCooldownUntilNano.Store(
		time.Now().Add(propagationObservabilityCooldown).UnixNano())
}

func disableInnodbTrxDumpPermanently(store *DataStore, err error) {
	if !innodbTrxDumpUnavailable.CompareAndSwap(false, true) {
		return
	}
	logging.EntryFromContext(store.ctx()).Warnf(
		"disabling INNODB_TRX dumps for this process (need GRANT PROCESS ON *.*): %v", err)
}

// disableInnodbTrxDump permanently disables dumps (tests / permanent-error path).
func disableInnodbTrxDump(store *DataStore, err error) {
	disableInnodbTrxDumpPermanently(store, err)
}

func formatInnodbTrxRowValues(
	trxID, trxState sql.NullString, trxStarted sql.NullTime, trxThread, trxRowsLocked, trxQuery sql.NullString,
) string {
	query := ""
	if trxQuery.Valid {
		query = truncateUTF8(trxQuery.String, maxInnodbTrxQueryLenInLog)
	}
	started := ""
	if trxStarted.Valid {
		started = trxStarted.Time.Format("2006-01-02 15:04:05")
	}
	return fmt.Sprintf(
		"trx_id=%s state=%s started=%s thread=%s rows_locked=%s query=%q",
		nullStringValue(trxID),
		nullStringValue(trxState),
		started,
		nullStringValue(trxThread),
		nullStringValue(trxRowsLocked),
		query,
	)
}

func nullStringValue(value sql.NullString) string {
	if !value.Valid {
		return ""
	}
	return value.String
}

func truncateUTF8(value string, maxBytes int) string {
	if len(value) <= maxBytes {
		return value
	}
	for maxBytes > 0 && !utf8.RuneStart(value[maxBytes]) {
		maxBytes--
	}
	return value[:maxBytes] + "…"
}

// logPropagationStepDurationf logs chunk duration at Debug, or Warn when slow.
// Counter deltas (if any) are folded into the same line. The INNODB_TRX dump is appended after
// Sprintf so arbitrary SQL text (including '%') cannot corrupt the printf format.
// Called while the chunk transaction may still be open: the dump is a small information_schema
// read considered acceptable vs. adding a second post-commit observability call site.
func logPropagationStepDurationf(
	store *DataStore, duration time.Duration, before *propagationStatusSnapshot, format string, args ...interface{},
) {
	entry := logging.EntryFromContext(store.ctx())
	slow := duration >= getResultsPropagationSlowChunkThreshold()
	debugEnabled := logging.LoggerFromContext(store.ctx()).IsDebugEnabled()

	// Fast chunks at info/warn (API production) discard Debug lines — skip after-snapshot and Sprintf.
	if !slow && !debugEnabled {
		return
	}

	if before != nil {
		after, err := readPropagationStatusSnapshot(store)
		if err != nil {
			noteCountersQueryError(err)
		} else if !after.ok {
			propagationStatusCountersUnavailable.Store(true)
		} else if deltas := formatPropagationCounterDeltas(*before, after); deltas != "" {
			format += deltas
		}
	}

	args = append(append([]interface{}{}, args...), duration)
	format += ", took %v"
	msg := fmt.Sprintf(format, args...)

	if slow {
		if trxDump := formatLongRunningInnodbTrxDump(store); trxDump != "" {
			msg += "; long-running transactions: " + trxDump
		}
		entry.Warn(msg)
		return
	}
	entry.Debug(msg)
}

func (conn *DB) queryRowWithoutLogging(query string, args ...interface{}) *sql.Row {
	//nolint:forcetypeassert // CommonDB is always one of our wrappers
	return conn.db.CommonDB().(queryRowWithoutLogging).queryRowWithoutLogging(query, args...)
}

func (conn *DB) queryRowsWithoutLogging(query string, args ...interface{}) (*sql.Rows, error) {
	//nolint:forcetypeassert // CommonDB is always one of our wrappers
	return conn.db.CommonDB().(queryRowsWithoutLogging).queryRowsWithoutLogging(query, args...)
}

type queryRowsWithoutLogging interface {
	queryRowsWithoutLogging(query string, args ...interface{}) (*sql.Rows, error)
}
