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
	dbLogger, logMode := logger.NewDBLogger()
	assert.IsType(gorm.Logger{}, dbLogger)
	assert.Equal(false, logMode)
}

func TestLoggerFromConfig_TextLog(t *testing.T) {
	assert := assertlib.New(t)
	logger := new()
	logger.Configure(config.Logging{
		Format: "text",
		Output: "file",
	})
	dbLogger, _ := logger.NewDBLogger()
	assert.IsType(gorm.Logger{}, dbLogger)
}

func TestLoggerFromConfig_JSONLog(t *testing.T) {
	assert := assertlib.New(t)
	logger := new()
	logger.Configure(config.Logging{
		Format: "json",
		Output: "file",
	})
	dbLogger, _ := logger.NewDBLogger()
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

func TestLoggerFromConfig_WithoutSQL(t *testing.T) {
	assert := assertlib.New(t)
	logger := new()
	logger.Configure(config.Logging{
		LogSQLQueries: false,
		Format:        "text",
		Output:        "file",
	})
	_, logMode := logger.NewDBLogger()
	assert.False(logMode)
}

func TestLoggerFromConfig_WithSQL(t *testing.T) {
	assert := assertlib.New(t)
	logger := new()
	logger.Configure(config.Logging{
		LogSQLQueries: true,
		Format:        "text",
		Output:        "file",
	})
	_, logMode := logger.NewDBLogger()
	assert.True(logMode)
}
