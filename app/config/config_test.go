package config

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	assertlib "github.com/stretchr/testify/assert"
)

func TestLoadConfig(t *testing.T) {
	assert := assertlib.New(t)

	ClearCachedConfig()
	defer ClearCachedConfig()

	// create a temp config file
	tmpDir := os.TempDir()
	tmpFile, err := writeConfigIntoTempFile(tmpDir, "server:\n  port: 1234\n")
	assert.NoError(err)
	if err != nil {
		return
	}
	defer func() { _ = os.Remove(tmpFile.Name()) }()

	// change default config values
	fileName := filepath.Base(tmpFile.Name())
	configName = fileName[:len(fileName)-5] // strip the ".yaml"
	configDir = tmpDir

	_ = os.Setenv("ALGOREA_SERVER.WRITETIMEOUT", "999")
	conf, err := Load()
	assert.NoError(err)

	// test config override
	assert.EqualValues(1234, conf.Server.Port)

	// test default
	assert.EqualValues(60, conf.Server.ReadTimeout)

	// test env
	assert.EqualValues(999, conf.Server.WriteTimeout)
}

func TestLoadConfig_Caches(t *testing.T) {
	assert := assertlib.New(t)

	ClearCachedConfig()
	defer ClearCachedConfig()

	// create a temp config file
	tmpDir := os.TempDir()
	tmpFile, err := writeConfigIntoTempFile(tmpDir, "server:\n  port: 5678\n")
	assert.NoError(err)
	if err != nil {
		return
	}
	defer func() { _ = os.Remove(tmpFile.Name()) }()

	// change default config values
	fileName := filepath.Base(tmpFile.Name())
	configName = fileName[:len(fileName)-5] // strip the ".yaml"
	configDir = tmpDir

	conf, err := Load()
	assert.NoError(err)

	// test config override
	assert.EqualValues(5678, conf.Server.Port)

	oldConfig := loadedConfig

	tmpFile, err = writeConfigIntoTempFile(tmpDir, "server:\n  port: 1234\n")
	assert.NoError(err)
	if err != nil {
		return
	}
	defer func() { _ = os.Remove(tmpFile.Name()) }()

	// change default config values
	fileName = filepath.Base(tmpFile.Name())
	configName = fileName[:len(fileName)-5] // strip the ".yaml"

	conf, err = Load()
	assert.NoError(err)
	newConfig := loadedConfig

	assert.EqualValues(5678, conf.Server.Port)
	assert.Equal(oldConfig, newConfig)
}

func writeConfigIntoTempFile(dir, content string) (*os.File, error) {
	tmpFile, err := ioutil.TempFile(dir, "config-*.yaml")
	if err != nil {
		return nil, err
	}
	defer func() { _ = tmpFile.Close() }()

	_, err = tmpFile.Write([]byte(content))
	if err != nil {
		return nil, err
	}
	return tmpFile, nil
}
