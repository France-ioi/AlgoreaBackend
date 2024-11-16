// Package database provides an interface for database operations.
package database

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"reflect"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"
	"unsafe"

	"github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"github.com/luna-duclos/instrumentedsql"

	log "github.com/France-ioi/AlgoreaBackend/v2/app/logging"
	"github.com/France-ioi/AlgoreaBackend/v2/app/rand"
	"github.com/France-ioi/AlgoreaBackend/v2/golang"
)

// LogConfig is the configuration for the database logs.
type LogConfig struct {
	Logger            log.DBLogger
	LogSQLQueries     bool
	AnalyzeSQLQueries bool
}

type cte struct {
	name     string
	subQuery interface{}
}

// DB contains information for current db connection (wraps *gorm.DB).
type DB struct {
	db        *gorm.DB
	ctx       context.Context
	ctes      []cte
	logConfig *LogConfig
}

// ErrLockWaitTimeoutExceeded is returned when we cannot acquire a lock.
var ErrLockWaitTimeoutExceeded = errors.New("lock wait timeout exceeded")

// newDB wraps *gorm.DB.
func newDB(ctx context.Context, db *gorm.DB, ctes []cte, logConfig *LogConfig) *DB {
	return &DB{db: db, ctx: ctx, ctes: ctes, logConfig: logConfig}
}

// Open connects to the database and tests the connection.
func Open(source interface{}) (*DB, error) {
	logger := log.SharedLogger.NewDBLogger()

	lc := LogConfig{
		Logger:            logger,
		LogSQLQueries:     log.SharedLogger.IsSQLQueriesLoggingEnabled(),
		AnalyzeSQLQueries: log.SharedLogger.IsSQLQueriesAnalyzingEnabled(),
	}

	rawSQLQueriesLoggingEnabled := log.SharedLogger.IsRawSQLQueriesLoggingEnabled()
	return OpenWithLogConfig(source, lc, rawSQLQueriesLoggingEnabled)
}

// OpenWithLogConfig connects to the database and tests the connection. It uses the given logging settings.
func OpenWithLogConfig(source interface{}, lc LogConfig, rawSQLQueriesLoggingEnabled bool) (*DB, error) {
	var err error
	var dbConn *gorm.DB
	driverName := "mysql"

	ctx := context.Background()

	var rawConnection gorm.SQLCommon
	var ownSQLDBConnection bool
	switch src := source.(type) {
	case string:
		var sqlDB *sql.DB
		sqlDB, err = OpenRawDBConnection(src, rawSQLQueriesLoggingEnabled)
		if err != nil {
			return nil, err
		}
		rawConnection = &sqlDBWrapper{sqlDB: sqlDB, ctx: ctx, logConfig: &lc}
		ownSQLDBConnection = true
	case *sql.DB:
		rawConnection = &sqlDBWrapper{sqlDB: src, ctx: ctx, logConfig: &lc}
	default:
		return nil, fmt.Errorf("unknown database source type: %T (%v)", src, src)
	}
	dbConn, _ = gorm.Open(driverName, rawConnection)

	// gorm.Open only pings the connection when it's sql.DB. So we need to ping it ourselves.
	if err = dbConn.CommonDB().(*sqlDBWrapper).sqlDB.Ping(); err != nil && ownSQLDBConnection {
		_ = dbConn.CommonDB().(*sqlDBWrapper).sqlDB.Close()
	}

	// we log queries and errors ourselves
	dbConn.LogMode(false)

	return newDB(ctx, dbConn, nil, &lc), err
}

// OpenRawDBConnection creates a new DB connection.
func OpenRawDBConnection(sourceDSN string, enableRawLevelLogging bool) (*sql.DB, error) {
	if enableRawLevelLogging {
		registerDriver := true
		for _, driverName := range sql.Drivers() {
			if driverName == "instrumented-mysql" {
				registerDriver = false
				break
			}
		}

		if registerDriver {
			logger := log.SharedLogger.NewDBLogger()
			rawDBLogger := log.NewRawDBLogger(logger, log.SharedLogger.IsRawSQLQueriesLoggingEnabled())
			sql.Register("instrumented-mysql",
				instrumentedsql.WrapDriver(&mysql.MySQLDriver{}, instrumentedsql.WithLogger(rawDBLogger)))
		}
	}
	return sql.Open(golang.IfElse(enableRawLevelLogging, "instrumented-mysql", "mysql"), sourceDSN)
}

