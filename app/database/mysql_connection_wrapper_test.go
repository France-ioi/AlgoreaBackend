package database

import (
	"database/sql/driver"
	"errors"
	"testing"
	"unsafe"

	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"
)

func Test_mysqlConnWrapper_Begin_Panics(t *testing.T) {
	conn := &mysqlConnWrapper{}
	assert.Panics(t, func() { _, _ = conn.Begin() }, "Begin() should panic")
}

type stmtMock struct{}

func (stmt *stmtMock) Close() error                               { return nil }
func (stmt *stmtMock) NumInput() int                              { return 0 }
func (stmt *stmtMock) Exec([]driver.Value) (driver.Result, error) { return nil, nil }
func (stmt *stmtMock) Query([]driver.Value) (driver.Rows, error)  { return nil, nil }

var _ driver.Stmt = &stmtMock{}

type connMock struct{}

func (conn *connMock) Prepare(string) (driver.Stmt, error) {
	return &stmtMock{}, errors.New("error")
}
func (conn *connMock) Begin() (driver.Tx, error) { return nil, nil }
func (conn *connMock) Close() error              { return nil }

var _ driver.Conn = &connMock{}

func Test_mysqlConnWrapper_Prepare(t *testing.T) {
	conn := &mysqlConnWrapper{conn: &connMock{}}
	expectedQuery := "SELECT * FROM table"
	expectedStmt := &stmtMock{}
	expectedErr := errors.New("error")
	monkey.Patch(mysqlConnPrepare, func(c unsafe.Pointer, query string) (driver.Stmt, error) {
		assert.Equal(t, conn, (*mysqlConnWrapper)(c))
		assert.Equal(t, expectedQuery, query)
		return expectedStmt, expectedErr
	})
	defer monkey.UnpatchAll()
	stmt, err := conn.Prepare(expectedQuery)
	assert.Equal(t, expectedStmt, stmt)
	assert.Equal(t, expectedErr, err)
}

//go:linkname mysqlConnPrepare github.com/go-sql-driver/mysql.(*mysqlConn).Prepare
func mysqlConnPrepare(unsafe.Pointer, string) (driver.Stmt, error)
