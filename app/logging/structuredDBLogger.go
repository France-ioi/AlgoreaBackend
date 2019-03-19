package logging

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// StructuredDBLogger is a database structured logger
type StructuredDBLogger struct {
	logger *logrus.Logger
}

// NewStructuredDBLogger created a database structured logger
func NewStructuredDBLogger(logger *logrus.Logger) *StructuredDBLogger {
	return &StructuredDBLogger{logger}
}

var (
	sqlRegexp                = regexp.MustCompile(`\?`)
	numericPlaceHolderRegexp = regexp.MustCompile(`\$\d+`)
)

// Print defines how StructuredDBLogger print log entries.
// values: 0: level, 1: source file, 2: duration in ns, 3: query, 4: slice of parameters, 5: rows affected or returned
func (l *StructuredDBLogger) Print(values ...interface{}) {
	level := values[0]
	logger := l.logger.WithField("type", "db")

	if level == "sql" {
		duration := float64(values[2].(time.Duration).Nanoseconds()) / float64(time.Second.Nanoseconds()) // to seconds
		sql := fillSQLPlaceholders(values[3].(string), values[4].([]interface{}))
		logger.WithFields(map[string]interface{}{
			"duration": duration,
			"ts":       time.Now().Format("2006-01-02 15:04:05"),
			"rows":     values[5].(int64),
		}).Println(strings.TrimSpace(sql))
	} else if level == "rawsql" {
		/*
		   values[0] - level (rawsql)
		   values[1] - ctx
		   values[2] - message
		   values[3] - values map
		*/
		valuesMap := values[3].(map[string]interface{})
		if valuesMap["duration"] != nil {
			valuesMap["duration"] = float64(valuesMap["duration"].(time.Duration).Nanoseconds()) / float64(time.Second.Nanoseconds()) // to seconds
		}
		valuesMap["ts"] = time.Now().Format("2006-01-02 15:04:05")
		logger.WithFields(valuesMap).Println(values[2])
	} else { // level is not "sql", so typically errors
		logger.Println(values[2:]...)
	}

}

func fillSQLPlaceholders(query string, values []interface{}) string {
	var sql string
	var formattedValues []string
	for _, value := range values {
		indirectValue := reflect.Indirect(reflect.ValueOf(value))
		var formattedValue string
		if indirectValue.IsValid() {
			value = indirectValue.Interface()
			switch typedValue := value.(type) {
			case time.Time:
				formattedValue = fmt.Sprintf("'%v'", typedValue.Format("2006-01-02 15:04:05"))
			case []byte, string:
				formattedValue = fmt.Sprintf("%q", typedValue)
			default:
				formattedValue = fmt.Sprintf("%v", typedValue)
			}
		} else {
			formattedValue = "NULL"
		}
		formattedValues = append(formattedValues, formattedValue)
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
