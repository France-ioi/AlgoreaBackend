package logging

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"bou.ke/monkey"
	"github.com/sirupsen/logrus" //nolint:depguard
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEntryFromContext(t *testing.T) {
	ctx, _, hook := NewContextWithNewMockLogger()
	entry := EntryFromContext(ctx)
	assert.NotNil(t, entry)
	assert.Equal(t, ctx, entry.Context)
	entry.Info("Hello World")
	require.NotNil(t, hook.LastEntry())
	assert.Equal(t, "Hello World", hook.LastEntry().Message)
}

func TestLoggerFromContext(t *testing.T) {
	ctx, logger, _ := NewContextWithNewMockLogger()
	retrievedLogger := LoggerFromContext(ctx)
	assert.Equal(t, logger, retrievedLogger)
}

type testContextKeyType int

const (
	testContextKey   testContextKeyType = iota
	testContextValue                    = "test"
)

func TestContextWithLogger(t *testing.T) {
	logger := createLogger()
	ctx := ContextWithLogger(context.WithValue(context.Background(), testContextKey, testContextValue), logger)
	retrievedLogger := LoggerFromContext(ctx)
	assert.Equal(t, logger, retrievedLogger)
	assert.Equal(t, testContextValue, ctx.Value(testContextKey))
}

func TestContextWithLoggerMiddleware(t *testing.T) {
	logger := createLogger()

	ctx := context.WithValue(context.Background(), testContextKey, testContextValue)
	middleware := ContextWithLoggerMiddleware(logger)
	var called bool
	handler := middleware(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		called = true
		assert.Equal(t, testContextValue, r.Context().Value(testContextKey))
		assert.Equal(t, logger, LoggerFromContext(r.Context()))
	}))

	req, err := http.NewRequest(http.MethodGet, "/", http.NoBody)
	require.NoError(t, err)
	req = req.WithContext(ctx)
	handler.ServeHTTP(httptest.NewRecorder(), req)
	assert.True(t, called)
}

func TestLogger_Configure_FormatText(t *testing.T) {
	conf := viper.New()
	conf.Set("Format", "text")
	conf.Set("Output", "stdout")
	logger := NewLoggerFromConfig(conf)
	assert.IsType(t, &textFormatter{}, logger.logrusLogger.Formatter)
}

func TestLogger_Configure_FormatJson(t *testing.T) {
	conf := viper.New()
	conf.Set("Format", "json")
	conf.Set("Output", "stdout")
	logger := NewLoggerFromConfig(conf)
	assert.IsType(t, &jsonFormatter{}, logger.logrusLogger.Formatter)
}

func TestLogger_Configure_FormatConsole(t *testing.T) {
	conf := viper.New()
	conf.Set("Format", "console")
	conf.Set("Output", "stdout")
	logger := NewLoggerFromConfig(conf)
	assert.IsType(t, &consoleFormatter{}, logger.logrusLogger.Formatter)
}

func TestLogger_Configure_FormatInvalid(t *testing.T) {
	conf := viper.New()
	conf.Set("Format", "yml")
	conf.Set("Output", "stdout")
	logger := createLogger()
	assert.Panics(t, func() { logger.Configure(conf) })
}

func TestLogger_Configure_OutputStdout(t *testing.T) {
	conf := viper.New()
	conf.Set("Format", "json")
	conf.Set("Output", "stdout")
	logger := NewLoggerFromConfig(conf)
	assert.Equal(t, os.Stdout, logger.logrusLogger.Out)
}

func TestLogger_Configure_OutputStderr(t *testing.T) {
	conf := viper.New()
	conf.Set("Format", "json")
	conf.Set("Output", "stderr")
	logger := NewLoggerFromConfig(conf)
	assert.Equal(t, os.Stderr, logger.logrusLogger.Out)
}

func TestLogger_Configure_OutputFile(t *testing.T) {
	conf := viper.New()
	conf.Set("Format", "json")
	conf.Set("Output", "file")

	_ = os.Remove("../../log/all.log.test")
	defer func() { _ = os.Remove("../../log/all.log.test") }()
	var patchGuard *monkey.PatchGuard
	patchGuard = monkey.Patch(os.OpenFile, func(name string, flag int, perm os.FileMode) (*os.File, error) {
		patchGuard.Unpatch()
		defer patchGuard.Restore()
		return os.OpenFile(name+".test", flag, perm) //nolint:gosec // No user input
	})
	defer patchGuard.Unpatch()

	// will append time to make sure not to match a prev exec of the test
	timestamp := time.Now().UnixNano()

	logger1 := NewLoggerFromConfig(conf)
	logger1.WithContext(context.Background()).Errorf("logexec1 %d", timestamp)

	// redo another init to check it will not override
	logger2 := NewLoggerFromConfig(conf)
	logger2.WithContext(context.Background()).Warnf("logexec2 %d", timestamp)

	// check the resulting file
	content, _ := os.ReadFile("../../log/all.log")
	assert.Contains(t, string(content), fmt.Sprintf("logexec1 %d", timestamp))
	assert.Contains(t, string(content), fmt.Sprintf("logexec2 %d", timestamp))
}

func TestLogger_Configure_OutputFileError(t *testing.T) {
	conf := viper.New()
	conf.Set("Format", "json")
	conf.Set("Output", "file")
	fakeFunc := func(_ string, _ int, _ os.FileMode) (*os.File, error) {
		return nil, errors.New("open error")
	}
	patch := monkey.Patch(os.OpenFile, fakeFunc)
	defer patch.Unpatch()
	logger := NewLoggerFromConfig(conf)
	assert.Equal(t, os.Stdout, logger.logrusLogger.Out)
}

func TestLogger_Configure_OutputInvalid(t *testing.T) {
	conf := viper.New()
	conf.Set("Format", "json")
	conf.Set("Output", "S3")
	logger := createLogger()
	assert.Panics(t, func() { logger.Configure(conf) })
}

func TestLogger_Configure_FormatConsoleAndFileOutput(t *testing.T) {
	conf := viper.New()
	conf.Set("Format", "console")
	conf.Set("Output", "file")
	logger := createLogger()
	assert.Panics(t, func() { logger.Configure(conf) })
}

func TestLogger_Configure_LevelDefault(t *testing.T) {
	conf := viper.New()
	conf.Set("Format", "text")
	conf.Set("Output", "stdout")
	conf.Set("Level", "")
	logger := NewLoggerFromConfig(conf)
	assert.Equal(t, logrus.InfoLevel, logger.logrusLogger.Level)
}

func TestLogger_Configure_LevelParsed(t *testing.T) {
	conf := viper.New()
	conf.Set("Format", "text")
	conf.Set("Output", "stdout")
	conf.Set("Level", "warn")
	logger := NewLoggerFromConfig(conf)
	assert.Equal(t, logrus.WarnLevel, logger.logrusLogger.Level)
}

func TestLogger_Configure_LevelInvalid(t *testing.T) {
	conf := viper.New()
	conf.Set("Format", "text")
	conf.Set("Output", "stdout")
	conf.Set("Level", "invalid_level")
	logger := NewLoggerFromConfig(conf)
	assert.Equal(t, logrus.InfoLevel, logger.logrusLogger.Level)
}
