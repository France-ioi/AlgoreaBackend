package database

import (
	"context"
	"database/sql"
	"unsafe"

	"github.com/jinzhu/gorm"
)

// SQLStmtWrapper is a wrapper around a [sql.Stmt] that logs queries and execution plans.
// An SQLStmtWrapper is safe for concurrent use by multiple goroutines.
//
// As the only possible way to get an object of this type is to call [DB.Prepare],
//
// If an SQLStmtWrapper is prepared on a Tx or Conn, it will be bound to a single
// underlying connection forever. If the Tx or Conn closes, the Stmt will
// become unusable and all operations will return an error.
// If a Stmt is prepared on a DB, it will remain usable for the lifetime of the
// DB. When the Stmt needs to execute on a new underlying connection, it will
// prepare itself on the new connection automatically.
type SQLStmtWrapper struct {
	db        queryRowWithoutLogging
	sql       string
	stmt      *sql.Stmt
	logConfig *LogConfig
}

// Note: We don't implement methods that don't accept a context because
// they would use a default context (like context.Background() or a maybe a context stored in the struct),
// which is bad practice. Although sqlDBWrapper and sqlTxWrapper store the context for compatibility reasons,
// we don't want to encourage this practice.

// ExecContext executes a prepared statement with the given arguments and
// returns a [sql.Result] summarizing the effect of the statement.
func (s *SQLStmtWrapper) ExecContext(ctx context.Context, args ...interface{}) (result sql.Result, err error) {
	defer getSQLExecutionPlanLoggingFunc(ctx, s.db, s.logConfig, s.sql, args...)()
	defer getSQLQueryLoggingFunc(ctx, func() *int64 {
		rowsAffected, _ := result.RowsAffected()
		return &rowsAffected
	}, &err, gorm.NowFunc(), s.sql, args...)(s.logConfig)

	result, err = s.stmt.ExecContext(ctx, args...)
	err = s.handleError(ctx, err)
	return result, err
}

/*
// Exec panics if called.
func (s *SQLStmtWrapper) Exec(...interface{}) (result sql.Result, err error) {
	panic("SQLStmtWrapper.Exec should not be called")
}
*/

// QueryContext executes a prepared query statement with the given arguments
// and returns the query results as a [*sql.Rows].
func (s *SQLStmtWrapper) QueryContext(ctx context.Context, args ...interface{}) (rows *sql.Rows, err error) {
	defer getSQLExecutionPlanLoggingFunc(ctx, s.db, s.logConfig, s.sql, args...)()
	defer getSQLQueryLoggingFunc(ctx, nil, &err, gorm.NowFunc(), s.sql, args...)(s.logConfig)

	rows, err = s.stmt.QueryContext(ctx, args...) //nolint:sqlclosecheck // The caller is responsible for closing the returned *sql.Rows.
	err = s.handleError(ctx, err)
	return rows, err
}

/*
// Query panics if called.
func (s *SQLStmtWrapper) Query(...interface{}) (_ *sql.Rows, err error) {
	panic("SQLStmtWrapper.Query should not be called")
}
*/

// QueryRowContext executes a prepared query statement with the given arguments.
// If an error occurs during the execution of the statement, that error will
// be returned by a call to Scan on the returned [*sql.Row], which is always non-nil.
// If the query selects no rows, the [*sql.Row.Scan] will return [sql.ErrNoRows].
// Otherwise, the [*sql.Row.Scan] scans the first selected row and discards
// the rest.
//
// Example usage:
//
//	var name string
//	err := nameByUseridStmt.QueryRow(ctx, id).Scan(&name)
func (s *SQLStmtWrapper) QueryRowContext(ctx context.Context, args ...interface{}) (row *sql.Row) {
	defer getSQLExecutionPlanLoggingFunc(ctx, s.db, s.logConfig, s.sql, args...)()
	startTime := gorm.NowFunc()
	defer func() {
		err := s.handleError(ctx, row.Err())
		if err != nil {
			row = (*sql.Row)(unsafe.Pointer(&sqlRowAccessor{err: err})) //nolint:gosec // G103: patch the error
		}
		getSQLQueryLoggingFunc(ctx, nil, &err, startTime, s.sql, args...)(s.logConfig)
	}()

	return s.stmt.QueryRowContext(ctx, args...)
}

type sqlRowAccessor struct {
	err  error
	rows *sql.Rows
}

/*
// QueryRow panics if called.
func (s *SQLStmtWrapper) QueryRow(...interface{}) (row *sql.Row) {
	panic("SQLStmtWrapper.QueryRow should not be called")
}
*/

// Close closes the statement.
func (s *SQLStmtWrapper) Close() error {
	return s.stmt.Close()
}

func (s *SQLStmtWrapper) handleError(ctx context.Context, err error) error {
	if err != nil && ctx.Err() != nil { // ignore the returned error if the context has been canceled before
		err = ctx.Err() // return the context error instead as it is the root cause
	}
	return err
}
