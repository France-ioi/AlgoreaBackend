package config

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	assert_lib "github.com/stretchr/testify/assert"
)

func TestLoadConfig(t *testing.T) {
	assert := assert_lib.New(t)

	// create a temp config file
	tmpDir := os.TempDir()
	tmpFile, err := ioutil.TempFile(tmpDir, "config-*.yaml")
	assert.NoError(err)
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	text := []byte("server:\n  port: 1234\n")
	_, err = tmpFile.Write(text)
	assert.NoError(err)

	// change default config values
	fileName := filepath.Base(tmpFile.Name())
	configName = fileName[:len(fileName)-5] // strip the ".yaml"
	configDir = tmpDir

	os.Setenv("ALGOREA_SERVER.WRITETIMEOUT", "999")
	conf, err := Load()
	assert.NoError(err)

	// test config override
	assert.EqualValues(1234, conf.Server.Port)

	// test default
	assert.EqualValues(60, conf.Server.ReadTimeout)

	// test env
	assert.EqualValues(999, conf.Server.WriteTimeout)
}
