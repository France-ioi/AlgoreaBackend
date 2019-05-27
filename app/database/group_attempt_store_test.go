package database

import (
	"errors"
	"reflect"
	"testing"

	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"
)

func TestGroupAttemptStore_After_MustBeInTransaction(t *testing.T) {
	db, dbMock := NewDBMock()
	defer func() { _ = db.Close() }()

	assert.PanicsWithValue(t, ErrNoTransaction, func() {
		_ = NewDataStore(db).GroupAttempts().After()
	})

	assert.NoError(t, dbMock.ExpectationsWereMet())
}

func TestGroupAttemptStore_HandlesErrorOfPropagateAttempts(t *testing.T) {
	expectedError := errors.New("some error")
	monkey.PatchInstanceMethod(reflect.TypeOf(&UserItemStore{}), "PropagateAttempts",
		func(*UserItemStore) error { return expectedError })
	defer monkey.UnpatchAll()

	db, dbMock := NewDBMock()
	defer func() { _ = db.Close() }()
	dbMock.ExpectBegin()
	dbMock.ExpectCommit()

	assert.NoError(t, db.inTransaction(func(trDB *DB) error {
		err := NewDataStore(trDB).GroupAttempts().After()
		assert.Equal(t, expectedError, err)
		return nil
	}))

	assert.NoError(t, dbMock.ExpectationsWereMet())
}

func TestGroupAttemptStore_HandlesErrorOfComputeAllUserItems(t *testing.T) {
	expectedError := errors.New("some error")
	monkey.PatchInstanceMethod(reflect.TypeOf(&UserItemStore{}), "PropagateAttempts",
		func(*UserItemStore) error { return nil })
	monkey.PatchInstanceMethod(reflect.TypeOf(&UserItemStore{}), "ComputeAllUserItems",
		func(*UserItemStore) error { return expectedError })
	defer monkey.UnpatchAll()

	db, dbMock := NewDBMock()
	defer func() { _ = db.Close() }()
	dbMock.ExpectBegin()
	dbMock.ExpectCommit()

	assert.NoError(t, db.inTransaction(func(trDB *DB) error {
		err := NewDataStore(trDB).GroupAttempts().After()
		assert.Equal(t, expectedError, err)
		return nil
	}))

	assert.NoError(t, dbMock.ExpectationsWereMet())
}

func TestGroupAttemptStore_CallBothMethods(t *testing.T) {
	var propagateAttemptsCalled, computeAllUserItemsCalled int
	monkey.PatchInstanceMethod(reflect.TypeOf(&UserItemStore{}), "PropagateAttempts",
		func(*UserItemStore) error { propagateAttemptsCalled++; return nil })
	monkey.PatchInstanceMethod(reflect.TypeOf(&UserItemStore{}), "ComputeAllUserItems",
		func(*UserItemStore) error { computeAllUserItemsCalled++; return nil })
	defer monkey.UnpatchAll()

	db, dbMock := NewDBMock()
	defer func() { _ = db.Close() }()
	dbMock.ExpectBegin()
	dbMock.ExpectCommit()

	assert.NoError(t, db.inTransaction(func(trDB *DB) error {
		err := NewDataStore(trDB).GroupAttempts().After()
		assert.NoError(t, err)
		return nil
	}))

	assert.Equal(t, 1, propagateAttemptsCalled)
	assert.Equal(t, 1, computeAllUserItemsCalled)
	assert.NoError(t, dbMock.ExpectationsWereMet())
}

func TestGroupAttemptStore_CreateNew_MustBeInTransaction(t *testing.T) {
	db, dbMock := NewDBMock()
	defer func() { _ = db.Close() }()

	assert.PanicsWithValue(t, ErrNoTransaction, func() {
		_, _ = NewDataStore(db).GroupAttempts().CreateNew(10, 20)
	})

	assert.NoError(t, dbMock.ExpectationsWereMet())
}
