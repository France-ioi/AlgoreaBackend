package app

import (
	"github.com/sirupsen/logrus"

	"github.com/France-ioi/AlgoreaBackend/app/config"
)

func setupLogger(logger *logrus.Logger, conf config.Logging) {

	if conf.TextLogging {
		logger.SetFormatter(&logrus.TextFormatter{DisableTimestamp: true})
	} else {
		logger.SetFormatter(&logrus.JSONFormatter{})
	}

	if level, err := logrus.ParseLevel(conf.Level); err != nil {
		logger.Errorf("Unable to parse logging level config, use default (%s)", logger.GetLevel().String())
	} else {
		logger.SetLevel(level)
	}
}
