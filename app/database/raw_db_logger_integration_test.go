//go:build !unit

package database_test

import (
	"reflect"
	"testing"
	"time"

	"bou.ke/monkey"
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

	monkey.PatchInstanceMethod(reflect.TypeOf(&viper.Viper{}), "GetBool",
		func(_ *viper.Viper, key string) bool { return key == "LogRawSQLQueries" || key == "LogSQLQueries" })
	defer monkey.UnpatchAll()

	sqlDB := testhelpers.OpenRawDBConnection()
	defer func() { _ = sqlDB.Close() }()

	db, err := database.OpenWithLogConfig(sqlDB, database.LogConfig{LogSQLQueries: true}, true)
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	loggerHook, loggerRestoreFunc := logging.MockSharedLoggerHook()
	defer loggerRestoreFunc()

	var result []interface{}
	require.NoError(t, db.Raw("SELECT 1").Scan(&result).Error())

	entries := loggerHook.AllEntries()
	expectedMessages := []string{"sql-conn-query", "SELECT 1", "sql-rows-next", "sql-rows-next"}
	require.Len(t, entries, len(expectedMessages))

	for i, entry := range entries {
		assert.Equal(t, expectedMessages[i], entry.Message)
		assert.Equal(t, "db", entry.Data["type"])
		assert.Contains(t, entry.Data, "duration")
		if duration, ok := entry.Data["duration"]; ok {
			assert.IsType(t, "", duration)
			if durationStr, ok := duration.(string); ok {
				parsedDuration, err := time.ParseDuration(durationStr)
				assert.NoError(t, err)
				if err == nil {
					assert.Greater(t, parsedDuration, 0*time.Second)
				}
			}
		}
	}
}

func Test_RawSQLQueryLogging_ResetSession(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	monkey.PatchInstanceMethod(reflect.TypeOf(&logging.Logger{}), "IsRawSQLQueriesLoggingEnabled",
		func(_ *logging.Logger) bool { return true })
	defer monkey.UnpatchAll()

	sqlDB := testhelpers.OpenRawDBConnection()
	defer func() { _ = sqlDB.Close() }()

	loggerHook, loggerRestoreFunc := logging.MockSharedLoggerHook()
	defer loggerRestoreFunc()

	db, err := database.OpenWithLogConfig(sqlDB, database.LogConfig{LogSQLQueries: true}, true)
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	var result []interface{}
	require.NoError(t, db.Raw("SELECT 1").Scan(&result).Error())
	require.NoError(t, db.Raw("SELECT 1").Scan(&result).Error())

	entries := loggerHook.AllEntries()
	expectedMessages := []string{
		"sql-conn-reset", "SELECT 1", "sql-conn-reset", "SELECT 1",
	}
	require.Len(t, entries, len(expectedMessages))

	for i, entry := range entries {
		assert.Equal(t, expectedMessages[i], entry.Message)
		assert.Equal(t, "db", entry.Data["type"])
		assert.Contains(t, entry.Data, "duration")
		if duration, ok := entry.Data["duration"]; ok {
			assert.IsType(t, "", duration)
			if durationStr, ok := duration.(string); ok {
				parsedDuration, err := time.ParseDuration(durationStr)
				assert.NoError(t, err)
				if err == nil {
					assert.Greater(t, parsedDuration, 0*time.Second)
				}
			}
		}
	}
}
