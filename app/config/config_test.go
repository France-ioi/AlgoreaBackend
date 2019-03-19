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

func TestLoadConfig_Concurrent(t *testing.T) {
	assert := assertlib.New(t)
	assert.NotPanics(func() {
		_, _ = Load()
		for i := 0; i < 1000; i++ {
			go func() { _, _ = Load() }()
		}
	})
}
