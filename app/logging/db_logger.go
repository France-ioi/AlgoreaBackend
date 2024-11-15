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

// NewDBLogger returns a logger for the database according to the config.
func (l *Logger) NewDBLogger() DBLogger {
	if l.config == nil {
		// if cannot parse config, log on error to stdout
		return gorm.Logger{LogWriter: &sharedLoggerWriter{}}
	}

	switch l.config.GetString("format") {
	case formatText:
		return gorm.Logger{LogWriter: &sharedLoggerWriter{}}
	case formatJSON:
		return NewStructuredDBLogger()
	default:
		panic("Logging format must be either 'text' or 'json'. Got: " + l.config.GetString("format"))
	}
}

// IsSQLQueriesLoggingEnabled returns whether the SQL queries logging is enabled in the config.
func (l *Logger) IsSQLQueriesLoggingEnabled() bool {
	return l.getBoolConfigFlagValue("LogSQLQueries")
}

// IsRawSQLQueriesLoggingEnabled returns whether the raw SQL queries logging is enabled in the config.
func (l *Logger) IsRawSQLQueriesLoggingEnabled() bool {
	return l.getBoolConfigFlagValue("LogRawSQLQueries")
}

// IsSQLQueriesAnalyzingEnabled returns whether the SQL queries analyzing is enabled in the config.
// Note: SQL queries analyzing is only enabled if the SQL queries logging is enabled as well.
func (l *Logger) IsSQLQueriesAnalyzingEnabled() bool {
	return l.getBoolConfigFlagValue("AnalyzeSQLQueries")
}

func (l *Logger) getBoolConfigFlagValue(flagName string) bool {
	if l.config == nil {
		return false
	}
	return l.config.GetBool(flagName)
}
