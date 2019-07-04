package common

import (
	"errors"
	"os"
	"testing"

	"bou.ke/monkey"

	"github.com/stretchr/testify/assert"
)

func TestEnv_Prod(t *testing.T) {
	_ = os.Setenv(envVarName, "prod")
	assert.Equal(t, "prod", Env())
	assert.False(t, IsEnvDev())
	assert.False(t, IsEnvTest())
	assert.True(t, IsEnvProd())
}

func TestEnv_Dev(t *testing.T) {
	_ = os.Setenv(envVarName, "dev")
	assert.Equal(t, "dev", Env())
	assert.True(t, IsEnvDev())
	assert.False(t, IsEnvTest())
	assert.False(t, IsEnvProd())
}

func TestEnv_Test(t *testing.T) {
	_ = os.Setenv(envVarName, "test")
	assert.Equal(t, "test", Env())
	assert.False(t, IsEnvDev())
	assert.True(t, IsEnvTest())
	assert.False(t, IsEnvProd())
}

func TestEnv_NotSet(t *testing.T) {
	_ = os.Unsetenv(envVarName)
	assert.Equal(t, "dev", Env())
	assert.True(t, IsEnvDev())
	assert.False(t, IsEnvTest())
	assert.False(t, IsEnvProd())
}

func TestEnv_Empty(t *testing.T) {
	_ = os.Setenv(envVarName, "")
	assert.Equal(t, "dev", Env())
	assert.True(t, IsEnvDev())
	assert.False(t, IsEnvTest())
	assert.False(t, IsEnvProd())
}

func TestEnv_Other(t *testing.T) {
	_ = os.Setenv(envVarName, "myownenv")
	assert.Equal(t, "myownenv", Env())
	assert.False(t, IsEnvDev())
	assert.False(t, IsEnvTest())
	assert.False(t, IsEnvProd())
}

func TestSetDefaultEnvToTest_NotSet(t *testing.T) {
	_ = os.Unsetenv(envVarName)
	SetDefaultEnvToTest()
	assert.Equal(t, "test", Env())
	assert.False(t, IsEnvDev())
	assert.True(t, IsEnvTest())
	assert.False(t, IsEnvProd())
}

func TestSetDefaultEnvToTest_Set(t *testing.T) {
	_ = os.Setenv(envVarName, "prod")
	SetDefaultEnvToTest()
	assert.Equal(t, "prod", Env())
	assert.False(t, IsEnvDev())
	assert.False(t, IsEnvTest())
	assert.True(t, IsEnvProd())
}

func TestSetDefaultEnvToTest_Panic(t *testing.T) {
	_ = os.Unsetenv(envVarName)
	monkey.Patch(os.Setenv, func(string, string) error {
		return errors.New("unexpected error")
	})
	defer monkey.UnpatchAll()
	assert.Panics(t, func() {
		SetDefaultEnv("prod")
	})
}
