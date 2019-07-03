package common

import "os"

const envVarName = "ALGOREA_ENV"
const productionEnv = "prod"
const developementEnv = "dev"
const testEnv = "test"

// Env returns the deployment environment set for this app ("prod", "dev", or "test"). Default to "dev".
func Env() string {
	switch os.Getenv(envVarName) {
	case productionEnv:
		return productionEnv
	case testEnv:
		return testEnv
	default:
		return developementEnv
	}
}

// SetDefaultEnv set the deployment environment to the given value if not set.
func SetDefaultEnv(newVal string) {
	if _, ok := os.LookupEnv(envVarName); ok {
		return // already set
	}
	if os.Setenv(envVarName, newVal) != nil {
		panic("unable to set env variable")
	}
}

// SetDefaultEnvToTest set the deployment environment to the "test" if not set.
func SetDefaultEnvToTest() {
	SetDefaultEnv(testEnv)
}

// IsEnvTest return whether the app is in "test" environment
func IsEnvTest() bool {
	return Env() == testEnv
}

// IsEnvDev return whether the app is in "dev" environment
func IsEnvDev() bool {
	return Env() == developementEnv
}
