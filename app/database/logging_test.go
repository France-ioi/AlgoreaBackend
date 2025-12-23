package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/sirupsen/logrus" //nolint:depguard // allow to use logrus here as an exception to reference logrus.Entry
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/France-ioi/AlgoreaBackend/v2/app/logging"
	"github.com/France-ioi/AlgoreaBackend/v2/app/loggingtest"
	"github.com/France-ioi/AlgoreaBackend/v2/golang"
	"github.com/France-ioi/AlgoreaBackend/v2/testhelpers/testoutput"
)

func expectAnalyzeForQuery(mock sqlmock.Sqlmock, query string, err error, withAnalyze bool) {
	if !withAnalyze {
		return
	}
	expectation := mock.ExpectQuery("^" + regexp.QuoteMeta("EXPLAIN ANALYZE "+query) + "$")
	if err != nil {
		expectation.WillReturnError(err)
	} else {
		expectation.WillReturnRows(sqlmock.NewRows([]string{"QUERY PLAN"}).AddRow("plan"))
	}
}

const (
	updateQueryForTesting         = "UPDATE users SET name = 'John' WHERE id = 1"
	selectQueryForTesting         = "SELECT id FROM users"
	selectQueryWithWithForTesting = "WITH users AS (SELECT 2 AS id) SELECT id FROM users"
)

type funcToRunInSQLQueryLoggingTests func(t *testing.T, db *DB, mock sqlmock.Sqlmock, withSQLAnalyze bool) (
	expectedQuery string, expectedAffectedRows *int64, expectedError error,
)

type sqlQueryLoggingTest struct {
	name                         string
	funcToRun                    funcToRunInSQLQueryLoggingTests
	skipAnalyzeSQLQueriesTesting bool
}

