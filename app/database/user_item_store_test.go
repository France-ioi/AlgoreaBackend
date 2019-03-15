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
	dbMock.ExpectQuery("^" + regexp.QuoteMeta("SELECT GET_LOCK('listener_computeAllUserItems', 1)") + "$").
		WillReturnError(expectedError)
	err := NewDataStore(db).UserItems().ComputeAllUserItems()
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
		_ = (&UserItemStore{}).ComputeAllUserItems()
		return false, nil
	}()

	assert.True(t, didPanic)
	assert.Implements(t, (*runtime.Error)(nil), panicValue)
	assert.Equal(t, "runtime error: invalid memory address or nil pointer dereference", panicValue.(error).Error())
}

func TestUserItemStore_ComputeAllUserItems_ReturnsSilentlyWhenGetLockTimeouts(t *testing.T) {
	db, dbMock := NewDBMock()
	dbMock.ExpectQuery("^" + regexp.QuoteMeta("SELECT GET_LOCK('listener_computeAllUserItems', 1)") + "$").
		WillReturnRows(sqlmock.NewRows([]string{"GET_LOCK('listener_computeAllUserItems', 1)"}).AddRow(int64(0)))
	err := NewDataStore(db).UserItems().ComputeAllUserItems()
	assert.Nil(t, err)
	assert.NoError(t, dbMock.ExpectationsWereMet())
}
