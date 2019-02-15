package database

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/jinzhu/gorm"

	"github.com/France-ioi/AlgoreaBackend/app/logging"
	"github.com/France-ioi/AlgoreaBackend/app/types"
)

// DB is the database connector that can be shared through the app
type DB interface {
	inTransaction(txFunc func(DB) error) error
	insert(tableName string, data interface{}) error
	table(string) DB

	Or(query interface{}, args ...interface{}) DB
	Select(query interface{}, args ...interface{}) DB
	Where(query interface{}, args ...interface{}) DB
	Joins(query string, args ...interface{}) DB
	Group(query string) DB
	Having(query interface{}, args ...interface{}) DB
	Order(value interface{}, reorder ...bool) DB
	Union(query interface{}) DB
	Raw(query string, args ...interface{}) DB

	Query() interface{}
	SubQuery() interface{}
	Scan(dest interface{}) DB
	Count(dest interface{}) DB
	Take(out interface{}, where ...interface{}) DB

	Error() error
}

type db struct {
	*gorm.DB
}

// Open connects to the database and tests the connection
// nolint: gosec
func Open(dsnConfig string) (DB, error) {
	var err error
	var dbConn *gorm.DB
	var driverName = "mysql"
	dbConn, err = gorm.Open(driverName, dsnConfig)
	dbConn.LogMode(true)

	// setup logging
	dbConn.SetLogger(logging.Logger.WithField("module", "database"))

	return &db{dbConn}, err
}

func (conn *db) inTransaction(txFunc func(DB) error) (err error) {
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
	err = txFunc(&db{txDB})
	return err
}

func (conn *db) table(tableName string) DB {
	return &db{conn.DB.Table(tableName)}
}

func (conn *db) Where(query interface{}, args ...interface{}) DB {
	return &db{conn.DB.Where(query, args...)}
}

func (conn *db) Joins(query string, args ...interface{}) DB {
	return &db{conn.DB.Joins(query, args...)}
}

func (conn *db) Or(query interface{}, args ...interface{}) DB {
	return &db{conn.DB.Or(query, args...)}
}

func (conn *db) Select(query interface{}, args ...interface{}) DB {
	return &db{conn.DB.Select(query, args...)}
}

func (conn *db) Group(query string) DB {
	return &db{conn.DB.Group(query)}
}

func (conn *db) Order(value interface{}, reorder ...bool) DB {
	return &db{conn.DB.Order(value, reorder...)}
}

func (conn *db) Having(query interface{}, args ...interface{}) DB {
	return &db{conn.DB.Having(query, args...)}
}

func (conn *db) Union(query interface{}) DB {
	return &db{conn.DB.New().Raw("? UNION (?)", conn.DB.QueryExpr(), query)}
}

func (conn *db) Raw(query string, args ...interface{}) DB {
	return &db{conn.DB.Raw(query, args...)}
}

func (conn *db) SubQuery() interface{} {
	return conn.DB.SubQuery()
}

func (conn *db) Query() interface{} {
	return conn.DB.QueryExpr()
}

func (conn *db) Scan(dest interface{}) DB {
	return &db{conn.DB.Scan(dest)}
}

func (conn *db) Count(dest interface{}) DB {
	return &db{conn.DB.Count(dest)}
}

func (conn *db) Take(out interface{}, where ...interface{}) DB {
	return &db{conn.DB.Take(out, where...)}
}

func (conn *db) Error() error {
	return conn.DB.Error
}

// insert reads fields from the data struct and insert the values which have been set
// into the given table
func (conn *db) insert(tableName string, data interface{}) error {
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
