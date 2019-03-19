package logging

import (
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/sirupsen/logrus"

	"github.com/France-ioi/AlgoreaBackend/app/config"
)

var (
	// Logger is the global logger
	// It should not be used directly by app which should prefer shorthands functions
	Logger = defaultGlobal()
)

// ConfigureGlobal applies the given logging configuration to the global logger
func ConfigureGlobal(conf config.Logging) {
	configure(Logger, conf)
}

// ResetGlobal reset the global logger to its default settings before its configuration
func ResetGlobal() {
	Logger = defaultGlobal()
}

func defaultGlobal() *logrus.Logger {
	logger := logrus.New()
	log.SetOutput(logger.Writer())
	return logger
}

func configure(logger *logrus.Logger, conf config.Logging) {

	// Format
	switch conf.Format {
	case "text":
		logger.SetFormatter(&logrus.TextFormatter{DisableTimestamp: true, ForceColors: conf.Output != "file"})
	case "json":
		logger.SetFormatter(&logrus.JSONFormatter{})
	default:
		panic("Logging format must be either 'text' or 'json'. Got: " + conf.Format)
	}

	// Output
	switch conf.Output {
	case "stdout":
		logger.SetOutput(os.Stdout)
	case "stderr":
		logger.SetOutput(os.Stderr)
	case "file":
		_, codeFilePath, _, _ := runtime.Caller(0)
		codeDir := filepath.Dir(codeFilePath)
		f, err := os.OpenFile(codeDir+"/../../log/all.log", os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0600)
		if err != nil {
			logger.SetOutput(os.Stdout)
			logger.Errorf("Unable to open file for logs, fallback to stdout: %v\n", err)
		} else {
			logger.SetOutput(f)
		}
	default:
		panic("Logging output must be either 'stdout', 'stderr' or 'file'. Got: " + conf.Output)
	}

	// Level
	if level, err := logrus.ParseLevel(conf.Level); err != nil {
		logger.Errorf("Unable to parse logging level config, use default (%s)", logger.GetLevel().String())
	} else {
		logger.SetLevel(level)
	}
}
