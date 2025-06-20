package database

import (
	"context"
	"database/sql"

	"github.com/jinzhu/gorm"
)

type sqlDBWrapper struct {
	sqlDB *sql.DB
	//nolint:containedctx // We store the context here because Gorm v1 does not support contexts
	//                    // as arguments for Exec, Query, and QueryRow methods.
	ctx       context.Context
	logConfig *LogConfig
}

// Exec executes a query without returning any rows.
// The args are for any placeholder parameters in the query.
//
// Exec uses the context of [sqlDBWrapper] internally.
func (sqlDB *sqlDBWrapper) Exec(query string, args ...interface{}) (result sql.Result, err error) {
	defer getSQLExecutionPlanLoggingFunc(sqlDB.ctx, sqlDB, sqlDB.logConfig, query, args...)()
	defer getSQLQueryLoggingFunc(sqlDB.ctx, func() *int64 {
		rowsAffected, _ := result.RowsAffected()
		return &rowsAffected
	}, &err, gorm.NowFunc(), query, args...)(sqlDB.logConfig)

	return sqlDB.sqlDB.ExecContext(sqlDB.ctx, query, args...)
}

// Prepare is not implemented intentionally and panics if called.
//
// Note: Gorm does not use this method, but it is required to implement the [gorm.SQLCommon] interface.
func (sqlDB *sqlDBWrapper) Prepare(_ string) (*sql.Stmt, error) {
	panic("sqlDBWrapper.Prepare is not implemented, should not be called")
}

// Query executes a query that returns rows, typically a SELECT.
// The args are for any placeholder parameters in the query.
//
// Query uses the context of [sqlDBWrapper] internally.
func (sqlDB *sqlDBWrapper) Query(query string, args ...interface{}) (_ *sql.Rows, err error) {
	defer getSQLExecutionPlanLoggingFunc(sqlDB.ctx, sqlDB, sqlDB.logConfig, query, args...)()
	defer getSQLQueryLoggingFunc(sqlDB.ctx, nil, &err, gorm.NowFunc(), query, args...)(sqlDB.logConfig)

	return sqlDB.sqlDB.QueryContext(sqlDB.ctx, query, args...)
}

// QueryRow executes a query that is expected to return at most one row.
// QueryRow always returns a non-nil value. Errors are deferred until
// [sql.Row]'s Scan method is called.
// If the query selects no rows, the [*Row.Scan] will return [sql.ErrNoRows].
// Otherwise, [*sql.Row.Scan] scans the first selected row and discards
// the rest.
//
// QueryRow uses the context of [sqlDBWrapper] internally.
func (sqlDB *sqlDBWrapper) QueryRow(query string, args ...interface{}) (row *sql.Row) {
	defer getSQLExecutionPlanLoggingFunc(sqlDB.ctx, sqlDB, sqlDB.logConfig, query, args...)()
	startTime := gorm.NowFunc()
	defer func() {
		err := row.Err()
		getSQLQueryLoggingFunc(sqlDB.ctx, nil, &err, startTime, query, args...)(sqlDB.logConfig)
	}()

	return sqlDB.sqlDB.QueryRowContext(sqlDB.ctx, query, args...)
}

var _ gorm.SQLCommon = &sqlDBWrapper{}

/*
type sqlDb interface {
	Begin() (*sql.Tx, error)
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
}
*/
// Begin is not implemented intentionally to make sqlDBWrapper incompatible with gorm.sqlDb interface.
//
// Note: Gorm calls this method from 'delete' and 'update' callbacks if the db implements 'sqlDb' interface.
// Otherwise, is silently skips the transaction start.
// The problem with this method is that it returns a raw sql.Tx which does not have
// the context of the sqlDBWrapper and logging.
// Happily, we do not need 'delete' and 'update' callbacks to be transactional as we
// never use Gorm models and never expose Gorm methods working with models. This means
// we do not have associations or timestamp fields that Gorm would update automatically.
// Actually, the only method triggering such callbacks is 'Delete' from the 'gorm.DB' struct
// and, in our cases, it only executes one DELETE statement, so no transaction is needed.
/*
func (sqlDB *sqlDBWrapper) Begin() (*sql.Tx, error) {
	panic("Begin is not implemented, should not be called")
}
*/

