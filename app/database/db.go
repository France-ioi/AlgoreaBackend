package database

import (
	"fmt"
	"log"
	"os"
	"reflect"
	"strings"

	"github.com/jinzhu/gorm"

	"github.com/France-ioi/AlgoreaBackend/app/types"
)

// DB contains information for current db connection (wraps *gorm.DB)
type DB struct {
	db *gorm.DB
}

// newDB wraps *gorm.DB
func newDB(db *gorm.DB) *DB {
	return &DB{db}
}

// Open connects to the database and tests the connection
// nolint: gosec
func Open(dsnConfig string) (*DB, error) {
	var err error
	var dbConn *gorm.DB
	var driverName = "mysql"
	dbConn, err = gorm.Open(driverName, dsnConfig)
	dbConn.LogMode(true)

	return newDB(dbConn), err
}

// DBLogger is the logger type for the DB logs
type DBLogger interface {
	Print(v ...interface{})
}

// SetLogger sets the logger to use for db and the verbosity level.
// Pass `nil` as logger to disable logging (only errors to stdout)
func (conn *DB) SetLogger(logger DBLogger) {
	if logger == nil {
		conn.db.LogMode(false)
		conn.db.SetLogger(gorm.Logger{LogWriter: log.New(os.Stdout, "\r\n", 0)})
	} else {
		conn.db.LogMode(true)
		conn.db.SetLogger(logger)
	}
}

func (conn *DB) inTransaction(txFunc func(*DB) error) (err error) {
	var txDB = conn.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if p := recover(); p != nil {
			// ensure rollback is executed even in case of panic
			txDB.Rollback()
			panic(p) // re-throw panic after rollback
		} else if err != nil {
			// do not change the err
			txDB = txDB.Rollback()
			if txDB.Error != nil {
				panic(p) // in case of error on rollback, panic
			}
		} else {
			txDB = txDB.Commit() // if err is nil, returns the potential error from commit
			err = txDB.Error
		}
	}()
	err = txFunc(newDB(txDB))
	return err
}

// Close close current db connection.  If database connection is not an io.Closer, returns an error.
func (conn *DB) Close() error {
	return conn.db.Close()
}

// Limit specifies the number of records to be retrieved
func (conn *DB) Limit(limit interface{}) *DB {
	return newDB(conn.db.Limit(limit))
}

// Where returns a new relation, filters records with given conditions, accepts `map`, `struct` or `string` as conditions, refer http://jinzhu.github.io/gorm/crud.html#query
func (conn *DB) Where(query interface{}, args ...interface{}) *DB {
	return newDB(conn.db.Where(query, args...))
}

// Joins specifies Joins conditions
//     db.Joins("JOIN emails ON emails.user_id = users.id AND emails.email = ?", "jinzhu@example.org").Find(&user)
func (conn *DB) Joins(query string, args ...interface{}) *DB {
	return newDB(conn.db.Joins(query, args...))
}

// Or filters records that match before conditions or this one, similar to `Where`
func (conn *DB) Or(query interface{}, args ...interface{}) *DB {
	return newDB(conn.db.Or(query, args...))
}

// Select specifies fields that you want to retrieve from database when querying, by default, will select all fields;
// When creating/updating, specify fields that you want to save to database
func (conn *DB) Select(query interface{}, args ...interface{}) *DB {
	return newDB(conn.db.Select(query, args...))
}

// Table specifies the table you would like to run db operations
func (conn *DB) Table(name string) *DB {
	return newDB(conn.db.Table(name))
}

// Group specifies the group method on the find
func (conn *DB) Group(query string) *DB {
	return newDB(conn.db.Group(query))
}

// Order specifies order when retrieve records from database, set reorder to `true` to overwrite defined conditions
//     db.Order("name DESC")
//     db.Order("name DESC", true) // reorder
//     db.Order(gorm.Expr("name = ? DESC", "first")) // sql expression
func (conn *DB) Order(value interface{}, reorder ...bool) *DB {
	return newDB(conn.db.Order(value, reorder...))
}

// Having specifies HAVING conditions for GROUP BY
func (conn *DB) Having(query interface{}, args ...interface{}) *DB {
	return newDB(conn.db.Having(query, args...))
}

// Union specifies UNION of two queries (receiver UNION query)
func (conn *DB) Union(query interface{}) *DB {
	return newDB(conn.db.New().Raw("? UNION ?", conn.db.QueryExpr(), query))
}

// UnionAll specifies UNION ALL of two queries (receiver UNION ALL query)
func (conn *DB) UnionAll(query interface{}) *DB {
	return newDB(conn.db.New().Raw("? UNION ALL ?", conn.db.QueryExpr(), query))
}

