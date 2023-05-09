package logging

import (
	"fmt"
	"testing"

	"github.com/jinzhu/gorm"
	"github.com/spf13/viper"
	assertlib "github.com/stretchr/testify/assert"
	gormLogger "gorm.io/gorm/logger"
)

func TestNewDBLogger_ErrorFallback(t *testing.T) {
	assert := assertlib.New(t)
	logger := createLogger() // no config
	dbLogger, logMode := logger.NewDBLogger()
	assert.IsType(gorm.Logger{}, dbLogger)
	assert.False(logMode)
}

func TestLogger_GetRawSQLLogMode_ErrorFallback(t *testing.T) {
	assert := assertlib.New(t)
	logger := createLogger() // no config
	rawLogMode := logger.GetRawSQLLogMode()
	assert.False(rawLogMode)
}

func TestLoggerFromConfig_TextLog(t *testing.T) {
	assert := assertlib.New(t)
	logger := createLogger()
	config := viper.New()
	config.Set("Format", "text")
	config.Set("Output", "file")
	logger.Configure(config)
	dbLogger, _ := logger.NewDBLogger()
	assert.Implements((*gormLogger.Interface)(nil), dbLogger)
}

func TestLoggerFromConfig_JSONLog(t *testing.T) {
	assert := assertlib.New(t)
	logger := createLogger()
	config := viper.New()
	config.Set("Format", "json")
	config.Set("Output", "file")
	logger.Configure(config)
	dbLogger, _ := logger.NewDBLogger()
	assert.IsType(&StructuredDBLogger{}, dbLogger)
}

func TestLoggerFromConfig_WrongFormat(t *testing.T) {
	assert := assertlib.New(t)
	logger := createLogger()
	config := viper.New()
	config.Set("Format", "yml")
	config.Set("Output", "file")
	logger.config = config
	assert.Panics(func() { logger.NewDBLogger() })
}

func TestNewDBLogger_LogMode(t *testing.T) {
	tests := []struct {
		name          string
		format        string
		logSQLQueries bool
	}{
		{name: "text: without SQL", format: "text", logSQLQueries: false},
		{name: "text: with SQL", format: "text", logSQLQueries: true},
		{name: "text: only raw SQL", format: "text", logSQLQueries: false},
		{name: "text: full SQL logging", format: "text", logSQLQueries: true},
		{name: "json: without SQL", format: "json", logSQLQueries: false},
		{name: "json: with SQL", format: "json", logSQLQueries: true},
		{name: "json: only raw SQL", format: "json", logSQLQueries: false},
		{name: "json: full SQL logging", format: "json", logSQLQueries: true},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			assert := assertlib.New(t)
			logger := createLogger()
			config := viper.New()
			config.Set("LogSQLQueries", test.logSQLQueries)
			config.Set("Format", test.format)
			config.Set("Output", "file")
			logger.Configure(config)
			_, logMode := logger.NewDBLogger()
			assert.Equal(test.logSQLQueries, logMode)
		})
	}
}

func TestLogger_GetRawSqlLog(t *testing.T) {
	tests := []bool{false, true}
	for _, test := range tests {
		test := test
		t.Run(fmt.Sprintf("%v", test), func(t *testing.T) {
			assert := assertlib.New(t)
			ourLogger := createLogger()
			config := viper.New()
			config.Set("LogRawSQLQueries", test)
			ourLogger.Configure(config)
			rawLogMode := ourLogger.GetRawSQLLogMode()
			assert.Equal(test, rawLogMode)
		})
	}
}
