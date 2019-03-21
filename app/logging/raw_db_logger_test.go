package logging

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/app/config"
	"github.com/France-ioi/AlgoreaBackend/app/loggingtest"
)

func TestNewRawDBLogger_TextLog(t *testing.T) {
	nulllogger, hook := loggingtest.NewNullLogger()
	logger := &Logger{nulllogger, &config.Logging{
		Format:        "text",
		LogSQLQueries: true,
	}}
	dbLogger, logMode := logger.NewDBLogger()

	rawLogger := NewRawDBLogger(dbLogger, logMode)
	rawLogger.Log(nil, "some message", "err", nil) //lint:ignore SA1012 sql often uses nil context
	assert.Contains(t, hook.GetAllStructuredLogs(), "some message map[err:<nil>]")
}

func TestNewRawDBLogger_HonoursLogMode(t *testing.T) {
	nulllogger, hook := loggingtest.NewNullLogger()
	logger := &Logger{nulllogger, &config.Logging{
		Format:        "text",
		LogSQLQueries: false,
	}}
	dbLogger, logMode := logger.NewDBLogger()
	rawLogger := NewRawDBLogger(dbLogger, logMode)
	rawLogger.Log(nil, "some message", "err", nil) //lint:ignore SA1012 sql often uses nil context
	assert.Empty(t, hook.GetAllStructuredLogs())
}

func TestNewRawDBLogger_JSONLog(t *testing.T) {
	nulllogger, hook := loggingtest.NewNullLogger()
	logger := &Logger{nulllogger, &config.Logging{
		Format:        "json",
		LogSQLQueries: true,
	}}
	dbLogger, logMode := logger.NewDBLogger()
	rawLogger := NewRawDBLogger(dbLogger, logMode)
	rawLogger.Log(nil, "some message", "err", nil) //lint:ignore SA1012 sql often uses nil context
	assert.Contains(t, hook.GetAllStructuredLogs(), `msg="some message"`)
	assert.Contains(t, hook.GetAllStructuredLogs(), `err="<nil>"`)
}

func TestRawDBLogger_ShouldSkipStmtExecWithNilContext(t *testing.T) {
	nulllogger, hook := loggingtest.NewNullLogger()
	logger := &Logger{nulllogger, &config.Logging{
		Format:        "json",
		LogSQLQueries: true,
	}}
	dbLogger, logMode := logger.NewDBLogger()
	rawLogger := NewRawDBLogger(dbLogger, logMode)
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
		{
			name: "with time",
			keyvals: []interface{}{
				"query", "UPDATE `users_items` SET `nbChildrenValidated` = ?, `nbTasksSolved` = ?, `nbTasksTried` = ?, `nbTasksWithHelp` = ?, `sLastActivityDate` = ? WHERE (ID=13)",
				"args", "{[int64 8], [int64 7], [int64 5], [int64 6], [time.Time 2019-03-18 16:24:01 +0000 UTC]}",
			},
			want: map[string]interface{}{
				"query": "UPDATE `users_items` SET `nbChildrenValidated` = 8, `nbTasksSolved` = 7, `nbTasksTried` = 5, `nbTasksWithHelp` = 6, `sLastActivityDate` = \"2019-03-18 16:24:01 +0000 UTC\" WHERE (ID=13)",
			},
		},
		{
			name: "with time and nils",
			keyvals: []interface{}{
				"query", "UPDATE some_table SET column1 = ?, column2 = ?, column3 = ?, column4 = ?",
				"args", "{[<nil> <nil>], [time.Time 2019-03-18 16:24:01 +0000 UTC], [<nil> <nil>], [time.Time 2018-03-18 16:24:01 +0000 UTC]}",
			},
			want: map[string]interface{}{
				"query": "UPDATE some_table SET column1 = NULL, column2 = \"2019-03-18 16:24:01 +0000 UTC\", column3 = NULL, column4 = \"2018-03-18 16:24:01 +0000 UTC\"",
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
