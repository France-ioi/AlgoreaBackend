package logging

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"bou.ke/monkey"
	"github.com/sirupsen/logrus"
	assertlib "github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/app/config"
)

func TestGlobal(t *testing.T) {
	assert := assertlib.New(t)
	ResetGlobal()
	assert.IsType(&logrus.TextFormatter{}, Logger.Formatter) // TextFormatter is logrus default
	conf := config.Logging{
		Format: "json",
		Output: "file",
	}
	ConfigureGlobal(conf)
	assert.IsType(&logrus.JSONFormatter{}, Logger.Formatter)
	ResetGlobal()
}

func TestConfigure_FormatText(t *testing.T) {
	assert := assertlib.New(t)
	conf := config.Logging{
		Format: "text",
		Output: "file",
	}
	logger := logrus.New()
	configure(logger, conf)
	assert.IsType(&logrus.TextFormatter{}, logger.Formatter)
}

func TestConfigure_FormatJson(t *testing.T) {
	assert := assertlib.New(t)
	conf := config.Logging{
		Format: "json",
		Output: "file",
	}
	logger := logrus.New()
	configure(logger, conf)
	assert.IsType(&logrus.JSONFormatter{}, logger.Formatter)
}

func TestConfigure_FormatInvalid(t *testing.T) {
	assert := assertlib.New(t)
	conf := config.Logging{
		Format: "yml",
		Output: "file",
	}
	logger := logrus.New()
	assert.Panics(func() { configure(logger, conf) })
}

func TestConfigure_OutputStdout(t *testing.T) {
	assert := assertlib.New(t)
	conf := config.Logging{
		Format: "json",
		Output: "stdout",
	}
	logger := logrus.New()
	configure(logger, conf)
	assert.Equal(os.Stdout, logger.Out)
}

func TestConfigure_OutputStderr(t *testing.T) {
	assert := assertlib.New(t)
	conf := config.Logging{
		Format: "json",
		Output: "stderr",
	}
	logger := logrus.New()
	configure(logger, conf)
	assert.Equal(os.Stderr, logger.Out)
}

func TestConfigure_OutputFile(t *testing.T) {
	assert := assertlib.New(t)
	conf := config.Logging{
		Format: "json",
		Output: "file",
	}
	// will append time to make sure not to match a prev exec of the test
	timestamp := time.Now().UnixNano()

	logger := logrus.New()
	configure(logger, conf)
	logger.Errorf("logexec1 %d", timestamp)

	// redo another init to check it will not override
	logger2 := logrus.New()
	configure(logger2, conf)
	logger2.Warnf("logexec2 %d", timestamp)

	// check the resulting file
	content, _ := ioutil.ReadFile("../../log/all.log")
	assert.Contains(string(content), fmt.Sprintf("logexec1 %d", timestamp))
	assert.Contains(string(content), fmt.Sprintf("logexec2 %d", timestamp))
}

func TestConfigure_OutputFileError(t *testing.T) {
	assert := assertlib.New(t)
	conf := config.Logging{
		Format: "json",
		Output: "file",
	}
	fakeFunc := func(name string, flag int, perm os.FileMode) (*os.File, error) {
		return nil, errors.New("open error")
	}
	patch := monkey.Patch(os.OpenFile, fakeFunc)
	defer patch.Unpatch()
	logger := logrus.New()
	configure(logger, conf)
	assert.Equal(os.Stdout, logger.Out)
}

func TestConfigure_OutputInvalid(t *testing.T) {
	assert := assertlib.New(t)
	conf := config.Logging{
		Format: "json",
		Output: "S3",
	}
	logger := logrus.New()
	assert.Panics(func() { configure(logger, conf) })
}

func TestConfigure_LevelDefault(t *testing.T) {
	assert := assertlib.New(t)
	conf := config.Logging{
		Level:  "",
		Format: "text",
		Output: "file",
	}
	logger := logrus.New()
	configure(logger, conf)
	assert.Equal(logrus.InfoLevel, logger.Level)
}

func TestConfigure_LevelParsed(t *testing.T) {
	assert := assertlib.New(t)
	conf := config.Logging{
		Level:  "warn",
		Format: "text",
		Output: "file",
	}
	logger := logrus.New()
	configure(logger, conf)
	assert.Equal(logrus.WarnLevel, logger.Level)
}

func TestConfigure_LevelInvalid(t *testing.T) {
	assert := assertlib.New(t)
	conf := config.Logging{
		Level:  "invalid_level",
		Format: "text",
		Output: "file",
	}
	logger := logrus.New()
	configure(logger, conf)
	assert.Equal(logrus.InfoLevel, logger.Level)
}
