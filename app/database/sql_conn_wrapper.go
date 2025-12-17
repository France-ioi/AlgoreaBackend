package database

import (
	"context"
	"database/sql"
	"unsafe"

	"github.com/jinzhu/gorm"
)

type sqlConnWrapper struct {
	conn *sql.Conn
	//nolint:containedctx // We store the context here because Gorm v1 does not support contexts
	//                    // as arguments for Exec, Query, and QueryRow methods.
	ctx       context.Context
	logConfig *LogConfig
}

// Exec executes a query without returning any rows.
// The args are for any placeholder parameters in the query.
//
// Exec uses the context of [sqlConnWrapper] internally.
func (c *sqlConnWrapper) Exec(query string, args ...interface{}) (result sql.Result, err error) {
	return execCallingExecContext(c.ctx, c.logConfig, c, c.conn, query, args...)
}

// Prepare is not implemented intentionally and panics if called.
//
// Note: Gorm does not use this method, but it is required to implement the [gorm.SQLCommon] interface.
func (c *sqlConnWrapper) Prepare(_ string) (*sql.Stmt, error) {
	panic("sqlConnWrapper.Prepare is not implemented, should not be called")
}

// Query executes a query that returns rows, typically a SELECT.
// The args are for any placeholder parameters in the query.
//
// Query uses the context of [sqlConnWrapper] internally.
func (c *sqlConnWrapper) Query(query string, args ...any) (rows *sql.Rows, err error) {
	err = retryOnRetriableError(c.ctx, func() error {
		defer getSQLExecutionPlanLoggingFunc(c.ctx, c, c.logConfig, query, args...)()
		defer getSQLQueryLoggingFunc(c.ctx, nil, &err, gorm.NowFunc(), query, args...)(c.logConfig)

		//nolint:rowserrcheck,sqlclosecheck // The caller is responsible for closing the returned *sql.Rows and checking rows.Err().
		rows, err = c.conn.QueryContext(c.ctx, query, args...)
		return err
	})
	if err != nil {
		return nil, err
	}
	return rows, nil
}

// QueryRow executes a query that is expected to return at most one row.
// QueryRow always returns a non-nil value. Errors are deferred until
// Row's Scan method is called.
// If the query selects no rows, the *Row's Scan will return ErrNoRows.
// Otherwise, the *Row's Scan scans the first selected row and discards
// the rest.
//
// QueryRow uses the context of [sqlConnWrapper] internally.
func (c *sqlConnWrapper) QueryRow(query string, args ...any) (row *sql.Row) {
	return c.QueryRowContext(c.ctx, query, args...)
}

var _ gorm.SQLCommon = &sqlConnWrapper{}

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
	return execCallingExecContext(ctx, c.logConfig, c, c.conn, query, args...)
}
*/

/*
// QueryContext executes a query that returns rows, typically a SELECT.
// The args are for any placeholder parameters in the query.
func (c *sqlConnWrapper) QueryContext(ctx context.Context, query string, args ...any) (rows *sql.Rows, err error) {
	err = retryOnRetriableError(ctx, func() error {
		defer getSQLExecutionPlanLoggingFunc(ctx, c, c.logConfig, query, args...)()
		defer getSQLQueryLoggingFunc(ctx, nil, &err, gorm.NowFunc(), query, args...)(c.logConfig)

		rows, err = c.conn.QueryContext(ctx, query, args...)
		return err
	})
	if err != nil {
		return nil, err
	}
	return rows, nil
}
*/

// QueryRowContext executes a query that is expected to return at most one row.
// QueryRowContext always returns a non-nil value. Errors are deferred until
// Row's Scan method is called.
// If the query selects no rows, the *Row's Scan will return ErrNoRows.
// Otherwise, the *Row's Scan scans the first selected row and discards
// the rest.
func (c *sqlConnWrapper) QueryRowContext(ctx context.Context, query string, args ...any) (row *sql.Row) {
	err := retryOnRetriableError(ctx, func() error {
		defer getSQLExecutionPlanLoggingFunc(ctx, c, c.logConfig, query, args...)()
		startTime := gorm.NowFunc()
		defer func() {
			err := row.Err()
			getSQLQueryLoggingFunc(ctx, nil, &err, startTime, query, args...)(c.logConfig)
		}()

		row = c.conn.QueryRowContext(ctx, query, args...)
		return row.Err()
	})
	(*struct{ err error })(unsafe.Pointer(row)).err = err //nolint:gosec // here we replace row.err
	return row
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
		logDBError(ctx, c.logConfig, err)
		return nil, err
	}
	return &SQLStmtWrapper{db: c, sql: query, stmt: stmt, logConfig: c.logConfig}, nil
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
	if c.logConfig.LogSQLQueries {
		logSQLQuery(ctx, gorm.NowFunc().Sub(startTime), beginTransactionLogMessage, nil, nil)
	}
	if err != nil {
		logDBError(ctx, err)
		return nil, err
	}
	return &sqlTxWrapper{sqlTx: tx, ctx: ctx, logConfig: c.logConfig}, nil
}

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

var _ txBeginner = &sqlConnWrapper{}

func (c *sqlConnWrapper) withContext(ctx context.Context) gorm.SQLCommon {
	return &sqlConnWrapper{conn: c.conn, ctx: ctx, logConfig: c.logConfig}
}

var _ withContexter = &sqlConnWrapper{}

func (c *sqlConnWrapper) getContext() context.Context {
	return c.ctx
}

var _ contextGetter = &sqlConnWrapper{}

func (c *sqlConnWrapper) getLogConfig() *LogConfig {
	return c.logConfig
}

var _ logConfigGetter = &sqlConnWrapper{}

func (c *sqlConnWrapper) close(err error) error {
	return sqlConnCloseWithError(c.conn, err)
}

//go:linkname sqlConnCloseWithError database/sql.(*Conn).close
func sqlConnCloseWithError(_ *sql.Conn, _ error) error
