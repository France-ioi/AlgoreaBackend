package database

import (
	"fmt"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/jinzhu/gorm"

	log "github.com/France-ioi/AlgoreaBackend/v2/app/logging"
)

func logSQLQuery(logger log.DBLogger, duration time.Duration, sql string, args []interface{}, rowsAffected *int64) {
	printArgs := []interface{}{
		"sql", fileWithLineNum(), duration, sql, args,
	}
	if rowsAffected != nil {
		printArgs = append(printArgs, *rowsAffected)
	}
	logger.Print(printArgs...)
}

var _ queryRowWithoutLogging = &sqlDBWrapper{}

var explainableStatementRegexp = regexp.MustCompile(`(?i)^\s*(SELECT|DELETE|INSERT|REPLACE|UPDATE|TABLE)\s`)

var emptyFunc = func() {}

func getSQLExecutionPlanLoggingFunc(db queryRowWithoutLogging, logConfig *LogConfig, query string, args ...interface{}) func() {
	if !logConfig.LogSQLQueries || !logConfig.AnalyzeSQLQueries || !explainableStatementRegexp.MatchString(query) {
		return emptyFunc
	}

	var plan string
	planStartTime := gorm.NowFunc()
	if err := db.queryRowWithoutLogging("EXPLAIN ANALYZE "+query, args...).Scan(&plan); err != nil {
		return func() {
			logConfig.Logger.Print("error", fileWithLineNum(), fmt.Sprintf("Failed to get an execution plan for a SQL query: %v", err))
		}
	}

	planDuration := gorm.NowFunc().Sub(planStartTime)
	return func() {
		logSQLQuery(logConfig.Logger, planDuration, "query execution plan:\n"+plan, nil, nil)
	}
}

func getSQLQueryLoggingFunc(
	rowsAffectedFunc func() *int64, err *error, startTime time.Time, query string, args ...interface{},
) func(logConfig *LogConfig) {
	return func(logConfig *LogConfig) {
		if *err != nil {
			// Log the error even if we don't log the query, but do it after logging the query
			defer logConfig.Logger.Print("error", fileWithLineNum(), *err)
		}
		if !logConfig.LogSQLQueries {
			return
		}
		var rowsAffected *int64
		if *err == nil && rowsAffectedFunc != nil {
			rowsAffected = rowsAffectedFunc()
		}
		logSQLQuery(logConfig.Logger, gorm.NowFunc().Sub(startTime), query, args, rowsAffected)
	}
}

func fileWithLineNum() string {
	for i := 4; ; i++ {
		_, file, line, ok := runtime.Caller(i)
		if !ok {
			return ""
		}
		if strings.Contains(file, "jinzhu/gorm") ||
			strings.HasSuffix(file, "/app/database/db.go") ||
			strings.HasSuffix(file, "/app/database/data_store.go") ||
			strings.HasSuffix(file, "/app/database/sql_conn_wrapper.go") ||
			strings.HasSuffix(file, "/app/database/sql_db_wrapper.go") ||
			strings.HasSuffix(file, "/app/database/sql_stmt_wrapper.go") ||
			strings.HasSuffix(file, "/app/database/sql_tx_wrapper.go") {
			continue
		}
		return fmt.Sprintf("%v:%v", file, line)
	}
}
