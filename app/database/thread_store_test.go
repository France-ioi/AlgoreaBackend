package database

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestThreadStore_UpdateHelperGroupID_MustBeInTransaction(t *testing.T) {
	db, dbMock := NewDBMock()
	defer func() { _ = db.Close() }()

	assert.PanicsWithValue(t, ErrNoTransaction, func() {
		NewDataStore(db).Threads().UpdateHelperGroupID(0, 1)
	})

	assert.NoError(t, dbMock.ExpectationsWereMet())
}