func sqlQueryLoggingTests() []sqlQueryLoggingTest {
	return []sqlQueryLoggingTest{
		{
			name: "sqlDBWrapper.Exec with error",
			funcToRun: generateTestFuncToCheckSQLDBWrapperQueryOrExecWithError(
				func(mock sqlmock.Sqlmock, expectedSQL string, expectedError error) {
					mock.ExpectExec(expectedSQL).WillReturnError(expectedError)
				},
				func(db *sqlDBWrapper, expectedQuery string) (interface{}, error) {
					return db.Exec(expectedQuery)
				}),
		},
		{
			name: "sqlDBWrapper.Exec with success",
			funcToRun: func(t *testing.T, db *DB, mock sqlmock.Sqlmock, withSQLAnalyze bool) (
				expectedQuery string, expectedAffectedRows *int64, expectedError error,
			) {
				t.Helper()

				expectedQuery = updateQueryForTesting
				expectedAffectedRows = golang.Ptr(int64(1))
				expectAnalyzeForQuery(mock, expectedQuery, expectedError, withSQLAnalyze)
				mock.ExpectExec("^" + regexp.QuoteMeta(expectedQuery) + "$").
					WillReturnResult(sqlmock.NewResult(-1, *expectedAffectedRows))
				dbWrapper, ok := db.db.CommonDB().(*sqlDBWrapper)
				require.True(t, ok)
				result, err := dbWrapper.Exec(expectedQuery)
				require.NoError(t, err)
				assert.NotNil(t, result)
				return expectedQuery, expectedAffectedRows, expectedError
			},
		},
		{
			name: "sqlDBWrapper.Query with error",
			funcToRun: generateTestFuncToCheckSQLDBWrapperQueryOrExecWithError(
				func(mock sqlmock.Sqlmock, expectedSQL string, expectedError error) {
					mock.ExpectQuery(expectedSQL).WillReturnError(expectedError)
				},
				func(db *sqlDBWrapper, expectedQuery string) (interface{}, error) {
					return db.Query(expectedQuery) //nolint:rowserrcheck // the caller is responsible for checking
				}),
		},
		{
			name: "sqlDBWrapper.Query with success",
			funcToRun: func(t *testing.T, db *DB, mock sqlmock.Sqlmock, withSQLAnalyze bool) (
				expectedQuery string, expectedAffectedRows *int64, expectedError error,
			) {
				t.Helper()

				expectedQuery = selectQueryForTesting
				expectAnalyzeForQuery(mock, expectedQuery, expectedError, withSQLAnalyze)
				mock.ExpectQuery("^" + regexp.QuoteMeta(expectedQuery) + "$").
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1).AddRow(2))
				dbWrapper, ok := db.db.CommonDB().(*sqlDBWrapper)
				require.True(t, ok)
				rows, err := dbWrapper.Query(expectedQuery)
				require.NoError(t, err)
				require.NotNil(t, rows)
				defer func() { _ = rows.Close() }()
				require.NoError(t, rows.Err())
				return expectedQuery, expectedAffectedRows, expectedError
			},
		},
		{
			name: "sqlDBWrapper.QueryRow with error",
			funcToRun: func(t *testing.T, db *DB, mock sqlmock.Sqlmock, withSQLAnalyze bool) (
				expectedQuery string, expectedAffectedRows *int64, expectedError error,
			) {
				t.Helper()

				expectedError = errors.New("some error")
				expectedQuery = selectQueryForTesting
				expectAnalyzeForQuery(mock, expectedQuery, expectedError, withSQLAnalyze)
				mock.ExpectQuery("^" + regexp.QuoteMeta(expectedQuery) + "$").WillReturnError(expectedError)
				dbWrapper, ok := db.db.CommonDB().(*sqlDBWrapper)
				require.True(t, ok)
				row := dbWrapper.QueryRow(expectedQuery)
				assert.Equal(t, expectedError, row.Err())
				return expectedQuery, expectedAffectedRows, expectedError
			},
		},
		{
			name: "sqlDBWrapper.QueryRow with success",
			funcToRun: func(t *testing.T, db *DB, mock sqlmock.Sqlmock, withSQLAnalyze bool) (
				expectedQuery string, expectedAffectedRows *int64, expectedError error,
			) {
				t.Helper()

				expectedQuery = selectQueryWithWithForTesting
				expectAnalyzeForQuery(mock, expectedQuery, expectedError, withSQLAnalyze)
				mock.ExpectQuery("^" + regexp.QuoteMeta(expectedQuery) + "$").
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1).AddRow(2))
				dbWrapper, ok := db.db.CommonDB().(*sqlDBWrapper)
				require.True(t, ok)
				row := dbWrapper.QueryRow(expectedQuery)
				require.NoError(t, row.Err())
				_ = row.Scan()
				return expectedQuery, expectedAffectedRows, expectedError
			},
		},
		{
			name: "sqlDBWrapper.BeginTx with error",
			funcToRun: func(t *testing.T, db *DB, mock sqlmock.Sqlmock, _ bool) (
				expectedQuery string, expectedAffectedRows *int64, expectedError error,
			) {
				t.Helper()

				expectedError = errors.New("some error")
				expectedQuery = beginTransactionLogMessage
				mock.ExpectBegin().WillReturnError(expectedError)
				dbWrapper, ok := db.db.CommonDB().(*sqlDBWrapper)
				require.True(t, ok)
				tx, err := dbWrapper.BeginTx(db.ctx(), &sql.TxOptions{})
				assert.Equal(t, expectedError, err)
				assert.Nil(t, tx)
				return expectedQuery, expectedAffectedRows, expectedError
			},
			skipAnalyzeSQLQueriesTesting: true,
		},
		{
			name: "sqlDBWrapper.BeginTx with success",
			funcToRun: func(t *testing.T, db *DB, mock sqlmock.Sqlmock, _ bool) (
				expectedQuery string, expectedAffectedRows *int64, expectedError error,
			) {
				t.Helper()

				expectedQuery = beginTransactionLogMessage
				mock.ExpectBegin()
				dbWrapper, ok := db.db.CommonDB().(*sqlDBWrapper)
				require.True(t, ok)
				tx, err := dbWrapper.BeginTx(db.ctx(), &sql.TxOptions{})
				require.NoError(t, err)
				assert.NotNil(t, tx)
				return expectedQuery, expectedAffectedRows, expectedError
			},
			skipAnalyzeSQLQueriesTesting: true,
		},
		{
			name: "sqlConnWrapper.Exec with error",
			funcToRun: generateTestFuncToCheckSQLConnWrapperQueryOrExecWithError(
				func(mock sqlmock.Sqlmock, expectedSQL string, expectedError error) {
					mock.ExpectExec(expectedSQL).WillReturnError(expectedError)
				},
				func(conn *sqlConnWrapper, expectedQuery string) (interface{}, error) {
					return conn.Exec(expectedQuery)
				}),
		},
		{
			name: "sqlConnWrapper.Exec with success",
			funcToRun: func(t *testing.T, db *DB, mock sqlmock.Sqlmock, withSQLAnalyze bool) (
				expectedQuery string, expectedAffectedRows *int64, expectedError error,
			) {
				t.Helper()

				expectedQuery = updateQueryForTesting
				expectedAffectedRows = golang.Ptr(int64(1))
				expectAnalyzeForQuery(mock, expectedQuery, expectedError, withSQLAnalyze)
				mock.ExpectExec("^" + regexp.QuoteMeta(expectedQuery) + "$").
					WillReturnResult(sqlmock.NewResult(-1, *expectedAffectedRows))
				var result sql.Result
				var err error
				err = db.WithFixedConnection(func(db *DB) error {
					connWrapper, ok := db.db.CommonDB().(*sqlConnWrapper)
					require.True(t, ok)
					result, err = connWrapper.Exec(expectedQuery)
					require.NoError(t, err)
					return err
				})
				require.NoError(t, err)
				assert.NotNil(t, result)
				return expectedQuery, expectedAffectedRows, expectedError
			},
		},
		{
			name: "sqlConnWrapper.Query with error",
			funcToRun: generateTestFuncToCheckSQLConnWrapperQueryOrExecWithError(
				func(mock sqlmock.Sqlmock, expectedSQL string, expectedError error) {
					mock.ExpectQuery(expectedSQL).WillReturnError(expectedError)
				},
				func(conn *sqlConnWrapper, expectedQuery string) (interface{}, error) {
					return conn.Query(expectedQuery) //nolint:rowserrcheck // the caller is responsible for checking
				}),
		},
		{
			name: "sqlConnWrapper.Query with success",
			funcToRun: func(t *testing.T, db *DB, mock sqlmock.Sqlmock, withSQLAnalyze bool) (
				expectedQuery string, expectedAffectedRows *int64, expectedError error,
			) {
				t.Helper()

				expectedQuery = selectQueryWithWithForTesting
				expectAnalyzeForQuery(mock, expectedQuery, expectedError, withSQLAnalyze)
				mock.ExpectQuery("^" + regexp.QuoteMeta(expectedQuery) + "$").
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1).AddRow(2))
				var rows *sql.Rows
				var err error
				require.NoError(t, db.WithFixedConnection(func(db *DB) error {
					connWrapper, ok := db.db.CommonDB().(*sqlConnWrapper)
					require.True(t, ok)
					rows, err = connWrapper.Query(expectedQuery)
					require.NoError(t, err)
					require.NotNil(t, rows)
					defer func() { _ = rows.Close() }()
					require.NoError(t, rows.Err())
					return nil
				}))
				return expectedQuery, expectedAffectedRows, expectedError
			},
		},
		{
			name: "sqlConnWrapper.QueryRow with error",
			funcToRun: func(t *testing.T, db *DB, mock sqlmock.Sqlmock, withSQLAnalyze bool) (
				expectedQuery string, expectedAffectedRows *int64, expectedError error,
			) {
				t.Helper()

				expectedError = errors.New("some error")
				expectedQuery = selectQueryForTesting
				expectAnalyzeForQuery(mock, expectedQuery, expectedError, withSQLAnalyze)
				mock.ExpectQuery("^" + regexp.QuoteMeta(expectedQuery) + "$").WillReturnError(expectedError)
				var row *sql.Row
				require.NoError(t, db.WithFixedConnection(func(db *DB) error {
					connWrapper, ok := db.db.CommonDB().(*sqlConnWrapper)
					require.True(t, ok)
					row = connWrapper.QueryRow(expectedQuery)
					return nil
				}))
				assert.Equal(t, expectedError, row.Err())
				return expectedQuery, expectedAffectedRows, expectedError
			},
		},
		{
			name: "sqlConnWrapper.QueryRow with success",
			funcToRun: func(t *testing.T, db *DB, mock sqlmock.Sqlmock, withSQLAnalyze bool) (
				expectedQuery string, expectedAffectedRows *int64, expectedError error,
			) {
				t.Helper()

				expectedQuery = selectQueryForTesting
				expectAnalyzeForQuery(mock, expectedQuery, expectedError, withSQLAnalyze)
				mock.ExpectQuery("^" + regexp.QuoteMeta(expectedQuery) + "$").
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1).AddRow(2))
				require.NoError(t, db.WithFixedConnection(func(db *DB) error {
					connWrapper, ok := db.db.CommonDB().(*sqlConnWrapper)
					require.True(t, ok)
					row := connWrapper.QueryRow(expectedQuery)
					require.NoError(t, row.Err())
					_ = row.Scan()
					return nil
				}))
				return expectedQuery, expectedAffectedRows, expectedError
			},
		},
		{
			name: "sqlConnWrapper.BeginTx with error",
			funcToRun: func(t *testing.T, db *DB, mock sqlmock.Sqlmock, _ bool) (
				expectedQuery string, expectedAffectedRows *int64, expectedError error,
			) {
				t.Helper()

				expectedError = errors.New("some error")
				expectedQuery = beginTransactionLogMessage
				mock.ExpectBegin().WillReturnError(expectedError)
				require.NoError(t, db.WithFixedConnection(func(db *DB) error {
					connWrapper, ok := db.db.CommonDB().(*sqlConnWrapper)
					require.True(t, ok)
					tx, err := connWrapper.BeginTx(db.ctx(), &sql.TxOptions{})
					assert.Equal(t, expectedError, err)
					assert.Nil(t, tx)
					return nil
				}))
				return expectedQuery, expectedAffectedRows, expectedError
			},
			skipAnalyzeSQLQueriesTesting: true,
		},
		{
			name: "sqlConnWrapper.BeginTx with success",
			funcToRun: func(t *testing.T, db *DB, mock sqlmock.Sqlmock, _ bool) (
				expectedQuery string, expectedAffectedRows *int64, expectedError error,
			) {
				t.Helper()

				expectedQuery = beginTransactionLogMessage
				mock.ExpectBegin()
				require.NoError(t, db.WithFixedConnection(func(db *DB) error {
					connWrapper, ok := db.db.CommonDB().(*sqlConnWrapper)
					require.True(t, ok)
					tx, err := connWrapper.BeginTx(db.ctx(), &sql.TxOptions{})
					require.NoError(t, err)
					assert.NotNil(t, tx)
					return nil
				}))
				return expectedQuery, expectedAffectedRows, expectedError
			},
			skipAnalyzeSQLQueriesTesting: true,
		},
		{
			name: "sqlConnWrapper.QueryRowContext with error",
			funcToRun: func(t *testing.T, db *DB, mock sqlmock.Sqlmock, withSQLAnalyze bool) (
				expectedQuery string, expectedAffectedRows *int64, expectedError error,
			) {
				t.Helper()

				expectedError = errors.New("some error")
				expectedQuery = selectQueryForTesting
				expectAnalyzeForQuery(mock, expectedQuery, expectedError, withSQLAnalyze)
				mock.ExpectQuery("^" + regexp.QuoteMeta(expectedQuery) + "$").WillReturnError(expectedError)
				dbWrapper, ok := db.db.CommonDB().(*sqlDBWrapper)
				require.True(t, ok)
				conn, err := dbWrapper.conn(db.ctx())
				require.NoError(t, err)
				defer func() { _ = conn.close(nil) }()
				row := conn.QueryRowContext(db.ctx(), expectedQuery)
				assert.Equal(t, expectedError, row.Err())
				return expectedQuery, expectedAffectedRows, expectedError
			},
		},
		{
			name: "sqlConnWrapper.QueryRowContext with success",
			funcToRun: func(t *testing.T, db *DB, mock sqlmock.Sqlmock, withSQLAnalyze bool) (
				expectedQuery string, expectedAffectedRows *int64, expectedError error,
			) {
				t.Helper()

				expectedQuery = selectQueryForTesting
				expectAnalyzeForQuery(mock, expectedQuery, expectedError, withSQLAnalyze)
				mock.ExpectQuery("^" + regexp.QuoteMeta(expectedQuery) + "$").
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1).AddRow(2))
				dbWrapper, ok := db.db.CommonDB().(*sqlDBWrapper)
				require.True(t, ok)
				conn, err := dbWrapper.conn(db.ctx())
				require.NoError(t, err)
				defer func() { _ = conn.close(nil) }()
				row := conn.QueryRowContext(db.ctx(), expectedQuery)
				require.NoError(t, row.Err())
				_ = row.Scan()
				return expectedQuery, expectedAffectedRows, expectedError
			},
		},
		{
			name: "SQLStmtWrapper.ExecContext with error",
			funcToRun: generateTestFuncToCheckSQLQueryLoggingForSQLStmtWrapperQueryContextOrExecContextWithError(
				func(prepareExpectation *sqlmock.ExpectedPrepare, expectedError error) {
					prepareExpectation.ExpectExec().WillReturnError(expectedError)
				},
				func(stmtWrapper *SQLStmtWrapper) (interface{}, error) {
					txWrapper, ok := stmtWrapper.db.(*sqlTxWrapper)
					if !ok {
						return nil, errors.New("the db of the statement is not an sqlTxWrapper")
					}
					return stmtWrapper.ExecContext(txWrapper.ctx)
				}),
		},
		{
			name: "SQLStmtWrapper.ExecContext with success",
			funcToRun: func(t *testing.T, db *DB, mock sqlmock.Sqlmock, withSQLAnalyze bool) (
				expectedQuery string, expectedAffectedRows *int64, expectedError error,
			) {
				t.Helper()

				expectedQuery = updateQueryForTesting
				expectedAffectedRows = golang.Ptr(int64(1))

				mock.ExpectBegin()
				prepareExpectation := mock.ExpectPrepare("^" + regexp.QuoteMeta(expectedQuery) + "$")
				expectAnalyzeForQuery(mock, expectedQuery, expectedError, withSQLAnalyze)
				prepareExpectation.ExpectExec().WillReturnResult(sqlmock.NewResult(-1, *expectedAffectedRows))
				mock.ExpectCommit()

				dbWrapper, ok := db.db.CommonDB().(*sqlDBWrapper)
				require.True(t, ok)
				tx, err := dbWrapper.sqlDB.BeginTx(db.ctx(), &sql.TxOptions{})
				require.NoError(t, err)
				defer func() { _ = tx.Commit() }()
				txWrapper := &sqlTxWrapper{sqlTx: tx, ctx: db.ctx(), logConfig: db.logConfig()}
				stmtWrapper, err := txWrapper.prepare(expectedQuery)
				require.NoError(t, err)
				defer func() { _ = stmtWrapper.Close() }()
				result, err := stmtWrapper.ExecContext(db.ctx())
				require.NoError(t, err)
				assert.NotNil(t, result)
				return expectedQuery, expectedAffectedRows, expectedError
			},
		},
		{
			name: "SQLStmtWrapper.QueryContext with error",
			funcToRun: generateTestFuncToCheckSQLQueryLoggingForSQLStmtWrapperQueryContextOrExecContextWithError(
				func(prepareExpectation *sqlmock.ExpectedPrepare, expectedError error) {
					prepareExpectation.ExpectQuery().WillReturnError(expectedError)
				},
				func(stmtWrapper *SQLStmtWrapper) (interface{}, error) {
					txWrapper, ok := stmtWrapper.db.(*sqlTxWrapper)
					if !ok {
						return nil, errors.New("the db of the statement is not an sqlTxWrapper")
					}
					//nolint:rowserrcheck // the caller is responsible for checking
					return stmtWrapper.QueryContext(txWrapper.ctx)
				}),
		},
		{
			name: "SQLStmtWrapper.QueryContext with success",
			funcToRun: func(t *testing.T, db *DB, mock sqlmock.Sqlmock, withSQLAnalyze bool) (
				expectedQuery string, expectedAffectedRows *int64, expectedError error,
			) {
				t.Helper()

				expectedQuery = updateQueryForTesting

				mock.ExpectBegin()
				prepareExpectation := mock.ExpectPrepare("^" + regexp.QuoteMeta(expectedQuery) + "$")
				expectAnalyzeForQuery(mock, expectedQuery, expectedError, withSQLAnalyze)
				prepareExpectation.ExpectQuery().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1).AddRow(2))
				mock.ExpectCommit()

				dbWrapper, ok := db.db.CommonDB().(*sqlDBWrapper)
				require.True(t, ok)
				tx, err := dbWrapper.sqlDB.BeginTx(db.ctx(), &sql.TxOptions{})
				require.NoError(t, err)
				defer func() { _ = tx.Commit() }()
				txWrapper := &sqlTxWrapper{sqlTx: tx, ctx: db.ctx(), logConfig: db.logConfig()}
				stmtWrapper, err := txWrapper.prepare(expectedQuery)
				require.NoError(t, err)
				defer func() { _ = stmtWrapper.Close() }()
				rows, err := stmtWrapper.QueryContext(db.ctx())
				require.NoError(t, err)
				if rows != nil {
					defer func() { _ = rows.Close() }()
					require.NoError(t, rows.Err())
				}
				return expectedQuery, expectedAffectedRows, expectedError
			},
		},
		{
			name: "SQLStmtWrapper.QueryRowContext with error",
			funcToRun: func(t *testing.T, db *DB, mock sqlmock.Sqlmock, withSQLAnalyze bool) (
				expectedQuery string, expectedAffectedRows *int64, expectedError error,
			) {
				t.Helper()

				expectedError = errors.New("some error")
				expectedQuery = updateQueryForTesting

				mock.ExpectBegin()
				prepareExpectation := mock.ExpectPrepare("^" + regexp.QuoteMeta(expectedQuery) + "$")
				expectAnalyzeForQuery(mock, expectedQuery, expectedError, withSQLAnalyze)
				prepareExpectation.ExpectQuery().WillReturnError(expectedError)
				mock.ExpectCommit()

				dbWrapper, ok := db.db.CommonDB().(*sqlDBWrapper)
				require.True(t, ok)
				tx, err := dbWrapper.sqlDB.BeginTx(db.ctx(), &sql.TxOptions{})
				require.NoError(t, err)
				defer func() { _ = tx.Commit() }()
				txWrapper := &sqlTxWrapper{sqlTx: tx, ctx: db.ctx(), logConfig: db.logConfig()}
				stmtWrapper, err := txWrapper.prepare(expectedQuery)
				require.NoError(t, err)
				defer func() { _ = stmtWrapper.Close() }()
				row := stmtWrapper.QueryRowContext(db.ctx())
				assert.Equal(t, expectedError, row.Err())
				return expectedQuery, expectedAffectedRows, expectedError
			},
		},
		{
			name: "SQLStmtWrapper.QueryRowContext with success",
			funcToRun: func(t *testing.T, db *DB, mock sqlmock.Sqlmock, withSQLAnalyze bool) (
				expectedQuery string, expectedAffectedRows *int64, expectedError error,
			) {
				t.Helper()

				expectedQuery = updateQueryForTesting

				mock.ExpectBegin()
				prepareExpectation := mock.ExpectPrepare("^" + regexp.QuoteMeta(expectedQuery) + "$")
				expectAnalyzeForQuery(mock, expectedQuery, expectedError, withSQLAnalyze)
				prepareExpectation.ExpectQuery().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1).AddRow(2))
				mock.ExpectCommit()

				dbWrapper, ok := db.db.CommonDB().(*sqlDBWrapper)
				require.True(t, ok)
				tx, err := dbWrapper.sqlDB.BeginTx(db.ctx(), &sql.TxOptions{})
				require.NoError(t, err)
				defer func() { _ = tx.Commit() }()
				txWrapper := &sqlTxWrapper{sqlTx: tx, ctx: db.ctx(), logConfig: db.logConfig()}
				stmtWrapper, err := txWrapper.prepare(expectedQuery)
				require.NoError(t, err)
				defer func() { _ = stmtWrapper.Close() }()
				row := stmtWrapper.QueryRowContext(db.ctx())
				require.NoError(t, row.Err())
				_ = row.Scan()
				return expectedQuery, expectedAffectedRows, expectedError
			},
		},
		{
			name: "sqlTxWrapper.Exec with error",
			funcToRun: generateTestFuncToCheckSQLQueryLoggingForSQLTxWrapperExecOrQueryWithError(
				func(mock sqlmock.Sqlmock, expectedSQL string, expectedError error) {
					mock.ExpectExec(expectedSQL).WillReturnError(expectedError)
				},
				func(txWrapper *sqlTxWrapper, expectedQuery string) (interface{}, error) {
					return txWrapper.Exec(expectedQuery)
				},
			),
		},
		{
			name: "sqlTxWrapper.Exec with success",
			funcToRun: func(t *testing.T, db *DB, mock sqlmock.Sqlmock, withSQLAnalyze bool) (
				expectedQuery string, expectedAffectedRows *int64, expectedError error,
			) {
				t.Helper()

				expectedQuery = updateQueryForTesting
				expectedAffectedRows = golang.Ptr(int64(1))

				mock.ExpectBegin()
				expectAnalyzeForQuery(mock, expectedQuery, expectedError, withSQLAnalyze)
				mock.ExpectExec("^" + regexp.QuoteMeta(expectedQuery) + "$").
					WillReturnResult(sqlmock.NewResult(-1, *expectedAffectedRows))
				mock.ExpectCommit()

				dbWrapper, ok := db.db.CommonDB().(*sqlDBWrapper)
				require.True(t, ok)
				tx, err := dbWrapper.sqlDB.BeginTx(db.ctx(), &sql.TxOptions{})
				require.NoError(t, err)
				defer func() { _ = tx.Commit() }()
				txWrapper := &sqlTxWrapper{sqlTx: tx, ctx: db.ctx(), logConfig: db.logConfig()}
				result, err := txWrapper.Exec(expectedQuery)
				require.NoError(t, err)
				assert.NotNil(t, result)
				return expectedQuery, expectedAffectedRows, expectedError
			},
		},
		{
			name: "sqlTxWrapper.Query with error",
			funcToRun: generateTestFuncToCheckSQLQueryLoggingForSQLTxWrapperExecOrQueryWithError(
				func(mock sqlmock.Sqlmock, expectedSQL string, expectedError error) {
					mock.ExpectQuery(expectedSQL).WillReturnError(expectedError)
				},
				func(txWrapper *sqlTxWrapper, expectedQuery string) (interface{}, error) {
					return txWrapper.Query(expectedQuery) //nolint:rowserrcheck // the caller is responsible for checking
				}),
		},
		{
			name: "sqlTxWrapper.Query with success",
			funcToRun: func(t *testing.T, db *DB, mock sqlmock.Sqlmock, withSQLAnalyze bool) (
				expectedQuery string, expectedAffectedRows *int64, expectedError error,
			) {
				t.Helper()

				expectedQuery = updateQueryForTesting

				mock.ExpectBegin()
				expectAnalyzeForQuery(mock, expectedQuery, expectedError, withSQLAnalyze)
				mock.ExpectQuery("^" + regexp.QuoteMeta(expectedQuery) + "$").
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1).AddRow(2))
				mock.ExpectCommit()

				dbWrapper, ok := db.db.CommonDB().(*sqlDBWrapper)
				require.True(t, ok)
				tx, err := dbWrapper.sqlDB.BeginTx(db.ctx(), &sql.TxOptions{})
				require.NoError(t, err)
				defer func() { _ = tx.Commit() }()
				txWrapper := &sqlTxWrapper{sqlTx: tx, ctx: db.ctx(), logConfig: db.logConfig()}
				rows, err := txWrapper.Query(expectedQuery)
				require.NoError(t, err)
				if rows != nil {
					defer func() { _ = rows.Close() }()
					require.NoError(t, rows.Err())
				}
				return expectedQuery, expectedAffectedRows, expectedError
			},
		},
		{
			name: "sqlTxWrapper.QueryRow with error",
			funcToRun: func(t *testing.T, db *DB, mock sqlmock.Sqlmock, withSQLAnalyze bool) (
				expectedQuery string, expectedAffectedRows *int64, expectedError error,
			) {
				t.Helper()

				expectedError = errors.New("some error")
				expectedQuery = updateQueryForTesting

				mock.ExpectBegin()
				expectAnalyzeForQuery(mock, expectedQuery, expectedError, withSQLAnalyze)
				mock.ExpectQuery("^" + regexp.QuoteMeta(expectedQuery) + "$").WillReturnError(expectedError)
				mock.ExpectCommit()

				dbWrapper, ok := db.db.CommonDB().(*sqlDBWrapper)
				require.True(t, ok)
				tx, err := dbWrapper.sqlDB.BeginTx(db.ctx(), &sql.TxOptions{})
				require.NoError(t, err)
				defer func() { _ = tx.Commit() }()
				txWrapper := &sqlTxWrapper{sqlTx: tx, ctx: db.ctx(), logConfig: db.logConfig()}
				row := txWrapper.QueryRow(expectedQuery)
				assert.Equal(t, expectedError, row.Err())
				return expectedQuery, expectedAffectedRows, expectedError
			},
		},
		{
			name: "sqlTxWrapper.QueryRow with success",
			funcToRun: func(t *testing.T, db *DB, mock sqlmock.Sqlmock, withSQLAnalyze bool) (
				expectedQuery string, expectedAffectedRows *int64, expectedError error,
			) {
				t.Helper()

				expectedQuery = updateQueryForTesting

				mock.ExpectBegin()
				expectAnalyzeForQuery(mock, expectedQuery, expectedError, withSQLAnalyze)
				mock.ExpectQuery("^" + regexp.QuoteMeta(expectedQuery) + "$").
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1).AddRow(2))
				mock.ExpectCommit()

				dbWrapper, ok := db.db.CommonDB().(*sqlDBWrapper)
				require.True(t, ok)
				tx, err := dbWrapper.sqlDB.BeginTx(db.ctx(), &sql.TxOptions{})
				require.NoError(t, err)
				defer func() { _ = tx.Commit() }()
				txWrapper := &sqlTxWrapper{sqlTx: tx, ctx: db.ctx(), logConfig: db.logConfig()}
				row := txWrapper.QueryRow(expectedQuery)
				require.NoError(t, row.Err())
				_ = row.Scan()
				return expectedQuery, expectedAffectedRows, expectedError
			},
		},
		{
			name: "sqlTxWrapper.Commit with error",
			funcToRun: generateTestFuncToCheckSQLQueryLoggingForSQLTxWrapperCommitOrRollbackWithError(
				func(mock sqlmock.Sqlmock, expectedError error) {
					mock.ExpectCommit().WillReturnError(expectedError)
				},
				func(txWrapper *sqlTxWrapper) error {
					return txWrapper.Commit()
				},
				commitTransactionLogMessage,
			),
			skipAnalyzeSQLQueriesTesting: true,
		},
		{
			name: "sqlTxWrapper.Commit with success",
			funcToRun: generateTestFuncToCheckSQLQueryLoggingForSQLTxWrapperSuccessfulCommitOrRollback(
				func(mock sqlmock.Sqlmock) {
					mock.ExpectCommit()
				},
				func(txWrapper *sqlTxWrapper) error {
					return txWrapper.Commit()
				},
				commitTransactionLogMessage),
			skipAnalyzeSQLQueriesTesting: true,
		},
		{
			name: "sqlTxWrapper.Rollback with error",
			funcToRun: generateTestFuncToCheckSQLQueryLoggingForSQLTxWrapperCommitOrRollbackWithError(
				func(mock sqlmock.Sqlmock, expectedError error) {
					mock.ExpectRollback().WillReturnError(expectedError)
				},
				func(txWrapper *sqlTxWrapper) error {
					return txWrapper.Rollback()
				},
				rollbackTransactionLogMessage),
			skipAnalyzeSQLQueriesTesting: true,
		},
		{
			name: "sqlTxWrapper.Rollback with success",
			funcToRun: generateTestFuncToCheckSQLQueryLoggingForSQLTxWrapperSuccessfulCommitOrRollback(
				func(mock sqlmock.Sqlmock) {
					mock.ExpectRollback()
				},
				func(txWrapper *sqlTxWrapper) error {
					return txWrapper.Rollback()
				},
				rollbackTransactionLogMessage),
			skipAnalyzeSQLQueriesTesting: true,
		},
		{
			name: "sqlTxWrapper.Rollback error because of a context cancellation",
			funcToRun: generateTestFuncToCheckSQLQueryLoggingForSQLTxWrapperCommitOrRollbackFailingBecauseOfContextCancellation(
				func(tx *sqlTxWrapper) error {
					return tx.Rollback()
				},
				rollbackTransactionLogMessage),
			skipAnalyzeSQLQueriesTesting: true,
		},
		{
			name: "sqlTxWrapper.Commit error because of a context cancellation",
			funcToRun: generateTestFuncToCheckSQLQueryLoggingForSQLTxWrapperCommitOrRollbackFailingBecauseOfContextCancellation(
				func(tx *sqlTxWrapper) error {
					return tx.Commit()
				},
				commitTransactionLogMessage),
			skipAnalyzeSQLQueriesTesting: true,
		},
	}
}

