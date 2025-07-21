//go:build !unit

package database_test

import (
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/app/logging"
	"github.com/France-ioi/AlgoreaBackend/v2/testhelpers"
	"github.com/France-ioi/AlgoreaBackend/v2/testhelpers/testoutput"
)

func Test_RawSQLQueryLogging_Duration(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	logger, loggerHook := logging.NewMockLogger()
	conf := viper.New()
	conf.Set("LogSQLQueries", true)
	conf.Set("LogRawSQLQueries", true)
	logger.Configure(conf)
	ctx := testhelpers.CreateTestContextWithLogger(logger)
	sqlDB := testhelpers.OpenRawDBConnection(ctx)
	defer func() { _ = sqlDB.Close() }()

	db, err := database.OpenWithLogConfig(ctx, sqlDB, database.LogConfig{LogSQLQueries: true}, true)
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	var result []interface{}
	require.NoError(t, db.Raw("SELECT 1").Scan(&result).Error())

	entries := loggerHook.AllEntries()
	expectedMessages := []string{
		"sql-connector-connect", "sql-ping",
		"sql-conn-reset", "sql-conn-query", "SELECT 1", "sql-rows-next", "sql-rows-next",
	}

	for i, entry := range entries {
		assert.Equal(t, expectedMessages[i], entry.Message)
		assert.Equal(t, "db", entry.Data["type"])
		assert.Contains(t, entry.Data, "duration")
		if duration, ok := entry.Data["duration"]; ok {
			assert.IsType(t, "", duration)
			if durationStr, ok := duration.(string); ok {
				parsedDuration, err := time.ParseDuration(durationStr)
				require.NoError(t, err)
				assert.Greater(t, parsedDuration, 0*time.Second)
			}
		}
	}
}

func Test_RawSQLQueryLogging_ResetSession(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	logger, loggerHook := logging.NewMockLogger()
	conf := viper.New()
	conf.Set("LogSQLQueries", true)
	conf.Set("LogRawSQLQueries", true)
	logger.Configure(conf)
	ctx := testhelpers.CreateTestContextWithLogger(logger)
	sqlDB := testhelpers.OpenRawDBConnection(ctx)
	defer func() { _ = sqlDB.Close() }()

	db, err := database.OpenWithLogConfig(ctx, sqlDB, database.LogConfig{LogSQLQueries: true}, true)
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	var result []interface{}
	require.NoError(t, db.Raw("SELECT 1").Scan(&result).Error())
	require.NoError(t, db.Raw("SELECT 1").Scan(&result).Error())

	expectedMessages := []string{
		"sql-connector-connect", "sql-ping",
		"sql-conn-reset", "sql-conn-query", "SELECT 1", "sql-rows-next", "sql-rows-next",
		"sql-conn-reset", "sql-conn-query", "SELECT 1", "sql-rows-next", "sql-rows-next",
	}
	entries := loggerHook.AllEntries()
	require.Len(t, entries, len(expectedMessages))

	for i, entry := range entries {
		assert.Equal(t, expectedMessages[i], entry.Message)
		assert.Equal(t, "db", entry.Data["type"])
		assert.Contains(t, entry.Data, "duration")
		if duration, ok := entry.Data["duration"]; ok {
			assert.IsType(t, "", duration)
			if durationStr, ok := duration.(string); ok {
				parsedDuration, err := time.ParseDuration(durationStr)
				require.NoError(t, err)
				assert.Greater(t, parsedDuration, 0*time.Second)
			}
		}
	}
}
