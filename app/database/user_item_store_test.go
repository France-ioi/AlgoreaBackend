package database

import (
	"errors"
	"regexp"
	"runtime"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestUserItemStore_ComputeAllUserItems_RecoverError(t *testing.T) {
	db, dbMock := NewDBMock()
	expectedError := errors.New("some error")
	dbMock.ExpectBegin()
	dbMock.ExpectQuery("^"+regexp.QuoteMeta("SELECT GET_LOCK(?, ?)")+"$").
		WithArgs("listener_computeAllUserItems", 1).WillReturnError(expectedError)
	dbMock.ExpectRollback()
	err := NewDataStore(db).InTransaction(func(s *DataStore) error {
		return s.UserItems().ComputeAllUserItems()
	})
	assert.Equal(t, expectedError, err)
	assert.NoError(t, dbMock.ExpectationsWereMet())
}

func TestUserItemStore_ComputeAllUserItems_RecoverRuntimeError(t *testing.T) {
	didPanic, panicValue := func() (didPanic bool, panicValue interface{}) {
		defer func() {
			if p := recover(); p != nil {
				didPanic = true
				panicValue = p
			}
		}()
		_ = (&UserItemStore{}).InTransaction(func(s *DataStore) error {
			return s.UserItems().ComputeAllUserItems()
		})
		return false, nil
	}()

	assert.True(t, didPanic)
	assert.Implements(t, (*runtime.Error)(nil), panicValue)
	assert.Equal(t, "runtime error: invalid memory address or nil pointer dereference", panicValue.(error).Error())
}

func TestUserItemStore_ComputeAllUserItems_ReturnsErrLockWaitTimeoutExceededWhenGetLockTimeouts(t *testing.T) {
	db, dbMock := NewDBMock()
	dbMock.ExpectBegin()
	dbMock.ExpectQuery("^"+regexp.QuoteMeta("SELECT GET_LOCK(?, ?)")+"$").
		WithArgs("listener_computeAllUserItems", 1).
		WillReturnRows(sqlmock.NewRows([]string{"GET_LOCK('listener_computeAllUserItems', 1)"}).AddRow(int64(0)))
	dbMock.ExpectRollback()
	err := NewDataStore(db).InTransaction(func(s *DataStore) error {
		return s.UserItems().ComputeAllUserItems()
	})
	assert.Equal(t, ErrLockWaitTimeoutExceeded, err)
	assert.NoError(t, dbMock.ExpectationsWereMet())
}
