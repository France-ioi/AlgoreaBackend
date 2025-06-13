package database

import (
	"context"
	"database/sql/driver"
	"regexp"
	"strconv"
	"time"

	"github.com/luna-duclos/instrumentedsql"

	log "github.com/France-ioi/AlgoreaBackend/v2/app/logging"
)

var rawArgsRegexp = regexp.MustCompile(`^\[(<nil>|[\w.]+) (.+?)\](?:(?:, \[(?:<nil>|[\w.]+) )|$)`)

// NewRawDBLogger returns a logger for raw database actions.
func NewRawDBLogger() instrumentedsql.Logger {
	return instrumentedsql.LoggerFunc(func(ctx context.Context, msg string, keyvals ...interface{}) {
		valuesMap := prepareRawDBLoggerValuesMap(keyvals)
		if valuesMap["err"] != nil && valuesMap["err"] == driver.ErrSkip { // duplicated message
			return
		}

		if valuesMap["duration"] != nil {
			valuesMap["duration"] = valuesMap["duration"].(time.Duration).String()
		}

		valuesMap["type"] = "db"

		log.SharedLogger.WithContext(ctx).WithFields(valuesMap).Info(msg)
	})
}

func prepareRawDBLoggerValuesMap(keyvals []interface{}) map[string]interface{} {
	valuesMap := make(map[string]interface{}, len(keyvals)/2) //nolint:gomnd // keyvals are key-value pairs
	for index := 0; index < len(keyvals); index += 2 {
		valuesMap[keyvals[index].(string)] = keyvals[index+1]
	}
	if valuesMap["query"] != nil && valuesMap["args"] != nil {
		argsString := valuesMap["args"].(string)
		argsString = argsString[1 : len(argsString)-1]
		var argsValues []interface{}
		for argsString != "" {
			indices := rawArgsRegexp.FindStringSubmatchIndex(argsString)
			typeStr := argsString[indices[2]:indices[3]]
			valueCopy := make([]byte, indices[5]-indices[4])
			copy(valueCopy, argsString[indices[4]:indices[5]])
			value := string(valueCopy)
			convertedValue := convertRawSQLArgValue(value, typeStr)
			argsValues = append(argsValues, convertedValue)
			if indices[5]+3 >= len(argsString) {
				break
			}
			argsString = argsString[indices[5]+3:]
		}
		valuesMap["query"] = fillSQLPlaceholders(valuesMap["query"].(string), argsValues)
		delete(valuesMap, "args")
	}
	return valuesMap
}

func convertRawSQLArgValue(value, typeStr string) interface{} {
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
