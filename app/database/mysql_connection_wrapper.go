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
	return conn.conn.(driver.ConnBeginTx).BeginTx(ctx, opts)
}

var _ driver.ConnBeginTx = &mysqlConnWrapper{}

func (conn *mysqlConnWrapper) PrepareContext(ctx context.Context, query string) (driver.Stmt, error) {
	return conn.conn.(driver.ConnPrepareContext).PrepareContext(ctx, query)
}

var _ driver.ConnPrepareContext = &mysqlConnWrapper{}

func (conn *mysqlConnWrapper) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	return conn.conn.(driver.ExecerContext).ExecContext(ctx, query, args)
}

var _ driver.ExecerContext = &mysqlConnWrapper{}

func (conn *mysqlConnWrapper) Ping(ctx context.Context) error {
	return conn.conn.(driver.Pinger).Ping(ctx)
}

var _ driver.Pinger = &mysqlConnWrapper{}

func (conn *mysqlConnWrapper) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	return conn.conn.(driver.QueryerContext).QueryContext(ctx, query, args)
}

var _ driver.QueryerContext = &mysqlConnWrapper{}

func (conn *mysqlConnWrapper) CheckNamedValue(nv *driver.NamedValue) error {
	return conn.conn.(driver.NamedValueChecker).CheckNamedValue(nv)
}

var _ driver.NamedValueChecker = &mysqlConnWrapper{}

type ifaceAccessor struct {
	_    unsafe.Pointer
	conn *mysqlConn
}

func (conn *mysqlConnWrapper) ResetSession(ctx context.Context) error {
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
	return conn.conn.(driver.Validator).IsValid()
}

var _ driver.Validator = &mysqlConnWrapper{}
