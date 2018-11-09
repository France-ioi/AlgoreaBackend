package database

import (
	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

// DB is a wrapper aroujnd the database connector that can be shared through the app
type DB struct {
	*sqlx.DB
}

// DBConn connects to the database and test the connection
func DBConn(dbconfig mysql.Config) (*DB, error) {

	var db *sqlx.DB
	db, _ = sqlx.Open("mysql", dbconfig.FormatDSN()) // failure not expected as it just prepares the database abstraction
	err := db.Ping()

	return &DB{db}, err
}

func (db *DB) inTransaction(txFunc func(*sqlx.Tx) error) (err error) {
	tx, err := db.Beginx()
	if err != nil {
		return err
	}
	defer func() {
		if p := recover(); p != nil {
			// ensure rollback is executed even in case of panic
			tx.Rollback()
			panic(p) // re-throw panic after rollback
		} else if err != nil {
			tx.Rollback() // do not change the err
		} else {
			err = tx.Commit() // if err is nil, returns the potential error from commit
		}
	}()
	err = txFunc(tx)
	return err
}
