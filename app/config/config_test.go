package config

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"bou.ke/monkey"
	assertlib "github.com/stretchr/testify/assert"
)

func TestLoadConfig(t *testing.T) {
	assert := assertlib.New(t)

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

	tmpTestFileName := tmpDir + "/" + configName + ".test.yaml"
	err = ioutil.WriteFile(tmpTestFileName, []byte("server:\n  rootpath: '/test/'"), 0644)
	assert.NoError(err)
	defer func() {
		_ = os.Remove(tmpTestFileName)
	}()

	_ = os.Setenv("ALGOREA_SERVER.WRITETIMEOUT", "999")
	conf, err := Load("test")
	assert.NoError(err)

	// test config override
	assert.EqualValues(1234, conf.Server.Port)

	// test default
	assert.EqualValues(60, conf.Server.ReadTimeout)

	// test env
	assert.EqualValues(999, conf.Server.WriteTimeout)

	// test 'test' section
	assert.EqualValues("/test/", conf.Server.RootPath)
}

func TestLoadConfig_CannotUnmarshal(t *testing.T) {
	assert := assertlib.New(t)

	// create a temp config file
	tmpDir := os.TempDir()
	tmpFile, err := ioutil.TempFile(tmpDir, "config-*.yaml")
	assert.NoError(err)
	defer func() {
		_ = os.Remove(tmpFile.Name())
		_ = tmpFile.Close()
	}()

	text := []byte("unknown: 1234\n")
	_, err = tmpFile.Write(text)
	assert.NoError(err)

	// change default config values
	fileName := filepath.Base(tmpFile.Name())
	configName = fileName[:len(fileName)-5] // strip the ".yaml"
	configDir = tmpDir

	called := false
	monkey.Patch(os.Exit, func(int) { called = true })
	defer monkey.UnpatchAll()
	_, err = Load("test")
	assert.Error(err)
	assert.True(called)
}

func TestLoadConfig_Concurrent(t *testing.T) {
	assert := assertlib.New(t)
	assert.NotPanics(func() {
		_, _ = Load("test")
		for i := 0; i < 1000; i++ {
			go func() { _, _ = Load("test") }()
		}
	})
}