// Raw uses raw sql as conditions
//    db.Raw("SELECT name, age FROM users WHERE name = ?", 3).Scan(&result)
func (conn *DB) Raw(query string, args ...interface{}) *DB {
	// db.Raw("").Joins(...) is a hack for making db.Raw("...").Joins(...) work better
	return newDB(conn.db.Raw("").Joins(query, args...))
}

// Updates update attributes with callbacks, refer: https://jinzhu.github.io/gorm/crud.html#update
func (conn *DB) Updates(values interface{}, ignoreProtectedAttrs ...bool) *DB {
	return newDB(conn.db.Updates(values, ignoreProtectedAttrs...))
}

// UpdateColumn updates attributes without callbacks, refer: https://jinzhu.github.io/gorm/crud.html#update
func (conn *DB) UpdateColumn(attrs ...interface{}) *DB {
	return newDB(conn.db.UpdateColumn(attrs...))
}

// SubQuery returns the query as sub query
func (conn *DB) SubQuery() interface{} {
	return conn.db.SubQuery()
}

// QueryExpr returns the query as expr object
func (conn *DB) QueryExpr() interface{} {
	return conn.db.QueryExpr()
}

// Scan scans value to a struct
func (conn *DB) Scan(dest interface{}) *DB {
	return newDB(conn.db.Scan(dest))
}

// ScanIntoSliceOfMaps scans value into a slice of maps
func (conn *DB) ScanIntoSliceOfMaps(dest *[]map[string]interface{}) *DB {
	rows, err := conn.db.Rows()
	if conn.db.AddError(err) != nil {
		return conn
	}
	cols, err := rows.Columns()
	if conn.db.AddError(err) != nil {
		return conn
	}

	if rows != nil {
		defer func() {
			if conn.db.AddError(rows.Close()) != nil {
				return
			}
		}()
	}

	for rows.Next() {
		columns := make([]interface{}, len(cols))
		columnPointers := make([]interface{}, len(cols))
		for i := range columns {
			columnPointers[i] = &columns[i]
		}

		if err := rows.Scan(columnPointers...); err != nil {
			return conn
		}

		rowMap := make(map[string]interface{})
		for i, columnName := range cols {
			if value, ok := columns[i].([]byte); ok {
				columns[i] = string(value)
			}
			rowMap[columnName] = columns[i]
		}
		*dest = append(*dest, rowMap)
	}

	return conn
}

// Count gets how many records for a model
func (conn *DB) Count(dest interface{}) *DB {
	return newDB(conn.db.Count(dest))
}

// Take returns a record that match given conditions, the order will depend on the database implementation
func (conn *DB) Take(out interface{}, where ...interface{}) *DB {
	return newDB(conn.db.Take(out, where...))
}

// Error returns current errors
func (conn *DB) Error() error {
	return conn.db.Error
}

// insert reads fields from the data struct and insert the values which have been set
// into the given table
func (conn *DB) insert(tableName string, data interface{}) error {
	// introspect data
	dataV := reflect.ValueOf(data)

	// extract data from pointer it is a pointer
	if dataV.Kind() == reflect.Ptr {
		dataV = dataV.Elem()
	}

	// we only accept structs
	if dataV.Kind() != reflect.Struct {
		return fmt.Errorf("insert only accepts structs; got %T", dataV)
	}

	// data for the building the SQL request
	// "INSERT INTO tablename (attributes... ) VALUES (?, ?, NULL, ?, ...)", values...
	var attributes = make([]string, 0, dataV.NumField())
	var valueMarks = make([]string, 0, dataV.NumField())
	var values = make([]interface{}, 0, dataV.NumField())

	typ := dataV.Type()
	for i := 0; i < dataV.NumField(); i++ {
		// gets us a StructField
		field := typ.Field(i)
		sqlParam := strings.Split(field.Tag.Get("sql"), ":")
		if len(sqlParam) == 2 && sqlParam[0] == "column" {
			attrName := sqlParam[1]
			value, null, set := dataV.Field(i).Interface(), false, true
			if val, ok := value.(types.NullableOptional); ok {
				value, null, set = val.AllAttributes()
			}

			// only add non optional value (we suppose they will be understandable by the
			// SQL lib, or optional which are set) and optional value which are set
			if set {
				attributes = append(attributes, attrName)
				if null {
					valueMarks = append(valueMarks, "NULL")
				} else {
					valueMarks = append(valueMarks, "?")
					values = append(values, value)
				}
			}
		}
	}
	query := fmt.Sprintf("INSERT INTO `%s` (%s) VALUES (%s)", tableName, strings.Join(attributes, ", "), strings.Join(valueMarks, ", ")) // nolint: gosec
	return conn.db.Exec(query, values...).Error
}

// Set sets setting by name, which could be used in callbacks, will clone a new db, and update its setting
func (conn *DB) Set(name string, value interface{}) *DB {
	return newDB(conn.db.Set(name, value))
}
