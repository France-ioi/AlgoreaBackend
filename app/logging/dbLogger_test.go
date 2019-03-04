package logging

import (
	"errors"
	"testing"

	"bou.ke/monkey"
	"github.com/jinzhu/gorm"
	"github.com/sirupsen/logrus"
	assertlib "github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/app/config"
)

func TestNewDBLogger_ErrorFallback(t *testing.T) {
	assert := assertlib.New(t)

	fakeFunc := func() (*config.Root, error) {
		return nil, errors.New("config loading error")
	}
	patch := monkey.Patch(config.Load, fakeFunc)
	defer patch.Unpatch()

	dbLogger, logMode := NewDBLogger()
	assert.IsType(gorm.Logger{}, dbLogger)
	assert.Equal(false, logMode)
}

func TestLoggerFromConfig_TextLog(t *testing.T) {
	assert := assertlib.New(t)
	logger := logrus.New()
	config := config.Logging{
		Format: "text",
		Output: "file",
	}
	dbLogger, _ := loggerFromConfig(config, logger)
	assert.IsType(gorm.Logger{}, dbLogger)
}

func TestLoggerFromConfig_JSONLog(t *testing.T) {
	assert := assertlib.New(t)
	logger := logrus.New()
	config := config.Logging{
		Format: "json",
		Output: "file",
	}
	dbLogger, _ := loggerFromConfig(config, logger)
	assert.IsType(&StructuredDBLogger{}, dbLogger)
}

func TestLoggerFromConfig_WrongFormat(t *testing.T) {
	assert := assertlib.New(t)
	logger := logrus.New()
	config := config.Logging{
		Format: "yml",
		Output: "file",
	}
	assert.Panics(func() { loggerFromConfig(config, logger) })
}

func TestLoggerFromConfig_WithoutSQL(t *testing.T) {
	assert := assertlib.New(t)
	logger := logrus.New()
	config := config.Logging{
		LogSQLQueries: false,
		Format:        "text",
		Output:        "file",
	}
	_, logMode := loggerFromConfig(config, logger)
	assert.False(logMode)
}

func TestLoggerFromConfig_WithSQL(t *testing.T) {
	assert := assertlib.New(t)
	logger := logrus.New()
	config := config.Logging{
		LogSQLQueries: true,
		Format:        "text",
		Output:        "file",
	}
	_, logMode := loggerFromConfig(config, logger)
	assert.True(logMode)
}
