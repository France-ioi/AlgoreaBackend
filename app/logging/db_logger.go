package logging

import (
	"github.com/jinzhu/gorm"
)

// DBLogger is the logger interface for the DB logs.
type DBLogger interface {
	Print(v ...interface{})
}

type sharedLoggerWriter struct{}

func (l *sharedLoggerWriter) Println(v ...interface{}) {
	SharedLogger.Println(v...)
}

// NewDBLogger returns a logger for the database and the `logMode` as well as the 'rawLogMode', according to the config.
func (l *Logger) NewDBLogger() (DBLogger, bool, bool) {
	if l.config == nil {
		// if cannot parse config, log on error to stdout
		return gorm.Logger{LogWriter: &sharedLoggerWriter{}}, false, false
	}

	logMode := l.config.GetBool("LogSQLQueries")
	rawLogMode := l.config.GetBool("LogRawSQLQueries")
	switch l.config.GetString("format") {
	case formatText:
		return gorm.Logger{LogWriter: &sharedLoggerWriter{}}, logMode, rawLogMode
	case formatJSON:
		return NewStructuredDBLogger(), logMode, rawLogMode
	default:
		panic("Logging format must be either 'text' or 'json'. Got: " + l.config.GetString("format"))
	}
}
