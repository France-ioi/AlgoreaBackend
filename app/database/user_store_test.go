package database

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUserStore_deleteWithTraps_DoesNothingWhenScopeReturnsNothing(t *testing.T) {
	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	mock.ExpectBegin()
	mock.ExpectQuery("^" + regexp.QuoteMeta("SELECT group_id, owned_group_id FROM `users` LIMIT 1000 FOR UPDATE") + "$").
		WillReturnRows(mock.NewRows([]string{"group_id", "owned_group_id"}))
	mock.ExpectCommit()

	assert.NoError(t, NewDataStore(db).InTransaction(func(store *DataStore) error {
		userStore := store.Users()
		cnt := userStore.deleteWithTraps(userStore.DB)
		assert.Zero(t, cnt)
		return nil
	}))
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserStore_executeBatchesInTransactions_ProcessesAllTheBatches(t *testing.T) {
	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	mock.ExpectBegin()
	mock.ExpectCommit()
	mock.ExpectBegin()
	mock.ExpectCommit()

	counts := []int{1000, 999}
	step := 0
	totalCount := 0
	NewDataStore(db).Users().executeBatchesInTransactions(func(store *DataStore) int {
		store.mustBeInTransaction()

		totalCount += counts[step]
		step++
		return counts[step-1]
	})
	assert.Equal(t, 1999, totalCount)
	assert.NoError(t, mock.ExpectationsWereMet())
}
