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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/France-ioi/AlgoreaBackend/v2/app/logging"
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
	updateQueryForTesting = "UPDATE users SET name = 'John' WHERE id = 1"
	selectQueryForTesting = "SELECT id FROM users"
)

type funcToRunInSQLQueryLoggingTests func(t *testing.T, db *DB, mock sqlmock.Sqlmock, withSQLAnalyze bool) (
	expectedQuery string, expectedAffectedRows *int64, expectedError error,
)

type sqlQueryLoggingTest struct {
	name                         string
	funcToRun                    funcToRunInSQLQueryLoggingTests
	skipAnalyzeSQLQueriesTesting bool
}

var sqlQueryLoggingTests = []sqlQueryLoggingTest{
	{
		name: "sqlDBWrapper.Exec with error",
		funcToRun: func(t *testing.T, db *DB, mock sqlmock.Sqlmock, withSQLAnalyze bool) (
			expectedQuery string, expectedAffectedRows *int64, expectedError error,
		) {
			expectedError = errors.New("some error")
			expectedQuery = updateQueryForTesting
			expectAnalyzeForQuery(mock, expectedQuery, expectedError, withSQLAnalyze)
			mock.ExpectExec("^" + regexp.QuoteMeta(expectedQuery) + "$").WillReturnError(expectedError)
			result, err := db.db.CommonDB().(*sqlDBWrapper).Exec(expectedQuery)
			assert.Equal(t, expectedError, err)
			assert.Nil(t, result)
			return expectedQuery, expectedAffectedRows, expectedError
		},
	},
	{
		name: "sqlDBWrapper.Exec with success",
		funcToRun: func(t *testing.T, db *DB, mock sqlmock.Sqlmock, withSQLAnalyze bool) (
			expectedQuery string, expectedAffectedRows *int64, expectedError error,
		) {
			expectedQuery = updateQueryForTesting
			expectedAffectedRows = golang.Ptr(int64(1))
			expectAnalyzeForQuery(mock, expectedQuery, expectedError, withSQLAnalyze)
			mock.ExpectExec("^" + regexp.QuoteMeta(expectedQuery) + "$").
				WillReturnResult(sqlmock.NewResult(-1, *expectedAffectedRows))
			result, err := db.db.CommonDB().(*sqlDBWrapper).Exec(expectedQuery)
			assert.NoError(t, err)
			assert.NotNil(t, result)
			return expectedQuery, expectedAffectedRows, expectedError
		},
	},
	{
		name: "sqlDBWrapper.Query with error",
		funcToRun: func(t *testing.T, db *DB, mock sqlmock.Sqlmock, withSQLAnalyze bool) (
			expectedQuery string, expectedAffectedRows *int64, expectedError error,
		) {
			expectedError = errors.New("some error")
			expectedQuery = selectQueryForTesting
			expectAnalyzeForQuery(mock, expectedQuery, expectedError, withSQLAnalyze)
			mock.ExpectQuery("^" + regexp.QuoteMeta(expectedQuery) + "$").WillReturnError(expectedError)
			result, err := db.db.CommonDB().(*sqlDBWrapper).Query(expectedQuery)
			assert.Equal(t, expectedError, err)
			assert.Nil(t, result)
			return expectedQuery, expectedAffectedRows, expectedError
		},
	},
	{
		name: "sqlDBWrapper.Query with success",
		funcToRun: func(t *testing.T, db *DB, mock sqlmock.Sqlmock, withSQLAnalyze bool) (
			expectedQuery string, expectedAffectedRows *int64, expectedError error,
		) {
			expectedQuery = selectQueryForTesting
			expectAnalyzeForQuery(mock, expectedQuery, expectedError, withSQLAnalyze)
			mock.ExpectQuery("^" + regexp.QuoteMeta(expectedQuery) + "$").
				WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1).AddRow(2))
			rows, err := db.db.CommonDB().(*sqlDBWrapper).Query(expectedQuery)
			assert.NoError(t, err)
			assert.NotNil(t, rows)
			if rows != nil {
				_ = rows.Close()
			}
			return expectedQuery, expectedAffectedRows, expectedError
		},
	},
	{
		name: "sqlDBWrapper.QueryRow with error",
		funcToRun: func(t *testing.T, db *DB, mock sqlmock.Sqlmock, withSQLAnalyze bool) (
			expectedQuery string, expectedAffectedRows *int64, expectedError error,
		) {
			expectedError = errors.New("some error")
			expectedQuery = selectQueryForTesting
			expectAnalyzeForQuery(mock, expectedQuery, expectedError, withSQLAnalyze)
			mock.ExpectQuery("^" + regexp.QuoteMeta(expectedQuery) + "$").WillReturnError(expectedError)
			row := db.db.CommonDB().(*sqlDBWrapper).QueryRow(expectedQuery)
			assert.Equal(t, expectedError, row.Err())
			return expectedQuery, expectedAffectedRows, expectedError
		},
	},
	{
		name: "sqlDBWrapper.QueryRow with success",
		funcToRun: func(t *testing.T, db *DB, mock sqlmock.Sqlmock, withSQLAnalyze bool) (
			expectedQuery string, expectedAffectedRows *int64, expectedError error,
		) {
			expectedQuery = selectQueryForTesting
			expectAnalyzeForQuery(mock, expectedQuery, expectedError, withSQLAnalyze)
			mock.ExpectQuery("^" + regexp.QuoteMeta(expectedQuery) + "$").
				WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1).AddRow(2))
			row := db.db.CommonDB().(*sqlDBWrapper).QueryRow(expectedQuery)
			assert.NoError(t, row.Err())
			_ = row.Scan()
			return expectedQuery, expectedAffectedRows, expectedError
		},
	},
	{
		name: "sqlDBWrapper.BeginTx with error",
		funcToRun: func(t *testing.T, db *DB, mock sqlmock.Sqlmock, _ bool) (
			expectedQuery string, expectedAffectedRows *int64, expectedError error,
		) {
			expectedError = errors.New("some error")
			expectedQuery = beginTransactionLogMessage
			mock.ExpectBegin().WillReturnError(expectedError)
			tx, err := db.db.CommonDB().(*sqlDBWrapper).BeginTx(context.Background(), &sql.TxOptions{})
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
			expectedQuery = beginTransactionLogMessage
			mock.ExpectBegin()
			tx, err := db.db.CommonDB().(*sqlDBWrapper).BeginTx(context.Background(), &sql.TxOptions{})
			assert.NoError(t, err)
			assert.NotNil(t, tx)
			return expectedQuery, expectedAffectedRows, expectedError
		},
		skipAnalyzeSQLQueriesTesting: true,
	},
	{
		name: "sqlConnWrapper.QueryRowContext with error",
		funcToRun: func(t *testing.T, db *DB, mock sqlmock.Sqlmock, withSQLAnalyze bool) (
			expectedQuery string, expectedAffectedRows *int64, expectedError error,
		) {
			expectedError = errors.New("some error")
			expectedQuery = selectQueryForTesting
			expectAnalyzeForQuery(mock, expectedQuery, expectedError, withSQLAnalyze)
			mock.ExpectQuery("^" + regexp.QuoteMeta(expectedQuery) + "$").WillReturnError(expectedError)
			conn, err := db.db.CommonDB().(*sqlDBWrapper).conn(context.Background())
			require.NoError(t, err)
			defer func() { _ = conn.close(nil) }()
			row := conn.QueryRowContext(context.Background(), expectedQuery)
			assert.Equal(t, expectedError, row.Err())
			return expectedQuery, expectedAffectedRows, expectedError
		},
	},
	{
		name: "sqlConnWrapper.QueryRowContext with success",
		funcToRun: func(t *testing.T, db *DB, mock sqlmock.Sqlmock, withSQLAnalyze bool) (
			expectedQuery string, expectedAffectedRows *int64, expectedError error,
		) {
			expectedQuery = selectQueryForTesting
			expectAnalyzeForQuery(mock, expectedQuery, expectedError, withSQLAnalyze)
			mock.ExpectQuery("^" + regexp.QuoteMeta(expectedQuery) + "$").
				WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1).AddRow(2))
			conn, err := db.db.CommonDB().(*sqlDBWrapper).conn(context.Background())
			require.NoError(t, err)
			defer func() { _ = conn.close(nil) }()
			row := conn.QueryRowContext(context.Background(), expectedQuery)
			assert.NoError(t, row.Err())
			_ = row.Scan()
			return expectedQuery, expectedAffectedRows, expectedError
		},
	},
	{
		name: "SQLStmtWrapper.ExecContext with error",
		funcToRun: func(t *testing.T, db *DB, mock sqlmock.Sqlmock, withSQLAnalyze bool) (
			expectedQuery string, expectedAffectedRows *int64, expectedError error,
		) {
			expectedError = errors.New("some error")
			expectedQuery = updateQueryForTesting

			mock.ExpectBegin()
			prepareExpectation := mock.ExpectPrepare("^" + regexp.QuoteMeta(expectedQuery) + "$")
			expectAnalyzeForQuery(mock, expectedQuery, expectedError, withSQLAnalyze)
			prepareExpectation.ExpectExec().WillReturnError(expectedError)
			mock.ExpectCommit()

			tx, err := db.db.CommonDB().(*sqlDBWrapper).sqlDB.BeginTx(context.Background(), &sql.TxOptions{})
			require.NoError(t, err)
			defer func() { _ = tx.Commit() }()
			txWrapper := &sqlTxWrapper{sqlTx: tx, ctx: context.Background(), logConfig: db.logConfig}
			stmtWrapper, err := txWrapper.prepare(expectedQuery)
			require.NoError(t, err)
			defer func() { _ = stmtWrapper.Close() }()
			result, err := stmtWrapper.ExecContext(context.Background())
			assert.Equal(t, expectedError, err)
			assert.Nil(t, result)
			return expectedQuery, expectedAffectedRows, expectedError
		},
	},
	{
		name: "SQLStmtWrapper.ExecContext with success",
		funcToRun: func(t *testing.T, db *DB, mock sqlmock.Sqlmock, withSQLAnalyze bool) (
			expectedQuery string, expectedAffectedRows *int64, expectedError error,
		) {
			expectedQuery = updateQueryForTesting
			expectedAffectedRows = golang.Ptr(int64(1))

			mock.ExpectBegin()
			prepareExpectation := mock.ExpectPrepare("^" + regexp.QuoteMeta(expectedQuery) + "$")
			expectAnalyzeForQuery(mock, expectedQuery, expectedError, withSQLAnalyze)
			prepareExpectation.ExpectExec().WillReturnResult(sqlmock.NewResult(-1, *expectedAffectedRows))
			mock.ExpectCommit()

			tx, err := db.db.CommonDB().(*sqlDBWrapper).sqlDB.BeginTx(context.Background(), &sql.TxOptions{})
			require.NoError(t, err)
			defer func() { _ = tx.Commit() }()
			txWrapper := &sqlTxWrapper{sqlTx: tx, ctx: context.Background(), logConfig: db.logConfig}
			stmtWrapper, err := txWrapper.prepare(expectedQuery)
			require.NoError(t, err)
			defer func() { _ = stmtWrapper.Close() }()
			result, err := stmtWrapper.ExecContext(context.Background())
			assert.NoError(t, err)
			assert.NotNil(t, result)
			return expectedQuery, expectedAffectedRows, expectedError
		},
	},
	{
		name: "SQLStmtWrapper.QueryContext with error",
		funcToRun: func(t *testing.T, db *DB, mock sqlmock.Sqlmock, withSQLAnalyze bool) (
			expectedQuery string, expectedAffectedRows *int64, expectedError error,
		) {
			expectedError = errors.New("some error")
			expectedQuery = updateQueryForTesting

			mock.ExpectBegin()
			prepareExpectation := mock.ExpectPrepare("^" + regexp.QuoteMeta(expectedQuery) + "$")
			expectAnalyzeForQuery(mock, expectedQuery, expectedError, withSQLAnalyze)
			prepareExpectation.ExpectQuery().WillReturnError(expectedError)
			mock.ExpectCommit()

			tx, err := db.db.CommonDB().(*sqlDBWrapper).sqlDB.BeginTx(context.Background(), &sql.TxOptions{})
			require.NoError(t, err)
			defer func() { _ = tx.Commit() }()
			txWrapper := &sqlTxWrapper{sqlTx: tx, ctx: context.Background(), logConfig: db.logConfig}
			stmtWrapper, err := txWrapper.prepare(expectedQuery)
			require.NoError(t, err)
			defer func() { _ = stmtWrapper.Close() }()
			rows, err := stmtWrapper.QueryContext(context.Background())
			assert.Equal(t, expectedError, err)
			assert.Nil(t, rows)
			return expectedQuery, expectedAffectedRows, expectedError
		},
	},
	{
		name: "SQLStmtWrapper.QueryContext with success",
		funcToRun: func(t *testing.T, db *DB, mock sqlmock.Sqlmock, withSQLAnalyze bool) (
			expectedQuery string, expectedAffectedRows *int64, expectedError error,
		) {
			expectedQuery = updateQueryForTesting

			mock.ExpectBegin()
			prepareExpectation := mock.ExpectPrepare("^" + regexp.QuoteMeta(expectedQuery) + "$")
			expectAnalyzeForQuery(mock, expectedQuery, expectedError, withSQLAnalyze)
			prepareExpectation.ExpectQuery().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1).AddRow(2))
			mock.ExpectCommit()

			tx, err := db.db.CommonDB().(*sqlDBWrapper).sqlDB.BeginTx(context.Background(), &sql.TxOptions{})
			require.NoError(t, err)
			defer func() { _ = tx.Commit() }()
			txWrapper := &sqlTxWrapper{sqlTx: tx, ctx: context.Background(), logConfig: db.logConfig}
			stmtWrapper, err := txWrapper.prepare(expectedQuery)
			require.NoError(t, err)
			defer func() { _ = stmtWrapper.Close() }()
			rows, err := stmtWrapper.QueryContext(context.Background())
			require.NoError(t, err)
			if rows != nil {
				_ = rows.Close()
			}
			return expectedQuery, expectedAffectedRows, expectedError
		},
	},
	{
		name: "SQLStmtWrapper.QueryRowContext with error",
		funcToRun: func(t *testing.T, db *DB, mock sqlmock.Sqlmock, withSQLAnalyze bool) (
			expectedQuery string, expectedAffectedRows *int64, expectedError error,
		) {
			expectedError = errors.New("some error")
			expectedQuery = updateQueryForTesting

			mock.ExpectBegin()
			prepareExpectation := mock.ExpectPrepare("^" + regexp.QuoteMeta(expectedQuery) + "$")
			expectAnalyzeForQuery(mock, expectedQuery, expectedError, withSQLAnalyze)
			prepareExpectation.ExpectQuery().WillReturnError(expectedError)
			mock.ExpectCommit()

			tx, err := db.db.CommonDB().(*sqlDBWrapper).sqlDB.BeginTx(context.Background(), &sql.TxOptions{})
			require.NoError(t, err)
			defer func() { _ = tx.Commit() }()
			txWrapper := &sqlTxWrapper{sqlTx: tx, ctx: context.Background(), logConfig: db.logConfig}
			stmtWrapper, err := txWrapper.prepare(expectedQuery)
			require.NoError(t, err)
			defer func() { _ = stmtWrapper.Close() }()
			row := stmtWrapper.QueryRowContext(context.Background())
			assert.Equal(t, expectedError, row.Err())
			return expectedQuery, expectedAffectedRows, expectedError
		},
	},
	{
		name: "SQLStmtWrapper.QueryRowContext with success",
		funcToRun: func(t *testing.T, db *DB, mock sqlmock.Sqlmock, withSQLAnalyze bool) (
			expectedQuery string, expectedAffectedRows *int64, expectedError error,
		) {
			expectedQuery = updateQueryForTesting

			mock.ExpectBegin()
			prepareExpectation := mock.ExpectPrepare("^" + regexp.QuoteMeta(expectedQuery) + "$")
			expectAnalyzeForQuery(mock, expectedQuery, expectedError, withSQLAnalyze)
			prepareExpectation.ExpectQuery().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1).AddRow(2))
			mock.ExpectCommit()

			tx, err := db.db.CommonDB().(*sqlDBWrapper).sqlDB.BeginTx(context.Background(), &sql.TxOptions{})
			require.NoError(t, err)
			defer func() { _ = tx.Commit() }()
			txWrapper := &sqlTxWrapper{sqlTx: tx, ctx: context.Background(), logConfig: db.logConfig}
			stmtWrapper, err := txWrapper.prepare(expectedQuery)
			require.NoError(t, err)
			defer func() { _ = stmtWrapper.Close() }()
			row := stmtWrapper.QueryRowContext(context.Background())
			require.NoError(t, row.Err())
			_ = row.Scan()
			return expectedQuery, expectedAffectedRows, expectedError
		},
	},
	{
		name: "sqlTxWrapper.Exec with error",
		funcToRun: func(t *testing.T, db *DB, mock sqlmock.Sqlmock, withSQLAnalyze bool) (
			expectedQuery string, expectedAffectedRows *int64, expectedError error,
		) {
			expectedError = errors.New("some error")
			expectedQuery = updateQueryForTesting

			mock.ExpectBegin()
			expectAnalyzeForQuery(mock, expectedQuery, expectedError, withSQLAnalyze)
			mock.ExpectExec("^" + regexp.QuoteMeta(expectedQuery) + "$").WillReturnError(expectedError)
			mock.ExpectCommit()

			tx, err := db.db.CommonDB().(*sqlDBWrapper).sqlDB.BeginTx(context.Background(), &sql.TxOptions{})
			require.NoError(t, err)
			defer func() { _ = tx.Commit() }()
			txWrapper := &sqlTxWrapper{sqlTx: tx, ctx: context.Background(), logConfig: db.logConfig}
			result, err := txWrapper.Exec(expectedQuery)
			assert.Equal(t, expectedError, err)
			assert.Nil(t, result)
			return expectedQuery, expectedAffectedRows, expectedError
		},
	},
	{
		name: "sqlTxWrapper.Exec with success",
		funcToRun: func(t *testing.T, db *DB, mock sqlmock.Sqlmock, withSQLAnalyze bool) (
			expectedQuery string, expectedAffectedRows *int64, expectedError error,
		) {
			expectedQuery = updateQueryForTesting
			expectedAffectedRows = golang.Ptr(int64(1))

			mock.ExpectBegin()
			expectAnalyzeForQuery(mock, expectedQuery, expectedError, withSQLAnalyze)
			mock.ExpectExec("^" + regexp.QuoteMeta(expectedQuery) + "$").
				WillReturnResult(sqlmock.NewResult(-1, *expectedAffectedRows))
			mock.ExpectCommit()

			tx, err := db.db.CommonDB().(*sqlDBWrapper).sqlDB.BeginTx(context.Background(), &sql.TxOptions{})
			require.NoError(t, err)
			defer func() { _ = tx.Commit() }()
			txWrapper := &sqlTxWrapper{sqlTx: tx, ctx: context.Background(), logConfig: db.logConfig}
			result, err := txWrapper.Exec(expectedQuery)
			assert.NoError(t, err)
			assert.NotNil(t, result)
			return expectedQuery, expectedAffectedRows, expectedError
		},
	},
	{
		name: "sqlTxWrapper.Query with error",
		funcToRun: func(t *testing.T, db *DB, mock sqlmock.Sqlmock, withSQLAnalyze bool) (
			expectedQuery string, expectedAffectedRows *int64, expectedError error,
		) {
			expectedError = errors.New("some error")
			expectedQuery = updateQueryForTesting

			mock.ExpectBegin()
			expectAnalyzeForQuery(mock, expectedQuery, expectedError, withSQLAnalyze)
			mock.ExpectQuery("^" + regexp.QuoteMeta(expectedQuery) + "$").WillReturnError(expectedError)
			mock.ExpectCommit()

			tx, err := db.db.CommonDB().(*sqlDBWrapper).sqlDB.BeginTx(context.Background(), &sql.TxOptions{})
			require.NoError(t, err)
			defer func() { _ = tx.Commit() }()
			txWrapper := &sqlTxWrapper{sqlTx: tx, ctx: context.Background(), logConfig: db.logConfig}
			rows, err := txWrapper.Query(expectedQuery)
			assert.Equal(t, expectedError, err)
			assert.Nil(t, rows)
			return expectedQuery, expectedAffectedRows, expectedError
		},
	},
	{
		name: "sqlTxWrapper.Query with success",
		funcToRun: func(t *testing.T, db *DB, mock sqlmock.Sqlmock, withSQLAnalyze bool) (
			expectedQuery string, expectedAffectedRows *int64, expectedError error,
		) {
			expectedQuery = updateQueryForTesting

			mock.ExpectBegin()
			expectAnalyzeForQuery(mock, expectedQuery, expectedError, withSQLAnalyze)
			mock.ExpectQuery("^" + regexp.QuoteMeta(expectedQuery) + "$").
				WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1).AddRow(2))
			mock.ExpectCommit()

			tx, err := db.db.CommonDB().(*sqlDBWrapper).sqlDB.BeginTx(context.Background(), &sql.TxOptions{})
			require.NoError(t, err)
			defer func() { _ = tx.Commit() }()
			txWrapper := &sqlTxWrapper{sqlTx: tx, ctx: context.Background(), logConfig: db.logConfig}
			rows, err := txWrapper.Query(expectedQuery)
			require.NoError(t, err)
			if rows != nil {
				_ = rows.Close()
			}
			return expectedQuery, expectedAffectedRows, expectedError
		},
	},
	{
		name: "sqlTxWrapper.QueryRow with error",
		funcToRun: func(t *testing.T, db *DB, mock sqlmock.Sqlmock, withSQLAnalyze bool) (
			expectedQuery string, expectedAffectedRows *int64, expectedError error,
		) {
			expectedError = errors.New("some error")
			expectedQuery = updateQueryForTesting

			mock.ExpectBegin()
			expectAnalyzeForQuery(mock, expectedQuery, expectedError, withSQLAnalyze)
			mock.ExpectQuery("^" + regexp.QuoteMeta(expectedQuery) + "$").WillReturnError(expectedError)
			mock.ExpectCommit()

			tx, err := db.db.CommonDB().(*sqlDBWrapper).sqlDB.BeginTx(context.Background(), &sql.TxOptions{})
			require.NoError(t, err)
			defer func() { _ = tx.Commit() }()
			txWrapper := &sqlTxWrapper{sqlTx: tx, ctx: context.Background(), logConfig: db.logConfig}
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
			expectedQuery = updateQueryForTesting

			mock.ExpectBegin()
			expectAnalyzeForQuery(mock, expectedQuery, expectedError, withSQLAnalyze)
			mock.ExpectQuery("^" + regexp.QuoteMeta(expectedQuery) + "$").
				WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1).AddRow(2))
			mock.ExpectCommit()

			tx, err := db.db.CommonDB().(*sqlDBWrapper).sqlDB.BeginTx(context.Background(), &sql.TxOptions{})
			require.NoError(t, err)
			defer func() { _ = tx.Commit() }()
			txWrapper := &sqlTxWrapper{sqlTx: tx, ctx: context.Background(), logConfig: db.logConfig}
			row := txWrapper.QueryRow(expectedQuery)
			require.NoError(t, row.Err())
			_ = row.Scan()
			return expectedQuery, expectedAffectedRows, expectedError
		},
	},
	{
		name: "sqlTxWrapper.Commit with error",
		funcToRun: func(t *testing.T, db *DB, mock sqlmock.Sqlmock, _ bool) (
			expectedQuery string, expectedAffectedRows *int64, expectedError error,
		) {
			expectedError = errors.New("some error")
			expectedQuery = commitTransactionLogMessage

			mock.ExpectBegin()
			mock.ExpectCommit().WillReturnError(expectedError)

			tx, err := db.db.CommonDB().(*sqlDBWrapper).sqlDB.BeginTx(context.Background(), &sql.TxOptions{})
			require.NoError(t, err)
			txWrapper := &sqlTxWrapper{sqlTx: tx, ctx: context.Background(), logConfig: db.logConfig}
			err = txWrapper.Commit()
			assert.Equal(t, expectedError, err)
			return expectedQuery, expectedAffectedRows, expectedError
		},
		skipAnalyzeSQLQueriesTesting: true,
	},
	{
		name: "sqlTxWrapper.Commit with success",
		funcToRun: func(t *testing.T, db *DB, mock sqlmock.Sqlmock, _ bool) (
			expectedQuery string, expectedAffectedRows *int64, expectedError error,
		) {
			expectedQuery = commitTransactionLogMessage

			mock.ExpectBegin()
			mock.ExpectCommit()

			tx, err := db.db.CommonDB().(*sqlDBWrapper).sqlDB.BeginTx(context.Background(), &sql.TxOptions{})
			require.NoError(t, err)
			txWrapper := &sqlTxWrapper{sqlTx: tx, ctx: context.Background(), logConfig: db.logConfig}
			err = txWrapper.Commit()
			require.NoError(t, err)
			return expectedQuery, expectedAffectedRows, expectedError
		},
		skipAnalyzeSQLQueriesTesting: true,
	},
	{
		name: "sqlTxWrapper.Rollback with error",
		funcToRun: func(t *testing.T, db *DB, mock sqlmock.Sqlmock, _ bool) (
			expectedQuery string, expectedAffectedRows *int64, expectedError error,
		) {
			expectedError = errors.New("some error")
			expectedQuery = rollbackTransactionLogMessage

			mock.ExpectBegin()
			mock.ExpectRollback().WillReturnError(expectedError)

			tx, err := db.db.CommonDB().(*sqlDBWrapper).sqlDB.BeginTx(context.Background(), &sql.TxOptions{})
			require.NoError(t, err)
			txWrapper := &sqlTxWrapper{sqlTx: tx, ctx: context.Background(), logConfig: db.logConfig}
			err = txWrapper.Rollback()
			assert.Equal(t, expectedError, err)
			return expectedQuery, expectedAffectedRows, expectedError
		},
		skipAnalyzeSQLQueriesTesting: true,
	},
	{
		name: "sqlTxWrapper.Rollback with success",
		funcToRun: func(t *testing.T, db *DB, mock sqlmock.Sqlmock, _ bool) (
			expectedQuery string, expectedAffectedRows *int64, expectedError error,
		) {
			expectedQuery = rollbackTransactionLogMessage

			mock.ExpectBegin()
			mock.ExpectRollback()

			tx, err := db.db.CommonDB().(*sqlDBWrapper).sqlDB.BeginTx(context.Background(), &sql.TxOptions{})
			require.NoError(t, err)
			txWrapper := &sqlTxWrapper{sqlTx: tx, ctx: context.Background(), logConfig: db.logConfig}
			err = txWrapper.Rollback()
			require.NoError(t, err)
			return expectedQuery, expectedAffectedRows, expectedError
		},
		skipAnalyzeSQLQueriesTesting: true,
	},
	{
		name: "sqlTxWrapper.Rollback error because of a context cancellation",
		funcToRun: func(t *testing.T, db *DB, mock sqlmock.Sqlmock, _ bool) (
			expectedQuery string, expectedAffectedRows *int64, expectedError error,
		) {
			expectedError = context.Canceled
			expectedQuery = rollbackTransactionLogMessage

			mock.ExpectBegin()
			mock.ExpectRollback()

			ctx, cancelFunc := context.WithCancel(context.Background())
			sqlDBWrapper := db.db.CommonDB().(*sqlDBWrapper)
			oldLogSQLQueriesValue := sqlDBWrapper.logConfig.LogSQLQueries
			sqlDBWrapper.logConfig.LogSQLQueries = false
			tx, err := sqlDBWrapper.BeginTx(ctx, &sql.TxOptions{})
			require.NoError(t, err)
			sqlDBWrapper.logConfig.LogSQLQueries = oldLogSQLQueriesValue
			cancelFunc()

			assert.Eventually(t, func() bool {
				return mock.ExpectationsWereMet() == nil
			}, 1*time.Second, 10*time.Millisecond)
			require.Nil(t, mock.ExpectationsWereMet(), "cancelFunc() should have caused the rollback to be called")
			err = tx.Rollback()
			require.Equal(t, expectedError, err)
			return expectedQuery, expectedAffectedRows, errors.New(expectedError.Error() + " (the transaction has been rolled back implicitly)")
		},
		skipAnalyzeSQLQueriesTesting: true,
	},
	{
		name: "sqlTxWrapper.Commit error because of a context cancellation",
		funcToRun: func(t *testing.T, db *DB, mock sqlmock.Sqlmock, _ bool) (
			expectedQuery string, expectedAffectedRows *int64, expectedError error,
		) {
			expectedError = context.Canceled
			expectedQuery = commitTransactionLogMessage

			mock.ExpectBegin()
			mock.ExpectRollback()

			ctx, cancelFunc := context.WithCancel(context.Background())
			sqlDBWrapper := db.db.CommonDB().(*sqlDBWrapper)
			oldLogSQLQueriesValue := sqlDBWrapper.logConfig.LogSQLQueries
			sqlDBWrapper.logConfig.LogSQLQueries = false
			tx, err := sqlDBWrapper.BeginTx(ctx, &sql.TxOptions{})
			require.NoError(t, err)
			sqlDBWrapper.logConfig.LogSQLQueries = oldLogSQLQueriesValue
			cancelFunc()

			assert.Eventually(t, func() bool {
				return mock.ExpectationsWereMet() == nil
			}, 1*time.Second, 10*time.Millisecond)
			require.Nil(t, mock.ExpectationsWereMet(), "cancelFunc() should have caused the rollback to be called")
			err = tx.Commit()
			assert.Equal(t, expectedError, err)
			return expectedQuery, expectedAffectedRows, errors.New(expectedError.Error() + " (the transaction has been rolled back implicitly)")
		},
		skipAnalyzeSQLQueriesTesting: true,
	},
}

func Test_SQLQueryLogging(t *testing.T) {
	for _, test := range sqlQueryLoggingTests {
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
	loggerHook, loggerRestoreFunc := logging.MockSharedLoggerHook()
	defer loggerRestoreFunc()

	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	db.logConfig.LogSQLQueries = logSQLQueries
	db.logConfig.AnalyzeSQLQueries = analyzeSQLQueries
	db.logConfig.Logger = logging.NewStructuredDBLogger()

	expectedQuery, expectedAffectedRows, expectedError := test.funcToRun(t, db, mock, analyzeSQLQueries)

	assert.NoError(t, mock.ExpectationsWereMet())

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
	require.Len(t, logEntries, expectedLen)
	index := 0
	if logSQLQueries {
		assert.Equal(t, "info", logEntries[0].Level.String())
		assert.Equal(t, expectedQuery, logEntries[0].Message)
		assert.Equal(t, "db", logEntries[0].Data["type"])
		assert.Greater(t, logEntries[0].Data["duration"], 0.0)
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
			assert.Equal(t, "query execution plan: plan", logEntries[index].Message)
			assert.Greater(t, logEntries[index].Data["duration"], 0.0)
		}
		assert.Equal(t, "db", logEntries[index].Data["type"])
	}
}
