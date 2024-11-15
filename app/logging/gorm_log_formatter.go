package logging

import (
	"database/sql/driver"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"time"
	"unicode"

	"github.com/jinzhu/gorm"
)

/*
	The code below is based on the original GORM log formatter from https://github.com/jinzhu/gorm under MIT License:

	The MIT License (MIT)

	Copyright (c) 2013-NOW  Jinzhu <wosmvp@gmail.com>

	Permission is hereby granted, free of charge, to any person obtaining a copy
	of this software and associated documentation files (the "Software"), to deal
	in the Software without restriction, including without limitation the rights
	to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
	copies of the Software, and to permit persons to whom the Software is
	furnished to do so, subject to the following conditions:

	The above copyright notice and this permission notice shall be included in
	all copies or substantial portions of the Software.

	THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
	IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
	FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
	AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
	LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
	OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
	THE SOFTWARE.
*/

func isPrintable(s string) bool {
	for _, r := range s {
		if !unicode.IsPrint(r) {
			return false
		}
	}
	return true
}

const (
	nullString = "NULL"
	sqlString  = "sql"
)

func formatValue(value interface{}) string {
	reflValue := reflect.ValueOf(value)
	if !reflValue.IsValid() {
		return nullString
	}

	switch v := value.(type) {
	case time.Time:
		return formatTime(v)
	case []byte:
		return formatBytes(v)
	case driver.Valuer:
		return formatValuer(v)
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

func formatTime(v time.Time) string {
	if v.IsZero() {
		return fmt.Sprintf("'%v'", "0000-00-00 00:00:00")
	}

	// Print fractions for time values
	return fmt.Sprintf("'%v'", v.Format("2006-01-02 15:04:05.999999999"))
}

func formatGormDBLog(values ...interface{}) (messages []interface{}) {
	if len(values) <= 1 {
		return
	}

	var (
		sql             string
		formattedValues []string
		level           = values[0]
		currentTime     = "\n\033[33m[" + gorm.NowFunc().Format(time.DateTime) + "]\033[0m"
		source          = fmt.Sprintf("\033[35m(%v)\033[0m", values[1])
	)

	if len(values) == 2 {
		// remove the line break
		currentTime = currentTime[1:]
		// remove the brackets
		source = fmt.Sprintf("\033[35m%v\033[0m", values[1])

		return []interface{}{currentTime, source}
	}

	messages = []interface{}{source, currentTime}

	if level == sqlString {
		// duration
		messages = append(messages,
			fmt.Sprintf(" \033[36;1m[%.2fms]\033[0m ",
				float64(values[2].(time.Duration).Nanoseconds()/1e4)/100.0))

		// sql
		for _, value := range values[4].([]interface{}) {
			formattedValues = append(formattedValues, formatValue(value))
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

		messages = append(messages, sql)

		if len(values) >= 6 {
			messages = append(messages,
				fmt.Sprintf(" \n\033[36;31m[%v]\033[0m ",
					strconv.FormatInt(values[5].(int64), 10)+" rows affected "),
			)
		}

		return messages
	}

	messages = append(messages, "\033[31;1m")
	messages = append(messages, values[2:]...)
	messages = append(messages, "\033[0m")

	return messages
}

// Override the default GORM log formatter to print values of pointer types correctly.
func init() {
	gorm.LogFormatter = formatGormDBLog
}
