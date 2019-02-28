package logging

import (
	"testing"

	"github.com/sirupsen/logrus"
	assertlib "github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/app/config"
)

func TestNew(t *testing.T) {
	assert := assertlib.New(t)
	// just verify it can load the config
	assert.NotNil(new())
}

func TestConfigureLoggerText(t *testing.T) {
	assert := assertlib.New(t)
	conf := config.Logging{
		TextLogging: true,
	}
	logger := logrus.New()
	configureLogger(logger, conf)
	assert.IsType(&logrus.TextFormatter{}, logger.Formatter)
}

func TestConfigureLoggerJson(t *testing.T) {
	assert := assertlib.New(t)
	conf := config.Logging{
		TextLogging: false,
	}
	logger := logrus.New()
	configureLogger(logger, conf)
	assert.IsType(&logrus.JSONFormatter{}, logger.Formatter)
}

func TestConfigureLoggerDefaultLevel(t *testing.T) {
	assert := assertlib.New(t)
	conf := config.Logging{
		LogLevel: "",
	}
	logger := logrus.New()
	configureLogger(logger, conf)
	assert.Equal(logrus.ErrorLevel, logger.Level)
}

func TestConfigureLoggerParsedLevel(t *testing.T) {
	assert := assertlib.New(t)
	conf := config.Logging{
		LogLevel: "warn",
	}
	logger := logrus.New()
	configureLogger(logger, conf)
	assert.Equal(logrus.WarnLevel, logger.Level)
}

func TestConfigureLoggerInvalidLevel(t *testing.T) {
	assert := assertlib.New(t)
	conf := config.Logging{
		LogLevel: "invalid_level",
	}
	logger := logrus.New()
	assert.Panics(func() {
		configureLogger(logger, conf)
	})
}
