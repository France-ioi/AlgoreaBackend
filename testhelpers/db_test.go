// +build !prod

package testhelpers_test

import (
	"github.com/France-ioi/AlgoreaBackend/app"
	"github.com/France-ioi/AlgoreaBackend/app/appenv"
	"github.com/France-ioi/AlgoreaBackend/app/logging"
)

func init() { // nolint:gochecknoinits
	appenv.SetDefaultEnvToTest()
	// Apply the config to the global logger
	logging.SharedLogger.Configure(app.LoggingConfig(app.LoadConfig()))
}
