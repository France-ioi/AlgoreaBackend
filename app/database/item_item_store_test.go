package database

import (
	"errors"
	"regexp"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestItemItemStore_ChildrenOf(t *testing.T) {
	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	const parentItemID = 123
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `items_items` WHERE (items_items.item_parent_id=?)")).
		WithArgs(parentItemID).
		WillReturnRows(mock.NewRows([]string{"id"}))

	store := NewDataStore(db).ItemItems()
	newStore := store.ChildrenOf(parentItemID)
	assert.NotEqual(t, store, newStore)
	assert.Equal(t, "items_items", newStore.DataStore.tableName)

	var result []interface{}
	err := newStore.Scan(&result).Error()
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestItemItemStore_After_MustBeInTransaction(t *testing.T) {
	db, dbMock := NewDBMock()
	defer func() { _ = db.Close() }()

	assert.PanicsWithValue(t, ErrNoTransaction, func() {
		_ = NewDataStore(db).ItemItems().After()
	})

	assert.NoError(t, dbMock.ExpectationsWereMet())
}

func TestItemItemStore_After_HandlesErrorOfCreateNewAncestors(t *testing.T) {
	expectedError := errors.New("some error")

	db, dbMock := NewDBMock()
	defer func() { _ = db.Close() }()
	dbMock.ExpectBegin()
	dbMock.ExpectExec("^INSERT INTO  items_propagate").WillReturnError(expectedError)
	dbMock.ExpectRollback()

	assert.Equal(t, expectedError, db.inTransaction(func(trDB *DB) error {
		return NewDataStore(trDB).ItemItems().After()
	}))

	assert.NoError(t, dbMock.ExpectationsWereMet())
}

func TestItemItemStore_After_HandlesErrorOfComputeAllAccess(t *testing.T) {
	expectedError := errors.New("some error")

	db, dbMock := NewDBMock()
	defer func() { _ = db.Close() }()
	dbMock.ExpectBegin()
	dbMock.ExpectExec("^INSERT INTO  items_propagate").WillReturnResult(sqlmock.NewResult(0, 0))
	dbMock.ExpectPrepare("UPDATE items_propagate")
	dbMock.ExpectPrepare("INSERT IGNORE INTO items_ancestors")
	dbMock.ExpectPrepare("INSERT IGNORE INTO items_ancestors")
	dbMock.ExpectPrepare("INSERT IGNORE INTO items_ancestors")
	dbMock.ExpectPrepare("UPDATE items_propagate")
	dbMock.ExpectExec("UPDATE items_propagate").WillReturnResult(sqlmock.NewResult(0, 0))
	dbMock.ExpectExec("INSERT IGNORE INTO items_ancestors").WillReturnResult(sqlmock.NewResult(0, 0))
	dbMock.ExpectExec("INSERT IGNORE INTO items_ancestors").WillReturnResult(sqlmock.NewResult(0, 0))
	dbMock.ExpectExec("INSERT IGNORE INTO items_ancestors").WillReturnResult(sqlmock.NewResult(0, 0))
	dbMock.ExpectExec("UPDATE items_propagate").WillReturnResult(sqlmock.NewResult(0, 0))
	dbMock.ExpectPrepare("INSERT IGNORE INTO groups_items").WillReturnError(expectedError)
	dbMock.ExpectRollback()

	assert.Equal(t, expectedError, db.inTransaction(func(trDB *DB) error {
		return NewDataStore(trDB).ItemItems().After()
	}))

	assert.NoError(t, dbMock.ExpectationsWereMet())
}
