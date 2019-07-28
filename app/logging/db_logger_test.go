package logging

import (
	"testing"

	"github.com/jinzhu/gorm"
	assertlib "github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/app/config"
)

func TestNewDBLogger_ErrorFallback(t *testing.T) {
	assert := assertlib.New(t)
	logger := new() // no config
	dbLogger, logMode, rawLogMode := logger.NewDBLogger()
	assert.IsType(gorm.Logger{}, dbLogger)
	assert.False(logMode)
	assert.False(rawLogMode)
}

func TestLoggerFromConfig_TextLog(t *testing.T) {
	assert := assertlib.New(t)
	logger := new()
	logger.Configure(config.Logging{
		Format: "text",
		Output: "file",
	})
	dbLogger, _, _ := logger.NewDBLogger()
	assert.IsType(gorm.Logger{}, dbLogger)
}

func TestLoggerFromConfig_JSONLog(t *testing.T) {
	assert := assertlib.New(t)
	logger := new()
	logger.Configure(config.Logging{
		Format: "json",
		Output: "file",
	})
	dbLogger, _, _ := logger.NewDBLogger()
	assert.IsType(&StructuredDBLogger{}, dbLogger)
}

func TestLoggerFromConfig_WrongFormat(t *testing.T) {
	assert := assertlib.New(t)
	logger := new()
	logger.config = &config.Logging{
		Format: "yml",
		Output: "file",
	}
	assert.Panics(func() { logger.NewDBLogger() })
}

func TestNewDBLogger_LogMode(t *testing.T) {
	tests := []struct {
		name             string
		format           string
		logSQLQueries    bool
		logRawSQLQueries bool
	}{
		{name: "text: without SQL", format: "text", logSQLQueries: false, logRawSQLQueries: false},
		{name: "text: with SQL", format: "text", logSQLQueries: true, logRawSQLQueries: false},
		{name: "text: only raw SQL", format: "text", logSQLQueries: false, logRawSQLQueries: true},
		{name: "text: full SQL logging", format: "text", logSQLQueries: true, logRawSQLQueries: true},
		{name: "json: without SQL", format: "json", logSQLQueries: false, logRawSQLQueries: false},
		{name: "json: with SQL", format: "json", logSQLQueries: true, logRawSQLQueries: false},
		{name: "json: only raw SQL", format: "json", logSQLQueries: false, logRawSQLQueries: true},
		{name: "json: full SQL logging", format: "json", logSQLQueries: true, logRawSQLQueries: true},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			assert := assertlib.New(t)
			logger := new()
			logger.Configure(config.Logging{
				LogSQLQueries:    test.logSQLQueries,
				LogRawSQLQueries: test.logRawSQLQueries,
				Format:           test.format,
				Output:           "file",
			})
			_, logMode, rawLogMode := logger.NewDBLogger()
			assert.Equal(test.logSQLQueries, logMode)
			assert.Equal(test.logRawSQLQueries, rawLogMode)
		})
	}
}
