package logging

import "github.com/sirupsen/logrus"

// Debug logs a message at level Debug on the standard Logger.
func Debug(args ...interface{}) {
	Logger.Debug(args...)
}

// Info logs a message at level Info on the standard Logger.
func Info(args ...interface{}) {
	Logger.Info(args...)
}

// Warn logs a message at level Warn on the standard Logger.
func Warn(args ...interface{}) {
	Logger.Warn(args...)
}

// Error logs a message at level Error on the standard Logger.
func Error(args ...interface{}) {
	Logger.Error(args...)
}

// Fatal logs a message at level Fatal on the standard logger then the process will exit with status set to 1.
func Fatal(args ...interface{}) {
	Logger.Fatal(args...)
}

// Panic logs a message at level Panic on the standard Logger.
func Panic(args ...interface{}) {
	Logger.Panic(args...)
}

// Debugf logs a message at level Debug on the standard Logger.
func Debugf(format string, args ...interface{}) {
	Logger.Debugf(format, args...)
}

// Infof logs a message at level Info on the standard Logger.
func Infof(format string, args ...interface{}) {
	Logger.Infof(format, args...)
}

// Warnf logs a message at level Warn on the standard Logger.
func Warnf(format string, args ...interface{}) {
	Logger.Warnf(format, args...)
}

// Errorf logs a message at level Error on the standard Logger.
func Errorf(format string, args ...interface{}) {
	Logger.Errorf(format, args...)
}

// Panicf logs a message at level Panic on the standard Logger.
func Panicf(format string, args ...interface{}) {
	Logger.Panicf(format, args...)
}

// Fatalf logs a message at level Fatal on the standard logger then the process will exit with status set to 1.
func Fatalf(format string, args ...interface{}) {
	Logger.Fatalf(format, args...)
}

// WithField creates an entry from the standard logger and adds a field to
// it. If you want multiple fields, use `WithFields`.
//
// Note that it doesn't log until you call Debug, Print, Info, Warn, Fatal
// or Panic on the Entry it returns.
func WithField(key string, value interface{}) *logrus.Entry {
	return Logger.WithField(key, value)
}

// WithFields creates an entry from the standard logger and adds multiple
// fields to it. This is simply a helper for `WithField`, invoking it
// once for each field.
//
// Note that it doesn't log until you call Debug, Print, Info, Warn, Fatal
// or Panic on the Entry it returns.
func WithFields(fields map[string]interface{}) *logrus.Entry {
	return Logger.WithFields(fields)
}
