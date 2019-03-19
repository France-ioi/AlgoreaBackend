package logging

import (
	"context"
	"regexp"
	"strconv"

	"github.com/luna-duclos/instrumentedsql"
)

var rawArgsRegexp = regexp.MustCompile(`^{\[(<nil>|[\w.]+) (.+?)\]((?:, \[(?:<nil>|[\w.]+) .+?\])*)}$`)

// NewRawDBLogger returns a logger for raw database actions using an existing dblogger and logMode setting
func NewRawDBLogger(logger DBLogger, logMode bool) instrumentedsql.Logger {
	return instrumentedsql.LoggerFunc(func(ctx context.Context, msg string, keyvals ...interface{}) {
		if !logMode {
			return
		}

		if ctx == nil && msg == "sql-stmt-exec" { // duplicated message
			return
		}

		valuesMap := prepareRawDBLoggerValuesMap(keyvals)
		args := make([]interface{}, 0, 4)

		args = append(args, "rawsql")
		args = append(args, ctx)
		args = append(args, msg)
		args = append(args, valuesMap)
		logger.Print(args...)
	})
}

func prepareRawDBLoggerValuesMap(keyvals []interface{}) map[string]interface{} {
	valuesMap := make(map[string]interface{}, len(keyvals)/2)
	for index := 0; index < len(keyvals); index += 2 {
		valuesMap[keyvals[index].(string)] = keyvals[index+1]
	}
	if valuesMap["query"] != nil && valuesMap["args"] != nil {
		argsString := valuesMap["args"].(string)
		var argsValues []interface{}
		for argsString != "{}" {
			subMatches := rawArgsRegexp.FindStringSubmatch(argsString)
			typeStr := subMatches[1]
			valueCopy := make([]byte, len(subMatches[2]))
			copy(valueCopy, subMatches[2])
			value := string(valueCopy)
			convertedValue := convertRawSQLArgValue(value, typeStr)
			nextStr := subMatches[3]
			if nextStr != "" {
				argsString = "{" + subMatches[3][2:] + "}"
			} else {
				argsString = "{}"
			}
			argsValues = append(argsValues, convertedValue)
		}
		valuesMap["query"] = fillSQLPlaceholders(valuesMap["query"].(string), argsValues)
		delete(valuesMap, "args")
	}
	return valuesMap
}

func convertRawSQLArgValue(value string, typeStr string) interface{} {
	var convertedValue interface{} = value
	switch typeStr {
	case "string":
		if unquoted, err := strconv.Unquote(value); err == nil {
			convertedValue = unquoted
		}
	case "int64", "int32", "int16", "int8", "int":
		if converted, err := strconv.ParseInt(value, 10, 64); err == nil {
			convertedValue = converted
		}
	case "float64", "float32":
		if converted, err := strconv.ParseFloat(value, 64); err == nil {
			convertedValue = converted
		}
	case "<nil>":
		convertedValue = nil
	}
	return convertedValue
}
