package database

import (
  "fmt"
  "reflect"
  "strings"

  t "github.com/France-ioi/AlgoreaBackend/app/types"
  "github.com/France-ioi/AlgoreaBackend/app/logging"
  "github.com/France-ioi/AlgoreaBackend/app/config"
  "github.com/jmoiron/sqlx"
  "github.com/jinzhu/gorm"
)

// DB is a wrapper around the database connector that can be shared through the app
type DB struct {
  *sqlx.DB
  gorm           *gorm.DB
  DriverName     string
  DataSourceName string
}

// Tx is a wrapper around a database transaction
type Tx struct {
  *sqlx.Tx
}

const dbStructTag string = "db"

// DBConn connects to the database and test the connection
// nolint: gosec
func DBConn(dbConfig config.Database) (*DB, error) {
  var err error
  var db *sqlx.DB
  var driverName = "mysql"
  var dataSourceName = dbConfig.Connection.FormatDSN()

  // failure not expected as it just prepares the database abstraction
  db, _ = sqlx.Open(driverName, dataSourceName)
  gorm, _ := gorm.Open(driverName, dataSourceName)
  err = db.Ping()

  // setup logging
  gorm.SetLogger(logging.Logger)

  return &DB{db, gorm, driverName, dataSourceName}, err
}

func (db *DB) inTransaction(txFunc func(Tx) error) (err error) {
  var tx *sqlx.Tx
  tx, err = db.Beginx()
  if err != nil {
    return err
  }
  defer func() {
    if p := recover(); p != nil {
      // ensure rollback is executed even in case of panic
      _ = tx.Rollback() // nolint: gosec
      panic(p)          // re-throw panic after rollback
    } else if err != nil {
      // do not change the err
      if tx.Rollback() != nil {
        panic(p) // in case of eror on rollback, panic
      }
    } else {
      err = tx.Commit() // if err is nil, returns the potential error from commit
    }
  }()
  err = txFunc(Tx{tx})
  return err
}

// insert reads fields from the data struct and insert the values which have been set
// into the given table
func (tx Tx) insert(tableName string, data interface{}) error {
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
    if attrName := field.Tag.Get(dbStructTag); attrName != "" {
      value := dataV.Field(i).Interface()
      null := false
      skip := false
      switch value.(type) {
      case t.Int64:
        val := value.(t.Int64)
        value, null, skip = val.Value, val.Null, !val.Set
      case t.String:
        val := value.(t.String)
        value, null, skip = val.Value, val.Null, !val.Set
      }
      // only add non optional value (we suppose they will be understandable by the
      // SQL lib, or optional which are set) and optional value which are set
      if !skip {
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
  _, err := tx.Exec(query, values...)
  return err
}