// New clones a new db connection without search conditions.
func (conn *DB) New() *DB {
	return newDB(conn.ctx, conn.db.New(), nil, conn.logConfig)
}

func (conn *DB) inTransaction(txFunc func(*DB) error, txOptions ...*sql.TxOptions) (err error) {
	return conn.inTransactionWithCount(txFunc, 0, txOptions...)
}

const (
	transactionRetriesLimit        = 30
	transactionDelayBetweenRetries = 50 * time.Millisecond
)

func (conn *DB) inTransactionWithCount(txFunc func(*DB) error, count int64, txOptions ...*sql.TxOptions) (err error) {
	if count > transactionRetriesLimit {
		return errors.New("transaction retries limit exceeded")
	}

	conn.sleepBeforeStartingTransactionIfNeeded(count)

	txOpts := &sql.TxOptions{}
	if len(txOptions) > 0 {
		txOpts = txOptions[0]
	}

	txDB := gormDBBeginTxReplacement(conn.ctx, conn.db, txOpts)
	if txDB.Error != nil {
		return txDB.Error
	}
	defer func() {
		p := recover()
		switch {
		case p != nil:
			// ensure rollback is executed even in case of panic
			rollbackErr := txDB.Rollback().Error
			// There are two possible causes of the rollback error: 1) the connection is broken, 2) the context is canceled.
			// In all cases, the DB library closes the connection on rollback failure and logs the error.
			// But still, in both cases, we should not retry the transaction.
			// If the panic was a deadlock/timeout error, we replace it with either the rollback error or the result of retrying.
			if conn.handleDeadlockAndLockWaitTimeout(txFunc, count, p, rollbackErr, &err, txOptions...) {
				return
			}
			panic(p) // re-throw panic after rollback if it was not a deadlock/timeout error
		case err != nil:
			// ensure the rollback is executed, do not change the err
			rollbackErr := txDB.Rollback().Error
			// There are two possible causes of the rollback error: 1) the connection is broken, 2) the context is canceled.
			// In all cases, the DB library closes the connection on rollback failure and logs the error.
			// But still, in both cases, we should not retry the transaction.
			// If the error was a deadlock/timeout error, we replace it with either the rollback error or the result of retrying.
			conn.handleDeadlockAndLockWaitTimeout(txFunc, count, err, rollbackErr, &err, txOptions...)
		default:
			err = txDB.Commit().Error // if err is nil, returns the potential error from commit
		}
	}()
	err = txFunc(newDB(conn.ctx, txDB, nil, conn.logConfig))
	return err
}

func (conn *DB) sleepBeforeStartingTransactionIfNeeded(count int64) {
	if count > 0 && conn.ctx.Value(retryEachTransactionContextKey) == nil {
		time.Sleep(time.Duration(float64(transactionDelayBetweenRetries) * (1.0 + (rand.Float64()-0.5)*0.1))) // ±5%
	}
}

func cloneGormDB(db *gorm.DB) *gorm.DB {
	return db.Model(db.Value) // clone the db
}

type gormDBAccessor struct {
	sync.RWMutex
	Value        interface{}
	Error        error
	RowsAffected int64

	// single db
	DB gorm.SQLCommon
}

func replaceDBInGormDB(db *gorm.DB, newDB gorm.SQLCommon) {
	(*gormDBAccessor)(unsafe.Pointer(db)).DB = newDB //nolint:gosec // G103: Here we write into a private field of a struct.
	db.Dialect().SetDB(newDB)
}

// gormDBBeginTxReplacement is a replacement for gorm.DB.BeginTx that uses our sqlDBWrapper.
// The code does absolutely the same as gorm.DB.BeginTx, but uses our sqlDBWrapper instead of sql.DB.
// Happily, the original gorm.DB.BeginTx is only called from gorm.DB.Begin which is only called from gorm.DB.Transaction,
// and we never use/expose gorm.DB.Transaction.
func gormDBBeginTxReplacement(ctx context.Context, db *gorm.DB, txOpts *sql.TxOptions) *gorm.DB {
	c := cloneGormDB(db)
	if db, ok := db.CommonDB().(*sqlDBWrapper); ok && db != nil {
		tx, err := db.BeginTx(ctx, txOpts)
		replaceDBInGormDB(c, tx)
		_ = c.AddError(err)
	} else {
		_ = c.AddError(gorm.ErrCantStartTransaction)
	}
	return c
}

