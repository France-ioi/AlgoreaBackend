package database

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"unsafe"

	"github.com/jinzhu/gorm"

	log "github.com/France-ioi/AlgoreaBackend/v2/app/logging"
)

type sqlTxWrapper struct {
	sqlTx *sql.Tx
	//nolint:containedctx // We store the context here because Gorm v1 does not support contexts
	//                    // as arguments for Exec, Query, and QueryRow methods.
	//                    //
	//                    // Actually, sql.Tx already has a context inside it, and the transaction gets rolled back
	//                    // when the context of sql.Tx is done (canceled or timed out). And it is the only purpose
	//                    // of the context inside sql.Tx.
	//                    //
	//                    // Our context is a copy of the sql.Tx context. We need this copy to be able to
	//                    // add some fields not related to the deadline in the context, but sql.Tx
	//                    // does not allow to change the context after it has been created.
	//                    //
	//                    // When a new sqlTxWrapper is created with different context for the same sql.Tx,
	//                    // (e.g., in withContext), the sqlTxWrapper.ctx is set to the new context,
	//                    // but the sqlTx.sqlTx still has the original context.
	//                    //
	//                    // At the same time, the sqlTxWrapper.ctx is used by Exec, Query, and QueryRow methods
	//                    // to provide deadline for each query execution.
	//                    //
	//                    // As we normally do not want to have different deadlines for the transaction and the queries,
	//                    // it doesn't seem to be a good idea to set a new deadline in the sqlTxWrapper.ctx.
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

	result, err = sqlTX.sqlTx.ExecContext(sqlTX.ctx, query, args...)
	err = sqlTX.handleError(err)
	if err != nil {
		return nil, err
	}

	return result, nil
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
	err = sqlTX.handleError(err)
	if err != nil {
		logDBError(sqlTX.ctx, sqlTX.logConfig, err)
		return nil, err
	}
	return &SQLStmtWrapper{db: sqlTX, sql: query, stmt: stmt, logConfig: sqlTX.logConfig}, nil
}

// Query executes a query that returns rows, typically a SELECT.
//
// Query uses the context of [sqlTxWrapper] internally.
func (sqlTX *sqlTxWrapper) Query(query string, args ...interface{}) (rows *sql.Rows, err error) {
	defer getSQLExecutionPlanLoggingFunc(sqlTX.ctx, sqlTX, sqlTX.logConfig, query, args...)()
	defer getSQLQueryLoggingFunc(sqlTX.ctx, nil, &err, gorm.NowFunc(), query, args...)(sqlTX.logConfig)

	//nolint:sqlclosecheck // The caller is responsible for closing the returned *sql.Rows.
	rows, err = sqlTX.sqlTx.QueryContext(sqlTX.ctx, query, args...)
	err = sqlTX.handleError(err)
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
// QueryRow uses the context of [sqlTxWrapper] internally.
func (sqlTX *sqlTxWrapper) QueryRow(query string, args ...interface{}) (row *sql.Row) {
	defer getSQLExecutionPlanLoggingFunc(sqlTX.ctx, sqlTX, sqlTX.logConfig, query, args...)()
	startTime := gorm.NowFunc()
	defer func() {
		err := sqlTX.handleError(row.Err())
		if err != nil {
			row = (*sql.Row)(unsafe.Pointer(&sqlRowAccessor{err: err})) //nolint:gosec // G103: patch the error
		}
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

func (sqlTX *sqlTxWrapper) handleError(err error) error {
	if err != nil && sqlTX.ctx.Err() != nil { // ignore the returned error if the context has been canceled before
		err = sqlTX.ctx.Err() // return the context error instead as it is the root cause
	}
	return err
}

func (sqlTX *sqlTxWrapper) handleCommitOrRollbackError(err error) error {
	newErr := sqlTX.handleError(err)
	errorWasPatched := newErr != err //nolint:errorlint // here we want to check if the error was patched by handleError
	err = newErr
	if err != nil {
		errString := err.Error()
		if errorWasPatched {
			errString += " (the transaction has been rolled back implicitly)"
		}
		log.EntryFromContext(sqlTX.ctx).WithFields(
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

func (sqlTX *sqlTxWrapper) getContext() context.Context {
	return sqlTX.ctx
}

var _ contextGetter = &sqlTxWrapper{}

func (sqlTX *sqlTxWrapper) getLogConfig() *LogConfig {
	return sqlTX.logConfig
}

var _ logConfigGetter = &sqlTxWrapper{}
