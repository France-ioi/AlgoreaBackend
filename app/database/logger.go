package database

// Part of the following code are inspired from the GORM logger, under MIT Licence:
//
// The MIT License (MIT)
//
// Copyright (c) 2013-NOW  Jinzhu <wosmvp@gmail.com>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

import (
	"fmt"
	"log"
	"os"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/sirupsen/logrus"
)

// StdOutColoredLogger is the default logger to be used to output colored logs to the console
var StdOutColoredLogger = gorm.Logger{LogWriter: log.New(os.Stdout, "\r\n", 0)}

// DBLogger is the logger interface for the DB logs
type DBLogger interface {
	Print(v ...interface{})
}

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
