package logging

import (
	"context"
	"regexp"
	"strconv"

	"github.com/jinzhu/gorm"
	"github.com/luna-duclos/instrumentedsql"

	"github.com/France-ioi/AlgoreaBackend/app/config"
)

var rawArgsRegexp = regexp.MustCompile(`^{\[(\w+) (.+?)\]((?:, \[\w+ .+?\])*)}$`)

// NewRawDBLogger returns a logger for raw database actions and the `logmode`, according to the config
func NewRawDBLogger() (instrumentedsql.Logger, bool) {
	var (
		err     error
		conf    *config.Root
		logger  DBLogger
		logMode bool
	)

	if conf, err = config.Load(); err != nil {
		// if cannot parse config, log on error to stdout
		logger = gorm.Logger{LogWriter: Logger}
	} else {
		logger, logMode = loggerFromConfig(conf.Logging, Logger)
	}

	sqlLogger := instrumentedsql.LoggerFunc(func(ctx context.Context, msg string, keyvals ...interface{}) {
		if ctx == nil && msg == "sql-stmt-exec" { // duplicated message
			return
		}

		valuesMap := prepareValuesMap(keyvals)
		args := make([]interface{}, 0, 4)

		args = append(args, "rawsql")
		args = append(args, ctx)
		args = append(args, msg)
		args = append(args, valuesMap)
		logger.Print(args...)
	})
	return sqlLogger, logMode
}

func prepareValuesMap(keyvals []interface{}) map[string]interface{} {
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
			if typeStr == "string" {
				if unquoted, err := strconv.Unquote(value); err == nil {
					value = unquoted
				}
			}
			nextStr := subMatches[3]
			if nextStr != "" {
				argsString = "{" + subMatches[3][2:] + "}"
			} else {
				argsString = "{}"
			}
			argsValues = append(argsValues, value)
		}
		valuesMap["query"] = fillSQLPlaceholders(valuesMap["query"].(string), argsValues)
		delete(valuesMap, "args")
	}
	return valuesMap
}
