// Package database provides an interface for database operations.
package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"runtime"
	"sort"
	"strings"
	"time"
	"unsafe"

	"github.com/go-sql-driver/mysql"
	"github.com/luna-duclos/instrumentedsql"
	gormMysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	log "github.com/France-ioi/AlgoreaBackend/app/logging"
	"github.com/France-ioi/AlgoreaBackend/app/rand"
)

// TODO:
// 1) wait until
//      https://github.com/go-gorm/gorm/issues/4525,
//      https://github.com/go-gorm/gorm/issues/4533,
//      https://github.com/go-gorm/gorm/issues/4534
//    are fixed
// 2) make everything work,
// 3) take full advantage of AddSelect() chaining (appending columns to the list)

// DB contains information for current db connection (wraps *gorm.DB).
type DB struct {
	db  *gorm.DB
	ctx context.Context
}

// ErrLockWaitTimeoutExceeded is returned when we cannot acquire a lock.
var ErrLockWaitTimeoutExceeded = errors.New("lock wait timeout exceeded")

type suppressFromClause struct{}

func (s *suppressFromClause) Build(clause.Builder) {}
func (s *suppressFromClause) ModifyStatement(statement *gorm.Statement) {
	statement.Clauses["FROM"] = clause.Clause{Expression: s, Builder: func(c clause.Clause, builder clause.Builder) {
		c.Expression.Build(builder)
	}}
}

var (
	_ gorm.StatementModifier = &suppressFromClause{}
	_ clause.Expression      = &suppressFromClause{}
)

type suppressSelectClause struct{}

func (s *suppressSelectClause) Build(clause.Builder) {}
func (s *suppressSelectClause) ModifyStatement(statement *gorm.Statement) {
	statement.Clauses["SELECT"] = clause.Clause{Expression: s}
}

var (
	_ gorm.StatementModifier = &suppressSelectClause{}
	_ clause.Expression      = &suppressSelectClause{}
)

// newDB wraps *gorm.DB.
func newDB(ctx context.Context, db *gorm.DB) *DB {
	return &DB{db: db, ctx: ctx}
}

