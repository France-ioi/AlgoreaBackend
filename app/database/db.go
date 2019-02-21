package database

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/jinzhu/gorm"

	"github.com/France-ioi/AlgoreaBackend/app/logging"
	"github.com/France-ioi/AlgoreaBackend/app/types"
)

// DB contains information for current db connection (wraps *gorm.DB)
type DB struct {
	*gorm.DB
}

// NewDB wraps *gorm.DB
func NewDB(db *gorm.DB) *DB {
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

	// setup logging
	dbConn.SetLogger(logging.Logger.WithField("module", "database"))

	return NewDB(dbConn), err
}

func (conn *DB) inTransaction(txFunc func(*DB) error) (err error) {
	var txDB = conn.Begin()
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
				panic(p) // in case of eror on rollback, panic
			}
		} else {
			txDB = txDB.Commit() // if err is nil, returns the potential error from commit
			err = txDB.Error
		}
	}()
	err = txFunc(NewDB(txDB))
	return err
}

// Limit specifies the number of records to be retrieved
func (conn *DB) Limit(limit interface{}) *DB {
	return NewDB(conn.DB.Limit(limit))
}

// Where returns a new relation, filters records with given conditions, accepts `map`, `struct` or `string` as conditions, refer http://jinzhu.github.io/gorm/crud.html#query
func (conn *DB) Where(query interface{}, args ...interface{}) *DB {
	return NewDB(conn.DB.Where(query, args...))
}

// Joins specifies Joins conditions
//     db.Joins("JOIN emails ON emails.user_id = users.id AND emails.email = ?", "jinzhu@example.org").Find(&user)
func (conn *DB) Joins(query string, args ...interface{}) *DB {
	return NewDB(conn.DB.Joins(query, args...))
}

// Or filters records that match before conditions or this one, similar to `Where`
func (conn *DB) Or(query interface{}, args ...interface{}) *DB {
	return NewDB(conn.DB.Or(query, args...))
}

// Select specifies fields that you want to retrieve from database when querying, by default, will select all fields;
// When creating/updating, specify fields that you want to save to database
func (conn *DB) Select(query interface{}, args ...interface{}) *DB {
	return NewDB(conn.DB.Select(query, args...))
}

// Table specifies the table you would like to run db operations
func (conn *DB) Table(name string) *DB {
	return NewDB(conn.DB.Table(name))
}

// Group specifies the group method on the find
func (conn *DB) Group(query string) *DB {
	return NewDB(conn.DB.Group(query))
}

// Order specifies order when retrieve records from database, set reorder to `true` to overwrite defined conditions
//     db.Order("name DESC")
//     db.Order("name DESC", true) // reorder
//     db.Order(gorm.Expr("name = ? DESC", "first")) // sql expression
func (conn *DB) Order(value interface{}, reorder ...bool) *DB {
	return NewDB(conn.DB.Order(value, reorder...))
}

// Having specifies HAVING conditions for GROUP BY
func (conn *DB) Having(query interface{}, args ...interface{}) *DB {
	return NewDB(conn.DB.Having(query, args...))
}

// Union specifies UNION of two queries (receiver UNION query)
func (conn *DB) Union(query interface{}) *DB {
	return NewDB(conn.DB.New().Raw("? UNION ?", conn.DB.QueryExpr(), query))
}

// UnionAll specifies UNION ALL of two queries (receiver UNION ALL query)
func (conn *DB) UnionAll(query interface{}) *DB {
	return NewDB(conn.DB.New().Raw("? UNION ALL ?", conn.DB.QueryExpr(), query))
}

// Raw uses raw sql as conditions
//    db.Raw("SELECT name, age FROM users WHERE name = ?", 3).Scan(&result)
func (conn *DB) Raw(query string, args ...interface{}) *DB {
	// db.Raw("").Joins(...) is a hack for making db.Raw("...").Joins(...) work better
	return NewDB(conn.DB.Raw("").Joins(query, args...))
}

// SubQuery returns the query as sub query
func (conn *DB) SubQuery() interface{} {
	return conn.DB.SubQuery()
}

// QueryExpr returns the query as expr object
func (conn *DB) QueryExpr() interface{} {
	return conn.DB.QueryExpr()
}

// Scan scans value to a struct
func (conn *DB) Scan(dest interface{}) *DB {
	return NewDB(conn.DB.Scan(dest))
}

// ScanIntoSliceOfMaps scans value into a slice of maps
func (conn *DB) ScanIntoSliceOfMaps(dest *[]map[string]interface{}) *DB {
	rows, err := conn.DB.Rows()
	if conn.DB.AddError(err) != nil {
		return conn
	}
	cols, err := rows.Columns()
	if conn.DB.AddError(err) != nil {
		return conn
	}

	if rows != nil {
		defer func() {
			if conn.DB.AddError(rows.Close()) != nil {
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
	return NewDB(conn.DB.Count(dest))
}

// Take returns a record that match given conditions, the order will depend on the database implementation
func (conn *DB) Take(out interface{}, where ...interface{}) *DB {
	return NewDB(conn.DB.Take(out, where...))
}

// Error returns current errors
func (conn *DB) Error() error {
	return conn.DB.Error
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
	return conn.Exec(query, values...).Error
}