func generateTestFuncToCheckSQLConnWrapperQueryOrExecWithError(
	mockCallOnDBFunc func(sqlmock.Sqlmock, string, error),
	callOnConnFunc func(*sqlConnWrapper, string) (interface{}, error),
) funcToRunInSQLQueryLoggingTests {
	return func(t *testing.T, db *DB, mock sqlmock.Sqlmock, withSQLAnalyze bool) (
		expectedQuery string, expectedAffectedRows *int64, expectedError error,
	) {
		t.Helper()

		expectedError = errors.New("some error")
		expectedQuery = updateQueryForTesting
		expectAnalyzeForQuery(mock, expectedQuery, expectedError, withSQLAnalyze)
		mockCallOnDBFunc(mock, "^"+regexp.QuoteMeta(expectedQuery)+"$", expectedError)
		var resultOrRows interface{}
		var err error
		err = db.WithFixedConnection(func(db *DB) error {
			connWrapper, ok := db.db.CommonDB().(*sqlConnWrapper)
			require.True(t, ok)
			resultOrRows, err = callOnConnFunc(connWrapper, expectedQuery)
			assert.Equal(t, expectedError, err)
			return err
		})
		assert.Equal(t, expectedError, err)
		assert.Nil(t, resultOrRows)
		return expectedQuery, expectedAffectedRows, expectedError
	}
}

