package database

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"runtime"
	"strconv"
	"testing"
	"time"

	"bou.ke/monkey"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/France-ioi/AlgoreaBackend/v2/app/logging"
	"github.com/France-ioi/AlgoreaBackend/v2/app/loggingtest"
	"github.com/France-ioi/AlgoreaBackend/v2/golang"
	"github.com/France-ioi/AlgoreaBackend/v2/testhelpers/testoutput"
)

const someName = "some name"

func TestDB_inTransaction_NoErrors(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT 1 AS id").
		WillReturnRows(mock.NewRows([]string{"id"}).AddRow(int64(1)))
	mock.ExpectCommit()

	type resultStruct struct {
		ID int64
	}
	var result []resultStruct
	err := db.inTransaction(func(db *DB) error {
		return db.Raw("SELECT 1 AS id").Scan(&result).Error()
	})

	assert.NoError(t, err)
	assert.Equal(t, []resultStruct{{1}}, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDB_inTransaction_DBErrorOnBegin(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	expectedError := errors.New("some error")

	mock.ExpectBegin().WillReturnError(expectedError)

	gotError := db.inTransaction(func(db *DB) error {
		var result []interface{}
		return db.Raw("SELECT 1").Scan(&result).Error()
	})
	assert.Equal(t, expectedError, gotError)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDB_inTransaction_DBError(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	expectedError := errors.New("some error")

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT 1").WillReturnError(expectedError)
	mock.ExpectRollback()

	gotError := db.inTransaction(func(db *DB) error {
		var result []interface{}
		return db.Raw("SELECT 1").Scan(&result).Error()
	})
	assert.Equal(t, expectedError, gotError)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDB_inTransaction_Panic(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	expectedError := errors.New("some error")

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT 1").WillReturnError(expectedError)
	mock.ExpectRollback()

	assert.PanicsWithValue(t, expectedError.(interface{}), func() {
		_ = db.inTransaction(func(db *DB) error {
			var result []interface{}
			db.Raw("SELECT 1").Scan(&result)
			panic(expectedError)
		})
	})
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDB_inTransaction_PanicWithString(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	expectedError := "some error"

	mock.ExpectBegin()
	mock.ExpectRollback()

	assert.PanicsWithValue(t, expectedError, func() {
		_ = db.inTransaction(func(db *DB) error {
			panic(expectedError)
		})
	})
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDB_inTransaction_ErrorOnRollback(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	expectedError := errors.New("some error")

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT 1").WillReturnError(expectedError)
	mock.ExpectRollback().WillReturnError(errors.New("rollback error"))

	err := db.inTransaction(func(db *DB) error {
		var result []interface{}
		err := db.Raw("SELECT 1").Scan(&result).Error()
		assert.Equal(t, expectedError, err)
		return err
	})
	assert.Equal(t, expectedError, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDB_inTransaction_ErrorOnCommit(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	expectedError := errors.New("commit error")

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT 1").
		WillReturnRows(mock.NewRows([]string{"1"}).AddRow(1))
	mock.ExpectCommit().WillReturnError(expectedError)

	assert.Equal(t, expectedError, db.inTransaction(func(db *DB) error {
		var result []interface{}
		return db.Raw("SELECT 1").Scan(&result).Error()
	}))
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDB_inTransaction_RetriesOnDeadlockAndLockWaitTimeoutErrors(t *testing.T) {
	for _, errorNumber := range []uint16{1213, 1205} {
		t.Run(fmt.Sprintf("error%d", errorNumber), func(t *testing.T) {
			testoutput.SuppressIfPasses(t)

			db, mock := NewDBMock()
			defer func() { _ = db.Close() }()

			var duration time.Duration
			monkey.Patch(time.Sleep, func(d time.Duration) { duration += d })
			defer monkey.UnpatchAll()

			logHook, restoreLoggerFunc := logging.MockSharedLoggerHook()
			defer restoreLoggerFunc()

			mock.ExpectBegin()
			mock.ExpectQuery("SELECT 1").
				WillReturnError(&mysql.MySQLError{Number: errorNumber})
			mock.ExpectRollback()
			mock.ExpectBegin()
			mock.ExpectQuery("SELECT 1").
				WillReturnRows(mock.NewRows([]string{"1"}).AddRow(1))
			mock.ExpectCommit()

			assert.NoError(t, db.inTransaction(func(db *DB) error {
				var result []interface{}
				return db.Raw("SELECT 1").Scan(&result).Error()
			}))
			assert.InEpsilon(t, transactionDelayBetweenRetries, duration, 0.05)
			assert.NoError(t, mock.ExpectationsWereMet())

			logs := (&loggingtest.Hook{Hook: logHook}).GetAllStructuredLogs()
			assert.Contains(t, logs, fmt.Sprintf("Retrying transaction (count: 1) after Error %d:", errorNumber))
		})
	}
}

func TestDB_inTransaction_SkipsRetryingOnDeadlockAndLockWaitTimeoutErrorsOnRollbackError(t *testing.T) {
	for _, errorNumber := range []uint16{1213, 1205} {
		t.Run(fmt.Sprintf("error%d", errorNumber), func(t *testing.T) {
			testoutput.SuppressIfPasses(t)

			db, mock := NewDBMock()
			defer func() { _ = db.Close() }()

			var duration time.Duration
			monkey.Patch(time.Sleep, func(d time.Duration) { duration += d })
			defer monkey.UnpatchAll()

			expectedError := errors.New("rollback error")
			mock.ExpectBegin()
			mock.ExpectQuery("SELECT 1").
				WillReturnError(&mysql.MySQLError{Number: errorNumber})
			mock.ExpectRollback().WillReturnError(expectedError)

			err := db.inTransaction(func(db *DB) error {
				var result []interface{}
				return db.Raw("SELECT 1").Scan(&result).Error()
			})
			assert.Equal(t, expectedError, err)

			assert.Zero(t, duration)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestDB_inTransaction_RetriesOnDeadlockAndLockWaitTimeoutPanic(t *testing.T) {
	for _, errorNumber := range []uint16{1213, 1205} {
		t.Run(fmt.Sprintf("error%d", errorNumber), func(t *testing.T) {
			testoutput.SuppressIfPasses(t)

			db, mock := NewDBMock()
			defer func() { _ = db.Close() }()

			var duration time.Duration
			monkey.Patch(time.Sleep, func(d time.Duration) { duration += d })
			defer monkey.UnpatchAll()

			mock.ExpectBegin()
			mock.ExpectQuery("SELECT 1").
				WillReturnError(&mysql.MySQLError{Number: errorNumber})
			mock.ExpectRollback()
			mock.ExpectBegin()
			mock.ExpectQuery("SELECT 1").
				WillReturnRows(mock.NewRows([]string{"1"}).AddRow(1))
			mock.ExpectCommit()

			assert.NoError(t, db.inTransaction(func(db *DB) error {
				var result []interface{}
				mustNotBeError(db.Raw("SELECT 1").Scan(&result).Error())
				return nil
			}))
			assert.InEpsilon(t, transactionDelayBetweenRetries, duration, 0.05)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func patchGormBeginTxWithVerifier(
	t *testing.T, callsCount *int, expectedTxOptions *sql.TxOptions, expectedContextVars map[interface{}]interface{},
) (patch *monkey.PatchGuard) {
	patch = monkey.Patch(gormDBBeginTxReplacement,
		func(ctx context.Context, db *gorm.DB, opts *sql.TxOptions) *gorm.DB {
			*callsCount++
			assert.Equal(t, expectedTxOptions, opts)
			if len(expectedContextVars) > 0 {
				assert.NotNil(t, ctx)
				for key, value := range expectedContextVars {
					assert.Equal(t, value, ctx.Value(key), "Wrong context value for key %#v", key)
				}
			}

			defer patch.Restore()
			patch.Unpatch()
			return gormDBBeginTxReplacement(ctx, db, opts)
		})
	return patch
}

func TestDB_inTransaction_RetriesOnDeadLockPanic_WithTxOptions(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	var calledCount int
	txOptions := &sql.TxOptions{Isolation: sql.LevelReadCommitted}
	patch := patchGormBeginTxWithVerifier(t, &calledCount, txOptions, nil)
	defer patch.Unpatch()

	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	monkey.Patch(time.Sleep, func(d time.Duration) {})
	defer monkey.UnpatchAll()

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT 1").
		WillReturnError(&mysql.MySQLError{Number: 1213})
	mock.ExpectRollback()
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT 1").
		WillReturnRows(mock.NewRows([]string{"1"}).AddRow(1))
	mock.ExpectCommit()

	assert.NoError(t, db.inTransaction(func(db *DB) error {
		var result []interface{}
		mustNotBeError(db.Raw("SELECT 1").Scan(&result).Error())
		return nil
	}, txOptions))
	assert.NoError(t, mock.ExpectationsWereMet())
	assert.Equal(t, 2, calledCount)
}

func TestDB_inTransaction_RetriesOnDeadLockError_WithTxOptions(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	var calledCount int
	txOptions := &sql.TxOptions{Isolation: sql.LevelReadCommitted}
	patch := patchGormBeginTxWithVerifier(t, &calledCount, txOptions, nil)
	defer patch.Unpatch()

	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	monkey.Patch(time.Sleep, func(d time.Duration) {})
	defer monkey.UnpatchAll()

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT 1").
		WillReturnError(&mysql.MySQLError{Number: 1213})
	mock.ExpectRollback()
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT 1").
		WillReturnRows(mock.NewRows([]string{"1"}).AddRow(1))
	mock.ExpectCommit()

	assert.NoError(t, db.inTransaction(func(db *DB) error {
		var result []interface{}
		return db.Raw("SELECT 1").Scan(&result).Error()
	}, txOptions))
	assert.NoError(t, mock.ExpectationsWereMet())
	assert.Equal(t, 2, calledCount)
}

func TestDB_inTransaction_SkipsRetryingOnDeadockAndLockWaitTimeoutPanicOnRollbackError(t *testing.T) {
	for _, errorNumber := range []uint16{1213, 1205} {
		t.Run(fmt.Sprintf("error%d", errorNumber), func(t *testing.T) {
			testoutput.SuppressIfPasses(t)

			db, mock := NewDBMock()
			defer func() { _ = db.Close() }()

			var duration time.Duration
			monkey.Patch(time.Sleep, func(d time.Duration) { duration += d })
			defer monkey.UnpatchAll()

			expectedError := errors.New("rollback error")
			mock.ExpectBegin()
			mock.ExpectQuery("SELECT 1").
				WillReturnError(&mysql.MySQLError{Number: errorNumber})
			mock.ExpectRollback().WillReturnError(expectedError)

			err := db.inTransaction(func(db *DB) error {
				var result []interface{}
				mustNotBeError(db.Raw("SELECT 1").Scan(&result).Error())
				return nil
			})
			assert.Equal(t, expectedError, err)

			assert.Zero(t, duration)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestDB_inTransaction_RetriesAllowedUpToTheLimit_Panic(t *testing.T) {
	for _, errorNumber := range []uint16{1213, 1205} {
		t.Run(fmt.Sprintf("error%d", errorNumber), func(t *testing.T) {
			testoutput.SuppressIfPasses(t)

			db, mock := NewDBMock()
			defer func() { _ = db.Close() }()

			var duration time.Duration
			monkey.Patch(time.Sleep, func(d time.Duration) { duration += d })
			defer monkey.UnpatchAll()

			for i := 0; i < transactionRetriesLimit; i++ {
				mock.ExpectBegin()
				mock.ExpectQuery("SELECT 1").
					WillReturnError(&mysql.MySQLError{Number: errorNumber})
				mock.ExpectRollback()
			}
			mock.ExpectBegin()
			mock.ExpectQuery("SELECT 1").
				WillReturnRows(mock.NewRows([]string{"1"}).AddRow(1))
			mock.ExpectCommit()

			assert.NoError(t, db.inTransaction(func(db *DB) error {
				var result []interface{}
				mustNotBeError(db.Raw("SELECT 1").Scan(&result).Error())
				return nil
			}))
			assert.InEpsilon(t, transactionRetriesLimit*transactionDelayBetweenRetries, duration, transactionRetriesLimit*0.05)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestDB_inTransaction_RetriesAllowedUpToTheLimit_Error(t *testing.T) {
	for _, errorNumber := range []uint16{1213, 1205} {
		t.Run(fmt.Sprintf("error%d", errorNumber), func(t *testing.T) {
			testoutput.SuppressIfPasses(t)

			db, mock := NewDBMock()
			defer func() { _ = db.Close() }()

			var duration time.Duration
			monkey.Patch(time.Sleep, func(d time.Duration) { duration += d })
			defer monkey.UnpatchAll()

			for i := 0; i < transactionRetriesLimit; i++ {
				mock.ExpectBegin()
				mock.ExpectQuery("SELECT 1").
					WillReturnError(&mysql.MySQLError{Number: errorNumber})
				mock.ExpectRollback()
			}
			mock.ExpectBegin()
			mock.ExpectQuery("SELECT 1").
				WillReturnRows(mock.NewRows([]string{"1"}).AddRow(1))
			mock.ExpectCommit()

			assert.NoError(t, db.inTransaction(func(db *DB) error {
				var result []interface{}
				return db.Raw("SELECT 1").Scan(&result).Error()
			}))
			assert.InEpsilon(t, transactionRetriesLimit*transactionDelayBetweenRetries, duration, transactionRetriesLimit*0.05)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestDB_inTransaction_RetriesAboveTheLimitAreDisallowed_Panic(t *testing.T) {
	for _, errorNumber := range []uint16{1213, 1205} {
		t.Run(fmt.Sprintf("error%d", errorNumber), func(t *testing.T) {
			testoutput.SuppressIfPasses(t)

			db, mock := NewDBMock()
			defer func() { _ = db.Close() }()

			loggerHook, restoreLoggerFunc := logging.MockSharedLoggerHook()
			defer restoreLoggerFunc()
			conf := viper.New()
			conf.Set("Level", "error")
			logging.SharedLogger.Configure(conf)

			var duration time.Duration
			monkey.Patch(time.Sleep, func(d time.Duration) { duration += d })
			defer monkey.UnpatchAll()

			for i := 0; i < transactionRetriesLimit+1; i++ {
				mock.ExpectBegin()
				mock.ExpectQuery("SELECT 1").
					WillReturnError(&mysql.MySQLError{Number: errorNumber})
				mock.ExpectRollback()
			}

			assert.Equal(t, errors.New("transaction retries limit exceeded"),
				db.inTransaction(func(db *DB) error {
					var result []interface{}
					mustNotBeError(db.Raw("SELECT 1").Scan(&result).Error())
					return nil
				}))
			assert.InEpsilon(t, transactionRetriesLimit*transactionDelayBetweenRetries, duration, transactionRetriesLimit*0.05)
			assert.NoError(t, mock.ExpectationsWereMet())

			require.Len(t, loggerHook.AllEntries(), 1)
			assert.Equal(t, "error", loggerHook.LastEntry().Level.String())
			assert.Equal(t,
				fmt.Sprintf("transaction retries limit has been exceeded, cannot retry the error: Error %d: ", errorNumber),
				loggerHook.LastEntry().Message)
		})
	}
}

func TestDB_inTransaction_RetriesAboveTheLimitAreDisallowed_Error(t *testing.T) {
	for _, errorNumber := range []uint16{1213, 1205} {
		t.Run(fmt.Sprintf("error%d", errorNumber), func(t *testing.T) {
			testoutput.SuppressIfPasses(t)

			db, mock := NewDBMock()
			defer func() { _ = db.Close() }()

			loggerHook, restoreLoggerFunc := logging.MockSharedLoggerHook()
			defer restoreLoggerFunc()
			conf := viper.New()
			conf.Set("Level", "error")
			logging.SharedLogger.Configure(conf)

			var duration time.Duration
			monkey.Patch(time.Sleep, func(d time.Duration) { duration += d })
			defer monkey.UnpatchAll()

			for i := 0; i < transactionRetriesLimit+1; i++ {
				mock.ExpectBegin()
				mock.ExpectQuery("SELECT 1").
					WillReturnError(&mysql.MySQLError{Number: errorNumber})
				mock.ExpectRollback()
			}

			assert.Equal(t, errors.New("transaction retries limit exceeded"),
				db.inTransaction(func(db *DB) error {
					var result []interface{}
					return db.Raw("SELECT 1").Scan(&result).Error()
				}))
			assert.InEpsilon(t, transactionRetriesLimit*transactionDelayBetweenRetries, duration, transactionRetriesLimit*0.05)
			assert.NoError(t, mock.ExpectationsWereMet())

			require.Len(t, loggerHook.AllEntries(), 1)
			assert.Equal(t, "error", loggerHook.LastEntry().Level.String())
			assert.Equal(t,
				fmt.Sprintf("transaction retries limit has been exceeded, cannot retry the error: Error %d: ", errorNumber),
				loggerHook.LastEntry().Message)
		})
	}
}

func TestDB_QueryConstructors(t *testing.T) {
	tests := []struct {
		name              string
		funcToPrepare     func(*DB, sqlmock.Sqlmock) *DB
		funcToCall        func(*DB) (*DB, []*DB)
		expectedQuery     string
		expectedQueryArgs []driver.Value
		skipCTEsChecking  bool
	}{
		{
			name:          "Limit",
			funcToCall:    func(db *DB) (*DB, []*DB) { return db.Limit(1), nil },
			expectedQuery: "SELECT * FROM `myTable` LIMIT 1",
		},
		{
			name:          "Limit and Offset",
			funcToCall:    func(db *DB) (*DB, []*DB) { return db.Limit(1).Offset(2), nil },
			expectedQuery: "SELECT * FROM `myTable` LIMIT 1 OFFSET 2",
		},
		{
			name:          "Order",
			funcToCall:    func(db *DB) (*DB, []*DB) { return db.Order("id"), nil },
			expectedQuery: "SELECT * FROM `myTable` ORDER BY `id`",
		},
		{
			name:          "Having",
			funcToCall:    func(db *DB) (*DB, []*DB) { return db.Having("id > 0"), nil },
			expectedQuery: "SELECT * FROM `myTable` HAVING (id > 0)",
		},
		{
			name:             "Raw",
			funcToCall:       func(db *DB) (*DB, []*DB) { return db.Raw("SELECT 1"), nil },
			expectedQuery:    "SELECT 1",
			skipCTEsChecking: true,
		},
		{
			name: "Union",
			funcToCall: func(db *DB) (newDB *DB, dbs []*DB) {
				dbTwo := db.Table("otherTable")
				dbs = append(dbs, dbTwo)
				return db.Union(dbTwo), dbs
			},
			expectedQuery:    "(SELECT * FROM `myTable` ) UNION (SELECT * FROM `otherTable` )",
			skipCTEsChecking: true,
		},
		{
			name: "UnionAll",
			funcToCall: func(db *DB) (newDB *DB, dbs []*DB) {
				dbTwo := db.Table("otherTable")
				dbs = append(dbs, dbTwo)
				return db.UnionAll(dbTwo), dbs
			},
			expectedQuery:    "(SELECT * FROM `myTable` ) UNION ALL (SELECT * FROM `otherTable` )",
			skipCTEsChecking: true,
		},
		{
			name: "With",
			funcToCall: func(db *DB) (newDB *DB, dbs []*DB) {
				dbTwo := db.Table("otherTable")
				dbs = append(dbs, dbTwo)
				dbThree := db.Table("thirdTable")
				dbs = append(dbs, dbThree)
				dbFour := db.With("t1", dbTwo)
				dbs = append(dbs, dbFour)
				return dbFour.With("t2", dbThree), dbs
			},
			expectedQuery:    "WITH `t1` AS (SELECT * FROM `otherTable` ), `t2` AS (SELECT * FROM `thirdTable` ) SELECT * FROM `myTable`",
			skipCTEsChecking: true,
		},
		{
			name: "With (with locking)",
			funcToPrepare: func(db *DB, s sqlmock.Sqlmock) *DB {
				s.ExpectBegin()
				s.ExpectCommit()
				return db
			},
			funcToCall: func(db *DB) (newDB *DB, dbs []*DB) {
				var queryExpr interface{}
				mustNotBeError(db.inTransaction(func(db *DB) error {
					dbTwo := db.Table("otherTable").WithExclusiveWriteLock()
					queryExpr = db.WithExclusiveWriteLock().With("t1", dbTwo).QueryExpr()
					return nil
				}))
				newDB = db.Raw("?", queryExpr)
				dbs = append(dbs, newDB)
				return newDB, dbs
			},
			expectedQuery:    "WITH `t1` AS (SELECT * FROM `otherTable` FOR UPDATE) SELECT * FROM `myTable` FOR UPDATE",
			skipCTEsChecking: true,
		},
	}
	for _, testCase := range tests {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			testoutput.SuppressIfPasses(t)

			db, mock := NewDBMock()
			defer func() { _ = db.Close() }()

			db = cloneDBWithNewContext(context.WithValue(context.Background(), testContextKey("k"), "v"), db)
			db = db.Table("myTable")
			db.ctes = make([]cte, 0, 1)
			if testCase.funcToPrepare != nil {
				db = testCase.funcToPrepare(db, mock)
			}

			mock.ExpectQuery("^" + regexp.QuoteMeta(testCase.expectedQuery) + "$").
				WithArgs(testCase.expectedQueryArgs...).
				WillReturnRows(mock.NewRows([]string{"1"}).AddRow(1))

			resultDB, oldDBObjects := testCase.funcToCall(db)
			assert.NotEqual(t, resultDB, db)
			assert.Equal(t, db.ctx(), resultDB.ctx())
			assert.Equal(t, db.logConfig, resultDB.logConfig)
			if !testCase.skipCTEsChecking {
				assert.Equal(t, db.ctes, resultDB.ctes)
			}
			for _, oldDBObject := range oldDBObjects {
				assert.NotEqual(t, oldDBObject, db)
				assert.Equal(t, db.ctx(), oldDBObject.ctx())
				assert.Equal(t, db.logConfig, oldDBObject.logConfig)
				if !testCase.skipCTEsChecking {
					assert.Equal(t, db.ctes, oldDBObject.ctes)
				}
			}
			assert.NoError(t, resultDB.Error())

			var result []interface{}
			assert.NoError(t, resultDB.Scan(&result).Error())
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestDB_With_Duplicate(t *testing.T) {
	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	assert.Panics(t, func() {
		db.Table("myTable").
			With("t1", db.Table("otherTable")).
			With("t1", db.Table("thirdTable"))
	})
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDB_Count(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT count(*) FROM `myTable`")).
		WillReturnRows(mock.NewRows([]string{"count"}).AddRow(1))

	db = cloneDBWithNewContext(context.WithValue(context.Background(), testContextKey("k"), "v"), db)
	db = db.Table("myTable")
	db.ctes = make([]cte, 0, 1)

	var result int
	countDB := db.Count(&result)

	assert.NotEqual(t, countDB, db)
	assert.NoError(t, countDB.Error())
	assert.Equal(t, db.ctx(), countDB.ctx())
	assert.Nil(t, countDB.ctes)
	assert.Equal(t, db.logConfig, countDB.logConfig)

	assert.Equal(t, 1, result)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDB_Count_DoesNothingIfDBContainsError(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	expectedError := errors.New("some error")
	db = db.Table("myTable")
	_ = db.db.AddError(expectedError)

	var result int
	countDB := db.Count(&result)

	assert.Equal(t, countDB, db)
	assert.Equal(t, expectedError, countDB.Error())
	assert.Zero(t, result)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDB_Take(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `myTable` WHERE (id = 1) LIMIT 1")).
		WillReturnRows(mock.NewRows([]string{"id"}).AddRow(1))

	db = cloneDBWithNewContext(context.WithValue(context.Background(), testContextKey("k"), "v"), db)
	db = db.Table("myTable")
	db.ctes = make([]cte, 0, 1)

	type resultType struct{ ID int }
	var result resultType
	takeDB := db.Take(&result, "id = 1")

	assert.NotEqual(t, takeDB, db)
	assert.NoError(t, takeDB.Error())
	assert.Equal(t, db.ctx(), takeDB.ctx())
	assert.Nil(t, takeDB.ctes)
	assert.Equal(t, db.logConfig, takeDB.logConfig)

	assert.Equal(t, resultType{1}, result)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDB_HasRows(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT 1 FROM `myTable` WHERE (id = 1) LIMIT 1")).
		WillReturnRows(mock.NewRows([]string{"1"}).AddRow(1))

	db = db.Table("myTable")

	found, err := db.Where("id = 1").HasRows()

	assert.NoError(t, err)
	assert.True(t, found)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDB_HasRows_NoRows(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT 1 FROM `myTable` WHERE (id = 1) LIMIT 1")).
		WillReturnRows(mock.NewRows([]string{"1"}))

	db = db.Table("myTable")

	found, err := db.Where("id = 1").HasRows()

	assert.NoError(t, err)
	assert.False(t, found)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDB_HasRows_Error(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	expectedError := errors.New("some error")
	mock.ExpectQuery(regexp.QuoteMeta("SELECT 1 FROM `myTable` WHERE (id = 1) LIMIT 1")).
		WillReturnError(expectedError)

	db = db.Table("myTable")

	found, err := db.Where("id = 1").HasRows()

	assert.Equal(t, expectedError, err)
	assert.False(t, found)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDB_Pluck(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id FROM `myTable`")).
		WillReturnRows(mock.NewRows([]string{"id"}).AddRow(1))

	db = cloneDBWithNewContext(context.WithValue(context.Background(), testContextKey("k"), "v"), db)
	db = db.Table("myTable")
	db.ctes = make([]cte, 0, 1)

	var result []int64
	pluckDB := db.Pluck("id", &result)

	assert.NotEqual(t, pluckDB, db)
	assert.NoError(t, pluckDB.Error())
	assert.Equal(t, db.ctx(), pluckDB.ctx())
	assert.Nil(t, pluckDB.ctes)
	assert.Equal(t, db.logConfig, pluckDB.logConfig)

	assert.Equal(t, []int64{1}, result)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDB_Pluck_DoesNothingIfErrorIsSet(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	db = db.Table("myTable")
	expectedError := errors.New("some error")
	db.db.Error = expectedError
	var result []int64
	pluckDB := db.Pluck("id", &result)
	assert.Equal(t, expectedError, pluckDB.Error())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDB_Pluck_WipesOldData(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id FROM `myTable`")).
		WillReturnRows(mock.NewRows([]string{"id"}).AddRow(1))

	db = db.Table("myTable")

	result := []int64{1, 2, 3}
	db.Pluck("id", &result)

	assert.Equal(t, []int64{1}, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDB_Pluck_NonSlicePointer(t *testing.T) {
	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	db = db.Table("myTable")

	result := 1
	assert.PanicsWithValue(t, "values should be a pointer to a slice, not a pointer to int", func() {
		db.Pluck("id", &result)
	})

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDB_Pluck_NonPointer(t *testing.T) {
	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	db = db.Table("myTable")

	result := 1
	assert.PanicsWithValue(t, "values should be a pointer to a slice, not int", func() {
		db.Pluck("id", result)
	})

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDB_PluckFirst(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id FROM `myTable` LIMIT 1")).
		WillReturnRows(mock.NewRows([]string{"id"}).AddRow(1))

	db = cloneDBWithNewContext(context.WithValue(context.Background(), testContextKey("k"), "v"), db)
	db = db.Table("myTable")
	db.ctes = make([]cte, 0, 1)

	var result int64
	pluckFirstDB := db.PluckFirst("id", &result)

	assert.NotEqual(t, pluckFirstDB, db)
	assert.NoError(t, pluckFirstDB.Error())
	assert.Equal(t, db.ctx(), pluckFirstDB.ctx())
	assert.Nil(t, pluckFirstDB.ctes)
	assert.Equal(t, db.logConfig, pluckFirstDB.logConfig)

	assert.Equal(t, int64(1), result)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDB_PluckFirst_NotFound(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id FROM `myTable` LIMIT 1")).
		WillReturnRows(mock.NewRows([]string{"id"}))

	db = db.Table("myTable")

	var result int64
	pluckFirstDB := db.PluckFirst("id", &result)

	assert.NotEqual(t, pluckFirstDB, db)
	assert.Equal(t, gorm.ErrRecordNotFound, pluckFirstDB.Error())

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDB_PluckFirst_Error(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	expectedError := errors.New("some error")
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id FROM `myTable` LIMIT 1")).
		WillReturnError(expectedError)

	db = db.Table("myTable")

	var result int64
	pluckFirstDB := db.PluckFirst("id", &result)

	assert.NotEqual(t, pluckFirstDB, db)
	assert.Equal(t, expectedError, pluckFirstDB.Error())

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDB_Scan(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `myTable`")).
		WillReturnRows(mock.NewRows([]string{"id", "value"}).AddRow(int64(1), "value"))

	db = cloneDBWithNewContext(context.WithValue(context.Background(), testContextKey("k"), "v"), db)
	db = db.Table("myTable")
	db.ctes = make([]cte, 0, 1)

	type resultType struct {
		ID    int64
		Value string
	}
	var result []resultType
	scanDB := db.Scan(&result)

	assert.NotEqual(t, scanDB, db)
	assert.NoError(t, scanDB.Error())
	assert.Equal(t, db.ctx(), scanDB.ctx())
	assert.Nil(t, scanDB.ctes)
	assert.Equal(t, db.logConfig, scanDB.logConfig)

	assert.Equal(t, []resultType{{ID: 1, Value: "value"}}, result)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDB_Scan_WipesOldData(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `myTable`")).
		WillReturnRows(mock.NewRows([]string{"id", "value"}).AddRow(int64(1), "value"))

	db = db.Table("myTable")

	type resultType struct {
		ID    int64
		Value string
	}
	result := []resultType{{ID: 2, Value: "another value"}, {ID: 3, Value: "third value"}}
	scanDB := db.Scan(&result)

	assert.NoError(t, scanDB.Error())
	assert.Equal(t, []resultType{{ID: 1, Value: "value"}}, result)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDB_Scan_NonSlicePointer(t *testing.T) {
	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	db = db.Table("myTable")

	result := 1
	scanDB := db.Scan(&result)

	assert.EqualError(t, scanDB.Error(), "unsupported destination, should be slice or struct")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDB_Scan_NonPointer(t *testing.T) {
	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	db = db.Table("myTable")

	type resultType struct {
		ID    int64
		Value string
	}
	var result []resultType
	assert.Panics(t, func() {
		db.Scan(result)
	})

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDB_RowsAffected(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	mock.ExpectExec("^" + regexp.QuoteMeta("UPDATE `myTable` SET `myColumn` = ?") + "$").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(-1, 123))

	rowsAffected := db.Table("myTable").UpdateColumn("myColumn", 1).RowsAffected()

	assert.Equal(t, int64(123), rowsAffected)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDB_Delete(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM `myTable`") + `\s+` +
		regexp.QuoteMeta("WHERE (id = 1)")).
		WillReturnResult(sqlmock.NewResult(-1, 1))

	db = cloneDBWithNewContext(context.WithValue(context.Background(), testContextKey("k"), "v"), db)
	db = db.Table("myTable")
	db.ctes = make([]cte, 0, 1)

	deleteDB := db.Delete("id = 1")

	assert.NotEqual(t, deleteDB, db)
	assert.Equal(t, db.ctx(), deleteDB.ctx())
	assert.Nil(t, deleteDB.ctes)
	assert.Equal(t, db.logConfig, deleteDB.logConfig)

	assert.NoError(t, deleteDB.Error())

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDB_Exec(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	query := "UPDATE users set id = ? WHERE id = ?"
	mock.ExpectExec("^"+regexp.QuoteMeta(query)+"$").
		WithArgs(1, 2).
		WillReturnResult(sqlmock.NewResult(-1, 1))

	db = cloneDBWithNewContext(context.WithValue(context.Background(), testContextKey("k"), "v"), db)
	db.ctes = make([]cte, 0, 1)

	execDB := db.Exec(query, 1, 2)

	assert.NotEqual(t, execDB, db)
	assert.NoError(t, execDB.Error())
	assert.Equal(t, db.ctx(), execDB.ctx())
	assert.Nil(t, execDB.ctes)
	assert.Equal(t, db.logConfig, execDB.logConfig)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDB_insertMaps(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	dataRows := []map[string]interface{}{{"id": int64(1), "sField": "some value", "sNullField": nil}}

	expectedError := errors.New("some error")
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `myTable` (`id`, `sField`, `sNullField`) VALUES (?, ?, ?)")).
		WithArgs(int64(1), "some value", nil).
		WillReturnError(expectedError)

	assert.Equal(t, expectedError, db.insertMaps("myTable", dataRows))
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDB_insertMaps_MultipleRows(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	dataRows := []map[string]interface{}{
		{"id": int64(1), "sField": "some value", "sNullField": "value"},
		{"id": int64(2), "sField": "another value", "sNullField": nil},
	}

	expectedError := errors.New("some error")
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `myTable` (`id`, `sField`, `sNullField`) VALUES (?, ?, ?), (?, ?, ?)")).
		WithArgs(int64(1), "some value", "value", int64(2), "another value", nil).
		WillReturnError(expectedError)

	assert.Equal(t, expectedError, db.insertMaps("myTable", dataRows))
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDB_insertMaps_WithEmptyArray(t *testing.T) {
	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()
	var dataRows []map[string]interface{}
	assert.NoError(t, db.insertMaps("myTable", dataRows))
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDB_insertOrUpdateMaps(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	dataRows := []map[string]interface{}{{"id": int64(1), "sField": "some value", "sNullField": nil}}

	expectedError := errors.New("some error")
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `myTable` (`id`, `sField`, `sNullField`) VALUES (?, ?, ?)"+
		" ON DUPLICATE KEY UPDATE `sField` = VALUES(`sField`), `sNullField` = VALUES(`sNullField`)")).
		WithArgs(int64(1), "some value", nil).
		WillReturnError(expectedError)

	assert.Equal(t, expectedError, db.insertOrUpdateMaps("myTable", dataRows, []string{"sField", "sNullField"}))
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDB_insertOrUpdateMaps_WithNilUpdateColumnsList(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	dataRows := []map[string]interface{}{{"id": int64(1), "sField": "some value", "sNullField": nil}}

	expectedError := errors.New("some error")
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `myTable` (`id`, `sField`, `sNullField`) VALUES (?, ?, ?)"+
		" ON DUPLICATE KEY UPDATE `id` = VALUES(`id`), `sField` = VALUES(`sField`), `sNullField` = VALUES(`sNullField`)")).
		WithArgs(int64(1), "some value", nil).
		WillReturnError(expectedError)

	assert.Equal(t, expectedError, db.insertOrUpdateMaps("myTable", dataRows, nil))
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDB_insertOrUpdateMaps_WithEmptyArray(t *testing.T) {
	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()
	var dataRows []map[string]interface{}
	assert.NoError(t, db.insertOrUpdateMaps("myTable", dataRows, []string{"sField", "sNullField"}))
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDB_ScanIntoSlices(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, Field FROM `myTable`")).
		WillReturnRows(
			mock.NewRows([]string{"id", "Field"}).
				AddRow(1, "value").AddRow(2, "another value").AddRow([]byte("3"), nil))

	db = db.Table("myTable").Select("id, Field")

	ids := make([]int64, 0, 3)
	fields := make([]*string, 0, 3)

	dbScan := db.ScanIntoSlices(&ids, &fields)
	assert.Equal(t, dbScan, db)
	assert.NoError(t, dbScan.Error())

	assert.Equal(t, []int64{1, 2, 3}, ids)
	assert.Equal(t, []*string{golang.Ptr("value"), golang.Ptr("another value"), nil}, fields)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDB_ScanIntoSlices_DoesNothingIfErrorIsSet(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	expectedError := errors.New("some error")
	_ = db.db.AddError(expectedError)
	var result []int64
	dbScan := db.ScanIntoSlices(&result)
	assert.Equal(t, dbScan, db)
	assert.Equal(t, expectedError, dbScan.Error())

	assert.Equal(t, []int64(nil), result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDB_ScanIntoSlices_WipesOldData(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `myTable`")).
		WillReturnRows(
			mock.NewRows([]string{"id", "Field"}).
				AddRow(1, "value").AddRow(2, "another value").AddRow([]byte("3"), nil))

	db = db.Table("myTable")

	ids := []int64{10, 20, 30}
	fields := []*string{golang.Ptr("old value1"), golang.Ptr("old value2"), golang.Ptr("old value3")}

	db.ScanIntoSlices(&ids, &fields)
	assert.Equal(t, []int64{1, 2, 3}, ids)
	assert.Equal(t, []*string{golang.Ptr("value"), golang.Ptr("another value"), nil}, fields)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDB_ScanIntoSlices_RowsError(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	expectedError := errors.New("some error")
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `myTable`")).
		WillReturnError(expectedError)
	db = db.Table("myTable")

	var result []int64
	dbScan := db.ScanIntoSlices(&result)
	assert.Equal(t, dbScan, db)
	assert.Equal(t, expectedError, dbScan.Error())

	assert.Nil(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDB_ScanIntoSlices_ErrorOnScan(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	expectedError := errors.New("some error")

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `myTable`")).WillReturnRows(mock.NewRows([]string{"id"}).AddRow(1))
	monkey.PatchInstanceMethod(reflect.TypeOf(&sql.Rows{}), "Scan", func(*sql.Rows, ...interface{}) error { return expectedError })
	defer monkey.UnpatchAll()
	db = db.Table("myTable")

	var result []int64
	dbScan := db.ScanIntoSlices(&result)
	assert.Equal(t, dbScan, db)
	assert.Equal(t, expectedError, dbScan.Error())

	assert.Equal(t, []int64(nil), result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDB_ScanIntoSliceOfMaps(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `myTable`")).
		WillReturnRows(
			mock.NewRows([]string{"id", "Field"}).
				AddRow(1, "value").AddRow(2, "another value").AddRow([]byte("3"), nil))

	db = db.Table("myTable")

	var result []map[string]interface{}
	dbScan := db.ScanIntoSliceOfMaps(&result)
	assert.Equal(t, dbScan, db)
	assert.NoError(t, dbScan.Error())

	assert.Equal(t, []map[string]interface{}{
		{"id": int64(1), "Field": "value"},
		{"id": int64(2), "Field": "another value"},
		{"id": "3", "Field": nil},
	}, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDB_ScanIntoSliceOfMaps_DoesNothingIfErrorIsSet(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	expectedError := errors.New("some error")
	_ = db.db.AddError(expectedError)
	var result []map[string]interface{}
	dbScan := db.ScanIntoSliceOfMaps(&result)
	assert.Equal(t, dbScan, db)
	assert.Equal(t, expectedError, dbScan.Error())

	assert.Equal(t, []map[string]interface{}(nil), result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDB_ScanIntoSliceOfMaps_WipesOldData(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `myTable`")).
		WillReturnRows(
			mock.NewRows([]string{"id", "Field"}).
				AddRow(1, "value").AddRow(2, "another value").AddRow([]byte("3"), nil))

	db = db.Table("myTable")

	result := []map[string]interface{}{
		{"column": "value"},
	}
	db.ScanIntoSliceOfMaps(&result)
	assert.Equal(t, []map[string]interface{}{
		{"id": int64(1), "Field": "value"},
		{"id": int64(2), "Field": "another value"},
		{"id": "3", "Field": nil},
	}, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDB_ScanIntoSliceOfMaps_RowsError(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	expectedError := errors.New("some error")
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `myTable`")).
		WillReturnError(expectedError)
	db = db.Table("myTable")

	var result []map[string]interface{}
	dbScan := db.ScanIntoSliceOfMaps(&result)
	assert.Equal(t, dbScan, db)
	assert.Equal(t, expectedError, dbScan.Error())

	assert.Nil(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDB_ScanIntoSliceOfMaps_ErrorOnGettingColumns(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	expectedError := errors.New("some error")

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `myTable`")).WillReturnRows(mock.NewRows([]string{"id"}).AddRow(1))
	monkey.PatchInstanceMethod(reflect.TypeOf(&sql.Rows{}), "Columns", func(*sql.Rows) ([]string, error) { return nil, expectedError })
	defer monkey.UnpatchAll()
	db = db.Table("myTable")

	var result []map[string]interface{}
	dbScan := db.ScanIntoSliceOfMaps(&result)
	assert.Equal(t, dbScan, db)
	assert.Equal(t, expectedError, dbScan.Error())

	assert.Equal(t, []map[string]interface{}(nil), result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDB_ScanIntoSliceOfMaps_ErrorOnScan(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	expectedError := errors.New("some error")

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `myTable`")).WillReturnRows(mock.NewRows([]string{"id"}).AddRow(1))
	monkey.PatchInstanceMethod(reflect.TypeOf(&sql.Rows{}), "Scan", func(*sql.Rows, ...interface{}) error { return expectedError })
	defer monkey.UnpatchAll()
	db = db.Table("myTable")

	var result []map[string]interface{}
	dbScan := db.ScanIntoSliceOfMaps(&result)
	assert.Equal(t, dbScan, db)
	assert.Equal(t, expectedError, dbScan.Error())

	assert.Equal(t, []map[string]interface{}(nil), result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDB_ScanAndHandleMaps_FailsIfHandlerReturnsError(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	expectedError := errors.New("some error")

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `myTable`")).WillReturnRows(mock.NewRows([]string{"id"}).
		AddRow(1).AddRow(2))
	db = db.Table("myTable")

	dbScan := db.ScanAndHandleMaps(func(map[string]interface{}) error {
		return expectedError
	})
	assert.Equal(t, dbScan, db)
	assert.Equal(t, expectedError, dbScan.Error())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDB_UpdateColumns(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	mock.ExpectExec(regexp.QuoteMeta("UPDATE `myTable` SET `id` = ?, `name` = ?")).
		WithArgs(1, someName).
		WillReturnResult(sqlmock.NewResult(0, 1))

	db = db.Table("myTable")
	updateDB := db.UpdateColumns(map[string]interface{}{"id": 1, "name": someName})
	assert.NotEqual(t, updateDB, db)
	assert.NoError(t, updateDB.Error())

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDB_UpdateColumn(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	mock.ExpectExec(regexp.QuoteMeta("UPDATE `myTable` SET `name` = ?")).
		WithArgs(someName).
		WillReturnResult(sqlmock.NewResult(0, 1))

	db = cloneDBWithNewContext(context.WithValue(context.Background(), testContextKey("k"), "v"), db)
	db = db.Table("myTable")
	db.ctes = make([]cte, 0, 1)

	updateDB := db.UpdateColumn("name", someName)
	assert.NotEqual(t, updateDB, db)
	assert.NoError(t, updateDB.Error())
	assert.Equal(t, db.ctx(), updateDB.ctx())
	assert.Nil(t, updateDB.ctes)
	assert.Equal(t, db.logConfig, updateDB.logConfig)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDB_Set(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	mock.ExpectQuery("SELECT \\* FROM `myTable` FOR UPDATE").
		WillReturnRows(mock.NewRows([]string{"1"}).AddRow(1))

	db = cloneDBWithNewContext(context.WithValue(context.Background(), testContextKey("k"), "v"), db)
	db = db.Table("myTable")
	db.ctes = make([]cte, 0, 1)

	setDB := db.Set("gorm:query_option", "FOR UPDATE")
	assert.NotEqual(t, setDB, db)
	assert.NoError(t, setDB.Error())
	assert.Equal(t, db.ctx(), setDB.ctx())
	assert.Equal(t, db.ctes, setDB.ctes)
	assert.Equal(t, db.logConfig, setDB.logConfig)

	var result []interface{}
	assert.NoError(t, setDB.Scan(&result).Error())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestOpenRawDBConnection(t *testing.T) {
	db, err := OpenRawDBConnection("/db", true)
	assert.NoError(t, err)
	assert.Contains(t, sql.Drivers(), "instrumented-mysql")
	assertRawDBIsOK(t, db)
}

func TestOpen_DSN(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	patchGuard := monkey.PatchInstanceMethod(reflect.TypeOf(&logging.Logger{}), "IsRawSQLQueriesLoggingEnabled",
		func(*logging.Logger) bool { return true })
	defer patchGuard.Unpatch()

	db, err := Open("/db")
	assert.Error(t, err) // we want an error since dsn is wrong, but other things should be ok
	assert.NotNil(t, db)
	assert.NotNil(t, db.db)
	assert.Contains(t, sql.Drivers(), "instrumented-mysql")
	rawDB := db.db.CommonDB().(*sqlDBWrapper).sqlDB
	assertRawDBIsOK(t, rawDB)
}

func TestOpen_WrongSourceType(t *testing.T) {
	db, err := Open(1234)
	assert.Equal(t, errors.New("unknown database source type: int (1234)"), err)
	assert.Nil(t, db)
}

func TestOpen_OpenRawDBConnectionError(t *testing.T) {
	expectedError := errors.New("some error")
	monkey.Patch(OpenRawDBConnection, func(string, bool) (*sql.DB, error) { return &sql.DB{}, expectedError })
	defer monkey.UnpatchAll()

	db, err := Open("mydsn")
	assert.Equal(t, expectedError, err)
	assert.Nil(t, db)
}

func assertRawDBIsOK(t *testing.T, rawDB *sql.DB) {
	assert.Equal(t, "*instrumentedsql.WrappedDriver", fmt.Sprintf("%T", rawDB.Driver()))
	assert.Contains(t, fmt.Sprintf("%#v", rawDB), "parent:(*database.mysqlConnectorWrapper)")
}

func TestDB_mustBeInTransaction_DoesNothingInTransaction(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	mock.ExpectBegin()
	mock.ExpectCommit()

	assert.NoError(t, db.inTransaction(func(db *DB) error {
		assert.NotPanics(t, func() { db.mustBeInTransaction() })
		return nil
	}))
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDB_mustBeInTransaction_PanicWhenNotInTransaction(t *testing.T) {
	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	assert.PanicsWithValue(t, ErrNoTransaction, func() { db.mustBeInTransaction() })
	assert.NoError(t, mock.ExpectationsWereMet())
}

func Test_recoverPanics_RecoversError(t *testing.T) {
	expectedError := errors.New("some error")
	err := func() (err error) {
		defer recoverPanics(&err)
		panic(expectedError)
	}()
	assert.Equal(t, expectedError, err)
}

func Test_recoverPanics_PanicsOnRuntimeError(t *testing.T) {
	didPanic, panicValue := func() (didPanic bool, panicValue interface{}) {
		defer func() {
			if p := recover(); p != nil {
				didPanic = true
				panicValue = p
			}
		}()

		_ = func() (err error) {
			defer recoverPanics(&err)
			var a []int
			a[0]++ // runtime error
			return nil
		}()

		return false, nil
	}()

	assert.True(t, didPanic)
	assert.Implements(t, (*runtime.Error)(nil), panicValue)
	assert.Equal(t, "runtime error: index out of range [0] with length 0", panicValue.(error).Error())
}

func Test_recoverPanics_PanicsOnRecoveringValueOfNonErrorType(t *testing.T) {
	expectedPanicValue := "some panic"
	didPanic, panicValue := func() (didPanic bool, panicValue interface{}) {
		defer func() {
			if p := recover(); p != nil {
				didPanic = true
				panicValue = p
			}
		}()

		_ = func() (err error) {
			defer recoverPanics(&err)
			panic(expectedPanicValue)
		}()

		return false, nil
	}()

	assert.True(t, didPanic)
	assert.Equal(t, expectedPanicValue, panicValue)
}

func TestDB_withNamedLock_ReturnsErrNamedLockWaitTimeoutExceededWhenGetLockTimeouts(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db, dbMock := NewDBMock()
	defer func() { _ = db.Close() }()

	lockName := "lock name"
	timeout := 1234 * time.Millisecond
	expectedTimeout := int(timeout.Round(time.Second).Seconds())

	dbMock.ExpectQuery("^"+regexp.QuoteMeta("SELECT GET_LOCK(?, ?)")+"$").
		WithArgs(lockName, expectedTimeout).
		WillReturnRows(sqlmock.NewRows([]string{"GET_LOCK('" + lockName + "', " + strconv.Itoa(expectedTimeout) + ")"}).AddRow(int64(0)))

	err := db.withNamedLock(lockName, timeout, func(*DB) error {
		return nil
	})
	assert.Equal(t, ErrNamedLockWaitTimeoutExceeded, err)
	assert.NoError(t, dbMock.ExpectationsWereMet())
}

func TestDB_withNamedLock_ReturnsErrorWhenDBFails(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db, dbMock := NewDBMock()
	defer func() { _ = db.Close() }()

	lockName := someName
	timeout := 1234 * time.Millisecond
	expectedTimeout := int(timeout.Round(time.Second).Seconds())
	expectedError := errors.New("some error")

	dbMock.ExpectQuery("^"+regexp.QuoteMeta("SELECT GET_LOCK(?, ?)")+"$").
		WithArgs(lockName, expectedTimeout).
		WillReturnError(expectedError)

	err := db.withNamedLock(lockName, timeout, func(*DB) error {
		return nil
	})
	assert.Equal(t, expectedError, err)
	assert.NoError(t, dbMock.ExpectationsWereMet())
}

func TestDB_withNamedLock_ReleasesLockOnSuccess(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db, dbMock := NewDBMock()
	defer func() { _ = db.Close() }()

	lockName := someName
	timeout := 1234 * time.Millisecond
	expectedTimeout := int(timeout.Round(time.Second).Seconds())

	dbMock.ExpectQuery("^"+regexp.QuoteMeta("SELECT GET_LOCK(?, ?)")+"$").
		WithArgs(lockName, expectedTimeout).
		WillReturnRows(sqlmock.NewRows([]string{"GET_LOCK(?, ?)"}).AddRow(int64(1)))
	dbMock.ExpectQuery("^" + regexp.QuoteMeta("SELECT RELEASE_LOCK(?)") + "$").
		WithArgs(lockName).
		WillReturnRows(sqlmock.NewRows([]string{"RELEASE_LOCK(?)"}).AddRow(int64(1)))

	err := db.withNamedLock(lockName, timeout, func(*DB) error {
		return nil
	})
	assert.NoError(t, err)
	assert.NoError(t, dbMock.ExpectationsWereMet())
}

func TestDB_withNamedLock_ReleasesLockOnError(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db, dbMock := NewDBMock()
	defer func() { _ = db.Close() }()

	lockName := someName
	timeout := 1234 * time.Millisecond
	expectedTimeout := int(timeout.Round(time.Second).Seconds())
	expectedError := errors.New("some error")

	dbMock.ExpectQuery("^"+regexp.QuoteMeta("SELECT GET_LOCK(?, ?)")+"$").
		WithArgs(lockName, expectedTimeout).
		WillReturnRows(sqlmock.NewRows([]string{"GET_LOCK(?, ?)"}).AddRow(int64(1)))
	dbMock.ExpectQuery("^" + regexp.QuoteMeta("SELECT RELEASE_LOCK(?)") + "$").
		WithArgs(lockName).
		WillReturnRows(sqlmock.NewRows([]string{"RELEASE_LOCK(?)"}).AddRow(int64(1)))

	err := db.withNamedLock(lockName, timeout, func(*DB) error {
		return expectedError
	})
	assert.Equal(t, expectedError, err)
	assert.NoError(t, dbMock.ExpectationsWereMet())
}

func TestDB_withNamedLock_ReleasesLockOnPanic(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db, dbMock := NewDBMock()
	defer func() { _ = db.Close() }()

	lockName := someName
	timeout := 1234 * time.Millisecond
	expectedTimeout := int(timeout.Round(time.Second).Seconds())
	expectedError := errors.New("some error")

	dbMock.ExpectQuery("^"+regexp.QuoteMeta("SELECT GET_LOCK(?, ?)")+"$").
		WithArgs(lockName, expectedTimeout).
		WillReturnRows(sqlmock.NewRows([]string{"GET_LOCK(?, ?)"}).AddRow(int64(1)))
	dbMock.ExpectQuery("^" + regexp.QuoteMeta("SELECT RELEASE_LOCK(?)") + "$").
		WithArgs(lockName).
		WillReturnRows(sqlmock.NewRows([]string{"RELEASE_LOCK(?)"}).AddRow(int64(1)))

	assert.PanicsWithValue(t, expectedError, func() {
		_ = db.withNamedLock(lockName, timeout, func(*DB) error {
			panic(expectedError)
		})
	})
	assert.NoError(t, dbMock.ExpectationsWereMet())
}

func TestDB_withNamedLock_HandlesReleaseError(t *testing.T) {
	for _, test := range []struct {
		name                 string
		setReleaseResultFunc func(*sqlmock.ExpectedQuery)
		logsShouldContain    string
	}{
		{
			name: "release error",
			setReleaseResultFunc: func(query *sqlmock.ExpectedQuery) {
				query.WillReturnError(errors.New("some error"))
			},
			logsShouldContain: "some error",
		},
		{
			name: "empty release result",
			setReleaseResultFunc: func(query *sqlmock.ExpectedQuery) {
				query.WillReturnRows(sqlmock.NewRows([]string{"RELEASE_LOCK(?)"}))
			},
		},
		{
			name: "wrong release result",
			setReleaseResultFunc: func(query *sqlmock.ExpectedQuery) {
				query.WillReturnRows(sqlmock.NewRows([]string{"RELEASE_LOCK(?)"}).AddRow(int64(0)))
			},
		},
	} {
		test := test
		t.Run(test.name, func(t *testing.T) {
			testoutput.SuppressIfPasses(t)

			db, dbMock := NewDBMock()
			defer func() { _ = db.Close() }()
			logHook, logRestoreFunc := logging.MockSharedLoggerHook()
			defer logRestoreFunc()

			lockName := someName
			timeout := 1234 * time.Millisecond
			expectedTimeout := int(timeout.Round(time.Second).Seconds())

			dbMock.ExpectBegin()
			dbMock.ExpectQuery("^"+regexp.QuoteMeta("SELECT GET_LOCK(?, ?)")+"$").
				WithArgs(lockName, expectedTimeout).
				WillReturnRows(sqlmock.NewRows([]string{"GET_LOCK(?, ?)"}).AddRow(int64(1)))
			test.setReleaseResultFunc(
				dbMock.ExpectQuery("^" + regexp.QuoteMeta("SELECT RELEASE_LOCK(?)") + "$").
					WithArgs(lockName))
			dbMock.ExpectClose()
			dbMock.ExpectCommit()

			err := db.inTransaction(func(db *DB) error {
				assert.NoError(t, db.withNamedLock(lockName, timeout, func(*DB) error {
					return nil
				}))
				return nil
			})

			assert.NoError(t, err)
			assert.NoError(t, dbMock.ExpectationsWereMet())

			logs := (&loggingtest.Hook{Hook: logHook}).GetAllLogs()
			if test.logsShouldContain != "" {
				assert.Contains(t, logs, test.logsShouldContain)
			}
			assert.Contains(t, logs, "failed to release the lock \"some name\", closing the DB connection to release the lock")
		})
	}
}

func TestDB_WithExclusiveWriteLock(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT \\* FROM `myTable` FOR UPDATE").
		WillReturnRows(mock.NewRows([]string{"1"}).AddRow(1))
	mock.ExpectCommit()

	db = db.Table("myTable")
	err := db.inTransaction(func(db *DB) error {
		newDB := db.WithExclusiveWriteLock()
		assert.NotEqual(t, newDB, db)
		assert.NoError(t, newDB.Error())
		var result []interface{}
		assert.NoError(t, newDB.Scan(&result).Error())
		return nil
	})

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDB_WithExclusiveWriteLock_PanicsWhenNotInTransaction(t *testing.T) {
	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	assert.PanicsWithValue(t, ErrNoTransaction, func() { db.WithExclusiveWriteLock() })
	assert.NoError(t, mock.ExpectationsWereMet())
}

var retryOnDuplicatePrimaryKeyErrorExpectedQueryRegexp = "^" + regexp.QuoteMeta("INSERT INTO users (id) VALUES (?)") + "$"

func TestDB_retryOnDuplicatePrimaryKeyError(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	for i := 1; i < keyTriesCount; i++ {
		mock.ExpectExec(retryOnDuplicatePrimaryKeyErrorExpectedQueryRegexp).WithArgs(i).
			WillReturnError(&mysql.MySQLError{Number: 1062, Message: "Duplicate entry '" + strconv.Itoa(i) + "' for key 'users.PRIMARY'"})
	}
	mock.ExpectExec(retryOnDuplicatePrimaryKeyErrorExpectedQueryRegexp).WithArgs(keyTriesCount).
		WillReturnResult(sqlmock.NewResult(keyTriesCount, 1))

	retryCount := 0
	err := db.retryOnDuplicatePrimaryKeyError("users", func(db *DB) error {
		retryCount++
		return db.Exec("INSERT INTO users (id) VALUES (?)", retryCount).Error()
	})
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDB_retryOnDuplicatePrimaryKeyError_ErrorsWhenLimitExceeded(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	for i := 1; i < keyTriesCount+1; i++ {
		mock.ExpectExec(retryOnDuplicatePrimaryKeyErrorExpectedQueryRegexp).WithArgs(i).
			WillReturnError(&mysql.MySQLError{Number: 1062, Message: "Duplicate entry '" + strconv.Itoa(i) + "' for key 'users.PRIMARY'"})
	}

	retryCount := 0
	err := db.retryOnDuplicatePrimaryKeyError("users", func(db *DB) error {
		retryCount++
		return db.Exec("INSERT INTO users (id) VALUES (?)", retryCount).Error()
	})
	assert.Equal(t, errors.New("cannot generate a new id"), err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDB_retryOnDuplicatePrimaryKeyError_ReturnsOtherErrors(t *testing.T) {
	tests := []struct {
		name          string
		expectedError *mysql.MySQLError
	}{
		{
			name:          "non-primary key duplicate",
			expectedError: &mysql.MySQLError{Number: 1062, Message: "Duplicate entry '' for key 'users.name'"},
		},
		{
			name:          "error code != 1062",
			expectedError: &mysql.MySQLError{Number: 1063, Message: "Duplicate entry '1' for key 'users.PRIMARY'"},
		},
	}
	for _, testCase := range tests {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			testoutput.SuppressIfPasses(t)

			db, mock := NewDBMock()
			defer func() { _ = db.Close() }()

			mock.ExpectExec(retryOnDuplicatePrimaryKeyErrorExpectedQueryRegexp).WithArgs(1).
				WillReturnError(testCase.expectedError)

			err := db.retryOnDuplicatePrimaryKeyError("users", func(db *DB) error {
				return db.Exec("INSERT INTO users (id) VALUES (?)", 1).Error()
			})
			assert.Equal(t, testCase.expectedError, err)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestDB_retryOnDuplicateKeyError_LogsRetryableErrorsAsInfo(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db, mock := NewDBMockWithLogConfig(LogConfig{LogSQLQueries: false, AnalyzeSQLQueries: false}, false)
	defer func() { _ = db.Close() }()
	logHook, logRestoreFunc := logging.MockSharedLoggerHook()
	defer logRestoreFunc()

	mock.ExpectExec(retryOnDuplicatePrimaryKeyErrorExpectedQueryRegexp).WithArgs(1).
		WillReturnError(&mysql.MySQLError{Number: 1062, Message: "Duplicate entry 'login-1' for key 'users.login'"})
	mock.ExpectExec(retryOnDuplicatePrimaryKeyErrorExpectedQueryRegexp).WithArgs(1).
		WillReturnError(&mysql.MySQLError{Number: 1062, Message: "Duplicate entry '1' for key 'users.PRIMARY'"})
	mock.ExpectExec(retryOnDuplicatePrimaryKeyErrorExpectedQueryRegexp).WithArgs(1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := db.retryOnDuplicateKeyError("users", "login", "user's login", func(db *DB) error {
		return db.retryOnDuplicatePrimaryKeyError("users", func(db *DB) error {
			return db.Exec("INSERT INTO users (id) VALUES (?)", 1).Error()
		})
	})
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())

	entries := logHook.Entries
	require.Len(t, entries, 2)

	assert.Equal(t, "info", entries[0].Level.String())
	assert.Equal(t, "Error 1062: Duplicate entry 'login-1' for key 'users.login'", entries[0].Message)
	assert.Equal(t, "info", entries[1].Level.String())
	assert.Equal(t, "Error 1062: Duplicate entry '1' for key 'users.PRIMARY'", entries[1].Message)
}

func TestDefault(t *testing.T) {
	assert.Equal(t, gorm.Expr("DEFAULT"), Default())
}

func Test_EscapeLikeString(t *testing.T) {
	type args struct {
		s               string
		escapeCharacter byte
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "all characters",
			args: args{s: "|some _string_ 100%|", escapeCharacter: '|'},
			want: "||some |_string|_ 100|%||",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got := EscapeLikeString(tt.args.s, tt.args.escapeCharacter)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDB_InsertIgnoreMaps_WithEmptyArray(t *testing.T) {
	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()
	var dataRows []map[string]interface{}
	assert.NoError(t, db.InsertIgnoreMaps("myTable", dataRows))
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDB_InsertIgnoreMaps(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	dataRows := []map[string]interface{}{
		{"a": int64(1), "b": "value"},
		{"c": int64(2), "d": "another value"},
	}

	expectedError := errors.New("some error")
	mock.ExpectExec(regexp.QuoteMeta("INSERT IGNORE INTO `myTable` (`a`, `b`) VALUES (?, ?), (?, ?)")).
		WithArgs(int64(1), "value", nil, nil).
		WillReturnError(expectedError)

	assert.Equal(t, expectedError, db.InsertIgnoreMaps("myTable", dataRows))
	assert.NoError(t, mock.ExpectationsWereMet())
}

func Test_gormDBBeginTxReplacement_ErrsWhenDBIsNotSQLDBWrapper(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	sqlDB := db.db.CommonDB().(*sqlDBWrapper).sqlDB
	gormDB, err := gorm.Open("", sqlDB)
	require.NoError(t, err)
	gormDB.LogMode(false)

	newGormDB := gormDBBeginTxReplacement(context.Background(), gormDB, &sql.TxOptions{})
	assert.Equal(t, gorm.ErrCantStartTransaction, newGormDB.Error)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDB_GetContext(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	mock.ExpectBegin()
	mock.ExpectCommit()

	expectedContext := context.WithValue(context.Background(), testContextKey("key"), "value")
	db = cloneDBWithNewContext(expectedContext, db)

	assert.Equal(t, expectedContext, db.GetContext())

	assert.NoError(t, db.inTransaction(func(db *DB) error {
		assert.Equal(t, expectedContext, db.GetContext())
		return nil
	}))
}
