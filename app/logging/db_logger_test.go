package logging

import (
	"testing"

	"github.com/jinzhu/gorm"
	assertlib "github.com/stretchr/testify/assert"

	"github.com/spf13/viper"
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
	config := viper.New()
	config.Set("Format", "text")
	config.Set("Output", "file")
	logger.Configure(config)
	dbLogger, _, _ := logger.NewDBLogger()
	assert.IsType(gorm.Logger{}, dbLogger)
}

func TestLoggerFromConfig_JSONLog(t *testing.T) {
	assert := assertlib.New(t)
	logger := new()
	config := viper.New()
	config.Set("Format", "json")
	config.Set("Output", "file")
	logger.Configure(config)
	dbLogger, _, _ := logger.NewDBLogger()
	assert.IsType(&StructuredDBLogger{}, dbLogger)
}

func TestLoggerFromConfig_WrongFormat(t *testing.T) {
	assert := assertlib.New(t)
	logger := new()
	config := viper.New()
	config.Set("Format", "yml")
	config.Set("Output", "file")
	assert.Panics(func() { logger.Configure(config) })
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
			config := viper.New()
			config.Set("LogSQLQueries", test.logSQLQueries)
			config.Set("LogRawSQLQueries", test.logRawSQLQueries)
			config.Set("Format", test.format)
			config.Set("Output", "file")
			logger.Configure(config)
			_, logMode, rawLogMode := logger.NewDBLogger()
			assert.Equal(test.logSQLQueries, logMode)
			assert.Equal(test.logRawSQLQueries, rawLogMode)
		})
	}
}
