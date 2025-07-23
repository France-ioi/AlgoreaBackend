package testhelpers

import (
	"context"

	"github.com/spf13/viper"

	"github.com/France-ioi/AlgoreaBackend/v2/app"
	"github.com/France-ioi/AlgoreaBackend/v2/app/appenv"
	"github.com/France-ioi/AlgoreaBackend/v2/app/logging"
)

type testContextKey int

const (
	testContextConfigKey testContextKey = iota
)

// CreateTestContext creates a context for tests with a logger and the config.
// It forces the test environment and loads the configuration from the config file.
// The logger is created from the configuration.
func CreateTestContext() context.Context {
	appenv.ForceTestEnv()

	// needs actual config for connection to DB
	config := app.LoadConfig()
	loggingConfig := app.LoggingConfig(config)
	return context.WithValue(
		logging.ContextWithLogger(context.Background(), logging.NewLoggerFromConfig(loggingConfig)),
		testContextConfigKey, config)
}

// CreateTestContextWithLogger creates a context for tests with the given logger and the config.
// It forces the test environment and loads the configuration from the config file.
// The logger is not created from the configuration, it is passed as an argument.
func CreateTestContextWithLogger(logger *logging.Logger) context.Context {
	appenv.ForceTestEnv()

	// needs actual config for connection to DB
	config := app.LoadConfig()
	return context.WithValue(
		logging.ContextWithLogger(context.Background(), logger),
		testContextConfigKey, config)
}

// GetConfigFromContext returns the configuration from the given context
// initialized by CreateTestContext or CreateTestContextWithLogger.
func GetConfigFromContext(ctx context.Context) *viper.Viper {
	return ctx.Value(testContextConfigKey).(*viper.Viper)
}
