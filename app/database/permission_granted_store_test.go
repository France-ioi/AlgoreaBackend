package database

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPermissionGrantedStore_After_MustBeInTransaction(t *testing.T) {
	db, dbMock := NewDBMock()
	defer func() { _ = db.Close() }()

	assert.PanicsWithValue(t, ErrNoTransaction, func() {
		_ = NewDataStore(db).PermissionsGranted().After()
	})

	assert.NoError(t, dbMock.ExpectationsWereMet())
}

func TestPermissionGrantedStore_After_HandlesErrorOfComputeAllAccess(t *testing.T) {
	expectedError := errors.New("some error")

	db, dbMock := NewDBMock()
	defer func() { _ = db.Close() }()
	dbMock.ExpectBegin()
	dbMock.ExpectPrepare("^DROP TEMPORARY TABLE IF EXISTS").WillReturnError(expectedError)
	dbMock.ExpectRollback()

	assert.Equal(t, expectedError, db.inTransaction(func(trDB *DB) error {
		return NewDataStore(trDB).PermissionsGranted().After()
	}))

	assert.NoError(t, dbMock.ExpectationsWereMet())
}
