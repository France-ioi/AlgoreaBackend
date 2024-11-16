package database

import (
	"context"
	"database/sql"
	_ "unsafe" // required to use //go:linkname

	"github.com/jinzhu/gorm"
)

type sqlConnWrapper struct {
	conn      *sql.Conn
	logConfig *LogConfig
}

// Note: all the public methods of sql.Conn are implemented in sqlConnWrapper,
// but only the ones we use are uncommented.

/*
// PingContext verifies the connection to the database is still alive.
func (c *sqlConnWrapper) PingContext(ctx context.Context) error {
	return c.conn.PingContext(ctx)
}
*/

/*
// ExecContext executes a query without returning any rows.
// The args are for any placeholder parameters in the query.
func (c *sqlConnWrapper) ExecContext(ctx context.Context, query string, args ...any) (result sql.Result, err error) {
	defer getSQLExecutionPlanLoggingFunc(c, c.LogConfig, query, args...)()
	defer getSQLQueryLoggingFunc(func() *int64 {
		rowsAffected, _ := result.RowsAffected()
		return &rowsAffected
	}, &err, gorm.NowFunc(), query, args...)(c.LogConfig)

	return c.conn.ExecContext(ctx, query, args...)
}
*/

/*
// QueryContext executes a query that returns rows, typically a SELECT.
// The args are for any placeholder parameters in the query.
func (c *sqlConnWrapper) QueryContext(ctx context.Context, query string, args ...any) (rows *sql.Rows, err error) {
	defer getSQLExecutionPlanLoggingFunc(c, c.LogConfig, query, args...)()
	defer getSQLQueryLoggingFunc(nil, &err, gorm.NowFunc(), query, args...)(c.LogConfig)

	return c.conn.QueryContext(ctx, query, args...)
}
*/

// QueryRowContext executes a query that is expected to return at most one row.
// QueryRowContext always returns a non-nil value. Errors are deferred until
// Row's Scan method is called.
// If the query selects no rows, the *Row's Scan will return ErrNoRows.
// Otherwise, the *Row's Scan scans the first selected row and discards
// the rest.
func (c *sqlConnWrapper) QueryRowContext(ctx context.Context, query string, args ...any) (row *sql.Row) {
	defer getSQLExecutionPlanLoggingFunc(c, c.logConfig, query, args...)()
	startTime := gorm.NowFunc()
	defer func() {
		err := row.Err()
		getSQLQueryLoggingFunc(nil, &err, startTime, query, args...)(c.logConfig)
	}()

	return c.conn.QueryRowContext(ctx, query, args...)
}

/*
// PrepareContext creates a prepared statement for later queries or executions.
// Multiple queries or executions may be run concurrently from the
// returned statement.
// The caller must call the statement's Close method
// when the statement is no longer needed.
//
// The provided context is used for the preparation of the statement, not for the
// execution of the statement.
func (c *sqlConnWrapper) PrepareContext(ctx context.Context, query string) (*SQLStmtWrapper, error) {
	stmt, err := c.conn.PrepareContext(ctx, query)
	if err != nil {
		c.LogConfig.Logger.Print("error", fileWithLineNum(), err)
		return nil, err
	}
	return &SQLStmtWrapper{db: c, sql: query, stmt: stmt, LogConfig: c.LogConfig}, nil
}
*/

/*
// Raw executes f exposing the underlying driver connection for the
// duration of f. The driverConn must not be used outside of f.
//
// Once f returns and err is not driver.ErrBadConn, the Conn will continue to be usable
// until Conn.Close is called.
func (c *sqlConnWrapper) Raw(f func(driverConn any) error) (err error) {
	return c.conn.Raw(f)
}
*/

/*
// BeginTx starts a transaction.
//
// The provided context is used until the transaction is committed or rolled back.
// If the context is canceled, the sql package will roll back
// the transaction. Tx.Commit will return an error if the context provided to
// BeginTx is canceled.
//
// The provided TxOptions is optional and may be nil if defaults should be used.
// If a non-default isolation level is used that the driver doesn't support,
// an error will be returned.
func (c *sqlConnWrapper) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sqlTxWrapper, error) {
	startTime := gorm.NowFunc()
	tx, err := c.conn.BeginTx(ctx, opts)
	if c.LogConfig.LogSQLQueries {
		logSQLQuery(c.LogConfig.Logger, gorm.NowFunc().Sub(startTime), beginTransactionLogMessage, nil, nil)
	}
	if err != nil {
		c.LogConfig.Logger.Print("error", fileWithLineNum(), err)
		return nil, err
	}
	return &sqlTxWrapper{sqlTx: tx, ctx: ctx, LogConfig: c.LogConfig}, nil
}
*/

/*
// Close returns the connection to the connection pool.
// All operations after a Close will return with [sql.ErrConnDone].
// Close is safe to call concurrently with other operations and will
// block until all other operations finish. It may be useful to first
// cancel any used context and then call close directly after.
func (c *sqlConnWrapper) Close() error {
	return c.conn.Close()
}
*/

func (c *sqlConnWrapper) queryRowWithoutLogging(query string, args ...any) *sql.Row {
	return c.conn.QueryRowContext(context.Background(), query, args...)
}

var _ queryRowWithoutLogging = &sqlConnWrapper{}

func (c *sqlConnWrapper) close(err error) error {
	return sqlConnCloseWithError(c.conn, err)
}

//go:linkname sqlConnCloseWithError database/sql.(*Conn).close
func sqlConnCloseWithError(_ *sql.Conn, _ error) error