func generateTestFuncToCheckSQLDBWrapperQueryOrExecWithError(
	mockCallOnDBFunc func(sqlmock.Sqlmock, string, error),
	callOnDBFunc func(*sqlDBWrapper, string) (interface{}, error),
) funcToRunInSQLQueryLoggingTests {
	return func(t *testing.T, db *DB, mock sqlmock.Sqlmock, withSQLAnalyze bool) (
		expectedQuery string, expectedAffectedRows *int64, expectedError error,
	) {
		t.Helper()

		expectedError = errors.New("some error")
		expectedQuery = updateQueryForTesting
		expectAnalyzeForQuery(mock, expectedQuery, expectedError, withSQLAnalyze)
		mockCallOnDBFunc(mock, "^"+regexp.QuoteMeta(expectedQuery)+"$", expectedError)
		dbWrapper, ok := db.db.CommonDB().(*sqlDBWrapper)
		require.True(t, ok)
		resultOrRows, err := callOnDBFunc(dbWrapper, expectedQuery)
		assert.Equal(t, expectedError, err)
		assert.Nil(t, resultOrRows)
		return expectedQuery, expectedAffectedRows, expectedError
	}
}

func generateTestFuncToCheckSQLQueryLoggingForSQLStmtWrapperQueryContextOrExecContextWithError(
	mockCallOnPreparedStmtFunc func(*sqlmock.ExpectedPrepare, error),
	callOnPreparedStmtFunc func(*SQLStmtWrapper) (interface{}, error),
) funcToRunInSQLQueryLoggingTests {
	return func(t *testing.T, db *DB, mock sqlmock.Sqlmock, withSQLAnalyze bool) (
		expectedQuery string, expectedAffectedRows *int64, expectedError error,
	) {
		t.Helper()

		expectedError = errors.New("some error")
		expectedQuery = updateQueryForTesting

		mock.ExpectBegin()
		prepareExpectation := mock.ExpectPrepare("^" + regexp.QuoteMeta(expectedQuery) + "$")
		expectAnalyzeForQuery(mock, expectedQuery, expectedError, withSQLAnalyze)
		mockCallOnPreparedStmtFunc(prepareExpectation, expectedError)
		mock.ExpectCommit()

		dbWrapper, ok := db.db.CommonDB().(*sqlDBWrapper)
		require.True(t, ok)
		tx, err := dbWrapper.sqlDB.BeginTx(db.ctx(), &sql.TxOptions{})
		require.NoError(t, err)
		defer func() { _ = tx.Commit() }()
		txWrapper := &sqlTxWrapper{sqlTx: tx, ctx: db.ctx(), logConfig: db.logConfig()}
		stmtWrapper, err := txWrapper.prepare(expectedQuery)
		require.NoError(t, err)
		defer func() { _ = stmtWrapper.Close() }()
		resultOrRows, err := callOnPreparedStmtFunc(stmtWrapper)
		assert.Equal(t, expectedError, err)
		assert.Nil(t, resultOrRows)
		return expectedQuery, expectedAffectedRows, expectedError
	}
}

