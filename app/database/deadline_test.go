package database

import (
	"context"
	"regexp"
	"sync"
	"sync/atomic"
	"testing"
	"time"
	"unsafe"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/France-ioi/AlgoreaBackend/v2/app/logging"
	"github.com/France-ioi/AlgoreaBackend/v2/app/loggingtest"
	"github.com/France-ioi/AlgoreaBackend/v2/golang"
)

// cancelCtx mirrors the layout of the unexported context.cancelCtx so that the
// tests below can inject a synthetic context.DeadlineExceeded error mid-operation.
// Layout matches Go 1.25: `err` became atomic.Value (previously error) and
// timerCtx value-embeds cancelCtx (previously a pointer embed).
type cancelCtx struct {
	context.Context //nolint:containedctx // mirroring an unexported runtime struct

	_     sync.Mutex
	_     atomic.Value             // done
	_     map[interface{}]struct{} // children
	err   atomic.Value             // mirrors context.cancelCtx.err
	cause error
}

type timerCtx struct {
	cancelCtx // value-embedded as in stdlib

	_ unsafe.Pointer    // *time.Timer
	_ [3]unsafe.Pointer // time.Time (3 words)
}

type cancelCtxInterface struct {
	t unsafe.Pointer
	p *timerCtx
}

// resetAtomicValue atomically clears the typ slot of an atomic.Value so that a
// subsequent .Store() can succeed even with a different concrete type. Race-free
// because (a) the write is atomic.StorePointer and (b) atomic.Value.Load checks
// typ first and returns nil when typ==nil — readers never observe stale data.
func resetAtomicValue(av *atomic.Value) {
	// atomic.Value is `struct{ v any }`; its first word is the eface type pointer.
	atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(av)), nil) //nolint:gosec // see comment
}

// TestCancelCtxMirrorLayout verifies that the unsafe layout mirror of
// context.cancelCtx / context.timerCtx above still matches the stdlib on the
// running Go toolchain. If a future Go upgrade changes the offsets, this test
// fails fast with a clear message instead of silently corrupting unrelated
// memory inside Test_Deadline. Compatible with Go 1.21+ (when err became
// atomic.Value); should be re-checked on every Go major bump.
func TestCancelCtxMirrorLayout(t *testing.T) {
	parent, parentCancel := context.WithCancel(context.Background())
	defer parentCancel()

	rawCtx, cancel := context.WithDeadline(parent, time.Now().Add(time.Hour))
	cancel() // stores context.Canceled into cancelCtx.err and cancelCtx.cause

	mirror := (*cancelCtxInterface)(unsafe.Pointer(&rawCtx)).p //nolint:gosec // layout-mirror sanity check

	require.Equal(t, context.Canceled, rawCtx.Err(),
		"sanity: cancel() should make ctx.Err() == context.Canceled")
	require.Equal(t, context.Canceled, mirror.err.Load(),
		"cancelCtx.err offset is wrong: mirror does not see the value the stdlib stored. "+
			"Re-check the layout against the current Go runtime's context.cancelCtx.")
	require.Equal(t, context.Canceled, mirror.cause,
		"cancelCtx.cause offset is wrong: mirror does not see the value the stdlib stored. "+
			"Re-check the layout against the current Go runtime's context.cancelCtx.")
}

