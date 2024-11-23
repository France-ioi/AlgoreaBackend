// Package logging provides utilities for logging.
package logging

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/sirupsen/logrus" //nolint:depguard
	"github.com/spf13/viper"
)

// Logger is wrapper around a logger and keeping the logging config so that it can be reused by other loggers.
type Logger struct {
	logrusLogger *logrus.Logger
	config       *viper.Viper
}

// SharedLogger is the global scope logger
// It should not be used directly by other packages (except for testing) which should prefer shorthands functions.
var SharedLogger = createLogger()

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
		f, err := os.OpenFile(codeDir+"/../../log/all.log", os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0o600) //nolint:gosec,gosec No user input.
		if err != nil {
			l.logrusLogger.SetOutput(os.Stdout)
			l.logrusLogger.Errorf("Unable to open file for logs, fallback to stdout: %v\n", err)
		} else {
			l.logrusLogger.SetOutput(f)
		}
	default:
		panic("Logging output must be either 'stdout', 'stderr' or 'file'. Got: " + config.GetString("output"))
	}

	// Level
	if level, err := logrus.ParseLevel(config.GetString("level")); err != nil {
		l.logrusLogger.Errorf("Unable to parse logging level config, use default (%s)", l.logrusLogger.GetLevel().String())
	} else {
		l.logrusLogger.SetLevel(level)
	}

	log.SetOutput(l.logrusLogger.Writer())
}

// ResetShared reset the global logger to its default settings before its configuration.
func ResetShared() {
	SharedLogger = createLogger()
}

// WithContext returns a new entry with the given context.
func (l *Logger) WithContext(ctx context.Context) *logrus.Entry {
	entry := l.logrusLogger.WithContext(ctx)

	return entry
}

func createLogger() *Logger { //nolint:gosec
	return &Logger{logrus.New(), nil}
}