func generateTestFuncToCheckSQLQueryLoggingForSQLTxWrapperExecOrQueryWithError(
	mockCallOnTxFunc func(sqlmock.Sqlmock, string, error),
	callOnTxFunc func(*sqlTxWrapper, string) (interface{}, error),
) funcToRunInSQLQueryLoggingTests {
	return func(t *testing.T, db *DB, mock sqlmock.Sqlmock, withSQLAnalyze bool) (
		expectedQuery string, expectedAffectedRows *int64, expectedError error,
	) {
		t.Helper()

		expectedError = errors.New("some error")
		expectedQuery = updateQueryForTesting

		mock.ExpectBegin()
		expectAnalyzeForQuery(mock, expectedQuery, expectedError, withSQLAnalyze)
		mockCallOnTxFunc(mock, "^"+regexp.QuoteMeta(expectedQuery)+"$", expectedError)
		mock.ExpectCommit()

		dbWrapper, ok := db.db.CommonDB().(*sqlDBWrapper)
		require.True(t, ok)
		tx, err := dbWrapper.sqlDB.BeginTx(db.ctx(), &sql.TxOptions{})
		require.NoError(t, err)
		defer func() { _ = tx.Commit() }()
		txWrapper := &sqlTxWrapper{sqlTx: tx, ctx: db.ctx(), logConfig: db.logConfig()}
		resultOrRows, err := callOnTxFunc(txWrapper, expectedQuery)
		assert.Equal(t, expectedError, err)
		assert.Nil(t, resultOrRows)
		return expectedQuery, expectedAffectedRows, expectedError
	}
}

