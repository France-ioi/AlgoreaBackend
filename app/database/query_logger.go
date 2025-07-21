package database

import (
	"context"
	"database/sql/driver"
	"fmt"
	"reflect"
	"regexp"
	"runtime"
	"strings"
	"time"
	"unicode"

	"github.com/jinzhu/gorm"

	log "github.com/France-ioi/AlgoreaBackend/v2/app/logging"
)

func logSQLQuery(ctx context.Context, duration time.Duration, sql string, args []interface{}, rowsAffected *int64) {
	fields := map[string]interface{}{
		"duration": duration.String(),
		"fileline": fileWithLineNum(),
		"type":     "db",
	}
	if rowsAffected != nil {
		fields["rows"] = *rowsAffected
	}
	log.EntryFromContext(ctx).WithFields(fields).Info(strings.TrimSpace(fillSQLPlaceholders(sql, args)))
}

func logDBError(ctx context.Context, logConfig *LogConfig, err error) {
	entry := log.EntryFromContext(ctx).WithFields(
		map[string]interface{}{"type": "db", "fileline": fileWithLineNum()})
	if logConfig.LogRetryableErrorsAsInfo && isRetryableError(err) {
		entry.Info(err)
	} else if f, ok := ctx.Value(logErrorAsInfoFuncContextKey).(func(error) bool); ok && f(err) {
		entry.Info(err)
	} else {
		entry.Error(err)
	}
}

var _ queryRowWithoutLogging = &sqlDBWrapper{}

var explainableStatementRegexp = regexp.MustCompile(`(?i)^\s*(SELECT|DELETE|INSERT|REPLACE|UPDATE|TABLE)\s`)

var emptyFunc = func() {}

func getSQLExecutionPlanLoggingFunc(
	ctx context.Context, db queryRowWithoutLogging, logConfig *LogConfig, query string, args ...interface{},
) func() {
	if !logConfig.LogSQLQueries || !logConfig.AnalyzeSQLQueries || !explainableStatementRegexp.MatchString(query) {
		return emptyFunc
	}

	var plan string
	planStartTime := gorm.NowFunc()
	if err := db.queryRowWithoutLogging("EXPLAIN ANALYZE "+query, args...).Scan(&plan); err != nil {
		return func() {
			log.EntryFromContext(ctx).WithFields(
				map[string]interface{}{"type": "db", "fileline": fileWithLineNum()}).
				Errorf("Failed to get an execution plan for a SQL query: %v", err)
		}
	}

	planDuration := gorm.NowFunc().Sub(planStartTime)
	return func() {
		log.EntryFromContext(ctx).WithFields(
			map[string]interface{}{
				"type":     "db",
				"fileline": fileWithLineNum(),
				"duration": planDuration.String(),
			}).
			Infof("query execution plan:\n%s\n", plan)
	}
}

func getSQLQueryLoggingFunc(
	ctx context.Context, rowsAffectedFunc func() *int64,
	err *error, //nolint:gocritic // we need the pointer as the constructor is called before the error is set
	startTime time.Time, query string, args ...interface{},
) func(logConfig *LogConfig) {
	return func(logConfig *LogConfig) {
		if *err != nil {
			// Log the error even if we don't log the query, but do it after logging the query
			defer logDBError(ctx, logConfig, *err)
		}
		if !logConfig.LogSQLQueries {
			return
		}
		var rowsAffected *int64
		if *err == nil && rowsAffectedFunc != nil {
			rowsAffected = rowsAffectedFunc()
		}
		logSQLQuery(ctx, gorm.NowFunc().Sub(startTime), query, args, rowsAffected)
	}
}

var appDatabaseWrapperRegexp = regexp.MustCompile(`/app/database/sql_.*_wrapper.go$`)

func fileWithLineNum() string {
	for i := 4; ; i++ {
		_, file, line, ok := runtime.Caller(i)
		if !ok {
			return ""
		}
		if strings.Contains(file, "jinzhu/gorm") ||
			strings.HasSuffix(file, "/app/database/db.go") ||
			strings.HasSuffix(file, "/app/database/data_store.go") ||
			strings.HasSuffix(file, "/src/runtime/panic.go") ||
			appDatabaseWrapperRegexp.MatchString(file) {
			continue
		}
		return fmt.Sprintf("%v:%v", file, line)
	}
}

func isPrintable(s string) bool {
	for _, r := range s {
		if !unicode.IsPrint(r) {
			return false
		}
	}
	return true
}

const nullString = "NULL"

var (
	sqlRegexp                = regexp.MustCompile(`\?`)
	spacesRegexp             = regexp.MustCompile(`\s+`)
	numericPlaceHolderRegexp = regexp.MustCompile(`\$\d+`)
)

func formatValue(value interface{}) string {
	reflValue := reflect.ValueOf(value)
	if !reflValue.IsValid() {
		return nullString
	}

	switch castValue := value.(type) {
	case time.Time:
		return formatTime(castValue)
	case []byte:
		return formatBytes(castValue)
	case driver.Valuer:
		return formatValuer(castValue)
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, bool:
		return fmt.Sprintf("%v", value)
	default:
		// try if &value implements driver.Valuer
		reflPtrToValue := reflect.New(reflValue.Type())
		reflPtrToValue.Elem().Set(reflValue)
		valuePtr := reflPtrToValue.Interface()
		if valuer, ok := valuePtr.(driver.Valuer); ok {
			return formatValuer(valuer)
		}

		if reflValue.Kind() == reflect.Ptr {
			if reflValue.IsNil() {
				return nullString
			}
			return formatValue(reflValue.Elem().Interface())
		}
		return fmt.Sprintf("'%v'", value)
	}
}

func formatValuer(v driver.Valuer) string {
	if value, err := v.Value(); err == nil && value != nil {
		return fmt.Sprintf("'%v'", value)
	}
	return nullString
}

func formatBytes(v []byte) string {
	if str := string(v); isPrintable(str) {
		return fmt.Sprintf("'%v'", str)
	}

	return "'<binary>'"
}

func formatTime(value time.Time) string {
	if value.IsZero() {
		return fmt.Sprintf("'%v'", "0000-00-00 00:00:00")
	}

	// Print fractions for time values
	return fmt.Sprintf("'%v'", value.Format("2006-01-02 15:04:05.999999999"))
}

func fillSQLPlaceholders(query string, values []interface{}) string {
	var sql string
	formattedValues := make([]string, 0, len(values))

	query = strings.TrimSpace(spacesRegexp.ReplaceAllString(query, " "))
	for _, value := range values {
		formattedValues = append(formattedValues, formatValue(value))
	}
	// differentiate between $n placeholders or else treat like ?
	if numericPlaceHolderRegexp.MatchString(query) {
		sql = query
		for index, value := range formattedValues {
			placeholder := fmt.Sprintf(`\$%d([^\d]|$)`, index+1)
			sql = regexp.MustCompile(placeholder).ReplaceAllString(sql, value+"$1")
		}
	} else {
		formattedValuesLength := len(formattedValues)
		for index, value := range sqlRegexp.Split(query, -1) {
			sql += value
			if index < formattedValuesLength {
				sql += formattedValues[index]
			}
		}
	}
	return sql
}
