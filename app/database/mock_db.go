//go:build !prod

package database

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"

	"github.com/DATA-DOG/go-sqlmock"

	"github.com/France-ioi/AlgoreaBackend/v2/app/logging"
)

// NewDBMock generates a DB mock for the database engine with the given context if provided,
// or a new context with a mock logger if no context is provided.
// Note that the context must have a logger (set by logging.ContextWithLogger),
// so if you pass a context without a logger, it will panic.
func NewDBMock(optionalCtx ...context.Context) (*DB, sqlmock.Sqlmock) {
	dbMock, mock := createSQLMock()

	var ctx context.Context
	if len(optionalCtx) == 0 {
		ctx, _, _ = logging.NewContextWithNewMockLogger()
	} else {
		ctx = optionalCtx[0]
	}
	db, err := Open(ctx, dbMock)
	if err != nil {
		panic(fmt.Errorf("unable to create the gorm connection to the mock: %w", err))
	}

	return db, mock
}

func createSQLMock() (*sql.DB, sqlmock.Sqlmock) {
	dbMock, mock, err := sqlmock.New()
	if err != nil {
		panic(fmt.Errorf("unable to create the mock db: %w", err))
	}
	return dbMock, mock
}

// NewDBMockWithLogConfig generate a DB mock the database engine with the given logging configuration.
// Note that the context must have a logger (set by logging.ContextWithLogger),
// so if you pass a context without a logger, it will panic.
func NewDBMockWithLogConfig(ctx context.Context, logConfig LogConfig, rawSQLQueriesLoggingEnabled bool) (*DB, sqlmock.Sqlmock) {
	dbMock, mock := createSQLMock()

	db, err := OpenWithLogConfig(ctx, dbMock, logConfig, rawSQLQueriesLoggingEnabled)
	if err != nil {
		panic(fmt.Errorf("unable to create the gorm connection to the mock: %w", err))
	}

	return db, mock
}

// MockDBEnumQueries stubs all the db queries for loading permission enums.
func MockDBEnumQueries(sqlMock sqlmock.Sqlmock) {
	sqlMock.ExpectQuery("^"+regexp.QuoteMeta(
		"SELECT SUBSTRING(COLUMN_TYPE, 6, LENGTH(COLUMN_TYPE)-6) FROM `information_schema`.`COLUMNS`  "+
			"WHERE (TABLE_SCHEMA = DATABASE()) AND (TABLE_NAME = ?) AND (COLUMN_NAME = ?) LIMIT 1")+"$").
		WithArgs("permissions_granted", "can_view").
		WillReturnRows(sqlMock.NewRows([]string{"value"}).
			AddRow("'none','info','content','content_with_descendants','solution'"))
	sqlMock.ExpectQuery("^"+regexp.QuoteMeta(
		"SELECT SUBSTRING(COLUMN_TYPE, 6, LENGTH(COLUMN_TYPE)-6) FROM `information_schema`.`COLUMNS`  "+
			"WHERE (TABLE_SCHEMA = DATABASE()) AND (TABLE_NAME = ?) AND (COLUMN_NAME = ?) LIMIT 1")+"$").
		WithArgs("permissions_granted", "can_grant_view").
		WillReturnRows(sqlMock.NewRows([]string{"value"}).
			AddRow("'none','enter','content','content_with_descendants','solution','solution_with_grant'"))
	sqlMock.ExpectQuery("^"+regexp.QuoteMeta(
		"SELECT SUBSTRING(COLUMN_TYPE, 6, LENGTH(COLUMN_TYPE)-6) FROM `information_schema`.`COLUMNS`  "+
			"WHERE (TABLE_SCHEMA = DATABASE()) AND (TABLE_NAME = ?) AND (COLUMN_NAME = ?) LIMIT 1")+"$").
		WithArgs("permissions_granted", "can_watch").
		WillReturnRows(sqlMock.NewRows([]string{"value"}).
			AddRow("'none','result','answer','answer_with_grant'"))
	sqlMock.ExpectQuery("^"+regexp.QuoteMeta(
		"SELECT SUBSTRING(COLUMN_TYPE, 6, LENGTH(COLUMN_TYPE)-6) FROM `information_schema`.`COLUMNS`  "+
			"WHERE (TABLE_SCHEMA = DATABASE()) AND (TABLE_NAME = ?) AND (COLUMN_NAME = ?) LIMIT 1")+"$").
		WithArgs("permissions_granted", "can_edit").
		WillReturnRows(sqlMock.NewRows([]string{"value"}).
			AddRow("'none','children','all','all_with_grant'"))
	sqlMock.ExpectQuery("^"+regexp.QuoteMeta(
		"SELECT SUBSTRING(COLUMN_TYPE, 6, LENGTH(COLUMN_TYPE)-6) FROM `information_schema`.`COLUMNS`  "+
			"WHERE (TABLE_SCHEMA = DATABASE()) AND (TABLE_NAME = ?) AND (COLUMN_NAME = ?) LIMIT 1")+"$").
		WithArgs("group_managers", "can_manage").
		WillReturnRows(sqlMock.NewRows([]string{"value"}).
			AddRow("'none','memberships','memberships_and_group'"))
	sqlMock.ExpectQuery("^"+regexp.QuoteMeta(
		"SELECT SUBSTRING(COLUMN_TYPE, 6, LENGTH(COLUMN_TYPE)-6) FROM `information_schema`.`COLUMNS`  "+
			"WHERE (TABLE_SCHEMA = DATABASE()) AND (TABLE_NAME = ?) AND (COLUMN_NAME = ?) LIMIT 1")+"$").
		WithArgs("items_items", "content_view_propagation").
		WillReturnRows(sqlMock.NewRows([]string{"value"}).
			AddRow("'none','as_info','as_content'"))
	sqlMock.ExpectQuery("^"+regexp.QuoteMeta(
		"SELECT SUBSTRING(COLUMN_TYPE, 6, LENGTH(COLUMN_TYPE)-6) FROM `information_schema`.`COLUMNS`  "+
			"WHERE (TABLE_SCHEMA = DATABASE()) AND (TABLE_NAME = ?) AND (COLUMN_NAME = ?) LIMIT 1")+"$").
		WithArgs("items_items", "upper_view_levels_propagation").
		WillReturnRows(sqlMock.NewRows([]string{"value"}).
			AddRow("'use_content_view_propagation','as_content_with_descendants','as_is'"))
}

// ClearAllDBEnums clears all cached permission enums.
func ClearAllDBEnums() {
	enumsMutex.Lock()
	defer enumsMutex.Unlock()

	enumValueIndex2Name = nil
	enumValueName2Index = nil
}
