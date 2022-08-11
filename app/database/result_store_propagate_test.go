package database

import (
	"errors"
	"reflect"
	"regexp"
	"runtime"
	"testing"
	"time"

	"bou.ke/monkey"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestResultStore_Propagate_RecoverError(t *testing.T) {
	db, dbMock := NewDBMock()
	defer func() { _ = db.Close() }()

	expectedError := errors.New("some error")
	dbMock.ExpectBegin()
	dbMock.ExpectQuery("^"+regexp.QuoteMeta("SELECT GET_LOCK(?, ?)")+"$").
		WithArgs("listener_propagate", 10).WillReturnError(expectedError)
	dbMock.ExpectRollback()
	err := NewDataStore(db).InTransaction(func(s *DataStore) error {
		return s.Results().propagate()
	})
	assert.Equal(t, expectedError, err)
	assert.NoError(t, dbMock.ExpectationsWereMet())
}

func TestResultStore_Propagate_RecoverRuntimeError(t *testing.T) {
	db, dbMock := NewDBMock()
	defer func() { _ = db.Close() }()
	dbMock.ExpectBegin()
	dbMock.ExpectRollback()
	monkey.PatchInstanceMethod(reflect.TypeOf(&DataStore{}), "WithNamedLock",
		func(*DataStore, string, time.Duration, func(*DataStore) error) error {
			var a []int
			a[0]++ // nolint:govet // runtime error
			return nil
		})
	defer monkey.UnpatchAll()

	didPanic, panicValue := func() (didPanic bool, panicValue interface{}) {
		defer func() {
			if p := recover(); p != nil {
				didPanic = true
				panicValue = p
			}
		}()

		_ = NewDataStore(db).InTransaction(func(s *DataStore) error {
			return s.Results().propagate()
		})

		return false, nil
	}()

	assert.True(t, didPanic)
	assert.Implements(t, (*runtime.Error)(nil), panicValue)
	assert.Equal(t, "runtime error: index out of range [0] with length 0", panicValue.(error).Error())
	assert.NoError(t, dbMock.ExpectationsWereMet())
}

func TestResultStore_Propagate_ReturnsErrLockWaitTimeoutExceededWhenGetLockTimeouts(t *testing.T) {
	db, dbMock := NewDBMock()
	defer func() { _ = db.Close() }()
	dbMock.ExpectBegin()
	dbMock.ExpectQuery("^"+regexp.QuoteMeta("SELECT GET_LOCK(?, ?)")+"$").
		WithArgs("listener_propagate", 10).
		WillReturnRows(sqlmock.NewRows([]string{"GET_LOCK('listener_propagate', 1)"}).AddRow(int64(0)))
	dbMock.ExpectRollback()

	err := NewDataStore(db).InTransaction(func(s *DataStore) error {
		return s.Results().propagate()
	})
	assert.Equal(t, ErrLockWaitTimeoutExceeded, err)
	assert.NoError(t, dbMock.ExpectationsWereMet())
}

func TestResultStore_Propagate_CannotBeCalledWithoutTransaction(t *testing.T) {
	db, dbMock := NewDBMock()
	defer func() { _ = db.Close() }()

	resultStore := NewDataStore(db).Results()
	didPanic, panicValue := func() (didPanic bool, panicValue interface{}) {
		defer func() {
			if p := recover(); p != nil {
				didPanic = true
				panicValue = p
			}
		}()
		_ = resultStore.propagate()
		return false, nil
	}()

	assert.True(t, didPanic)
	assert.Equal(t, errors.New("should be executed in a transaction"), panicValue)
	assert.NoError(t, dbMock.ExpectationsWereMet())
}
