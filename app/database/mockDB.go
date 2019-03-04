package database

import (
	"fmt"
	"os"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

// NewDBMock generate a DB mock the database engine
func NewDBMock() (*DB, sqlmock.Sqlmock) {
	dbMock, mock, err := sqlmock.New()
	if err != nil {
		fmt.Println("Unable to create the mock db: ", err)
		os.Exit(1)
	}

	db, err := Open(dbMock)
	if err != nil {
		fmt.Println("Unable to create the gorm connection to the mock: ", err)
		os.Exit(1)
	}

	return db, mock
}
