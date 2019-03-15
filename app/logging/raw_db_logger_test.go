package logging

import (
	"errors"
	"testing"

	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/app/config"
	"github.com/France-ioi/AlgoreaBackend/app/loggingtest"
)

func TestNewRawDBLogger_ErrorFallback(t *testing.T) {
	patch := monkey.Patch(config.Load, func() (*config.Root, error) {
		return nil, errors.New("config loading error")
	})
	defer patch.Unpatch()

	var hook *loggingtest.Hook
	Logger, hook = loggingtest.NewNullLogger()

	rawLogger, logMode := NewRawDBLogger()
	assert.False(t, logMode)
	rawLogger.Log(nil, "some message", "err", nil) //lint:ignore SA1012 sql often uses nil context
	assert.Empty(t, hook.GetAllLogs())
}

func TestNewRawDBLogger_TextLog(t *testing.T) {
	conf := &config.Root{Logging: config.Logging{
		Format:        "text",
		LogSQLQueries: true,
	}}

	patch := monkey.Patch(config.Load, func() (*config.Root, error) {
		return conf, nil
	})
	defer patch.Unpatch()

	var hook *loggingtest.Hook
	Logger, hook = loggingtest.NewNullLogger()

	rawLogger, logMode := NewRawDBLogger()
	assert.True(t, logMode)
	rawLogger.Log(nil, "some message", "err", nil) //lint:ignore SA1012 sql often uses nil context
	assert.Contains(t, hook.GetAllStructuredLogs(), "some message map[err:<nil>]")
}

func TestNewRawDBLogger_HonoursLogMode(t *testing.T) {
	conf := &config.Root{Logging: config.Logging{
		Format:        "text",
		LogSQLQueries: false,
	}}

	patch := monkey.Patch(config.Load, func() (*config.Root, error) {
		return conf, nil
	})
	defer patch.Unpatch()

	var hook *loggingtest.Hook
	Logger, hook = loggingtest.NewNullLogger()

	rawLogger, logMode := NewRawDBLogger()
	assert.False(t, logMode)
	rawLogger.Log(nil, "some message", "err", nil) //lint:ignore SA1012 sql often uses nil context
	assert.Empty(t, hook.GetAllStructuredLogs())
}

func TestNewRawDBLogger_JSONLog(t *testing.T) {
	conf := &config.Root{Logging: config.Logging{
		Format:        "json",
		LogSQLQueries: true,
	}}

	patch := monkey.Patch(config.Load, func() (*config.Root, error) {
		return conf, nil
	})
	defer patch.Unpatch()

	var hook *loggingtest.Hook
	Logger, hook = loggingtest.NewNullLogger()

	rawLogger, logMode := NewRawDBLogger()
	assert.True(t, logMode)
	rawLogger.Log(nil, "some message", "err", nil) //lint:ignore SA1012 sql often uses nil context
	assert.Contains(t, hook.GetAllStructuredLogs(), `msg="some message"`)
	assert.Contains(t, hook.GetAllStructuredLogs(), `err="<nil>"`)
}

func TestRawDBLogger_ShouldSkipStmtExecWithNilContext(t *testing.T) {
	var hook *loggingtest.Hook
	Logger, hook = loggingtest.NewNullLogger()

	rawLogger, logMode := NewRawDBLogger()
	assert.True(t, logMode)
	rawLogger.Log(nil, "sql-stmt-exec", "err", nil) //lint:ignore SA1012 we check the nil context here
	assert.Empty(t, hook.GetAllStructuredLogs())
}

func Test_prepareRawDBLoggerValuesMap(t *testing.T) {
	tests := []struct {
		name    string
		keyvals []interface{}
		want    map[string]interface{}
	}{
		{
			name: "simple",
			keyvals: []interface{}{
				"query", "SELECT * FROM users WHERE users.ID=? and users.sName=?",
				"args", "{[int64 1], [string \"Joe\"]}",
			},
			want: map[string]interface{}{
				"query": `SELECT * FROM users WHERE users.ID=1 and users.sName="Joe"`,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := prepareRawDBLoggerValuesMap(tt.keyvals)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_convertRawSQLArgValue(t *testing.T) {
	tests := []struct {
		value   string
		typeStr string
		want    interface{}
	}{
		{value: `"some string \"' "`, typeStr: "string", want: `some string "' `},
		{value: `1234`, typeStr: "int", want: int64(1234)},
		{value: `1234`, typeStr: "int8", want: int64(1234)},
		{value: `1234`, typeStr: "int16", want: int64(1234)},
		{value: `1234`, typeStr: "int32", want: int64(1234)},
		{value: `1234`, typeStr: "int64", want: int64(1234)},
		{value: `1234`, typeStr: "int64", want: int64(1234)},
		{value: `12.34`, typeStr: "float64", want: float64(12.34)},
		{value: `12.34`, typeStr: "float32", want: float64(12.34)},
		{value: `some_value`, typeStr: "<nil>", want: nil},
		{value: `some_value`, typeStr: "unknown type", want: `some_value`},
	}
	for _, tt := range tests {
		t.Run(tt.typeStr, func(t *testing.T) {
			got := convertRawSQLArgValue(tt.value, tt.typeStr)
			assert.Equal(t, tt.want, got)
		})
	}
}
