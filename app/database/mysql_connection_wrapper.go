package database

import (
	"context"
	"database/sql/driver"
	"errors"
	"fmt"
	"unsafe"

	log "github.com/France-ioi/AlgoreaBackend/v2/app/logging"
)

type mysqlConnWrapper struct {
	conn          driver.Conn
	sessionParams map[string]string
}

func (conn *mysqlConnWrapper) Begin() (driver.Tx, error) {
	panic("should not be called")
}

func (conn *mysqlConnWrapper) Prepare(query string) (driver.Stmt, error) {
	return conn.conn.Prepare(query)
}

func (conn *mysqlConnWrapper) Close() error {
	return conn.conn.Close()
}

var _ driver.Conn = &mysqlConnWrapper{}

func (conn *mysqlConnWrapper) BeginTx(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
	//nolint:forcetypeassert // panic if conn.conn does not implement driver.ConnBeginTx
	return conn.conn.(driver.ConnBeginTx).BeginTx(ctx, opts)
}

var _ driver.ConnBeginTx = &mysqlConnWrapper{}

func (conn *mysqlConnWrapper) PrepareContext(ctx context.Context, query string) (driver.Stmt, error) {
	//nolint:forcetypeassert // panic if conn.conn does not implement driver.ConnPrepareContext
	return conn.conn.(driver.ConnPrepareContext).PrepareContext(ctx, query)
}

var _ driver.ConnPrepareContext = &mysqlConnWrapper{}

func (conn *mysqlConnWrapper) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	//nolint:forcetypeassert // panic if conn.conn does not implement driver.ExecerContext
	return conn.conn.(driver.ExecerContext).ExecContext(ctx, query, args)
}

var _ driver.ExecerContext = &mysqlConnWrapper{}

func (conn *mysqlConnWrapper) Ping(ctx context.Context) error {
	//nolint:forcetypeassert // panic if conn.conn does not implement driver.Pinger
	return conn.conn.(driver.Pinger).Ping(ctx)
}

var _ driver.Pinger = &mysqlConnWrapper{}

func (conn *mysqlConnWrapper) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	//nolint:forcetypeassert // panic if conn.conn does not implement driver.QueryerContext
	return conn.conn.(driver.QueryerContext).QueryContext(ctx, query, args)
}

var _ driver.QueryerContext = &mysqlConnWrapper{}

func (conn *mysqlConnWrapper) CheckNamedValue(nv *driver.NamedValue) error {
	//nolint:forcetypeassert // panic if conn.conn does not implement driver.NamedValueChecker
	return conn.conn.(driver.NamedValueChecker).CheckNamedValue(nv)
}

var _ driver.NamedValueChecker = &mysqlConnWrapper{}

type ifaceAccessor struct {
	_    unsafe.Pointer
	conn *mysqlConn
}

func (conn *mysqlConnWrapper) ResetSession(ctx context.Context) error {
	//nolint:forcetypeassert // panic if conn.conn does not implement driver.SessionResetter
	err := conn.conn.(driver.SessionResetter).ResetSession(ctx)
	if err != nil {
		return discardConnectionAfterResetFailure(ctx, "driver reset session", err)
	}

	// really reset the session
	if err := conn.Reset(ctx); err != nil {
		return discardConnectionAfterResetFailure(ctx, "resetting session", err)
	}

	// COM_RESET_CONNECTION restores session variables to their global values, so anything the
	// application relies on must be re-applied here. DSN params are not an option: the driver
	// sends them only at connect time (handleParams, called from connector.Connect) and never
	// after a reset.
	if err := conn.applySessionParams(ctx); err != nil {
		return discardConnectionAfterResetFailure(ctx, "re-applying session params", err)
	}
	return nil
}

// discardConnectionAfterResetFailure wraps err as driver.ErrBadConn so database/sql discards the
// connection. Any other error is ignored by the pool and the caller would silently get global session values.
func discardConnectionAfterResetFailure(ctx context.Context, what string, err error) error {
	entry := log.EntryFromContext(ctx).WithField("type", "db")
	msg := fmt.Sprintf("discarding DB connection after failure %s: %s", what, err)
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		entry.Info(msg)
	} else {
		entry.Warn(msg)
	}
	if errors.Is(err, driver.ErrBadConn) {
		return err
	}
	return fmt.Errorf("%w (%s: %s)", driver.ErrBadConn, what, err.Error())
}

func (conn *mysqlConnWrapper) Reset(ctx context.Context) error {
	//nolint:gosec // G103: Here we convert driver.Conn to go-sql-driver/mysql.Conn which we make accessible via mysqlConn struct.
	return ((*ifaceAccessor)(unsafe.Pointer(&conn.conn)).conn).Reset(ctx)
}

var _ driver.SessionResetter = &mysqlConnWrapper{}

func (conn *mysqlConnWrapper) IsValid() bool {
	//nolint:forcetypeassert // panic if conn.conn does not implement driver.Validator
	return conn.conn.(driver.Validator).IsValid()
}

var _ driver.Validator = &mysqlConnWrapper{}
