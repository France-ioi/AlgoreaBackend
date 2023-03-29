package logging

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"bou.ke/monkey"
	"github.com/sirupsen/logrus" //nolint:depguard
	"github.com/spf13/viper"
	assertlib "github.com/stretchr/testify/assert"
)

func TestGlobal(t *testing.T) {
	assert := assertlib.New(t)
	ResetShared()
	assert.IsType(&logrus.TextFormatter{}, SharedLogger.Formatter) // TextFormatter is logrus default
	conf := viper.New()
	conf.Set("Format", "json")
	conf.Set("Output", "file")
	SharedLogger.Configure(conf)
	assert.IsType(&logrus.JSONFormatter{}, SharedLogger.Formatter)
	ResetShared()
}

func TestConfigure_FormatText(t *testing.T) {
	assert := assertlib.New(t)
	conf := viper.New()
	conf.Set("Format", "text")
	conf.Set("Output", "file")
	logger := createLogger()
	logger.Configure(conf)
	assert.IsType(&logrus.TextFormatter{}, logger.Formatter)
}

func TestConfigure_FormatJson(t *testing.T) {
	assert := assertlib.New(t)
	conf := viper.New()
	conf.Set("Format", "json")
	conf.Set("Output", "file")
	logger := createLogger()
	logger.Configure(conf)
	assert.IsType(&logrus.JSONFormatter{}, logger.Formatter)
}

func TestConfigure_FormatInvalid(t *testing.T) {
	assert := assertlib.New(t)
	conf := viper.New()
	conf.Set("Format", "yml")
	conf.Set("Output", "file")
	logger := createLogger()
	assert.Panics(func() { logger.Configure(conf) })
}

func TestConfigure_OutputStdout(t *testing.T) {
	assert := assertlib.New(t)
	conf := viper.New()
	conf.Set("Format", "json")
	conf.Set("Output", "stdout")
	logger := createLogger()
	logger.Configure(conf)
	assert.Equal(os.Stdout, logger.Out)
}

func TestConfigure_OutputStderr(t *testing.T) {
	assert := assertlib.New(t)
	conf := viper.New()
	conf.Set("Format", "json")
	conf.Set("Output", "stderr")
	logger := createLogger()
	logger.Configure(conf)
	assert.Equal(os.Stderr, logger.Out)
}

func TestConfigure_OutputFile(t *testing.T) {
	assert := assertlib.New(t)
	conf := viper.New()
	conf.Set("Format", "json")
	conf.Set("Output", "file")

	// will append time to make sure not to match a prev exec of the test
	timestamp := time.Now().UnixNano()

	logger1 := createLogger()
	logger1.Configure(conf)
	logger1.Errorf("logexec1 %d", timestamp)

	// redo another init to check it will not override
	logger2 := createLogger()
	logger2.Configure(conf)
	logger2.Warnf("logexec2 %d", timestamp)

	// check the resulting file
	content, _ := ioutil.ReadFile("../../log/all.log")
	assert.Contains(string(content), fmt.Sprintf("logexec1 %d", timestamp))
	assert.Contains(string(content), fmt.Sprintf("logexec2 %d", timestamp))
}

func TestConfigure_OutputFileError(t *testing.T) {
	assert := assertlib.New(t)
	conf := viper.New()
	conf.Set("Format", "json")
	conf.Set("Output", "file")
	fakeFunc := func(name string, flag int, perm os.FileMode) (*os.File, error) {
		return nil, errors.New("open error")
	}
	patch := monkey.Patch(os.OpenFile, fakeFunc)
	defer patch.Unpatch()
	logger := createLogger()
	logger.Configure(conf)
	assert.Equal(os.Stdout, logger.Out)
}

func TestConfigure_OutputInvalid(t *testing.T) {
	assert := assertlib.New(t)
	conf := viper.New()
	conf.Set("Format", "json")
	conf.Set("Output", "S3")
	logger := createLogger()
	assert.Panics(func() { logger.Configure(conf) })
}

func TestConfigure_LevelDefault(t *testing.T) {
	assert := assertlib.New(t)
	conf := viper.New()
	conf.Set("Format", "text")
	conf.Set("Output", "file")
	conf.Set("Level", "")
	logger := createLogger()
	logger.Configure(conf)
	assert.Equal(logrus.InfoLevel, logger.Level)
}

func TestConfigure_LevelParsed(t *testing.T) {
	assert := assertlib.New(t)
	conf := viper.New()
	conf.Set("Format", "text")
	conf.Set("Output", "file")
	conf.Set("Level", "warn")
	logger := createLogger()
	logger.Configure(conf)
	assert.Equal(logrus.WarnLevel, logger.Level)
}

func TestConfigure_LevelInvalid(t *testing.T) {
	assert := assertlib.New(t)
	conf := viper.New()
	conf.Set("Format", "text")
	conf.Set("Output", "file")
	conf.Set("Level", "invalid_level")
	logger := createLogger()
	logger.Configure(conf)
	assert.Equal(logrus.InfoLevel, logger.Level)
}
