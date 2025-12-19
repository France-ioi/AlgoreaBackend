//go:build !unit

package database_test

import (
	"testing"

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

func Test_SQLQueryAnalyzing(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	for _, testCase := range []struct {
		name       string
		funcToCall func(db *database.DB) error
	}{
		{
			name: "fixed conn Exec",
			funcToCall: func(db *database.DB) error {
				return db.WithFixedConnection(func(db *database.DB) error {
					return db.Exec("SELECT 1").Error()
				})
			},
		},
		{
			name: "fixed conn Query",
			funcToCall: func(db *database.DB) error {
				return db.WithFixedConnection(func(db *database.DB) error {
					var result []map[string]interface{}
					return db.Raw("SELECT 1").ScanIntoSliceOfMaps(&result).Error()
				})
			},
		},
		{
			name: "usual conn Exec",
			funcToCall: func(db *database.DB) error {
				return db.Exec("SELECT 1").Error()
			},
		},
		{
			name: "usual conn Query",
			funcToCall: func(db *database.DB) error {
				var result []map[string]interface{}
				return db.Raw("SELECT 1").ScanIntoSliceOfMaps(&result).Error()
			},
		},
		{
			name: "tx Exec",
			funcToCall: func(db *database.DB) error {
				return database.NewDataStore(db).InTransaction(func(store *database.DataStore) error {
					return store.Exec("SELECT 1").Error()
				})
			},
		},
		{
			name: "tx Query",
			funcToCall: func(db *database.DB) error {
				return database.NewDataStore(db).InTransaction(func(store *database.DataStore) error {
					var result []map[string]interface{}
					return store.Raw("SELECT 1").ScanIntoSliceOfMaps(&result).Error()
				})
			},
		},
	} {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			logger, loggerHook := logging.NewMockLogger()
			ctx := testhelpers.CreateTestContextWithLogger(logger)
			appenv.ForceTestEnv()
			config := testhelpers.GetConfigFromContext(ctx)
			dbConfig, _ := app.DBConfig(config)
			rawDB, err := database.OpenRawDBConnection(dbConfig.FormatDSN(), true)
			require.NoError(t, err)
			defer func() { _ = rawDB.Close() }()
			db, err := database.OpenWithLogConfig(ctx, rawDB, database.LogConfig{LogSQLQueries: true, AnalyzeSQLQueries: true}, true)
			require.NoError(t, err)

			require.NoError(t, testCase.funcToCall(db))

			logs := (&loggingtest.Hook{Hook: loggerHook}).GetAllLogs()
			assert.Contains(t, logs, "SELECT 1")
			assert.Contains(t, logs, "query execution plan:")
		})
	}
}
