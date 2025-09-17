package database

import (
	"context"
	"database/sql/driver"
	"unsafe"
)

type mysqlConnWrapper struct {
	conn driver.Conn
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
		return err
	}

	// really reset the session
	return conn.Reset(ctx)
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
