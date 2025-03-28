package database

import (
	"context"
	"database/sql/driver"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/France-ioi/AlgoreaBackend/v2/app/logging"
	"github.com/France-ioi/AlgoreaBackend/v2/app/loggingtest"
)

func TestNewRawDBLogger(t *testing.T) {
	hook, restoreFunc := logging.MockSharedLoggerHook()
	defer restoreFunc()

	rawLogger := NewRawDBLogger()
	rawLogger.Log(context.TODO(), "some message", "err", nil)
	entries := hook.AllEntries()
	require.Len(t, entries, 1)
	assert.Equal(t, "info", entries[0].Level.String())
	assert.Equal(t, "some message", entries[0].Message)
	require.Contains(t, entries[0].Data, "err")
	assert.Nil(t, entries[0].Data["err"])
}

func TestRawDBLogger_ShouldSkipSkippedActions(t *testing.T) {
	hook, restoreFunc := logging.MockSharedLoggerHook()
	defer restoreFunc()
	rawLogger := NewRawDBLogger()
	rawLogger.Log(context.TODO(), "sql-stmt-exec", "err", driver.ErrSkip)
	assert.Empty(t, (&loggingtest.Hook{Hook: hook}).GetAllStructuredLogs())
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
				"query", "SELECT * FROM users WHERE users.id=? and users.name=?",
				"args", "{[int64 1], [string \"Joe\"]}",
			},
			want: map[string]interface{}{
				"query": `SELECT * FROM users WHERE users.id=1 and users.name='Joe'`,
			},
		},
		{
			name: "with time",
			keyvals: []interface{}{
				"query", "UPDATE `users_items` SET `children_validated` = ?, `tasks_solved` = ?, `tasks_tried` = ?, " +
					"`tasks_with_help` = ?, `latest_activity_at` = ? WHERE (id=13)",
				"args", "{[int64 8], [int64 7], [int64 5], [int64 6], [time.Time 2019-03-18 16:24:01 +0000 UTC]}",
			},
			want: map[string]interface{}{
				"query": "UPDATE `users_items` SET `children_validated` = 8, `tasks_solved` = 7, `tasks_tried` = 5, " +
					"`tasks_with_help` = 6, `latest_activity_at` = '2019-03-18 16:24:01 +0000 UTC' WHERE (id=13)",
			},
		},
		{
			name: "with time and nils",
			keyvals: []interface{}{
				"query", "UPDATE some_table SET column1 = ?, column2 = ?, column3 = ?, column4 = ?",
				"args", "{[<nil> <nil>], [time.Time 2019-03-18 16:24:01 +0000 UTC], [<nil> <nil>], [time.Time 2018-03-18 16:24:01 +0000 UTC]}",
			},
			want: map[string]interface{}{
				"query": "UPDATE some_table SET column1 = NULL, column2 = '2019-03-18 16:24:01 +0000 UTC', column3 = NULL, " +
					"column4 = '2018-03-18 16:24:01 +0000 UTC'",
			},
		},
	}
	for _, tt := range tests {
		tt := tt
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
		tt := tt
		t.Run(tt.typeStr, func(t *testing.T) {
			got := convertRawSQLArgValue(tt.value, tt.typeStr)
			assert.Equal(t, tt.want, got)
		})
	}
}