func (conn *DB) handleDeadlockAndLockWaitTimeout(txFunc func(*DB) error, count int64, errToHandle interface{}, rollbackErr error,
	returnErr *error, txOptions ...*sql.TxOptions,
) (shouldIgnoreInitialError bool) {
	errToHandleError, _ := errToHandle.(error)

	// Deadlock found / lock wait timeout exceeded
	if errToHandle != nil && (IsDeadlockError(errToHandleError) || IsLockWaitTimeoutExceededError(errToHandleError)) {
		if rollbackErr != nil { // do not retry if rollback failed
			// as the previous error was a retryable error, we should return the rollback error, as it is more important
			*returnErr = rollbackErr
			return true
		}
		// retry
		log.Infof("Retrying transaction (count: %d) after %s", count+1, errToHandleError.Error())
		*returnErr = conn.inTransactionWithCount(txFunc, count+1, txOptions...)
		return true
	}
	return false
}

func (conn *DB) isInTransaction() bool {
	if _, ok := interface{}(conn.db.CommonDB()).(driver.Tx); ok {
		return true
	}
	return false
}

func (conn *DB) withNamedLock(lockName string, timeout time.Duration, txFunc func(*DB) error) (err error) {
	initGetLockTime := time.Now()

	// Use a lock so that we don't execute the listener multiple times in parallel
	var getLockResult int64
	err = conn.db.Raw("SELECT GET_LOCK(?, ?)", lockName, int64(timeout/time.Second)).Row().Scan(&getLockResult)
	if err != nil {
		return err
	}
	if getLockResult != 1 {
		return ErrLockWaitTimeoutExceeded
	}

	log.Debugf("Duration for GET_LOCK(%s, %v): %v", lockName, timeout, time.Since(initGetLockTime))

	defer func() {
		releaseErr := conn.db.Exec("SELECT RELEASE_LOCK(?)", lockName).Error
		if err == nil {
			err = releaseErr
		}
	}()
	err = txFunc(conn)
	return
}

// Close closes current db connection.  If database connection is not an io.Closer, returns an error.
func (conn *DB) Close() error {
	return conn.db.Close()
}

// Limit specifies the number of records to be retrieved.
func (conn *DB) Limit(limit interface{}) *DB {
	return newDB(conn.ctx, conn.db.Limit(limit), conn.ctes, conn.logConfig)
}

// Offset specifies the offset of the records to be retrieved.
func (conn *DB) Offset(offset interface{}) *DB {
	return newDB(conn.ctx, conn.db.Offset(offset), conn.ctes, conn.logConfig)
}

// Where returns a new relation, filters records with given conditions, accepts `map`,
// `struct` or `string` as conditions, refer http://jinzhu.github.io/gorm/crud.html#query
func (conn *DB) Where(query interface{}, args ...interface{}) *DB {
	return newDB(conn.ctx, conn.db.Where(query, args...), conn.ctes, conn.logConfig)
}

// Joins specifies Joins conditions
//
//	db.Joins("JOIN emails ON emails.user_id = users.id AND emails.email = ?", "jinzhu@example.org").Find(&user)
func (conn *DB) Joins(query string, args ...interface{}) *DB {
	return newDB(conn.ctx, conn.db.Joins(query, args...), conn.ctes, conn.logConfig)
}

// Select specifies fields that you want to retrieve from database when querying, by default, will select all fields;
// When creating/updating, specify fields that you want to save to database.
func (conn *DB) Select(query interface{}, args ...interface{}) *DB {
	return newDB(conn.ctx, conn.db.Select(query, args...), conn.ctes, conn.logConfig)
}

// Table specifies the table you would like to run db operations.
func (conn *DB) Table(name string) *DB {
	return newDB(conn.ctx, conn.db.Table(name), conn.ctes, conn.logConfig)
}

// Group specifies the group method on the find.
func (conn *DB) Group(query string) *DB {
	return newDB(conn.ctx, conn.db.Group(query), conn.ctes, conn.logConfig)
}

