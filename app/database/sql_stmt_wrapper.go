package database

import (
	"context"
	"database/sql"

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
	defer getSQLExecutionPlanLoggingFunc(s.db, s.logConfig, s.sql, args...)()
	defer getSQLQueryLoggingFunc(func() *int64 {
		rowsAffected, _ := result.RowsAffected()
		return &rowsAffected
	}, &err, gorm.NowFunc(), s.sql, args...)(s.logConfig)

	return s.stmt.ExecContext(ctx, args...)
}

/*
// Exec panics if called.
func (s *SQLStmtWrapper) Exec(...interface{}) (result sql.Result, err error) {
	panic("SQLStmtWrapper.Exec should not be called")
}
*/

// QueryContext executes a prepared query statement with the given arguments
// and returns the query results as a [*sql.Rows].
func (s *SQLStmtWrapper) QueryContext(ctx context.Context, args ...interface{}) (_ *sql.Rows, err error) {
	defer getSQLExecutionPlanLoggingFunc(s.db, s.logConfig, s.sql, args...)()
	defer getSQLQueryLoggingFunc(nil, &err, gorm.NowFunc(), s.sql, args...)(s.logConfig)

	return s.stmt.QueryContext(ctx, args...)
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
	defer getSQLExecutionPlanLoggingFunc(s.db, s.logConfig, s.sql, args...)()
	startTime := gorm.NowFunc()
	defer func() {
		err := row.Err()
		getSQLQueryLoggingFunc(nil, &err, startTime, s.sql, args...)(s.logConfig)
	}()

	return s.stmt.QueryRowContext(ctx, args...)
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
