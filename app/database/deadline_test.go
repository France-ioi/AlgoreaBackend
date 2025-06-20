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

	"github.com/France-ioi/AlgoreaBackend/v2/app/logging"
	"github.com/France-ioi/AlgoreaBackend/v2/app/loggingtest"
	"github.com/France-ioi/AlgoreaBackend/v2/golang"
)

type cancelCtx struct {
	context.Context //nolint:containedctx // it is not us who store the context in the structure
	_               sync.Mutex
	_               atomic.Value
	_               map[interface{}]struct{}
	err             error
	cause           error
}

type timerCtx struct {
	*cancelCtx
}

type cancelCtxInterface struct {
	t unsafe.Pointer
	p *timerCtx
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
				return s.InTransaction(func(s *DataStore) error {
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
						_ = rows.Close()
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
					return s.WithNamedLock("test_lock", time.Second, func(s *DataStore) error {
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
					return s.WithNamedLock("test_lock", time.Second, func(s *DataStore) error {
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
		test := test
		t.Run(test.name, func(t *testing.T) {
			db, mock := NewDBMock()
			defer func() { _ = db.Close() }()
			logHook, logRestoreFunc := logging.MockSharedLoggerHook()
			defer logRestoreFunc()

			if test.setupMock != nil {
				test.setupMock(mock)
			}
			ctx, cancel := context.WithDeadline(context.Background(),
				time.Now().Add(golang.IfElse(test.timeoutHasPassed, 0, 1*time.Hour)))
			defer cancel()
			dataStore := NewDataStoreWithContext(ctx, db)

			err := test.funcToCall(dataStore, func() {
				cancel()
				//nolint:gosec // access the context directly to set the error
				(*cancelCtxInterface)(unsafe.Pointer(&ctx)).p.err = context.DeadlineExceeded
				//nolint:gosec // access the context directly to set the cause to nil
				(*cancelCtxInterface)(unsafe.Pointer(&ctx)).p.cause = nil
			})

			assert.EqualError(t, err, context.DeadlineExceeded.Error())

			assert.Eventually(t, func() bool {
				return mock.ExpectationsWereMet() == nil
			}, 3*time.Second, 10*time.Millisecond)
			assert.NoError(t, mock.ExpectationsWereMet())

			logs := (&loggingtest.Hook{Hook: logHook}).GetAllStructuredLogs()
			assert.Contains(t, logs, "context deadline exceeded")
		})
	}
}