// Order specifies order when retrieve records from database, set reorder to `true` to overwrite defined conditions
//
//	db.Order("name DESC")
//	db.Order("name DESC", true) // reorder
//	db.Order(gorm.SqlExpr("name = ? DESC", "first")) // sql expression
func (conn *DB) Order(value interface{}, reorder ...bool) *DB {
	return newDB(conn.ctx, conn.db.Order(value, reorder...), conn.ctes, conn.logConfig)
}

// Having specifies HAVING conditions for GROUP BY.
func (conn *DB) Having(query interface{}, args ...interface{}) *DB {
	return newDB(conn.ctx, conn.db.Having(query, args...), conn.ctes, conn.logConfig)
}

// Union specifies UNION of two queries (receiver UNION query).
func (conn *DB) Union(query *DB) *DB {
	return conn.New().Raw("? UNION ?", conn.SubQuery(), query.SubQuery())
}

// UnionAll specifies UNION ALL of two queries (receiver UNION ALL query).
func (conn *DB) UnionAll(query *DB) *DB {
	return conn.New().Raw("? UNION ALL ?", conn.SubQuery(), query.SubQuery())
}

// With adds a common table expression (CTE) to the query.
func (conn *DB) With(name string, query *DB) *DB {
	if conn.ctes != nil {
		for _, cte := range conn.ctes {
			if cte.name == name {
				panic(fmt.Sprintf("CTE with name %q already exists", name))
			}
		}
	}

	newCTEs := make([]cte, 0, len(conn.ctes)+1)
	newCTEs = append(newCTEs, conn.ctes...)
	newCTEs = append(newCTEs, cte{name: name, subQuery: query.SubQuery()})
	return newDB(conn.ctx, conn.db, newCTEs, conn.logConfig)
}

// Raw uses raw sql as conditions
//
//	db.Raw("SELECT name, age FROM users WHERE name = ?", 3).Scan(&result)
func (conn *DB) Raw(query string, args ...interface{}) *DB {
	// db.Raw("").Joins(...) is a hack for making db.Raw("...").Joins(...) work better
	return newDB(conn.ctx,
		conn.db.New().Set("gorm:query_option", "").
			Raw("").Joins(query, args...), nil, conn.logConfig)
}

// UpdateColumns is a synonym for UpdateColumn.
func (conn *DB) UpdateColumns(attrs ...interface{}) *DB {
	return conn.UpdateColumn(attrs...)
}

// UpdateColumn updates attributes without callbacks, refer: https://jinzhu.github.io/gorm/crud.html#update
func (conn *DB) UpdateColumn(attrs ...interface{}) *DB {
	return newDB(conn.ctx, conn.toQuery().UpdateColumn(attrs...), nil, conn.logConfig)
}

// SubQuery returns the query as sub query.
func (conn *DB) SubQuery() interface{} {
	mustNotBeError(conn.Error())
	return conn.toQuery().SubQuery()
}

// QueryExpr returns the query as expr object.
func (conn *DB) QueryExpr() interface{} {
	mustNotBeError(conn.Error())
	return conn.toQuery().QueryExpr()
}

func (conn *DB) toQuery() *gorm.DB {
	if len(conn.ctes) == 0 {
		return conn.db
	}
	strBuilder := new(strings.Builder)
	strBuilder.WriteString("WITH ")
	cteSubQueries := make([]interface{}, 0, len(conn.ctes)+1)
	isFirstCTE := true
	for _, cte := range conn.ctes {
		if !isFirstCTE {
			strBuilder.WriteString(", ")
		}
		isFirstCTE = false
		strBuilder.WriteString(QuoteName(cte.name))
		strBuilder.WriteString(" AS ?")
		cteSubQueries = append(cteSubQueries, cte.subQuery)
	}
	strBuilder.WriteString(" ?")
	cteSubQueries = append(cteSubQueries, conn.db.QueryExpr())
	return conn.Raw(strBuilder.String(), cteSubQueries...).db
}

// Scan scans value to a struct.
func (conn *DB) Scan(dest interface{}) *DB {
	return newDB(conn.ctx, conn.toQuery().Scan(dest), nil, conn.logConfig)
}