func generateTestFuncToCheckSQLQueryLoggingForSQLTxWrapperSuccessfulCommitOrRollback(
	mockCallOnTxFunc func(sqlmock.Sqlmock),
	callOnTxFunc func(*sqlTxWrapper) error,
	expectedQueryLogMessage string,
) funcToRunInSQLQueryLoggingTests {
	return func(t *testing.T, db *DB, mock sqlmock.Sqlmock, _ bool) (
		expectedQuery string, expectedAffectedRows *int64, expectedError error,
	) {
		t.Helper()

		mock.ExpectBegin()
		mockCallOnTxFunc(mock)

		dbWrapper, ok := db.db.CommonDB().(*sqlDBWrapper)
		require.True(t, ok)
		tx, err := dbWrapper.sqlDB.BeginTx(db.ctx(), &sql.TxOptions{})
		require.NoError(t, err)
		txWrapper := &sqlTxWrapper{sqlTx: tx, ctx: db.ctx(), logConfig: db.logConfig()}
		err = callOnTxFunc(txWrapper)
		require.NoError(t, err)
		return expectedQueryLogMessage, expectedAffectedRows, expectedError
	}
}

func generateTestFuncToCheckSQLQueryLoggingForSQLTxWrapperCommitOrRollbackWithError(
	mockCallOnTxFunc func(sqlmock.Sqlmock, error),
	callOnTxFunc func(*sqlTxWrapper) error,
	expectedQueryLogMessage string,
) funcToRunInSQLQueryLoggingTests {
	return func(t *testing.T, db *DB, mock sqlmock.Sqlmock, _ bool) (
		expectedQuery string, expectedAffectedRows *int64, expectedError error,
	) {
		t.Helper()

		expectedError = errors.New("some error")

		mock.ExpectBegin()
		mockCallOnTxFunc(mock, expectedError)

		dbWrapper, ok := db.db.CommonDB().(*sqlDBWrapper)
		require.True(t, ok)
		tx, err := dbWrapper.sqlDB.BeginTx(db.ctx(), &sql.TxOptions{})
		require.NoError(t, err)
		txWrapper := &sqlTxWrapper{sqlTx: tx, ctx: db.ctx(), logConfig: db.logConfig()}
		err = callOnTxFunc(txWrapper)
		assert.Equal(t, expectedError, err)
		return expectedQueryLogMessage, expectedAffectedRows, expectedError
	}
}