func Test_Deadline(t *testing.T) {
	for _, test := range []struct {
		name             string
		timeoutHasPassed bool
		funcToCall       func(*DataStore, context.CancelFunc) error
		setupMock        func(sqlmock.Sqlmock)
	}{
		{
			name:             "Before_Exec",
			timeoutHasPassed: true,
			funcToCall: func(s *DataStore, _ context.CancelFunc) error {
				return s.Exec("INSERT INTO t1 (id) VALUES (1)").Error()
			},
		},
		{
			name:             "Before_Query",
			timeoutHasPassed: true,
			funcToCall: func(s *DataStore, _ context.CancelFunc) error {
				var ids []int64
				return s.Table("t1").Select("id").Scan(&ids).Error()
			},
		},
		{
			name:             "Before_QueryRow",
			timeoutHasPassed: true,
			funcToCall: func(s *DataStore, _ context.CancelFunc) error {
				var cnt int64
				return s.Table("t1").Count(&cnt).Error()
			},
		},
		{
			name:             "Before_Transaction",
			timeoutHasPassed: true,
			funcToCall: func(s *DataStore, _ context.CancelFunc) error {
				return s.InTransaction(func(_ *DataStore) error {
					return nil
				})
			},
		},
		{
			name: "Before_Exec_InTransaction",
			funcToCall: func(s *DataStore, cancel context.CancelFunc) error {
				return s.InTransaction(func(s *DataStore) error {
					cancel()
					return s.Exec("INSERT INTO t1 (id) VALUES (1)").Error()
				})
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectRollback()
			},
		},
		{
			name: "Before_Prepare_InTransaction",
			funcToCall: func(s *DataStore, cancel context.CancelFunc) error {
				return s.InTransaction(func(s *DataStore) error {
					cancel()
					_, err := s.Prepare("INSERT INTO t1 (id) VALUES (1)")
					return err
				})
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectRollback()
			},
		},
		{
			name: "Before_Query_InTransaction",
			funcToCall: func(s *DataStore, cancel context.CancelFunc) error {
				return s.InTransaction(func(s *DataStore) error {
					cancel()
					var ids []int64
					return s.Table("t1").Select("id").Scan(&ids).Error()
				})
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectRollback()
			},
		},
		{
			name: "Before_QueryRow_InTransaction",
			funcToCall: func(s *DataStore, cancel context.CancelFunc) error {
				return s.InTransaction(func(s *DataStore) error {
					cancel()
					var cnt int64
					return s.Table("t1").Count(&cnt).Error()
				})
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectRollback()
			},
		},
		{
			name: "Before_Commit_InTransaction",
			funcToCall: func(s *DataStore, cancel context.CancelFunc) error {
				return s.InTransaction(func(_ *DataStore) error {
					cancel()
					return nil
				})
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectRollback()
			},
		},
		{
			name: "Before_Rollback_InTransaction",
			funcToCall: func(s *DataStore, cancel context.CancelFunc) error {
				return s.InTransaction(func(s *DataStore) error {
					cancel()
					return s.db.Rollback().Error
				})
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectRollback()
			},
		},
		{
			name: "Before_Stmt_ExecContext",
			funcToCall: func(s *DataStore, cancel context.CancelFunc) error {
				return s.InTransaction(func(s *DataStore) error {
					stmt, err := s.Prepare("INSERT INTO t1 (id) VALUES (1)")
					mustNotBeError(err)
					defer func() { _ = stmt.Close() }()
					cancel()
					_, err = stmt.ExecContext(s.ctx())
					return err
				})
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectPrepare("^" + regexp.QuoteMeta("INSERT INTO t1 (id) VALUES (1)") + "$")
				mock.ExpectRollback()
			},
		},
		{
			name: "Before_Stmt_QueryContext",
			funcToCall: func(s *DataStore, cancel context.CancelFunc) error {
				return s.InTransaction(func(s *DataStore) error {
					stmt, err := s.Prepare("SELECT * FROM t1")
					mustNotBeError(err)
					defer func() { _ = stmt.Close() }()
					cancel()
					rows, err := stmt.QueryContext(s.ctx())
					if rows != nil {
						_ = rows.Err() // ignore the error as err is expected to be non-nil
						defer func() { _ = rows.Close() }()
					}
					return err
				})
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectPrepare("^" + regexp.QuoteMeta("SELECT * FROM t1") + "$")
				mock.ExpectRollback()
			},
		},
		{
			name: "Before_Stmt_QueryRowContext",
			funcToCall: func(s *DataStore, cancel context.CancelFunc) error {
				return s.InTransaction(func(s *DataStore) error {
					stmt, err := s.Prepare("SELECT * FROM t1")
					mustNotBeError(err)
					defer func() { _ = stmt.Close() }()
					cancel()
					row := stmt.QueryRowContext(s.ctx())
					return row.Err()
				})
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectPrepare("^" + regexp.QuoteMeta("SELECT * FROM t1") + "$")
				mock.ExpectRollback()
			},
		},
		{
			name: "Before releasing a named lock inside a transaction",
			funcToCall: func(s *DataStore, cancel context.CancelFunc) error {
				return s.InTransaction(func(s *DataStore) error {
					return s.WithNamedLock("test_lock", time.Second, func(_ *DataStore) error {
						cancel()
						return nil
					})
				})
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.MatchExpectationsInOrder(false)
				mock.ExpectBegin()
				mock.ExpectQuery("^"+regexp.QuoteMeta("SELECT GET_LOCK(?, ?)")+"$").WithArgs("test_lock", 1).
					WillReturnRows(sqlmock.NewRows([]string{"GET_LOCK('test_lock', 1)"}).AddRow(1))
				mock.ExpectQuery("^" + regexp.QuoteMeta("SELECT RELEASE_LOCK(?)") + "$").WithArgs("test_lock").
					WillReturnRows(sqlmock.NewRows([]string{"RELEASE_LOCK('test_lock')"}).AddRow(1))
				mock.ExpectRollback()
			},
		},
		{
			name: "Before acquiring a named lock inside a transaction",
			funcToCall: func(s *DataStore, cancel context.CancelFunc) error {
				return s.InTransaction(func(s *DataStore) error {
					cancel()
					return s.WithNamedLock("test_lock", time.Second, func(_ *DataStore) error {
						return nil
					})
				})
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectRollback()
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			ctx, _, logHook := logging.NewContextWithNewMockLogger()
			db, mock := NewDBMock(ctx)
			defer func() { _ = db.Close() }()

			if test.setupMock != nil {
				test.setupMock(mock)
			}
			ctx, cancel := context.WithDeadline(ctx,
				time.Now().Add(golang.IfElse(test.timeoutHasPassed, 0, 1*time.Hour)))
			defer cancel()
			dataStore := NewDataStoreWithContext(ctx, db)

			err := test.funcToCall(dataStore, func() {
				cancel()
				ctxPtr := (*cancelCtxInterface)(unsafe.Pointer(&ctx)).p //nolint:gosec // imitate a deadline-exceeded
				// cancel() above already Stored context.Canceled into err.
				// atomic.Value.Store with a different concrete type panics, and a plain
				// non-atomic interface assignment races with database/sql's awaitDone
				// goroutine doing atomic.Value.Load via ctx.Err() (caught by `-race`).
				// Reset the type slot atomically (concurrent Loaders see typ=nil and
				// return nil — never a torn typ/data pair), then Store normally.
				resetAtomicValue(&ctxPtr.err)
				ctxPtr.err.Store(context.DeadlineExceeded)
				ctxPtr.cause = nil
			})

			require.EqualError(t, err, context.DeadlineExceeded.Error())

			assert.Eventually(t, func() bool {
				return mock.ExpectationsWereMet() == nil
			}, 3*time.Second, 10*time.Millisecond)
			require.NoError(t, mock.ExpectationsWereMet())

			logs := (&loggingtest.Hook{Hook: logHook}).GetAllStructuredLogs()
			assert.Contains(t, logs, "context deadline exceeded")
		})
	}
}
