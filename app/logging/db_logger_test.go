package logging

import (
	"strconv"
	"testing"

	"github.com/jinzhu/gorm"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestLogger_NewDBLogger_ErrorFallback(t *testing.T) {
	logger := createLogger() // no config
	dbLogger := logger.NewDBLogger()
	assert.IsType(t, gorm.Logger{}, dbLogger)
}

func TestLogger_NewDBLogger_TextLog(t *testing.T) {
	logger := createLogger()
	config := viper.New()
	config.Set("Format", "text")
	config.Set("Output", "file")
	logger.Configure(config)
	dbLogger := logger.NewDBLogger()
	assert.IsType(t, gorm.Logger{}, dbLogger)
}

func TestLogger_NewDBLogger_JSONLog(t *testing.T) {
	logger := createLogger()
	config := viper.New()
	config.Set("Format", "json")
	config.Set("Output", "file")
	logger.Configure(config)
	dbLogger := logger.NewDBLogger()
	assert.IsType(t, &StructuredDBLogger{}, dbLogger)
}

func TestLogger_NewDBLogger_WrongFormat(t *testing.T) {
	logger := createLogger()
	config := viper.New()
	config.Set("Format", "yml")
	config.Set("Output", "file")
	logger.config = config
	assert.Panics(t, func() { logger.NewDBLogger() })
}

func TestLogger_BooleanConfigFlags(t *testing.T) {
	for _, test := range []struct {
		flagName   string
		funcToCall func(*Logger) bool
	}{
		{"LogSQLQueries", (*Logger).IsSQLQueriesLoggingEnabled},
		{"LogRawSQLQueries", (*Logger).IsRawSQLQueriesLoggingEnabled},
		{"AnalyzeSQLQueries", (*Logger).IsSQLQueriesAnalyzingEnabled},
	} {
		test := test
		t.Run(test.flagName, func(t *testing.T) {
			t.Run("nil config", func(t *testing.T) {
				logger := createLogger()
				assert.False(t, test.funcToCall(logger))
			})

			for _, value := range []bool{true, false} {
				t.Run("value="+strconv.FormatBool(value), func(t *testing.T) {
					logger := createLogger()
					config := viper.New()
					config.Set(test.flagName, value)
					config.Set("Output", "file")
					logger.Configure(config)
					assert.Equal(t, value, test.funcToCall(logger))
				})
			}
		})
	}
}
