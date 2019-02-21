package database

import (
	"fmt"
	"os"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/jinzhu/gorm"
)

// NewDBMock generate a DB mock the database engine
func NewDBMock() (*DB, sqlmock.Sqlmock) {
	dbMock, mock, err := sqlmock.New()
	if err != nil {
		fmt.Println("Unable to create the mock db: ", err)
		os.Exit(1)
	}

	dbConn, err := gorm.Open("mysql", dbMock)
	if err != nil {
		fmt.Println("Unable to create the gorm connection to the mock: ", err)
		os.Exit(1)
	}

	return &DB{dbConn}, mock
}
