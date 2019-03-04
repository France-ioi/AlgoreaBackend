package logging

import (
	"github.com/jinzhu/gorm"
	"github.com/sirupsen/logrus"

	"github.com/France-ioi/AlgoreaBackend/app/config"
)

// DBLogger is the logger interface for the DB logs
type DBLogger interface {
	Print(v ...interface{})
}

// NewDBLogger returns a logger for the database and the `logmode`, according to the config
func NewDBLogger() (DBLogger, bool) {
	var (
		err  error
		conf *config.Root
	)

	if conf, err = config.Load(); err != nil {
		// if cannot parse config, log on error to stdout
		return gorm.Logger{LogWriter: Logger}, false
	}
	return loggerFromConfig(conf.Logging, Logger)
}

func loggerFromConfig(conf config.Logging, logger *logrus.Logger) (DBLogger, bool) {
	logMode := conf.LogSQLQueries
	switch conf.Format {
	case "text":
		return gorm.Logger{LogWriter: logger}, logMode
	case "json":
		return NewStructuredDBLogger(logger), logMode
	default:
		panic("Logging format must be either 'text' or 'json'. Got: " + conf.Format)
	}
}