// Open connects to the database and tests the connection.
// nolint: gosec
func Open(source interface{}) (*DB, error) {
	var err error
	var dbConn *gorm.DB
	driverName := "mysql"
	logger, _ := log.SharedLogger.NewDBLogger()

	var rawConnection gorm.ConnPool
	switch src := source.(type) {
	case string:
		rawConnection, err = OpenRawDBConnection(src)
		if err != nil {
			return nil, err
		}
	case gorm.ConnPool:
		rawConnection = src
	default:
		return nil, fmt.Errorf("unknown database source type: %T (%v)", src, src)
	}
	dbConn, err = gorm.Open(gormMysql.New(gormMysql.Config{
		Conn:                      rawConnection,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{
		Logger:            logger,
		AllowGlobalUpdate: true,
	})
	if err != nil {
		return nil, err
	}

	return newDB(context.Background(), dbConn), err
}

// OpenRawDBConnection creates a new DB connection.
func OpenRawDBConnection(sourceDSN string) (*sql.DB, error) {
	registerDriver := true
	for _, driverName := range sql.Drivers() {
		if driverName == "instrumented-mysql" {
			registerDriver = false
			break
		}
	}

	if registerDriver {
		rawLogMode := log.SharedLogger.GetRawSQLLogMode()
		rawDBLogger := log.NewRawDBLogger(log.SharedLogger.Logger, rawLogMode)
		sql.Register("instrumented-mysql",
			instrumentedsql.WrapDriver(&mysql.MySQLDriver{}, instrumentedsql.WithLogger(rawDBLogger)))
	}
	return sql.Open("instrumented-mysql", sourceDSN)
}

// New clones a new db connection without search conditions.
func (conn *DB) New() *DB {
	return newDB(conn.ctx, conn.db.Session(&gorm.Session{
		NewDB: true,
	}))
}

func (conn *DB) inTransaction(txFunc func(*DB) error) (err error) {
	return conn.inTransactionWithCount(txFunc, 0)
}

const (
	transactionRetriesLimit        = 30
	transactionDelayBetweenRetries = 1000 * time.Millisecond
)

func (conn *DB) inTransactionWithCount(txFunc func(*DB) error, count int64) (err error) {
	if count > transactionRetriesLimit {
		return errors.New("transaction retries limit exceeded")
	}

	if count > 0 {
		time.Sleep(time.Duration(float64(transactionDelayBetweenRetries) * (1.0 + (rand.Float64()-0.5)*0.1))) // ±5%
	}

	txDB := conn.db.Begin()
	if txDB.Error != nil {
		return txDB.Error
	}
	defer func() {
		p := recover()
		switch {
		case p != nil:
			// ensure rollback is executed even in case of panic
			rollbackErr := txDB.Rollback().Error
			if conn.handleDeadLock(txFunc, count, p.(error), rollbackErr, &err) {
				return
			}
			panic(p) // re-throw panic after rollback
		case err != nil:
			// do not change the err
			rollbackErr := txDB.Rollback().Error
			if conn.handleDeadLock(txFunc, count, err, rollbackErr, &err) {
				return
			}
			if rollbackErr != nil {
				panic(err) // in case of error on rollback, panic
			}
		default:
			err = txDB.Commit().Error // if err is nil, returns the potential error from commit
		}
	}()
	err = txFunc(newDB(conn.ctx, txDB))
	return err
}

func (conn *DB) handleDeadLock(txFunc func(*DB) error, count int64, errToHandle, rollbackErr error,
	returnErr *error,
) bool {
	// Error 1213: Deadlock found when trying to get lock; try restarting transaction
	if errToHandle != nil && IsLockDeadlockError(errToHandle) {
		if rollbackErr != nil {
			panic(rollbackErr)
		}
		// retry
		*returnErr = conn.inTransactionWithCount(txFunc, count+1)
		return true
	}
	return false
}

func (conn *DB) isInTransaction() bool {
	if _, ok := conn.db.Statement.ConnPool.(gorm.TxCommitter); ok {
		return true
	}
	return false
}

func (conn *DB) withNamedLock(lockName string, timeout time.Duration, txFunc func(*DB) error) (err error) {
	// Use a lock so that we don't execute the listener multiple times in parallel
	var getLockResult int64
	err = conn.db.Raw("SELECT GET_LOCK(?, ?)", lockName, int64(timeout/time.Second)).Row().Scan(&getLockResult)
	if err != nil {
		return err
	}
	if getLockResult != 1 {
		return ErrLockWaitTimeoutExceeded
	}
	defer func() {
		releaseErr := conn.db.Exec("SELECT RELEASE_LOCK(?)", lockName).Error
		if err == nil {
			err = releaseErr
		}
	}()
	err = txFunc(conn)
	return
}

// Close close current db connection.  If database connection is not an io.Closer, returns an error.
func (conn *DB) Close() error {
	sqlDB, err := conn.db.DB()
	if err == nil {
		err = sqlDB.Close()
	}
	return err
}

// Limit specifies the number of records to be retrieved.
func (conn *DB) Limit(limit int) *DB {
	return newDB(conn.ctx, conn.db.Limit(limit).Session(&gorm.Session{}))
}

// Where returns a new relation, filters records with given conditions, accepts `map`,
// `struct` or `string` as conditions, refer http://jinzhu.github.io/gorm/crud.html#query
func (conn *DB) Where(query interface{}, args ...interface{}) *DB {
	return newDB(conn.ctx, conn.db.Where(query, args...).Session(&gorm.Session{}))
}

// Joins specifies Joins conditions
//
//	db.Joins("JOIN emails ON emails.user_id = users.id AND emails.email = ?", "jinzhu@example.org").Find(&user)
func (conn *DB) Joins(query string, args ...interface{}) *DB {
	return newDB(conn.ctx, conn.db.Joins(query, args...).Session(&gorm.Session{}))
}

// Or filters records that match before conditions or this one, similar to `Where`
func (conn *DB) Or(query interface{}, args ...interface{}) *DB {
	return newDB(conn.ctx, conn.db.Or(query, args...).Session(&gorm.Session{}))
}

// Select specifies fields that you want to retrieve from database when querying, by default, will select all fields;
// When creating/updating, specify fields that you want to save to database.
func (conn *DB) Select(query interface{}, args ...interface{}) *DB {
	oldSelectClause, hadPreviousSelectClause := conn.db.Statement.Clauses["SELECT"]
	newGormDB := conn.db.Select(query, args...).Session(&gorm.Session{})
	if hadPreviousSelectClause {
		newSelectClause := newGormDB.Statement.Clauses["SELECT"]
		newSelectClause.BeforeExpression = oldSelectClause.BeforeExpression
		newGormDB.Statement.Clauses["SELECT"] = newSelectClause
	}
	return newDB(newGormDB)
}

// AddSelect appends fields that you want to retrieve from database when querying
func (conn *DB) AddSelect(query interface{}, args ...interface{}) *DB {
	previousSelects := conn.db.Statement.Selects
	newDB := conn.Select(query, args...)
	previousSelects = append(previousSelects, newDB.db.Statement.Selects...)
	newDB.db.Statement.Selects = previousSelects
	return newDB
}

// With adds a CTE
func (conn *DB) With(query interface{}, args ...interface{}) *DB {
	if conditions := conn.ctx, conn.db.Statement.BuildCondition(query, args...); len(conditions) > 0 {
		return conn.Clauses(&withClause{expressions: conditions})
	}
	return conn
}

// Table specifies the table you would like to run db operations.
func (conn *DB) Table(name string, args ...interface{}) *DB {
	return newDB(conn.ctx, conn.db.Table(name, args...).Session(&gorm.Session{}))
}

// Group specifies the group method on the find.
func (conn *DB) Group(query string) *DB {
	return newDB(conn.ctx, conn.db.Group(query).Session(&gorm.Session{}))
}

// Order specify order when retrieve records from database
//
//	db.Order("name DESC")
//	db.Order(clause.OrderByColumn{Column: clause.Column{Name: "name"}, Desc: true})
func (conn *DB) Order(value interface{}) *DB {
	return newDB(conn.ctx, conn.db.Order(value).Session(&gorm.Session{}))
}

// Clauses adds clauses
func (conn *DB) Clauses(conds ...clause.Expression) *DB {
	return newDB(conn.db.Clauses(conds...).Session(&gorm.Session{}))
}

// Having specifies HAVING conditions for GROUP BY.
func (conn *DB) Having(query interface{}, args ...interface{}) *DB {
	return newDB(conn.ctx, conn.db.Having(query, args...).Session(&gorm.Session{}))
}

// Union specifies UNION of two queries (receiver UNION query).
func (conn *DB) Union(query interface{}) *DB {
	return conn.New().Table("(?) UNION (?)", conn.SubQuery(), query).Joins("").
		Clauses(&suppressFromClause{}, &suppressSelectClause{})
}

// UnionAll specifies UNION ALL of two queries (receiver UNION ALL query).
func (conn *DB) UnionAll(query interface{}) *DB {
	return conn.New().Table("(?) UNION ALL (?)", conn.SubQuery(), query).Joins("").
		Clauses(&suppressFromClause{}, &suppressSelectClause{})
}

// Raw uses raw sql as conditions
//
//	db.Raw("SELECT name, age FROM users WHERE name = ?", 3).Scan(&result)
func (conn *DB) Raw(query string, args ...interface{}) *DB {
	// db.Raw("").Joins(...) is a hack for making db.Raw("...").Joins(...) work better
	return newDB(conn.ctx, conn.db.Raw(query, args...).Session(&gorm.Session{}))
}

// Updates update attributes with callbacks, refer: https://gorm.io/docs/update.html#Update-Changed-Fields
func (conn *DB) Updates(values interface{}) *DB {
	return newDB(conn.db.Updates(values))
}

// UpdateColumn updates one attribute without callbacks
func (conn *DB) UpdateColumn(column string, value interface{}) *DB {
	return newDB(conn.ctx, conn.db.UpdateColumn(column, value))
}

// UpdateColumns updates attributes without callbacks
func (conn *DB) UpdateColumns(values interface{}) *DB {
	return newDB(conn.ctx, conn.db.UpdateColumns(values))
}

// SubQuery returns the query as sub query.
func (conn *DB) SubQuery() interface{} {
	mustNotBeError(conn.Error())
	return conn.db
}

// QueryExpr is just a synonym for SubQuery().
func (conn *DB) QueryExpr() interface{} {
	return conn.SubQuery()
}

// Scan scans value to a struct.
func (conn *DB) Scan(dest interface{}) *DB {
	result := newDB(conn.ctx, conn.db.Scan(dest))

	if result.Error() == nil {
		reflSlice := reflect.ValueOf(dest).Elem()
		if reflSlice.Type().Kind() == reflect.Slice && reflSlice.IsNil() {
			reflSlice.Set(reflect.MakeSlice(reflSlice.Type(), 0, 0))
		}
	}

	return result
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

	rows, err := conn.db.Rows() //nolint:rowserrcheck rows.Err() is checked before return.
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

	rows, err := conn.db.Rows() //nolint:rowserrcheck rows.Err() is checked before return.
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
func (conn *DB) Count(dest *int64) *DB {
	if conn.Error() != nil {
		return conn
	}
	return newDB(conn.ctx, conn.db.Count(dest))
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
	return newDB(conn.ctx, conn.db.Pluck(column, values))
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
	result := newDB(conn.ctx, conn.db.Limit(1).Pluck(column, values))
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
func (conn *DB) Take(dest interface{}, conds ...interface{}) *DB {
	return newDB(conn.ctx, conn.db.Take(dest, conds...))
}

// HasRows returns true if at least one row is found.
func (conn *DB) HasRows() (bool, error) {
	var result int64
	if conn.db.Statement.Table == "" {
		conn = conn.Clauses(&suppressFromClause{}).Table("to_be_ignored")
	}
	err := conn.Select("1").Limit(1).PluckFirst("1 AS `1`", &result).Error()
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	return err == nil, err
}

// Delete deletes value matching given conditions.
func (conn *DB) Delete(where ...interface{}) *DB {
	return newDB(conn.ctx, conn.db.Delete(nil, where...))
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
	return newDB(conn.ctx, conn.db.Exec(sqlQuery, values...))
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
	return newDB(conn.ctx, conn.db.Set(name, value))
}

// ErrNoTransaction means that a called method/function cannot work outside of a transaction.
var ErrNoTransaction = errors.New("should be executed in a transaction")

func (conn *DB) mustBeInTransaction() {
	if !conn.isInTransaction() {
		panic(ErrNoTransaction)
	}
}

// WithWriteLock converts "SELECT ..." statement into "SELECT ... FOR UPDATE" statement.
func (conn *DB) WithWriteLock() *DB {
	conn.mustBeInTransaction()
	return newDB(conn.db.Clauses(clause.Locking{Strength: "UPDATE"}).Session(&gorm.Session{}))
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

func (conn *DB) withForeignKeyChecksDisabled(blockFunc func(*DB) error) (err error) {
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
		return conn.inTransaction(innerFunc)
	} // else {
	return innerFunc(conn)
	//}
}

func mustNotBeError(err error) {
	if err != nil {
		panic(err)
	}
}

func recoverPanics(returnErr *error) { // nolint:gocritic
	if p := recover(); p != nil {
		switch e := p.(type) {
		case runtime.Error:
			panic(e)
		default:
			*returnErr = p.(error)
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