func generateTestFuncToCheckSQLQueryLoggingForSQLTxWrapperCommitOrRollbackFailingBecauseOfContextCancellation(
	callOnTxFunc func(*sqlTxWrapper) error,
	expectedQueryLogMessage string,
) funcToRunInSQLQueryLoggingTests {
	return func(t *testing.T, db *DB, mock sqlmock.Sqlmock, _ bool) (
		expectedQuery string, expectedAffectedRows *int64, expectedError error,
	) {
		t.Helper()

		expectedError = context.Canceled

		mock.ExpectBegin()
		mock.ExpectRollback()

		ctx, cancelFunc := context.WithCancel(db.ctx())
		sqlDBWrapper, ok := db.db.CommonDB().(*sqlDBWrapper)
		require.True(t, ok)
		oldLogSQLQueriesValue := sqlDBWrapper.logConfig.LogSQLQueries
		sqlDBWrapper.logConfig.LogSQLQueries = false
		tx, err := sqlDBWrapper.BeginTx(ctx, &sql.TxOptions{})
		require.NoError(t, err)
		sqlDBWrapper.logConfig.LogSQLQueries = oldLogSQLQueriesValue
		tx.logConfig.LogSQLQueries = oldLogSQLQueriesValue
		cancelFunc()

		assert.Eventually(t, func() bool {
			return mock.ExpectationsWereMet() == nil
		}, 1*time.Second, 10*time.Millisecond)
		require.NoError(t, mock.ExpectationsWereMet(), "cancelFunc() should have caused the rollback to be called")
		err = callOnTxFunc(tx)
		require.Equal(t, expectedError, err)
		return expectedQueryLogMessage, expectedAffectedRows,
			errors.New(expectedError.Error() + " (the transaction has been rolled back implicitly)")
	}
}

func Test_SQLQueryLogging(t *testing.T) {
	for _, test := range sqlQueryLoggingTests() {
		test := test
		t.Run(test.name, func(t *testing.T) {
			for _, logSQLQueries := range []bool{false, true} {
				for _, analyzeSQLQueries := range []bool{false, true} {
					if test.skipAnalyzeSQLQueriesTesting && analyzeSQLQueries || analyzeSQLQueries && !logSQLQueries {
						continue
					}
					t.Run(fmt.Sprintf("logSQLQueries=%v,analyzeSQLQueries=%v", logSQLQueries, analyzeSQLQueries), func(t *testing.T) {
						testoutput.SuppressIfPasses(t)

						verifySQLLogs(t, logSQLQueries, analyzeSQLQueries, test)
					})
				}
			}
		})
	}
}

