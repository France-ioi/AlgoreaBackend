package database

import (
	"fmt"
	"os"

	"github.com/jinzhu/gorm"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

// NewDBMock generate a DB mock the database engine
func NewDBMock() (DB, sqlmock.Sqlmock) {
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

	return &db{dbConn}, mock
}
