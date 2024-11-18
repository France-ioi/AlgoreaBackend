//go:build !unit

package database_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/app/logging"
	"github.com/France-ioi/AlgoreaBackend/v2/app/loggingtest"
	"github.com/France-ioi/AlgoreaBackend/v2/testhelpers"
	"github.com/France-ioi/AlgoreaBackend/v2/testhelpers/testoutput"
)

func TestDataStore_WithForeignKeyChecksDisabled(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	rawDB, err := testhelpers.OpenRawDBConnection()
	require.NoError(t, err)
	db, err := database.Open(rawDB)
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

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

func TestDataStore_WithNamedLock_WorksOutsideOfTransaction(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	testHook, restoreFunc := logging.MockSharedLoggerHook()
	defer restoreFunc()
	rawDB, err := testhelpers.OpenRawDBConnection()
	require.NoError(t, err)
	db, err := database.OpenWithLogConfig(rawDB,
		database.LogConfig{Logger: logging.NewStructuredDBLogger(), LogSQLQueries: false}, true)
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	s := database.NewDataStore(db)
	resultCh := make(chan struct{}, 100)
	defer close(resultCh)
	for i := 0; i < 100; i++ {
		go func() {
			assert.NoError(t, s.WithNamedLock("test", 1*time.Second, func(store *database.DataStore) error {
				return nil
			}))
			resultCh <- struct{}{}
		}()
	}
	for i := 0; i < 100; i++ {
		<-resultCh
	}

	logEntries := (&loggingtest.Hook{Hook: testHook}).GetAllStructuredLogs()
	assert.Empty(t, logEntries)
}
