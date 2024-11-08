package logging_test

import (
	"errors"
	"testing"
	"time"

	"github.com/sirupsen/logrus/hooks/test" //nolint:depguard
	"github.com/spf13/viper"
	assertlib "github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/app/logging"
	"github.com/France-ioi/AlgoreaBackend/v2/testhelpers/testoutput"
)

func TestStructuredDBLogger_Print_SQL(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	assert := assertlib.New(t)
	var hook *test.Hook
	logging.SharedLogger, hook = logging.NewMockLogger()
	defer logging.ResetShared()
	conf := viper.New()
	conf.Set("Format", "json")
	conf.Set("Output", "stdout")
	conf.Set("LogSQLQueries", true)
	logging.SharedLogger.Configure(conf)
	db, mock := database.NewDBMock()
	defer func() { _ = db.Close() }()

	timeParam := time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)

	mock.ExpectQuery(`^SELECT \$1, \$2, \$3, \$4, \$5$`).
		WithArgs(1, timeParam, "foo", []byte("bar"), nil).
		WillReturnRows(mock.NewRows([]string{"1"}).AddRow(1))

	var result []interface{}
	db.Raw("SELECT $1, $2, $3, $4, $5", 1, timeParam, "foo", []byte("bar"), nil).Scan(&result)
	assert.Equal(`SELECT 1, '2009-11-10 23:00:00', "foo", "bar", NULL`, hook.LastEntry().Message)
	data := hook.LastEntry().Data
	assert.Equal("db", data["type"])
	assert.True(data["duration"].(float64) < 0.1, "unexpected duration: %v", data["duration"])
	assert.NotNil(data["ts"])
	assert.Equal(int64(1), data["rows"].(int64))
	assert.NoError(mock.ExpectationsWereMet())
}

func TestStructuredDBLogger_Print_SQLWithInterrogationMark(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	assert := assertlib.New(t)
	var hook *test.Hook
	logging.SharedLogger, hook = logging.NewMockLogger()
	defer logging.ResetShared()
	conf := viper.New()
	conf.Set("Format", "json")
	conf.Set("Output", "stdout")
	conf.Set("LogSQLQueries", true)
	logging.SharedLogger.Configure(conf)
	db, mock := database.NewDBMock()
	defer func() { _ = db.Close() }()

	mock.ExpectQuery(`^SELECT \?$`).WithArgs(1).WillReturnRows(mock.NewRows([]string{"1"}).AddRow(1))

	var result []interface{}
	db.Raw("SELECT ?", 1).Scan(&result)
	assert.Equal("SELECT 1", hook.LastEntry().Message)
	assert.NoError(mock.ExpectationsWereMet())
}

func TestStructuredDBLogger_Print_SQLError(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	assert := assertlib.New(t)
	var hook *test.Hook
	logging.SharedLogger, hook = logging.NewMockLogger()
	defer logging.ResetShared()
	conf := viper.New()
	conf.Set("Format", "json")
	conf.Set("Output", "stdout")
	conf.Set("LogSQLQueries", true)
	conf.Set("Level", "debug")
	logging.SharedLogger.Configure(conf)
	db, mock := database.NewDBMock()
	defer func() { _ = db.Close() }()

	mock.ExpectQuery("SELECT 2").WillReturnError(errors.New("a query error"))

	var result []interface{}
	db.Raw("SELECT 2").Scan(&result)

	assert.Equal("a query error", hook.Entries[0].Message)
	data := hook.Entries[0].Data
	assert.Equal("db", data["type"])

	assert.Equal("SELECT 2", hook.Entries[1].Message)
	data = hook.Entries[1].Data
	assert.Equal("db", data["type"])
	assert.True(data["duration"].(float64) < 1.0, "unexpected duration: %v", data["duration"])
	assert.NotNil(data["ts"])
	assert.Equal(int64(0), data["rows"].(int64))
	assert.NoError(mock.ExpectationsWereMet())
}

func TestStructuredDBLogger_Print_RawSQLWithDuration(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	assert := assertlib.New(t)
	var hook *test.Hook
	logging.SharedLogger.Logger, hook = test.NewNullLogger()
	defer logging.ResetShared()
	structuredLogger := logging.NewStructuredDBLogger()
	structuredLogger.Print("rawsql", nil, "sql-stmt-exec",
		map[string]interface{}{
			"query":    "SELECT 1",
			"duration": 500 * time.Millisecond,
		})

	assert.Equal("sql-stmt-exec", hook.Entries[0].Message)
	data := hook.Entries[0].Data
	assert.Equal("db", data["type"])
	assert.Equal(data["duration"].(float64), 0.5)
	assert.NotNil(data["ts"])
}
