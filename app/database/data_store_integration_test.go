//go:build !unit

package database_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/testhelpers"
)

func TestDataStore_WithForeignKeyChecksDisabled(t *testing.T) {
	rawDB, err := testhelpers.OpenRawDBConnection()
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	db, err := database.Open(rawDB)
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	s := database.NewDataStore(db)
	assert.NoError(t, s.WithForeignKeyChecksDisabled(func(store *database.DataStore) error {
		assertForeignKeysDBVars(t, store, 1, 0)
		assert.NoError(t, store.WithForeignKeyChecksDisabled(func(innerStore *database.DataStore) error {
			assertForeignKeysDBVars(t, innerStore, 2, 0)
			return nil
		}))
		assertForeignKeysDBVars(t, store, 1, 0)
		return nil
	}))
	assertForeignKeysDBVars(t, s, 0, 1)
	assert.NoError(t, s.InTransaction(func(store *database.DataStore) error {
		assertForeignKeysDBVars(t, store, 0, 1)
		assert.NoError(t, store.WithForeignKeyChecksDisabled(func(innerStore *database.DataStore) error {
			assertForeignKeysDBVars(t, innerStore, 1, 0)
			return nil
		}))
		assertForeignKeysDBVars(t, store, 0, 1)
		return nil
	}))
}

func assertForeignKeysDBVars(t *testing.T, store *database.DataStore, expectedStackCount, expectedForeignKeyChecks int64) {
	var result struct {
		StackCount       int64
		ForeignKeyChecks int64
	}
	assert.NoError(t,
		store.Raw("SELECT @foreign_key_checks_stack_count AS stack_count, @@SESSION.foreign_key_checks AS foreign_key_checks").
			Scan(&result).Error())
	assert.Equal(t, expectedStackCount, result.StackCount, "wrong @foreign_key_checks_stack_count")
	assert.Equal(t, expectedForeignKeyChecks, result.ForeignKeyChecks, "wrong @@SESSION.foreign_key_checks")
}
