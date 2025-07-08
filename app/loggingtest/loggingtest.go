//go:build !prod

// Package loggingtest provides functions to test the content of the logs.
package loggingtest

import (
	"strings"

	"github.com/sirupsen/logrus/hooks/test" //nolint

	"github.com/France-ioi/AlgoreaBackend/v2/app/logging"
)

// Hook is a hook designed for dealing with logs in test scenarios. It wraps logrus/hooks/test.Hook.
type Hook struct {
	*test.Hook
}

const newLine = "\n"

// GetAllLogs returns all the log messages (entry.Message) collected by the hook as a string.
func (hook *Hook) GetAllLogs() string {
	logs := ""

	for _, entry := range hook.AllEntries() {
		if logs != "" {
			logs += newLine
		}

		logs += strings.TrimSpace(entry.Message)
	}

	return logs
}

// GetAllStructuredLogs returns all the structured logs collected by the hook as a string.
func (hook *Hook) GetAllStructuredLogs() string {
	logs := ""
	formatter := logging.NewTextFormatterForTests()

	for _, entry := range hook.AllEntries() {
		if logs != "" {
			logs += newLine
		}
		logBytes, err := formatter.Format(entry)
		if err != nil {
			logs += strings.TrimSpace(err.Error())
		} else {
			logs += strings.TrimSpace(string(logBytes))
		}
	}

	return logs
}
