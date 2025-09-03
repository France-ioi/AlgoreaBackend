// Package logging provides utilities for logging.
package logging

import (
	"context"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	"github.com/go-chi/chi/middleware"
	"github.com/sirupsen/logrus" //nolint:depguard
	"github.com/spf13/viper"
)

// Logger is wrapper around a logger and keeping the logging config so that it can be reused by other loggers.
type Logger struct {
	logrusLogger *logrus.Logger
	config       *viper.Viper
}

const (
	formatJSON    = "json"
	formatText    = "text"
	formatConsole = "console"
)

const (
	outputStdout = "stdout"
	outputStderr = "stderr"
	outputFile   = "file"
)

type loggerContextKeyType int

const loggerContextKey loggerContextKeyType = iota

// EntryFromContext returns a new logrus entry of the logger from the given context with the context set.
// The context must have been created with ContextWithLogger, otherwise EntryFromContext will panic.
func EntryFromContext(ctx context.Context) *logrus.Entry {
	return LoggerFromContext(ctx).WithContext(ctx)
}

// LoggerFromContext returns the logger from the given context.
// The context must have been created with ContextWithLogger, otherwise LoggerFromContext will panic.
func LoggerFromContext(ctx context.Context) *Logger {
	return ctx.Value(loggerContextKey).(*Logger)
}

// ContextWithLogger returns a copy of the given context with the logger set.
func ContextWithLogger(ctx context.Context, logger *Logger) context.Context {
	return context.WithValue(ctx, loggerContextKey, logger)
}

// ContextWithLoggerMiddleware returns a middleware that sets the logger in the request context.
func ContextWithLoggerMiddleware(logger *Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r.WithContext(ContextWithLogger(r.Context(), logger)))
		}
		return http.HandlerFunc(fn)
	}
}

// NewLoggerFromConfig creates a new logger from the given configuration.
func NewLoggerFromConfig(config *viper.Viper) *Logger {
	logger := createLogger()
	logger.Configure(config)
	return logger
}

// Configure applies the given logging configuration to the logger
// (may panic if the configuration is invalid).
func (l *Logger) Configure(config *viper.Viper) {
	l.config = config

	// set default values (if not configured)
	config.SetDefault("format", "json")
	config.SetDefault("output", "file")
	config.SetDefault("level", "info")
	config.SetDefault("logSqlQueries", true)

	// Format
	switch config.GetString("format") {
	case formatText:
		l.logrusLogger.SetFormatter(newTextFormatter(config.GetString("output") != outputFile))
	case formatJSON:
		l.logrusLogger.SetFormatter(newJSONFormatter())
	case formatConsole:
		l.logrusLogger.SetFormatter(newConsoleFormatter())
	default:
		panic("Logging format must be one of 'text'/'json'/'console'. Got: " + config.GetString("format"))
	}

	// Output
	l.setOutput(config)

	// Level
	if level, err := logrus.ParseLevel(config.GetString("level")); err != nil {
		l.logrusLogger.Errorf("Unable to parse logging level config, use default (%s)", l.logrusLogger.GetLevel().String())
	} else {
		l.logrusLogger.SetLevel(level)
	}

	log.SetOutput(l.logrusLogger.Writer())
}

func (l *Logger) setOutput(config *viper.Viper) {
	switch config.GetString("output") {
	case outputStdout:
		l.logrusLogger.SetOutput(os.Stdout)
	case outputStderr:
		l.logrusLogger.SetOutput(os.Stderr)
	case outputFile:
		if config.GetString("format") == formatConsole {
			panic("Logging format 'console' is not supported with output 'file'")
		}
		_, codeFilePath, _, _ := runtime.Caller(0)
		codeDir := filepath.Dir(codeFilePath)
		const logFilePermissions = 0o600
		//nolint:gosec,gosec // No user input
		f, err := os.OpenFile(codeDir+"/../../log/all.log", os.O_WRONLY|os.O_APPEND|os.O_CREATE, logFilePermissions)
		if err != nil {
			l.logrusLogger.SetOutput(os.Stdout)
			l.logrusLogger.Errorf("Unable to open file for logs, fallback to stdout: %v\n", err)
		} else {
			l.logrusLogger.SetOutput(f)
		}
	default:
		panic("Logging output must be either 'stdout', 'stderr' or 'file'. Got: " + config.GetString("output"))
	}
}

// WithContext returns a new entry with the given context.
func (l *Logger) WithContext(ctx context.Context) *logrus.Entry {
	entry := l.logrusLogger.WithContext(ctx)

	requestID := middleware.GetReqID(ctx)
	if requestID != "" {
		entry = entry.WithField("req_id", requestID)
	}

	return entry
}

func createLogger() *Logger {
	return &Logger{logrus.New(), nil}
}
