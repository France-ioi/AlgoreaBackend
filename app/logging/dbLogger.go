package logging

import (
	"log"
	"os"

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
	logMode := conf.LogSQL
	if conf.TextLogging {
		return gorm.Logger{LogWriter: log.New(os.Stdout, "\r\n", 0)}, logMode
	}
	return NewStructuredDBLogger(logger), logMode
}
