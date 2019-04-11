package database

import (
	"github.com/DATA-DOG/go-sqlmock"
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGroupGroupStore_transition_MustBeInTransaction(t *testing.T) {
	db, dbMock := NewDBMock()
	defer func() { _ = db.Close() }()

	assert.PanicsWithValue(t, ErrNoTransaction, func() {
		NewDataStore(db).GroupGroups().Transition(
			AdminCreatesInvitation, 20, []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
		)
	})

	assert.NoError(t, dbMock.ExpectationsWereMet())
}

func TestGroupGroupStore_transition_UsesNamedLock(t *testing.T) {
	db, dbMock := NewDBMock()
	defer func() { _ = db.Close() }()

	dbMock.ExpectBegin()
	dbMock.ExpectQuery("^"+regexp.QuoteMeta("SELECT GET_LOCK(?, ?)")+"$").
		WithArgs("groups_groups", groupsRelationsLockTimeout/time.Second).
		WillReturnRows(sqlmock.NewRows([]string{"SELECT GET_LOCK(?, ?)"}).AddRow(int64(0)))
	dbMock.ExpectRollback()

	_ = NewDataStore(db).InTransaction(func(dataStore *DataStore) (err error) {
		defer recoverPanics(&err)
		dataStore.GroupGroups().Transition(
			AdminCreatesInvitation, 20, []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
		)
		return nil
	})

	assert.NoError(t, dbMock.ExpectationsWereMet())
}
