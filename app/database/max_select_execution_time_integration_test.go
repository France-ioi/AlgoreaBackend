//go:build !unit

package database_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/France-ioi/AlgoreaBackend/v2/app"
	"github.com/France-ioi/AlgoreaBackend/v2/app/appenv"
	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/app/logging"
	"github.com/France-ioi/AlgoreaBackend/v2/app/loggingtest"
	"github.com/France-ioi/AlgoreaBackend/v2/testhelpers"
	"github.com/France-ioi/AlgoreaBackend/v2/testhelpers/testoutput"
)

func TestMaxSelectExecutionTime_CappedQueryReturns3024(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db := openDBWithMaxOpenConns(t, 1)
	defer func() { _ = db.Close() }()

	ctx := database.ContextWithMaxSelectExecutionTime(db.GetContext(), 500*time.Millisecond)
	db = database.NewDataStoreWithContext(ctx, db).DB

	var result []map[string]interface{}
	err := db.Raw("SELECT * FROM (SELECT SLEEP(3) AS s) x").ScanIntoSliceOfMaps(&result).Error()
	require.True(t, database.IsQueryTimeoutError(err), "got: %v", err)
}

func TestMaxSelectExecutionTime_CTEQueryReturns3024(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db := openDBWithMaxOpenConns(t, 1)
	defer func() { _ = db.Close() }()

	ctx := database.ContextWithMaxSelectExecutionTime(db.GetContext(), 500*time.Millisecond)
	db = database.NewDataStoreWithContext(ctx, db).DB

	var result []map[string]interface{}
	err := db.With("t", db.Raw("SELECT SLEEP(3) AS s")).Table("t").
		Select("*").ScanIntoSliceOfMaps(&result).Error()
	// A hint misplaced inside the CTE body is silently ignored and SLEEP(3) completes without error.
	require.True(t, database.IsQueryTimeoutError(err), "got: %v", err)
}

func TestMaxSelectExecutionTime_NoHintWithoutCap(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	logger, loggerHook := logging.NewMockLogger()
	ctx := testhelpers.CreateTestContextWithLogger(logger)
	appenv.ForceTestEnv()
	config := testhelpers.GetConfigFromContext(ctx)
	dbConfig, _ := app.DBConfig(config)
	rawDB, err := database.OpenRawDBConnection(dbConfig.FormatDSN(), true)
	require.NoError(t, err)
	defer func() { _ = rawDB.Close() }()
	rawDB.SetMaxOpenConns(1)

	db, err := database.OpenWithLogConfig(ctx, rawDB, database.LogConfig{LogSQLQueries: true}, true)
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	var result []map[string]interface{}
	require.NoError(t, db.Raw("SELECT 1 AS n").ScanIntoSliceOfMaps(&result).Error())

	logs := (&loggingtest.Hook{Hook: loggerHook}).GetAllLogs()
	assert.Contains(t, logs, "SELECT 1 AS n")
	assert.NotContains(t, logs, "MAX_EXECUTION_TIME")
}

func TestMaxSelectExecutionTime_NoHintInsideTransaction(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	logger, loggerHook := logging.NewMockLogger()
	ctx := testhelpers.CreateTestContextWithLogger(logger)
	appenv.ForceTestEnv()
	config := testhelpers.GetConfigFromContext(ctx)
	dbConfig, _ := app.DBConfig(config)
	rawDB, err := database.OpenRawDBConnection(dbConfig.FormatDSN(), true)
	require.NoError(t, err)
	defer func() { _ = rawDB.Close() }()

	db, err := database.OpenWithLogConfig(ctx, rawDB, database.LogConfig{LogSQLQueries: true}, true)
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	store := database.NewDataStoreWithContext(
		database.ContextWithMaxSelectExecutionTime(db.GetContext(), 100*time.Millisecond), db)

	require.NoError(t, store.InTransaction(func(s *database.DataStore) error {
		var result []map[string]interface{}
		// Would raise 3024 if the hint were injected inside the transaction.
		return s.Raw("SELECT * FROM (SELECT SLEEP(1) AS s) x").ScanIntoSliceOfMaps(&result).Error()
	}))

	logs := (&loggingtest.Hook{Hook: loggerHook}).GetAllLogs()
	assert.NotContains(t, logs, "MAX_EXECUTION_TIME")
}

func openDBWithMaxOpenConns(t *testing.T, maxOpen int) *database.DB {
	t.Helper()
	ctx := testhelpers.CreateTestContext()
	appenv.ForceTestEnv()
	config := testhelpers.GetConfigFromContext(ctx)
	dbConfig, _ := app.DBConfig(config)
	rawDB, err := database.OpenRawDBConnection(dbConfig.FormatDSN(), true)
	require.NoError(t, err)
	rawDB.SetMaxOpenConns(maxOpen)
	db, err := database.Open(ctx, rawDB)
	require.NoError(t, err)
	return db
}
