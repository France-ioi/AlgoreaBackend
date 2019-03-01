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

		duration := float64(values[2].(time.Duration).Nanoseconds()/1e4) / 100000.0 // to seconds

		var sql string
		var formattedValues []string
		for _, value := range values[4].([]interface{}) {
			indirectValue := reflect.Indirect(reflect.ValueOf(value))
			if indirectValue.IsValid() {
				value = indirectValue.Interface()
				if t, ok := value.(time.Time); ok {
					formattedValues = append(formattedValues, fmt.Sprintf("'%v'", t.Format("2006-01-02 15:04:05")))
				} else if b, ok := value.([]byte); ok {
					formattedValues = append(formattedValues, fmt.Sprintf("'%v'", string(b)))
				} else {
					formattedValues = append(formattedValues, fmt.Sprintf("'%v'", value))
				}
			} else {
				formattedValues = append(formattedValues, "NULL")
			}
		}

		// differentiate between $n placeholders or else treat like ?
		if numericPlaceHolderRegexp.MatchString(values[3].(string)) {
			sql = values[3].(string)
			for index, value := range formattedValues {
				placeholder := fmt.Sprintf(`\$%d([^\d]|$)`, index+1)
				sql = regexp.MustCompile(placeholder).ReplaceAllString(sql, value+"$1")
			}
		} else {
			formattedValuesLength := len(formattedValues)
			for index, value := range sqlRegexp.Split(values[3].(string), -1) {
				sql += value
				if index < formattedValuesLength {
					sql += formattedValues[index]
				}
			}
		}
		logger.WithFields(map[string]interface{}{
			"duration": duration,
			"ts":       time.Now().Format("2006-01-02 15:04:05"),
			"rows":     values[5].(int64),
		}).Println(strings.TrimSpace(sql))

	} else { // level is not "sql", so typically errors
		logger.Println(values[2:]...)
	}

}
