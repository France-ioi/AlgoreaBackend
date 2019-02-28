package app

import (
	"fmt"
	"log"
	"os"

	"github.com/sirupsen/logrus"

	"github.com/France-ioi/AlgoreaBackend/app/config"
)

func setupLogger(logger *logrus.Logger, conf config.Logging) {

	// Format
	switch conf.Format {
	case "text":
		logger.SetFormatter(&logrus.TextFormatter{DisableTimestamp: true})
	case "json":
		logger.SetFormatter(&logrus.JSONFormatter{})
	default:
		panic("Logging format must be either 'text' or 'json'")
	}

	// Output
	switch conf.Output {
	case "stdout":
		log.SetOutput(os.Stdout)
	case "stderr":
		log.SetOutput(os.Stderr)
	case "file":
		f, err := os.OpenFile("bdd_test.log", os.O_WRONLY|os.O_CREATE, 0755)
		if err != nil {
			fmt.Printf("Unable to open file for logs, fallback to default: %v\n", err)
		} else {
			log.SetOutput(f)
		}
	default:
		panic("Logging output must be either 'stdout' or 'file'")
	}

	// Level
	if level, err := logrus.ParseLevel(conf.Level); err != nil {
		logger.Errorf("Unable to parse logging level config, use default (%s)", logger.GetLevel().String())
	} else {
		logger.SetLevel(level)
	}
}
