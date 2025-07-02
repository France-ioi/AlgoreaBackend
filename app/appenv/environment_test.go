package appenv

import (
	"errors"
	"os"
	"testing"

	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"
)

func TestEnv_Prod(t *testing.T) {
	t.Setenv(envVarName, "prod")

	assert.Equal(t, "prod", Env())
	assert.False(t, IsEnvDev())
	assert.False(t, IsEnvTest())
	assert.True(t, IsEnvProd())
}

func TestEnv_Dev(t *testing.T) {
	t.Setenv(envVarName, "dev")

	assert.Equal(t, "dev", Env())
	assert.True(t, IsEnvDev())
	assert.False(t, IsEnvTest())
	assert.False(t, IsEnvProd())
}

func TestEnv_Test(t *testing.T) {
	t.Setenv(envVarName, "test")

	assert.Equal(t, "test", Env())
	assert.False(t, IsEnvDev())
	assert.True(t, IsEnvTest())
	assert.False(t, IsEnvProd())
}

func TestEnv_NotSet(t *testing.T) {
	t.Setenv(envVarName, "") // make the test restore the environment variable on cleanup
	_ = os.Unsetenv(envVarName)

	assert.Equal(t, "dev", Env())
	assert.True(t, IsEnvDev())
	assert.False(t, IsEnvTest())
	assert.False(t, IsEnvProd())
}

func TestEnv_Empty(t *testing.T) {
	t.Setenv(envVarName, "")

	assert.Equal(t, "dev", Env())
	assert.True(t, IsEnvDev())
	assert.False(t, IsEnvTest())
	assert.False(t, IsEnvProd())
}

func TestEnv_Other(t *testing.T) {
	t.Setenv(envVarName, "myownenv")

	assert.Equal(t, "myownenv", Env())
	assert.False(t, IsEnvDev())
	assert.False(t, IsEnvTest())
	assert.False(t, IsEnvProd())
}

func TestSetDefaultEnvToTest_NotSet(t *testing.T) {
	t.Setenv(envVarName, "") // make the test restore the environment variable on cleanup
	_ = os.Unsetenv(envVarName)
	SetDefaultEnvToTest()

	assert.Equal(t, "test", Env())
	assert.False(t, IsEnvDev())
	assert.True(t, IsEnvTest())
	assert.False(t, IsEnvProd())
}

func TestSetDefaultEnvToTest_Set(t *testing.T) {
	t.Setenv(envVarName, "prod")
	SetDefaultEnvToTest()

	assert.Equal(t, "prod", Env())
	assert.False(t, IsEnvDev())
	assert.False(t, IsEnvTest())
	assert.True(t, IsEnvProd())
}

func TestSetDefaultEnvToTest_Panic(t *testing.T) {
	t.Setenv(envVarName, "") // make the test restore the environment variable on cleanup
	_ = os.Unsetenv(envVarName)

	//nolint:usetesting // here we patch os.Setenv to simulate an error
	monkey.Patch(os.Setenv, func(string, string) error {
		return errors.New("unexpected error")
	})
	defer monkey.UnpatchAll()

	assert.Panics(t, func() {
		SetDefaultEnv("prod")
	})
}

func TestSetEnv_Ok(t *testing.T) {
	t.Setenv(envVarName, "prod")
	SetEnv("myEnv")

	assert.Equal(t, "myEnv", Env())
}

func TestSetEnv_Panic(t *testing.T) {
	t.Setenv(envVarName, "prod")

	//nolint:usetesting // here we patch os.Setenv to simulate an error
	monkey.Patch(os.Setenv, func(string, string) error {
		return errors.New("unexpected error")
	})
	defer monkey.UnpatchAll()

	assert.Panics(t, func() {
		SetEnv("myEnv")
	})
}

func TestForceTestEnv(t *testing.T) {
	t.Setenv(envVarName, "prod")
	ForceTestEnv()

	assert.Equal(t, "test", Env())

	assert.Panics(t, func() {
		SetEnv("prod")
	})
	assert.Panics(t, func() {
		SetEnv("dev")
	})
	assert.NotPanics(t, func() {
		SetEnv("test")
	})
}
