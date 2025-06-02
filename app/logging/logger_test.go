package logging

import (
	"context"
	"errors"
	"fmt"
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
	assert.IsType(&logrus.TextFormatter{}, SharedLogger.logrusLogger.Formatter) // TextFormatter is logrus default
	conf := viper.New()
	conf.Set("Format", "json")
	conf.Set("Output", "stdout")
	SharedLogger.Configure(conf)
	assert.IsType(&jsonFormatter{}, SharedLogger.logrusLogger.Formatter)
	ResetShared()
}

func TestLogger_Configure_FormatText(t *testing.T) {
	assert := assertlib.New(t)
	conf := viper.New()
	conf.Set("Format", "text")
	conf.Set("Output", "stdout")
	logger := createLogger()
	logger.Configure(conf)
	assert.IsType(&textFormatter{}, logger.logrusLogger.Formatter)
}

func TestLogger_Configure_FormatJson(t *testing.T) {
	assert := assertlib.New(t)
	conf := viper.New()
	conf.Set("Format", "json")
	conf.Set("Output", "stdout")
	logger := createLogger()
	logger.Configure(conf)
	assert.IsType(&jsonFormatter{}, logger.logrusLogger.Formatter)
}

func TestLogger_Configure_FormatConsole(t *testing.T) {
	assert := assertlib.New(t)
	conf := viper.New()
	conf.Set("Format", "console")
	conf.Set("Output", "stdout")
	logger := createLogger()
	logger.Configure(conf)
	assert.IsType(&consoleFormatter{}, logger.logrusLogger.Formatter)
}

func TestLogger_Configure_FormatInvalid(t *testing.T) {
	assert := assertlib.New(t)
	conf := viper.New()
	conf.Set("Format", "yml")
	conf.Set("Output", "stdout")
	logger := createLogger()
	assert.Panics(func() { logger.Configure(conf) })
}

func TestLogger_Configure_OutputStdout(t *testing.T) {
	assert := assertlib.New(t)
	conf := viper.New()
	conf.Set("Format", "json")
	conf.Set("Output", "stdout")
	logger := createLogger()
	logger.Configure(conf)
	assert.Equal(os.Stdout, logger.logrusLogger.Out)
}

func TestLogger_Configure_OutputStderr(t *testing.T) {
	assert := assertlib.New(t)
	conf := viper.New()
	conf.Set("Format", "json")
	conf.Set("Output", "stderr")
	logger := createLogger()
	logger.Configure(conf)
	assert.Equal(os.Stderr, logger.logrusLogger.Out)
}

func TestLogger_Configure_OutputFile(t *testing.T) {
	assert := assertlib.New(t)
	conf := viper.New()
	conf.Set("Format", "json")
	conf.Set("Output", "file")

	_ = os.Remove("../../log/all.log.test")
	defer func() { _ = os.Remove("../../log/all.log.test") }()
	var patchGuard *monkey.PatchGuard
	patchGuard = monkey.Patch(os.OpenFile, func(name string, flag int, perm os.FileMode) (*os.File, error) {
		patchGuard.Unpatch()
		defer patchGuard.Restore()
		return os.OpenFile(name+".test", flag, perm)
	})
	defer patchGuard.Unpatch()

	// will append time to make sure not to match a prev exec of the test
	timestamp := time.Now().UnixNano()

	logger1 := createLogger()
	logger1.Configure(conf)
	logger1.WithContext(context.Background()).Errorf("logexec1 %d", timestamp)

	// redo another init to check it will not override
	logger2 := createLogger()
	logger2.Configure(conf)
	logger2.WithContext(context.Background()).Warnf("logexec2 %d", timestamp)

	// check the resulting file
	content, _ := os.ReadFile("../../log/all.log")
	assert.Contains(string(content), fmt.Sprintf("logexec1 %d", timestamp))
	assert.Contains(string(content), fmt.Sprintf("logexec2 %d", timestamp))
}

func TestLogger_Configure_OutputFileError(t *testing.T) {
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
	assert.Equal(os.Stdout, logger.logrusLogger.Out)
}

func TestLogger_Configure_OutputInvalid(t *testing.T) {
	assert := assertlib.New(t)
	conf := viper.New()
	conf.Set("Format", "json")
	conf.Set("Output", "S3")
	logger := createLogger()
	assert.Panics(func() { logger.Configure(conf) })
}

func TestLogger_Configure_FormatConsoleAndFileOutput(t *testing.T) {
	assert := assertlib.New(t)
	conf := viper.New()
	conf.Set("Format", "console")
	conf.Set("Output", "file")
	logger := createLogger()
	assert.Panics(func() { logger.Configure(conf) })
}

func TestLogger_Configure_LevelDefault(t *testing.T) {
	assert := assertlib.New(t)
	conf := viper.New()
	conf.Set("Format", "text")
	conf.Set("Output", "stdout")
	conf.Set("Level", "")
	logger := createLogger()
	logger.Configure(conf)
	assert.Equal(logrus.InfoLevel, logger.logrusLogger.Level)
}

func TestLogger_Configure_LevelParsed(t *testing.T) {
	assert := assertlib.New(t)
	conf := viper.New()
	conf.Set("Format", "text")
	conf.Set("Output", "stdout")
	conf.Set("Level", "warn")
	logger := createLogger()
	logger.Configure(conf)
	assert.Equal(logrus.WarnLevel, logger.logrusLogger.Level)
}

func TestLogger_Configure_LevelInvalid(t *testing.T) {
	assert := assertlib.New(t)
	conf := viper.New()
	conf.Set("Format", "text")
	conf.Set("Output", "stdout")
	conf.Set("Level", "invalid_level")
	logger := createLogger()
	logger.Configure(conf)
	assert.Equal(logrus.InfoLevel, logger.logrusLogger.Level)
}
