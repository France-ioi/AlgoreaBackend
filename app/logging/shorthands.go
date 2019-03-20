package logging

import "github.com/sirupsen/logrus"

// Debug logs a message at level Debug on the shared logger.
func Debug(args ...interface{}) {
	SharedLogger.Debug(args...)
}

// Info logs a message at level Info on the shared logger.
func Info(args ...interface{}) {
	SharedLogger.Info(args...)
}

// Warn logs a message at level Warn on the shared logger.
func Warn(args ...interface{}) {
	SharedLogger.Warn(args...)
}

// Error logs a message at level Error on the shared logger.
func Error(args ...interface{}) {
	SharedLogger.Error(args...)
}

// Fatal logs a message at level Fatal on the standard logger then the process will exit with status set to 1.
func Fatal(args ...interface{}) {
	SharedLogger.Fatal(args...)
}

// Panic logs a message at level Panic on the shared logger.
func Panic(args ...interface{}) {
	SharedLogger.Panic(args...)
}

// Debugf logs a message at level Debug on the shared logger.
func Debugf(format string, args ...interface{}) {
	SharedLogger.Debugf(format, args...)
}

// Infof logs a message at level Info on the shared logger.
func Infof(format string, args ...interface{}) {
	SharedLogger.Infof(format, args...)
}

// Warnf logs a message at level Warn on the shared logger.
func Warnf(format string, args ...interface{}) {
	SharedLogger.Warnf(format, args...)
}

// Errorf logs a message at level Error on the shared logger.
func Errorf(format string, args ...interface{}) {
	SharedLogger.Errorf(format, args...)
}

// Panicf logs a message at level Panic on the shared logger.
func Panicf(format string, args ...interface{}) {
	SharedLogger.Panicf(format, args...)
}

// Fatalf logs a message at level Fatal on the standard logger then the process will exit with status set to 1.
func Fatalf(format string, args ...interface{}) {
	SharedLogger.Fatalf(format, args...)
}

// WithField creates an entry from the standard logger and adds a field to
// it. If you want multiple fields, use `WithFields`.
//
// Note that it doesn't log until you call Debug, Print, Info, Warn, Fatal
// or Panic on the Entry it returns.
func WithField(key string, value interface{}) *logrus.Entry {
	return SharedLogger.WithField(key, value)
}

// WithFields creates an entry from the standard logger and adds multiple
// fields to it. This is simply a helper for `WithField`, invoking it
// once for each field.
//
// Note that it doesn't log until you call Debug, Print, Info, Warn, Fatal
// or Panic on the Entry it returns.
func WithFields(fields map[string]interface{}) *logrus.Entry {
	return SharedLogger.WithFields(fields)
}