func verifySQLLogs(t *testing.T, logSQLQueries, analyzeSQLQueries bool, test sqlQueryLoggingTest) {
	t.Helper()

	ctx, _, loggerHook := logging.NewContextWithNewMockLogger()
	db, mock := NewDBMock(ctx)
	defer func() { _ = db.Close() }()

	db.logConfig().LogSQLQueries = logSQLQueries
	db.logConfig().AnalyzeSQLQueries = analyzeSQLQueries

	expectedQuery, expectedAffectedRows, expectedError := test.funcToRun(t, db, mock, analyzeSQLQueries)

	require.NoError(t, mock.ExpectationsWereMet())

	logEntries := loggerHook.AllEntries()
	expectedLen := 2
	if !logSQLQueries {
		expectedLen--
	}
	if expectedError == nil {
		expectedLen--
	}
	if analyzeSQLQueries {
		expectedLen++
	}
	require.Len(t, logEntries, expectedLen, "logs: %s", (&loggingtest.Hook{Hook: loggerHook}).GetAllLogs())
	index := 0
	if logSQLQueries {
		assert.Equal(t, "info", logEntries[0].Level.String())
		assert.Equal(t, expectedQuery, logEntries[0].Message)
		assert.Equal(t, "db", logEntries[0].Data["type"])
		assertDurationIsOK(t, logEntries[0])
		if expectedAffectedRows != nil {
			assert.Equal(t, *expectedAffectedRows, logEntries[0].Data["rows"])
		} else {
			assert.Nil(t, logEntries[0].Data["rows"])
		}
		index++
	}

	if expectedError != nil {
		assert.Equal(t, "error", logEntries[index].Level.String())
		assert.Equal(t, expectedError.Error(), logEntries[index].Message)
		assert.Equal(t, "db", logEntries[index].Data["type"])
		index++
	}

	if analyzeSQLQueries {
		if expectedError != nil {
			assert.Equal(t, "error", logEntries[index].Level.String())
			assert.Equal(t, "Failed to get an execution plan for a SQL query: "+expectedError.Error(), logEntries[index].Message)
		} else {
			assert.Equal(t, "info", logEntries[index].Level.String())
			assert.Equal(t, "query execution plan:\nplan\n", logEntries[index].Message)
			assertDurationIsOK(t, logEntries[index])
		}
		assert.Equal(t, "db", logEntries[index].Data["type"])
	}
}

func assertDurationIsOK(t *testing.T, entry *logrus.Entry) {
	t.Helper()

	assert.Contains(t, entry.Data, "duration")
	if duration, ok := entry.Data["duration"]; ok {
		assert.IsType(t, "", duration)
		if durationStr, ok := duration.(string); ok {
			parsedDuration, err := time.ParseDuration(durationStr)
			require.NoError(t, err)
			assert.Positive(t, parsedDuration)
		}
	}
}

func Test_SQLQueryLogging_Select(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	ctx, logger, loggerHook := logging.NewContextWithNewMockLogger()
	conf := viper.New()
	conf.Set("Format", "json")
	conf.Set("Output", "stdout")
	conf.Set("LogSQLQueries", true)
	logger.Configure(conf)
	db, mock := NewDBMock(ctx)
	defer func() { _ = db.Close() }()

	timeParam := time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)

	mock.ExpectQuery(`^SELECT \$1, \$2, \$3, \$4, \$5$`).
		WithArgs(1, timeParam, "foo", []byte("bar"), nil).
		WillReturnRows(mock.NewRows([]string{"1"}).AddRow(1))

	var result []interface{}
	db.Raw("SELECT $1, $2, $3, $4, $5", 1, timeParam, "foo", []byte("bar"), nil).Scan(&result)
	require.NotNil(t, loggerHook.LastEntry())
	assert.Equal(t, `SELECT 1, '2009-11-10 23:00:00', 'foo', 'bar', NULL`, loggerHook.LastEntry().Message)
	data := loggerHook.LastEntry().Data
	assert.Equal(t, "db", data["type"])
	assertDurationIsOK(t, loggerHook.LastEntry())
	assert.NotContains(t, data, "rows")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func Test_SQLQueryLogging_Update(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	ctx, logger, loggerHook := logging.NewContextWithNewMockLogger()
	conf := viper.New()
	conf.Set("Format", "json")
	conf.Set("Output", "stdout")
	conf.Set("LogSQLQueries", true)
	logger.Configure(conf)
	db, mock := NewDBMock(ctx)
	defer func() { _ = db.Close() }()

	timeParam := time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)

	mock.ExpectExec(`^UPDATE t1 SET c1=\$1, c2=\$2, c3=\$3, c4=\$4, c5=\$5$`).
		WithArgs(1, timeParam, "foo", []byte("bar"), nil).
		WillReturnResult(sqlmock.NewResult(-1, 123))

	db.Exec("UPDATE t1 SET c1=$1, c2=$2, c3=$3, c4=$4, c5=$5", 1, timeParam, "foo", []byte("bar"), nil)
	assert.Equal(t, `UPDATE t1 SET c1=1, c2='2009-11-10 23:00:00', c3='foo', c4='bar', c5=NULL`, loggerHook.LastEntry().Message)
	data := loggerHook.LastEntry().Data
	assert.Equal(t, "db", data["type"])
	assertDurationIsOK(t, loggerHook.LastEntry())
	assert.Equal(t, int64(123), data["rows"])
	assert.NoError(t, mock.ExpectationsWereMet())
}

func Test_SQLQueryLogging_SQLWithInterrogationMark(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	ctx, logger, loggerHook := logging.NewContextWithNewMockLogger()
	conf := viper.New()
	conf.Set("Format", "json")
	conf.Set("Output", "stdout")
	conf.Set("LogSQLQueries", true)
	logger.Configure(conf)
	db, mock := NewDBMock(ctx)
	defer func() { _ = db.Close() }()

	mock.ExpectQuery(`^SELECT \?$`).WithArgs(1).WillReturnRows(mock.NewRows([]string{"1"}).AddRow(1))

	var result []interface{}
	db.Raw("SELECT ?", 1).Scan(&result)
	assert.Equal(t, "SELECT 1", loggerHook.LastEntry().Message)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func Test_SQLQueryLogging_SQLError(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	ctx, logger, loggerHook := logging.NewContextWithNewMockLogger()
	conf := viper.New()
	conf.Set("Format", "json")
	conf.Set("Output", "stdout")
	conf.Set("LogSQLQueries", true)
	conf.Set("Level", "debug")
	logger.Configure(conf)
	db, mock := NewDBMock(ctx)
	defer func() { _ = db.Close() }()

	mock.ExpectQuery("SELECT 2").WillReturnError(errors.New("a query error"))

	var result []interface{}
	db.Raw("SELECT 2").Scan(&result)

	assert.Equal(t, "SELECT 2", loggerHook.Entries[0].Message)
	entry := loggerHook.Entries[0]
	data := entry.Data
	assert.Equal(t, "db", data["type"])
	assertDurationIsOK(t, &entry)
	assert.Nil(t, data["rows"])
	require.NoError(t, mock.ExpectationsWereMet())

	assert.Equal(t, "a query error", loggerHook.Entries[1].Message)
	assert.Equal(t, "error", loggerHook.Entries[1].Level.String())
	assert.NotNil(t, loggerHook.Entries[1].Time)
	assert.Equal(t, "db", loggerHook.Entries[1].Data["type"])
}
