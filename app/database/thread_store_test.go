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

func TestThreadStore_UserCanChangeStatus_EdgeCases(t *testing.T) {
	db, _ := NewDBMock()
	user := User{}

	assert.Equal(t, false, NewDataStore(db).Threads().UserCanChangeStatus(&user, "", "", 1, 1))
	assert.Equal(t, true, NewDataStore(db).Threads().UserCanChangeStatus(
		&user, "waiting_for_trainer", "waiting_for_trainer", 1, 1))
	assert.Equal(t, true, NewDataStore(db).Threads().UserCanChangeStatus(
		&user, "closed", "closed", 1, 1))
}