// ScanIntoSlices scans multiple columns into slices.
func (conn *DB) ScanIntoSlices(pointersToSlices ...interface{}) *DB {
	if conn.db.Error != nil {
		return conn
	}

	valuesPointers := make([]interface{}, len(pointersToSlices))
	for index := range pointersToSlices {
		reflSlice := reflect.ValueOf(pointersToSlices[index]).Elem()
		if reflSlice.Len() > 0 {
			reflSlice.Set(reflect.MakeSlice(reflSlice.Type(), 0, reflSlice.Cap()))
		}
		valuesPointers[index] = reflect.New(reflSlice.Type().Elem()).Interface()
	}

	rows, err := conn.toQuery().Rows() //nolint:rowserrcheck rows.Err() is checked before return.
	if rows != nil {
		defer func() {
			_ = conn.db.AddError(rows.Close())
		}()
	}
	if conn.db.AddError(err) != nil {
		return conn
	}

	for rows.Next() {
		if err := rows.Scan(valuesPointers...); conn.db.AddError(err) != nil { //nolint:gocritic Err is checked with AddError.
			return conn
		}
		for index, valuePointer := range valuesPointers {
			reflSlice := reflect.ValueOf(pointersToSlices[index]).Elem()
			reflSlice.Set(reflect.Append(reflSlice, reflect.ValueOf(valuePointer).Elem()))
		}
	}
	_ = conn.db.AddError(rows.Err())
	return conn
}

// ScanIntoSliceOfMaps scans value into a slice of maps.
func (conn *DB) ScanIntoSliceOfMaps(dest *[]map[string]interface{}) *DB {
	*dest = []map[string]interface{}(nil)

	return conn.ScanAndHandleMaps(func(rowMap map[string]interface{}) error {
		*dest = append(*dest, rowMap)
		return nil
	})
}

// ScanAndHandleMaps scans values into maps and calls the given handler for each row.
func (conn *DB) ScanAndHandleMaps(handler func(map[string]interface{}) error) *DB {
	if conn.db.Error != nil {
		return conn
	}

	rows, err := conn.toQuery().Rows() //nolint:rowserrcheck rows.Err() is checked before return.
	if rows != nil {
		defer func() {
			_ = conn.db.AddError(rows.Close())
		}()
	}
	if conn.db.AddError(err) != nil {
		return conn
	}
	cols, err := rows.Columns()
	if conn.db.AddError(err) != nil {
		return conn
	}

	for rows.Next() {
		mapValue := conn.readRowIntoMap(cols, rows)
		if conn.db.Error != nil {
			return conn
		}
		err = handler(mapValue)
		if conn.db.AddError(err) != nil {
			return conn
		}
	}
	_ = conn.db.AddError(rows.Err())
	return conn
}

func (conn *DB) readRowIntoMap(cols []string, rows *sql.Rows) map[string]interface{} {
	columns := make([]interface{}, len(cols))
	columnPointers := make([]interface{}, len(cols))
	for i := range columns {
		columnPointers[i] = &columns[i]
	}

	if err := rows.Scan(columnPointers...); conn.db.AddError(err) != nil { //nolint:gocritic Err is checked with AddError.
		return nil
	}

	rowMap := make(map[string]interface{})
	for i, columnName := range cols {
		if value, ok := columns[i].([]byte); ok {
			columns[i] = *(*string)(unsafe.Pointer(&value)) //nolint:gosec
		}
		rowMap[columnName] = columns[i]
	}
	return rowMap
}

// Count gets how many records for a model.
func (conn *DB) Count(dest interface{}) *DB {
	if conn.Error() != nil {
		return conn
	}
	return newDB(conn.ctx, conn.toQuery().Count(dest), nil, conn.logConfig)
}

// Pluck is used to query a single column into a slice of values
//
//	var ids []int64
//	db.Table("users").Pluck("id", &ids)
//
// The 'values' parameter should be a pointer to a slice.
func (conn *DB) Pluck(column string, values interface{}) *DB {
	if conn.db.Error != nil {
		return conn
	}
	// If 'values' is not empty, wipe its content by replacing it with an empty slice.
	// Otherwise we would get new values mixed with old values.
	reflectPtr := reflect.ValueOf(values)
	if reflectPtr.Kind() != reflect.Ptr {
		panic(fmt.Sprintf("values should be a pointer to a slice, not %s", reflectPtr.Kind()))
	}
	reflectValue := reflectPtr.Elem()
	if reflectValue.Kind() != reflect.Slice {
		panic(fmt.Sprintf("values should be a pointer to a slice, not a pointer to %s", reflectValue.Kind()))
	}
	return newDB(conn.ctx, conn.toQuery().Pluck(column, values), nil, conn.logConfig)
}

