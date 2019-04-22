package logging

import (
	"github.com/jinzhu/gorm"
)

// DBLogger is the logger interface for the DB logs
type DBLogger interface {
	Print(v ...interface{})
}

// NewDBLogger returns a logger for the database and the `logmode`, according to the config
func (l *Logger) NewDBLogger() (DBLogger, bool) {
	if l.config == nil {
		// if cannot parse config, log on error to stdout
		return gorm.Logger{LogWriter: l}, false
	}

	logMode := l.config.LogSQLQueries
	switch l.config.Format {
	case formatText:
		return gorm.Logger{LogWriter: l}, logMode
	case formatJSON:
		return NewStructuredDBLogger(l.Logger), logMode
	default:
		panic("Logging format must be either 'text' or 'json'. Got: " + l.config.Format)
	}
}
