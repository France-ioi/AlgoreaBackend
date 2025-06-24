package database

import (
	"database/sql/driver"
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/France-ioi/AlgoreaBackend/v2/app/logging"
	"github.com/France-ioi/AlgoreaBackend/v2/testhelpers/testoutput"
)

func Test_fileWithLineNum(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	mock.ExpectBegin()
	mock.ExpectCommit()

	require.NoError(t, NewDataStore(db).InTransaction(func(store *DataStore) error {
		assert.Contains(t, fileWithLineNum(), "/query_logger_test.go:")
		return nil
	}))

	assert.Nil(t, mock.ExpectationsWereMet())
}

func Test_fileWithLineNum_ReturnsEmptyStringWhenNoSuitableCallerFound(t *testing.T) {
	assert.Equal(t, "", fileWithLineNum())
}

func Test_getSQLExecutionPlanLoggingFunc_DoesNothingDependingOnParameters(t *testing.T) {
	tests := []struct {
		name              string
		query             string
		logSQLQueries     bool
		analyzeSQLQueries bool
	}{
		{
			name:  "logSQLQueries is false",
			query: "SELECT 1", logSQLQueries: false, analyzeSQLQueries: true,
		},
		{
			name:  "analyzeSQLQueries is false",
			query: "SELECT 1", logSQLQueries: true, analyzeSQLQueries: false,
		},
		{
			name:  "query is not explainable",
			query: "SET @a=1", logSQLQueries: true, analyzeSQLQueries: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			testoutput.SuppressIfPasses(t)

			db, mock := NewDBMock()
			defer func() { _ = db.Close() }()

			var called bool
			oldEmptyFunc := emptyFunc
			defer func() { emptyFunc = oldEmptyFunc }()
			emptyFunc = func() { called = true }

			logConfig := &LogConfig{
				LogSQLQueries:     tt.logSQLQueries,
				AnalyzeSQLQueries: tt.analyzeSQLQueries,
			}

			getSQLExecutionPlanLoggingFunc(db.ctx(), db.db.CommonDB().(*sqlDBWrapper), logConfig, tt.query)()
			assert.True(t, called)

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func Test_getSQLExecutionPlanLoggingFunc_RunsExplainAnalyzeForSuitableQueries(t *testing.T) {
	for _, query := range []string{
		"SELECT 1",
		"DELETE mytable",
		"INSERT INTO mytable (a, b) VALUES (1, 2)",
		"REPLACE INTO mytable (a, b) VALUES (1, 2)",
		"UPDATE mytable SET a = 1",
		"TABLE mytable",
		"\n\t sElEcT\n\t 1",
		"\n\t DelEte\n\t mytable",
		"\n\t iNSerT\n\t InTO mytable (a, b) VALUES (1, 2)",
		"\n\t RePlacE\n\t InTO mytable (a, b) VALUES (1, 2)",
		"\n\t uPdAte\n\t mytable sEt a = 1",
		"\n\t tAblE\n\t mytable",
	} {
		t.Run(query, func(t *testing.T) {
			testoutput.SuppressIfPasses(t)

			db, mock := NewDBMock()
			defer func() { _ = db.Close() }()
			loggerHook, loggerRestoreFunc := logging.MockSharedLoggerHook()
			defer loggerRestoreFunc()

			logConfig := &LogConfig{
				LogSQLQueries:     true,
				AnalyzeSQLQueries: true,
			}

			mock.ExpectQuery("^" + regexp.QuoteMeta("EXPLAIN ANALYZE "+query) + "$").
				WillReturnRows(mock.NewRows([]string{"plan"}).AddRow("plan"))
			getSQLExecutionPlanLoggingFunc(db.ctx(), db.db.CommonDB().(*sqlDBWrapper), logConfig, query)()
			require.NoError(t, mock.ExpectationsWereMet())

			printsCalled := loggerHook.AllEntries()
			require.Len(t, printsCalled, 1)
			printCalled := printsCalled[0]
			require.Len(t, printCalled.Data, 3)

			assertDurationIsOK(t, printCalled)
			assert.Equal(t, "db", printCalled.Data["type"])
			assert.IsType(t, "", printCalled.Data["fileline"])
			assert.Equal(t, "query execution plan:\nplan\n", printCalled.Message)
			assert.NotContains(t, printCalled.Data, "rows")
		})
	}
}

type timeType time.Time

// Value returns a timeType Value (*time.Time).
func (t *timeType) Value() (driver.Value, error) {
	if t == nil {
		return nil, nil
	}
	return (*time.Time)(t).UTC().Format("2006-01-02 15:04:05.999999"), nil
}

func Test_fillSQLPlaceholders(t *testing.T) {
	timeValue := time.Date(2022, 8, 2, 15, 4, 5, 123456789, time.UTC)
	timeValuePtr := timeValue
	timeTypeValue := timeType(timeValue)
	timeTypeValuePtr := &timeTypeValue
	int64Value := int64(12345)
	int64Ptr := &int64Value
	nilValue := (*int64)(nil)
	nilValuePtr := &nilValue

	type testArgs struct {
		query  string
		values []interface{}
	}

	tests := []struct {
		name     string
		args     testArgs
		expected string
	}{
		{
			name: "scalar numbers and booleans",
			args: testArgs{
				query: "INSERT INTO `t` (`c1`, `c2`, `c3`, `c4`, `c5`, `c6`, `c7`, `c8`, `c9`, `c10`, `c11`, `c12`, `c13`) VALUES " +
					"(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
				values: []interface{}{
					1, int8(2), int16(3), int32(4), int64(5), uint(6), uint8(7), uint16(8), uint32(9), uint64(10),
					float32(0.11), float64(0.12), true,
				},
			},
			expected: "INSERT INTO `t` (`c1`, `c2`, `c3`, `c4`, `c5`, `c6`, `c7`, `c8`, `c9`, `c10`, `c11`, `c12`, `c13`) VALUES " +
				"(1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 0.11, 0.12, true)",
		},
		{
			name: "time values",
			args: testArgs{
				query:  "INSERT INTO `t` (`c1`, `c2`, `c3`, `c4`, `c5`, `c6`) VALUES (?, ?, ?, ?, ?, ?)",
				values: []interface{}{timeValue, timeValuePtr, timeTypeValue, timeTypeValuePtr, (*timeType)(nil), time.Time{}},
			},
			expected: "INSERT INTO `t` (`c1`, `c2`, `c3`, `c4`, `c5`, `c6`) VALUES " +
				"('2022-08-02 15:04:05.123456789', '2022-08-02 15:04:05.123456789', " +
				"'2022-08-02 15:04:05.123456', '2022-08-02 15:04:05.123456', NULL, '0000-00-00 00:00:00')",
		},
		{
			name: "$n placeholders",
			args: testArgs{
				query:  "INSERT INTO `t` (`c1`, `c2`, `c3`, `c4`) VALUES ($1, $2, $3, $3)",
				values: []interface{}{int64(123451234512345), "testsessiontestsessiontestsessio", int64(1), int64(2)},
			},
			expected: "INSERT INTO `t` (`c1`, `c2`, `c3`, `c4`) VALUES (123451234512345, 'testsessiontestsessiontestsessio', 1, 1)",
		},
		{
			name: "bytes values",
			args: testArgs{
				query:  "INSERT INTO `t` (`c1`, `c2`) VALUES ($1, $2)",
				values: []interface{}{[]byte("test"), []byte{0x01, 0x02, 0x03}},
			},
			expected: "INSERT INTO `t` (`c1`, `c2`) VALUES ('test', '<binary>')",
		},
		{
			name: "pointer values",
			args: testArgs{
				query:  "INSERT INTO `t` (`c1`, `c2`) VALUES (?, ?)",
				values: []interface{}{int64Ptr, nilValuePtr},
			},
			expected: "INSERT INTO `t` (`c1`, `c2`) VALUES (12345, NULL)",
		},
		{
			name: "nil value",
			args: testArgs{
				query:  "INSERT INTO `t` (`c1`) VALUES (?)",
				values: []interface{}{nil},
			},
			expected: "INSERT INTO `t` (`c1`) VALUES (NULL)",
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, fillSQLPlaceholders(test.args.query, test.args.values))
		})
	}
}
