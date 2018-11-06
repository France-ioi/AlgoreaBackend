package database

import (
	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

// DBConn connects to the database and test the connection
func DBConn(dbconfig mysql.Config) (*sqlx.DB, error) {

	var db *sqlx.DB
	db, _ = sqlx.Open("mysql", dbconfig.FormatDSN()) // failure not expected as it just prepares the database abstraction
	err := db.Ping()

	return db, err
}
