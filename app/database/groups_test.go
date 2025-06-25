package database

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/v2/testhelpers/testoutput"
)

func TestDataStore_GetGroupJoiningByCodeInfoByCode_WithLock(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	mock.ExpectBegin()
	mock.ExpectQuery("^SELECT .+ FOR UPDATE$").WillReturnRows(mock.NewRows([]string{}))
	mock.ExpectCommit()
	assert.NoError(t, NewDataStore(db).InTransaction(func(store *DataStore) error {
		_, _, err := store.GetGroupJoiningByCodeInfoByCode("abc", true)
		return err
	}))
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDataStore_GetGroupJoiningByCodeInfoByCode_Error(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	expectedError := errors.New("some error")
	mock.ExpectQuery("^SELECT ").WillReturnError(expectedError)
	group, ok, err := NewDataStore(db).GetGroupJoiningByCodeInfoByCode("abc", false)
	assert.Equal(t, expectedError, err)
	assert.False(t, ok)
	assert.Empty(t, group)
	assert.NoError(t, mock.ExpectationsWereMet())
}
