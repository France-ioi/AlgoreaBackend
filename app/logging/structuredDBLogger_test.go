package logging_test

import (
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/sirupsen/logrus/hooks/test" //nolint:depguard
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/app/logging"
	"github.com/France-ioi/AlgoreaBackend/v2/testhelpers/testoutput"
)

func TestStructuredDBLogger_Print_SQL_Select(t *testing.T) {
	testoutput.SuppressIfPasses(t)

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
	assert.Equal(t, `SELECT 1, '2009-11-10 23:00:00', "foo", "bar", NULL`, hook.LastEntry().Message)
	data := hook.LastEntry().Data
	assert.Equal(t, "db", data["type"])
	assert.True(t, data["duration"].(float64) < 0.1, "unexpected duration: %v", data["duration"])
	assert.NotNil(t, data["ts"])
	assert.NotContains(t, data, "rows")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestStructuredDBLogger_Print_SQL_Update(t *testing.T) {
	testoutput.SuppressIfPasses(t)

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

	mock.ExpectExec(`^UPDATE t1 SET c1=\$1, c2=\$2, c3=\$3, c4=\$4, c5=\$5$`).
		WithArgs(1, timeParam, "foo", []byte("bar"), nil).
		WillReturnResult(sqlmock.NewResult(-1, 123))

	db.Exec("UPDATE t1 SET c1=$1, c2=$2, c3=$3, c4=$4, c5=$5", 1, timeParam, "foo", []byte("bar"), nil)
	assert.Equal(t, `UPDATE t1 SET c1=1, c2='2009-11-10 23:00:00', c3="foo", c4="bar", c5=NULL`, hook.LastEntry().Message)
	data := hook.LastEntry().Data
	assert.Equal(t, "db", data["type"])
	assert.True(t, data["duration"].(float64) < 0.1, "unexpected duration: %v", data["duration"])
	assert.NotNil(t, data["ts"])
	assert.Equal(t, int64(123), data["rows"].(int64))
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestStructuredDBLogger_Print_SQLWithInterrogationMark(t *testing.T) {
	testoutput.SuppressIfPasses(t)

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
	assert.Equal(t, "SELECT 1", hook.LastEntry().Message)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestStructuredDBLogger_Print_SQLError(t *testing.T) {
	testoutput.SuppressIfPasses(t)

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

	assert.Equal(t, "SELECT 2", hook.Entries[0].Message)
	data := hook.Entries[0].Data
	assert.Equal(t, "db", data["type"])
	assert.True(t, data["duration"].(float64) < 1.0, "unexpected duration: %v", data["duration"])
	assert.NotNil(t, data["ts"])
	assert.Nil(t, data["rows"])
	assert.NoError(t, mock.ExpectationsWereMet())

	assert.Equal(t, "a query error", hook.Entries[1].Message)
	assert.Equal(t, "error", hook.Entries[1].Level.String())
	assert.NotNil(t, hook.Entries[1].Time)
	assert.Equal(t, "db", hook.Entries[1].Data["type"])
}

func TestStructuredDBLogger_Print_RawSQLWithDuration(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	var hook *test.Hook
	logging.SharedLogger.Logger, hook = test.NewNullLogger()
	defer logging.ResetShared()
	structuredLogger := logging.NewStructuredDBLogger()
	structuredLogger.Print("rawsql", nil, "sql-stmt-exec",
		map[string]interface{}{
			"query":    "SELECT 1",
			"duration": 500 * time.Millisecond,
		})

	assert.Equal(t, "sql-stmt-exec", hook.Entries[0].Message)
	data := hook.Entries[0].Data
	assert.Equal(t, "db", data["type"])
	assert.Equal(t, data["duration"].(float64), 0.5)
	assert.NotNil(t, data["ts"])
}
