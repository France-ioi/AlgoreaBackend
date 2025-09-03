package logging

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
