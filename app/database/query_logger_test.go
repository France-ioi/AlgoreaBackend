package database

import (
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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
				Logger:            nil,
				LogSQLQueries:     tt.logSQLQueries,
				AnalyzeSQLQueries: tt.analyzeSQLQueries,
			}

			getSQLExecutionPlanLoggingFunc(db.db.CommonDB().(*sqlDBWrapper), logConfig, tt.query)()
			assert.True(t, called)

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

type getSQLExecutionPlanLoggingFuncTestLogger struct {
	printsCalled [][]interface{}
}

func (l *getSQLExecutionPlanLoggingFuncTestLogger) Print(args ...interface{}) {
	l.printsCalled = append(l.printsCalled, args)
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

			logConfig := &LogConfig{
				Logger:            &getSQLExecutionPlanLoggingFuncTestLogger{},
				LogSQLQueries:     true,
				AnalyzeSQLQueries: true,
			}

			mock.ExpectQuery("^" + regexp.QuoteMeta("EXPLAIN ANALYZE "+query) + "$").
				WillReturnRows(mock.NewRows([]string{"plan"}).AddRow("plan"))
			getSQLExecutionPlanLoggingFunc(db.db.CommonDB().(*sqlDBWrapper), logConfig, query)()
			require.NoError(t, mock.ExpectationsWereMet())

			printsCalled := logConfig.Logger.(*getSQLExecutionPlanLoggingFuncTestLogger).printsCalled
			require.Len(t, printsCalled, 1)
			printCalled := printsCalled[0]
			require.Len(t, printCalled, 5)

			assert.Equal(t, "sql", printCalled[0])
			assert.IsType(t, "", printCalled[1])
			assert.IsType(t, time.Duration(0), printCalled[2])
			assert.Equal(t, "query execution plan:\nplan", printCalled[3])
			assert.Nil(t, printCalled[4])
		})
	}
}
