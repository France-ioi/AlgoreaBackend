package database

import (
	"context"
	"database/sql"
	"database/sql/driver"

	"github.com/jinzhu/gorm"

	log "github.com/France-ioi/AlgoreaBackend/v2/app/logging"
)

type sqlTxWrapper struct {
	sqlTx     *sql.Tx
	ctx       context.Context
	logConfig *LogConfig
}

// Exec executes a query that doesn't return rows.
// For example: an INSERT and UPDATE.
//
// Exec uses the context of [sqlTxWrapper] internally.
func (sqlTX *sqlTxWrapper) Exec(query string, args ...interface{}) (result sql.Result, err error) {
	defer getSQLExecutionPlanLoggingFunc(sqlTX.ctx, sqlTX, sqlTX.logConfig, query, args...)()
	defer getSQLQueryLoggingFunc(sqlTX.ctx, func() *int64 {
		affectedRows, _ := result.RowsAffected()
		return &affectedRows
	}, &err, gorm.NowFunc(), query, args...)(sqlTX.logConfig)

	return sqlTX.sqlTx.ExecContext(sqlTX.ctx, query, args...)
}

// Prepare is not implemented intentionally and panics if called. Instead, use [prepare].
//
// Note: Gorm does not use this method, but it is required to implement the [gorm.SQLCommon] interface.
func (sqlTX *sqlTxWrapper) Prepare(_ string) (*sql.Stmt, error) {
	panic("sqlTxWrapper.Prepare is not implemented, should not be called")
}

// prepare creates a prepared statement for use within a transaction.
//
// The returned statement operates within the transaction and will be closed
// when the transaction has been committed or rolled back.
//
// To use an existing prepared statement on this transaction, see [sql.Tx.Stmt].
//
// prepare uses the context of [sqlTxWrapper] internally.
func (sqlTX *sqlTxWrapper) prepare(query string) (*SQLStmtWrapper, error) {
	stmt, err := sqlTX.sqlTx.PrepareContext(sqlTX.ctx, query)
	if err != nil {
		logDBError(sqlTX.ctx, sqlTX.logConfig, err)
		return nil, err
	}
	return &SQLStmtWrapper{db: sqlTX, sql: query, stmt: stmt, logConfig: sqlTX.logConfig}, nil
}

// Query executes a query that returns rows, typically a SELECT.
//
// Query uses the context of [sqlTxWrapper] internally.
func (sqlTX *sqlTxWrapper) Query(query string, args ...interface{}) (_ *sql.Rows, err error) {
	defer getSQLExecutionPlanLoggingFunc(sqlTX.ctx, sqlTX, sqlTX.logConfig, query, args...)()
	defer getSQLQueryLoggingFunc(sqlTX.ctx, nil, &err, gorm.NowFunc(), query, args...)(sqlTX.logConfig)

	return sqlTX.sqlTx.QueryContext(sqlTX.ctx, query, args...)
}

// QueryRow executes a query that is expected to return at most one row.
// QueryRow always returns a non-nil value. Errors are deferred until
// Row's Scan method is called.
// If the query selects no rows, the *Row's Scan will return ErrNoRows.
// Otherwise, the *Row's Scan scans the first selected row and discards
// the rest.
//
// QueryRow uses the context of [sqlTxWrapper] internally.
func (sqlTX *sqlTxWrapper) QueryRow(query string, args ...interface{}) (row *sql.Row) {
	defer getSQLExecutionPlanLoggingFunc(sqlTX.ctx, sqlTX, sqlTX.logConfig, query, args...)()
	startTime := gorm.NowFunc()
	defer func() {
		err := row.Err()
		getSQLQueryLoggingFunc(sqlTX.ctx, nil, &err, startTime, query, args...)(sqlTX.logConfig)
	}()

	return sqlTX.sqlTx.QueryRowContext(sqlTX.ctx, query, args...)
}

var _ gorm.SQLCommon = &sqlTxWrapper{}

const commitTransactionLogMessage = "**COMMIT TRANSACTION**"

// Commit commits the transaction.
func (sqlTX *sqlTxWrapper) Commit() (err error) {
	startTime := gorm.NowFunc()
	err = sqlTX.sqlTx.Commit()
	if sqlTX.logConfig.LogSQLQueries {
		logSQLQuery(sqlTX.ctx, gorm.NowFunc().Sub(startTime), commitTransactionLogMessage, nil, nil)
	}
	return sqlTX.handleCommitOrRollbackError(err)
}

const rollbackTransactionLogMessage = "**ROLLBACK TRANSACTION**"

// Rollback aborts the transaction.
func (sqlTX *sqlTxWrapper) Rollback() (err error) {
	startTime := gorm.NowFunc()
	err = sqlTX.sqlTx.Rollback()
	if sqlTX.logConfig.LogSQLQueries {
		logSQLQuery(sqlTX.ctx, gorm.NowFunc().Sub(startTime), rollbackTransactionLogMessage, nil, nil)
	}
	return sqlTX.handleCommitOrRollbackError(err)
}

func (sqlTX *sqlTxWrapper) handleCommitOrRollbackError(err error) error {
	var errorWasPatched bool
	if err != nil && sqlTX.ctx.Err() != nil { // ignore the returned error if the context has been canceled before
		err = sqlTX.ctx.Err() // return the context error instead as it is the root cause
		errorWasPatched = true
	}
	if err != nil {
		errString := err.Error()
		if errorWasPatched {
			errString += " (the transaction has been rolled back implicitly)"
		}
		log.SharedLogger.WithContext(sqlTX.ctx).WithFields(
			map[string]interface{}{"type": "db", "fileline": fileWithLineNum()}).Error(errString)
	}
	return err
}

var (
	_ gorm.SQLCommon = &sqlTxWrapper{}
	_ driver.Tx      = &sqlTxWrapper{}
)

func (sqlTX *sqlTxWrapper) queryRowWithoutLogging(query string, args ...interface{}) *sql.Row {
	return sqlTX.sqlTx.QueryRowContext(context.Background(), query, args...)
}

var _ queryRowWithoutLogging = &sqlTxWrapper{}

func (sqlTX *sqlTxWrapper) withContext(ctx context.Context) gorm.SQLCommon {
	return &sqlTxWrapper{sqlTx: sqlTX.sqlTx, ctx: ctx, logConfig: sqlTX.logConfig}
}

var _ withContexter = &sqlTxWrapper{}
