package config

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"bou.ke/monkey"
	assertlib "github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/app/appenv"
)

func TestLoad(t *testing.T) {
	assert := assertlib.New(t)
	appenv.SetDefaultEnvToTest() // to ensure it tries to find the config.test file

	// create a temp config file
	tmpDir := os.TempDir()
	tmpFile, err := ioutil.TempFile(tmpDir, "config-*.yaml")
	assert.NoError(err)
	defer func() {
		_ = os.Remove(tmpFile.Name())
		_ = tmpFile.Close()
	}()

	text := []byte("server:\n  port: 1234\n")
	_, err = tmpFile.Write(text)
	assert.NoError(err)

	// change default config values
	fileName := filepath.Base(tmpFile.Name())
	configName = fileName[:len(fileName)-5] // strip the ".yaml"
	configDir = tmpDir
	defer func() {
		configName = defaultConfigName
		configDir = configDirectory()
	}()

	tmpTestFileName := tmpDir + "/" + configName + ".test.yaml"
	err = ioutil.WriteFile(tmpTestFileName, []byte("server:\n  rootpath: '/test/'"), 0644)
	assert.NoError(err)
	defer func() {
		_ = os.Remove(tmpTestFileName)
	}()

	_ = os.Setenv("ALGOREA_SERVER__WRITETIMEOUT", "999")
	conf := Load()

	// test config override
	assert.EqualValues(1234, conf.Server.Port)

	// test default
	assert.EqualValues(60, conf.Server.ReadTimeout)

	// test env
	assert.EqualValues(999, conf.Server.WriteTimeout)

	// test 'test' section
	assert.EqualValues("/test/", conf.Server.RootPath)
}

func TestLoad_IgnoresMainConfigFileIfMissing(t *testing.T) {
	assert := assertlib.New(t)
	appenv.SetDefaultEnvToTest() // to ensure it tries to find the config.test file

	// create a temp config file
	tmpFile, deferFunc := createTmpFile("config-*.test.yaml", assert)
	defer deferFunc()

	fileName := filepath.Base(tmpFile.Name())
	configName = fileName[:len(fileName)-10] // strip the ".test.yaml"
	configDir = os.TempDir()
	defer func() {
		configName = defaultConfigName
		configDir = configDirectory()
	}()

	conf := Load()
	assert.NotNil(conf)
}

func TestLoad_IgnoresEnvConfigFileIfMissing(t *testing.T) {
	assert := assertlib.New(t)
	appenv.SetDefaultEnvToTest() // to ensure it tries to find the config.test file

	// create a temp config file
	tmpDir := os.TempDir()
	tmpFile, deferFunc := createTmpFile("config-*.yaml", assert)
	defer deferFunc()

	fileName := filepath.Base(tmpFile.Name())
	configName = fileName[:len(fileName)-5] // strip the ".yaml"
	configDir = tmpDir
	defer func() {
		configName = defaultConfigName
		configDir = configDirectory()
	}()

	conf := Load()
	assert.NotNil(conf)
}

func TestLoad_CannotUnmarshal(t *testing.T) {
	assert := assertlib.New(t)

	// create a temp config file
	tmpFile, deferFunc := createTmpFile("config-*.yaml", assert)
	defer deferFunc()

	text := []byte("unknown: 1234\n")
	_, err := tmpFile.Write(text)
	assert.NoError(err)

	// change default config values
	fileName := filepath.Base(tmpFile.Name())
	configName = fileName[:len(fileName)-5] // strip the ".yaml"
	configDir = os.TempDir()
	defer func() {
		configName = defaultConfigName
		configDir = configDirectory()
	}()

	called := false
	monkey.Patch(os.Exit, func(int) { called = true })
	defer monkey.UnpatchAll()
	Load()
	assert.True(called)
}

func TestLoad_Concurrent(t *testing.T) {
	_ = os.Unsetenv("ALGOREA_ENV")
	appenv.SetDefaultEnvToTest()
	assert := assertlib.New(t)
	assert.NotPanics(func() {
		Load()
		for i := 0; i < 1000; i++ {
			go func() { Load() }()
		}
	})
}

func createTmpFile(pattern string, assert *assertlib.Assertions) (tmpFile *os.File, deferFunc func()) {
	// create a temp config file
	tmpFile, err := ioutil.TempFile(os.TempDir(), pattern)
	assert.NoError(err)
	return tmpFile, func() {
		_ = os.Remove(tmpFile.Name())
		_ = tmpFile.Close()
	}
}