// PluckFirst is used to query a single column and take the first value
//
//	var id int64
//	db.Table("users").PluckFirst("id", &id)
//
// The 'values' parameter should be a pointer to a value.
func (conn *DB) PluckFirst(column string, value interface{}) *DB {
	valuesReflValue := reflect.MakeSlice(reflect.SliceOf(reflect.TypeOf(value).Elem()), 0, 1)
	valuesPtrReflValue := reflect.New(reflect.SliceOf(reflect.TypeOf(value).Elem()))
	valuesPtrReflValue.Elem().Set(valuesReflValue)
	valuesReflValue = valuesPtrReflValue.Elem()
	values := valuesPtrReflValue.Interface()
	result := newDB(conn.ctx, conn.Limit(1).toQuery().Pluck(column, values), nil, conn.logConfig)
	if result.Error() != nil {
		return result
	}
	if valuesReflValue.Len() == 0 {
		_ = result.db.AddError(gorm.ErrRecordNotFound) // nolint:gosec
		return result
	}
	reflect.ValueOf(value).Elem().Set(valuesReflValue.Index(0))
	return result
}

// Take returns a record that match given conditions, the order will depend on the database implementation.
func (conn *DB) Take(out interface{}, where ...interface{}) *DB {
	return newDB(conn.ctx, conn.toQuery().Take(out, where...), nil, conn.logConfig)
}

// HasRows returns true if at least one row is found.
func (conn *DB) HasRows() (bool, error) {
	var result int64
	err := conn.PluckFirst("1", &result).Error()
	if gorm.IsRecordNotFoundError(err) {
		return false, nil
	}
	return err == nil, err
}

// Delete deletes value matching given conditions, if the value has primary key, then will including the primary key as condition.
func (conn *DB) Delete(where ...interface{}) *DB {
	return newDB(conn.ctx, conn.toQuery().Delete(nil, where...), nil, conn.logConfig)
}

// RowsAffected returns the number of rows affected by the last INSERT/UPDATE/DELETE statement.
func (conn *DB) RowsAffected() int64 {
	return conn.db.RowsAffected
}

// Error returns current errors.
func (conn *DB) Error() error {
	return conn.db.Error
}

// Exec executes raw sql.
func (conn *DB) Exec(sqlQuery string, values ...interface{}) *DB {
	return newDB(conn.ctx, conn.db.Exec(sqlQuery, values...), nil, conn.logConfig)
}

// insertMaps reads fields from the given maps and inserts the values set in the first row (so keys in all maps should be same)
// into the given table.
func (conn *DB) insertMaps(tableName string, dataMaps []map[string]interface{}) error {
	if len(dataMaps) == 0 {
		return nil
	}
	query, values := conn.constructInsertMapsStatement(dataMaps, tableName, false)
	return conn.db.Exec(query, values...).Error
}

// InsertIgnoreMaps reads fields from the given maps and inserts the values set in the first row (so keys in all maps should be same)
// into the given table ignoring errors (such as duplicates).
func (conn *DB) InsertIgnoreMaps(tableName string, dataMaps []map[string]interface{}) error {
	if len(dataMaps) == 0 {
		return nil
	}
	query, values := conn.constructInsertMapsStatement(dataMaps, tableName, true)
	return conn.db.Exec(query, values...).Error
}