const beginTransactionLogMessage = "**BEGIN TRANSACTION**"

// BeginTx starts a transaction.
//
// The provided context is used until the transaction is committed or rolled back.
// If the context is canceled, the sql package will roll back
// the transaction. [sql.Tx.Commit] will return an error if the context provided to
// BeginTx is canceled.
//
// The provided [sql.TxOptions] is optional and may be nil if defaults should be used.
// If a non-default isolation level is used that the driver doesn't support,
// an error will be returned.
//
// Note: Gorm uses this method only in [gorm.DB.Transaction] method and only if the db implements
// the 'sqlDb' interface. We intentionally do not implement this interface to avoid Gorm from
// calling this method. We never use [gorm.DB.Transaction] method as we do not expose it.
// So, we intentionally implement this method with a different return value to make sqlDBWrapper
// incompatible with the 'sqlDb' interface.
func (sqlDB *sqlDBWrapper) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sqlTxWrapper, error) {
	startTime := gorm.NowFunc()
	tx, err := sqlDB.sqlDB.BeginTx(ctx, opts)
	if sqlDB.logConfig.LogSQLQueries {
		logSQLQuery(sqlDB.ctx, gorm.NowFunc().Sub(startTime), beginTransactionLogMessage, nil, nil)
	}
	if err != nil {
		logDBError(sqlDB.ctx, sqlDB.logConfig, err)
		return nil, err
	}
	newLogConfig := *sqlDB.logConfig // clone logConfig to avoid changing the original one when setting LogRetryableErrorsAsInfo
	return &sqlTxWrapper{sqlTx: tx, ctx: ctx, logConfig: &newLogConfig}, nil
}

// We intentionally do not implement the 'sqlDb' interface to avoid Gorm from calling 'Begin' method.
// var _ sqlDb = &sqlDBWrapper{}

type queryRowWithoutLogging interface {
	queryRowWithoutLogging(query string, args ...interface{}) *sql.Row
}

func (sqlDB *sqlDBWrapper) queryRowWithoutLogging(query string, args ...interface{}) *sql.Row {
	return sqlDB.sqlDB.QueryRowContext(context.Background(), query, args...)
}

var _ queryRowWithoutLogging = &sqlDBWrapper{}

type withContexter interface {
	withContext(ctx context.Context) gorm.SQLCommon
}

func (sqlDB *sqlDBWrapper) withContext(ctx context.Context) gorm.SQLCommon {
	return &sqlDBWrapper{sqlDB: sqlDB.sqlDB, ctx: ctx, logConfig: sqlDB.logConfig}
}

var _ withContexter = &sqlDBWrapper{}

type contextGetter interface {
	getContext() context.Context
}

func (sqlDB *sqlDBWrapper) getContext() context.Context {
	return sqlDB.ctx
}

var _ contextGetter = &sqlDBWrapper{}

type logConfigGetter interface {
	getLogConfig() *LogConfig
}

func (sqlDB *sqlDBWrapper) getLogConfig() *LogConfig {
	return sqlDB.logConfig
}

var _ logConfigGetter = &sqlDBWrapper{}

func (sqlDB *sqlDBWrapper) Close() error {
	return sqlDB.sqlDB.Close()
}

var _ interface{ Close() error } = &sqlDBWrapper{}

func (sqlDB *sqlDBWrapper) conn(ctx context.Context) (*sqlConnWrapper, error) {
	conn, err := sqlDB.sqlDB.Conn(ctx)
	if err != nil {
		logDBError(sqlDB.ctx, sqlDB.logConfig, err)
		return nil, err
	}
	return &sqlConnWrapper{conn: conn, logConfig: sqlDB.logConfig}, nil
}
