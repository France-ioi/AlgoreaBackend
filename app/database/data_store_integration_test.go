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

	testContext := testhelpers.CreateTestContext()
	rawDB := testhelpers.OpenRawDBConnection(testContext)
	db, err := database.Open(testContext, rawDB)
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	store := database.NewDataStore(db)
	require.NoError(t, store.WithForeignKeyChecksDisabled(func(store *database.DataStore) error {
		assertForeignKeysDBVars(t, store, 1, 0)
		require.NoError(t, store.WithForeignKeyChecksDisabled(func(innerStore *database.DataStore) error {
			assertForeignKeysDBVars(t, innerStore, 2, 0)
			return nil
		}))
		assertForeignKeysDBVars(t, store, 1, 0)
		return nil
	}))
	assertForeignKeysDBVars(t, store, 0, 1)
	require.NoError(t, store.InTransaction(func(store *database.DataStore) error {
		assertForeignKeysDBVars(t, store, 0, 1)
		require.NoError(t, store.WithForeignKeyChecksDisabled(func(innerStore *database.DataStore) error {
			assertForeignKeysDBVars(t, innerStore, 1, 0)
			return nil
		}))
		assertForeignKeysDBVars(t, store, 0, 1)
		return nil
	}))
}

func assertForeignKeysDBVars(t *testing.T, store *database.DataStore, expectedStackCount, expectedForeignKeyChecks int64) {
	t.Helper()

	var result struct {
		StackCount       int64
		ForeignKeyChecks int64
	}
	require.NoError(t,
		store.Raw("SELECT @foreign_key_checks_stack_count AS stack_count, @@SESSION.foreign_key_checks AS foreign_key_checks").
			Scan(&result).Error())
	assert.Equal(t, expectedStackCount, result.StackCount, "wrong @foreign_key_checks_stack_count")
	assert.Equal(t, expectedForeignKeyChecks, result.ForeignKeyChecks, "wrong @@SESSION.foreign_key_checks")
}

func TestDataStore_WithNamedLock_WorksOutsideOfTransaction(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	logger, testHook := logging.NewMockLogger()
	ctx := testhelpers.CreateTestContextWithLogger(logger)
	rawDB := testhelpers.OpenRawDBConnection(ctx)
	db, err := database.OpenWithLogConfig(ctx, rawDB, database.LogConfig{LogSQLQueries: false}, true)
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	s := database.NewDataStore(db)
	resultCh := make(chan struct{}, 100)
	defer close(resultCh)
	for i := 0; i < 100; i++ {
		go func() {
			assert.NoError(t, s.WithNamedLock("test", 1*time.Second, func(_ *database.DataStore) error {
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
