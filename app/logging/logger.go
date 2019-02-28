package logging

import (
	"log"

	"github.com/sirupsen/logrus"

	"github.com/France-ioi/AlgoreaBackend/app/config"
)

var (
	// Logger is the actual logrus logger which is used
	Logger = new()
)

func new() *logrus.Logger {
	logger := logrus.New()
	if conf, err := config.Load(); err == nil {
		// If config, configure logger. Otherwise, use default logger
		configureLogger(logger, conf.Logging)
	}
	log.SetOutput(logger.Writer())
	return logger
}

func configureLogger(logger *logrus.Logger, conf config.Logging) {

	if conf.TextLogging {
		logger.Formatter = &logrus.TextFormatter{
			DisableTimestamp: true,
		}
	} else {
		logger.Formatter = &logrus.JSONFormatter{}
	}

	level := conf.LogLevel
	if level == "" {
		level = "error"
	}
	l, err := logrus.ParseLevel(level)
	if err != nil {
		panic(err)
	} else {
		logger.Level = l
	}
}