func (conn *DB) constructInsertMapsStatement(
	dataMaps []map[string]interface{}, tableName string, ignore bool,
) (query string, values []interface{}) {
	// data for the building the SQL request
	// "INSERT INTO tablename (keys... ) VALUES (?, ?, NULL, ?, ...), ...", values...
	values = make([]interface{}, 0, len(dataMaps)*len(dataMaps[0]))
	keys := make([]string, 0, len(dataMaps[0]))
	for key := range dataMaps[0] {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	escapedKeys := make([]string, 0, len(keys))
	for _, key := range keys {
		escapedKeys = append(escapedKeys, QuoteName(key))
	}
	var builder strings.Builder
	var ignoreString string
	if ignore {
		ignoreString = "IGNORE "
	}
	// nolint:gosec
	_, _ = builder.WriteString(fmt.Sprintf("INSERT %sINTO `%s` (%s) VALUES ", ignoreString, tableName, strings.Join(escapedKeys, ", ")))
	for index, dataMap := range dataMaps {
		_, _ = builder.WriteRune('(')
		for keyIndex, key := range keys {
			_, _ = builder.WriteRune('?')
			values = append(values, dataMap[key])
			if keyIndex != len(keys)-1 {
				_, _ = builder.WriteString(", ")
			}
		}
		_, _ = builder.WriteRune(')')
		if index != len(dataMaps)-1 {
			_, _ = builder.WriteString(", ")
		}
	}
	query = builder.String()
	return query, values
}

// insertOrUpdateMaps reads fields from the given maps and inserts the values set in the first row
// (so all the maps should have the same keys)
// into the given table (like insertMaps does). If it is a duplicate, the listed columns will be updated.
// If updateColumns is nil, all the columns in dataMaps will be updated.
func (conn *DB) insertOrUpdateMaps(tableName string, dataMaps []map[string]interface{}, updateColumns []string) error {
	if len(dataMaps) == 0 {
		return nil
	}
	query, values := conn.constructInsertMapsStatement(dataMaps, tableName, false)

	if updateColumns == nil {
		updateColumns = make([]string, 0, len(dataMaps))
		for key := range dataMaps[0] {
			updateColumns = append(updateColumns, key)
		}
		sort.Strings(updateColumns)
	}

	var builder strings.Builder
	_, _ = builder.WriteString(query)
	_, _ = builder.WriteString(" ON DUPLICATE KEY UPDATE ")
	for index, column := range updateColumns {
		quotedColumn := QuoteName(column)
		if index != 0 {
			_, _ = builder.WriteString(", ")
		}
		_, _ = builder.WriteString(quotedColumn)
		_, _ = builder.WriteString(" = VALUES(")
		_, _ = builder.WriteString(quotedColumn)
		_, _ = builder.WriteRune(')')
	}
	return conn.db.Exec(builder.String(), values...).Error
}

// Set sets setting by name, which could be used in callbacks, will clone a new db, and update its setting.
func (conn *DB) Set(name string, value interface{}) *DB {
	return newDB(conn.ctx, conn.db.Set(name, value), conn.ctes, conn.logConfig)
}

// ErrNoTransaction means that a called method/function cannot work outside of a transaction.
var ErrNoTransaction = errors.New("should be executed in a transaction")

func (conn *DB) mustBeInTransaction() {
	if !conn.isInTransaction() {
		panic(ErrNoTransaction)
	}
}

// WithExclusiveWriteLock converts "SELECT ..." statement into "SELECT ... FOR UPDATE" statement.
// For existing rows, it will read the latest committed data (instead of the data from the repeatable-read snapshot)
// and acquire an exclusive lock on them, preventing other transactions from modifying them and
// even from getting exclusive/shared locks on them. For non-existing rows, it works similarly to a shared lock (FOR SHARE).
func (conn *DB) WithExclusiveWriteLock() *DB {
	conn.mustBeInTransaction()
	return conn.Set("gorm:query_option", "FOR UPDATE")
}

// WithSharedWriteLock converts "SELECT ..." statement into "SELECT ... FOR SHARE" statement.
// For existing rows, it will read the latest committed data (instead of the data from the repeatable-read snapshot)
// and acquire a shared lock on them, preventing other transactions from modifying them.
func (conn *DB) WithSharedWriteLock() *DB {
	conn.mustBeInTransaction()
	return conn.Set("gorm:query_option", "FOR SHARE")
}

// WithCustomWriteLocks converts "SELECT ..." statement into "SELECT ... FOR SHARE OF ... FOR UPDATE ..." statement.
// For existing rows, it will read the latest committed data for the listed tables
// (instead of the data from the repeatable-read snapshot) and acquire shared/exclusive locks on them,
// preventing other transactions from modifying them.
func (conn *DB) WithCustomWriteLocks(shared, exclusive *golang.Set[string]) *DB {
	conn.mustBeInTransaction()

	var builder strings.Builder
	if shared.Size() > 0 {
		builder.WriteString("FOR SHARE OF ")
		tables := shared.Values()
		sort.Strings(tables)
		for index, sharedTable := range tables {
			if index != 0 {
				builder.WriteString(", ")
			}
			builder.WriteString(QuoteName(sharedTable))
		}
	}
	if exclusive.Size() > 0 {
		if shared.Size() > 0 {
			builder.WriteString(" ")
		}
		builder.WriteString("FOR UPDATE OF ")
		tables := exclusive.Values()
		sort.Strings(tables)
		for index, exclusiveTable := range tables {
			if index != 0 {
				builder.WriteString(", ")
			}
			builder.WriteString(QuoteName(exclusiveTable))
		}
	}

	return conn.Set("gorm:query_option", builder.String())
}

// Prepare creates a prepared statement for later queries or executions.
// As the method must be executed in a transaction, the returned statement is bound to the transaction.
func (conn *DB) Prepare(query string) (*SQLStmtWrapper, error) {
	conn.mustBeInTransaction()

	tx := conn.db.CommonDB().(*sqlTxWrapper)
	return tx.prepare(query)
}

const keyTriesCount = 10

func (conn *DB) retryOnDuplicatePrimaryKeyError(f func(db *DB) error) error {
	return conn.retryOnDuplicateKeyError("PRIMARY", "id", f)
}

func (conn *DB) retryOnDuplicateKeyError(keyName, nameInError string, f func(db *DB) error) error {
	i := 0
	for ; i < keyTriesCount; i++ {
		err := f(conn)
		if err != nil {
			if IsDuplicateEntryErrorForKey(err, keyName) {
				continue // retry with a new id
			}
			return err
		}
		return nil
	}
	err := fmt.Errorf("cannot generate a new %s", nameInError)
	log.Error(err)
	return err
}

func (conn *DB) withForeignKeyChecksDisabled(blockFunc func(*DB) error, txOptions ...*sql.TxOptions) (err error) {
	defer recoverPanics(&err)

	innerFunc := func(db *DB) error {
		mustNotBeError(
			db.Exec(`
				SET @foreign_key_checks_stack_count = IFNULL(@foreign_key_checks_stack_count, 0) + 1,
				    FOREIGN_KEY_CHECKS = IF(IFNULL(@foreign_key_checks_stack_count, 0) = 0, 0, @@SESSION.FOREIGN_KEY_CHECKS)`).
				Error())
		defer func() {
			mustNotBeError(
				db.Exec(`
					SET @foreign_key_checks_stack_count = @foreign_key_checks_stack_count - 1,
					    FOREIGN_KEY_CHECKS = IF(@foreign_key_checks_stack_count = 1, 1, @@SESSION.FOREIGN_KEY_CHECKS)`).
					Error())
		}()
		return blockFunc(db)
	}
	if !conn.isInTransaction() {
		return conn.inTransaction(innerFunc, txOptions...)
	} // else {
	return innerFunc(conn)
	//}
}

func mustNotBeError(err error) {
	if err != nil {
		panic(err)
	}
}

func recoverPanics(returnErr *error) {
	if p := recover(); p != nil {
		switch e := p.(type) {
		case runtime.Error:
			panic(e)
		case error:
			*returnErr = e
		default:
			panic(p)
		}
	}
}

// QuoteName surrounds a given table/column name in backtick quotes and escapes the content.
func QuoteName(name string) string {
	return "`" + strings.Replace(name, "`", "``", -1) + "`"
}

// Default returns gorm.Expr("DEFAULT").
func Default() interface{} {
	return gorm.Expr("DEFAULT")
}

// EscapeLikeString escapes string with the given escape character.
// This escapes the contents of a string (provided as string)
// by adding the escape character before percent signs (%), and underscore signs (_).
func EscapeLikeString(v string, escapeCharacter byte) string {
	pos := 0
	buf := make([]byte, len(v)*3)

	for i := 0; i < len(v); i++ {
		c := v[i]
		switch c {
		case escapeCharacter:
			buf[pos] = escapeCharacter
			buf[pos+1] = escapeCharacter
			pos += 2
		case '%':
			buf[pos] = escapeCharacter
			buf[pos+1] = '%'
			pos += 2
		case '_':
			buf[pos] = escapeCharacter
			buf[pos+1] = '_'
			pos += 2
		default:
			buf[pos] = c
			pos++
		}
	}

	result := buf[:pos]
	return *(*string)(unsafe.Pointer(&result)) //nolint:gosec
}
