package logging

import (
	"strconv"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

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
					config.Set("Output", "stdout")
					logger.Configure(config)
					assert.Equal(t, value, test.funcToCall(logger))
				})
			}
		})
	}
}
