package logging

import (
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/sirupsen/logrus" //nolint:depguard
	"github.com/spf13/viper"
)

type configuration struct {
	Format           string
	Output           string
	Level            string
	LogSQLQueries    bool
	LogRawSQLQueries bool
}

// Logger is wrapper around a logger and keeping the logging config so that it can be reused by other loggers
type Logger struct {
	*logrus.Logger
	config *configuration
}

var (
	// SharedLogger is the global scope logger
	// It should not be used directly by other packages (except for testing) which should prefer shorthands functions
	SharedLogger = new()
)

const formatJSON = "json"
const formatText = "text"

const outputStdout = "stdout"
const outputStderr = "stderr"
const outputFile = "file"

// Configure applies the given logging configuration to the logger
// (may panic if the configuration is invalid)
func (l *Logger) Configure(config *viper.Viper) {
	unmarshalConfig(config, &l.config)

	// Format
	switch l.config.Format {
	case formatText:
		l.SetFormatter(&logrus.TextFormatter{DisableTimestamp: true, ForceColors: l.config.Output != outputFile})
	case formatJSON:
		l.SetFormatter(&logrus.JSONFormatter{})
	default:
		panic("Logging format must be either 'text' or 'json'. Got: " + l.config.Format)
	}

	// Output
	switch l.config.Output {
	case outputStdout:
		l.SetOutput(os.Stdout)
	case outputStderr:
		l.SetOutput(os.Stderr)
	case outputFile:
		_, codeFilePath, _, _ := runtime.Caller(0)
		codeDir := filepath.Dir(codeFilePath)
		f, err := os.OpenFile(codeDir+"/../../log/all.log", os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0600)
		if err != nil {
			l.SetOutput(os.Stdout)
			l.Errorf("Unable to open file for logs, fallback to stdout: %v\n", err)
		} else {
			l.SetOutput(f)
		}
	default:
		panic("Logging output must be either 'stdout', 'stderr' or 'file'. Got: " + l.config.Output)
	}

	// Level
	if level, err := logrus.ParseLevel(l.config.Level); err != nil {
		l.Errorf("Unable to parse logging level config, use default (%s)", l.GetLevel().String())
	} else {
		l.SetLevel(level)
	}

	log.SetOutput(l.Logger.Writer())
}

func unmarshalConfig(config *viper.Viper, dest **configuration) {
	// set default values (if not configured)
	config.SetDefault("format", "json")
	config.SetDefault("output", "file")
	config.SetDefault("level", "info")
	config.SetDefault("logSqlQueries", true)

	if err := config.Unmarshal(dest); err != nil {
		panic("Enable to load the 'logging' configuration")
	}
}

// ResetShared reset the global logger to its default settings before its configuration
func ResetShared() {
	SharedLogger = new()
}

func new() *Logger {
	return &Logger{logrus.New(), nil}
}
