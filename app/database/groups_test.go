package database

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDataStore_GetGroupJoiningByCodeInfoByCode_WithLock(t *testing.T) {
	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	mock.ExpectBegin()
	mock.ExpectQuery("^SELECT .+ FOR UPDATE$").WillReturnRows(mock.NewRows([]string{}))
	mock.ExpectCommit()
	assert.NoError(t, NewDataStore(db).InTransaction(func(store *DataStore) error {
		_, err := store.GetGroupJoiningByCodeInfoByCode("abc", true)
		return err
	}))
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDataStore_GetGroupJoiningByCodeInfoByCode_Error(t *testing.T) {
	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	expectedError := errors.New("some error")
	mock.ExpectQuery("^SELECT ").WillReturnError(expectedError)
	_, err := NewDataStore(db).GetGroupJoiningByCodeInfoByCode("abc", false)
	assert.Equal(t, expectedError, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
