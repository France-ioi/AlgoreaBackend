// Package appenv provides utilities to configure the app environment (prod, dev, test).
package appenv

import "os"

const (
	envVarName     = "ALGOREA_ENV"
	productionEnv  = "prod"
	developmentEnv = "dev"
	testEnv        = "test"
)

// Env returns the deployment environment set for this app ("prod", "dev", or "test"). Default to "dev".
func Env() string {
	env := os.Getenv(envVarName)
	if env != "" {
		return env
	}
	// else, not set or set to empty string
	return developmentEnv
}

// SetDefaultEnv sets the deployment environment to the given value if not set.
func SetDefaultEnv(newVal string) {
	if _, ok := os.LookupEnv(envVarName); ok {
		return // already set
	}

	SetEnv(newVal)
}

// SetEnv sets the deployment environment to the given value.
func SetEnv(newVal string) {
	if os.Setenv(envVarName, newVal) != nil {
		panic("unable to set env variable")
	}
}

// SetDefaultEnvToTest set the deployment environment to the "test" if not set.
func SetDefaultEnvToTest() {
	SetDefaultEnv(testEnv)
}

// IsEnvTest return whether the app is in "test" environment.
func IsEnvTest() bool {
	return Env() == testEnv
}

// IsEnvDev return whether the app is in "dev" environment.
func IsEnvDev() bool {
	return Env() == developmentEnv
}

// IsEnvProd return whether the app is in "prod" environment.
func IsEnvProd() bool {
	return Env() == productionEnv
}
