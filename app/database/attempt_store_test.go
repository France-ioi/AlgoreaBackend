package database

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAttemptStore_CreateNew_MustBeInTransaction(t *testing.T) {
	db, dbMock := NewDBMock()
	defer func() { _ = db.Close() }()

	assert.PanicsWithValue(t, ErrNoTransaction, func() {
		_, _ = NewDataStore(db).Attempts().CreateNew(10, 20, 100)
	})

	assert.NoError(t, dbMock.ExpectationsWereMet())
}
