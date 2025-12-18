package database

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPermissionGrantedStore_CreateTemporaryTablesForPermissionsExplanation_RequiresFixedConnection(t *testing.T) {
	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	assert.PanicsWithValue(t, ErrConnNotFixed, func() {
		_, _ = NewDataStore(db).PermissionsGranted().CreateTemporaryTablesForPermissionsExplanation()
	})

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPermissionGrantedStore_ComputePermissionsExplanation_RequiresFixedConnection(t *testing.T) {
	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	assert.PanicsWithValue(t, ErrConnNotFixed, func() {
		_ = NewDataStore(db).PermissionsGranted().ComputePermissionsExplanation(nil)
	})

	assert.NoError(t, mock.ExpectationsWereMet())
}
