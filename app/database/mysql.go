package database

import (
	"database/sql"

	"github.com/go-sql-driver/mysql"
)

// DBConn connects to the database and test the connection
func DBConn(dbconfig mysql.Config) (*sql.DB, error) {

	var db *sql.DB
	db, _ = sql.Open("mysql", dbconfig.FormatDSN()) // failure not expected as it just prepares the database abstraction
	err := db.Ping()

	return db, err
}
