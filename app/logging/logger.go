package logging

import (
	"log"

	"github.com/sirupsen/logrus"

	"github.com/France-ioi/AlgoreaBackend/app/config"
)

var (
	// Logger is a configured logrus.Logger.
	Logger *logrus.Logger
)

// New creates and configures a new logrus Logger.
func New(conf config.Logging) *logrus.Logger {
	var err error

	Logger = logrus.New()
	if conf.TextLogging {
		Logger.Formatter = &logrus.TextFormatter{
			DisableTimestamp: true,
		}
	} else {
		Logger.Formatter = &logrus.JSONFormatter{}
	}

	level := conf.LogLevel
	if level == "" {
		level = "error"
	}
	var l logrus.Level
	l, err = logrus.ParseLevel(level)
	if err != nil {
		log.Fatal(err)
	}
	Logger.Level = l
	return Logger
}
