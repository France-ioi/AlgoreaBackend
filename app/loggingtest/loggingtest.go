// +build !prod

// Package loggingtest provides functions to test the content of the logs.
package loggingtest

import (
	"strings"

	"github.com/sirupsen/logrus"            //nolint
	"github.com/sirupsen/logrus/hooks/test" //nolint
)

// Hook is a hook designed for dealing with logs in test scenarios. It wraps logrus/hooks/test.Hook.
type Hook struct {
	*test.Hook
}

const newLine = "\n"

// GetAllLogs returns all the logs collected by the hook as a string.
func (hook *Hook) GetAllLogs() string {
	logs := ""

	for _, entry := range hook.AllEntries() {
		if len(logs) > 0 {
			logs += newLine
		}

		logs += strings.TrimSpace(entry.Message)
	}

	return logs
}

// GetAllStructuredLogs returns all the structured logs collected by the hook as a string.
func (hook *Hook) GetAllStructuredLogs() string {
	logs := ""

	for _, entry := range hook.AllEntries() {
		if len(logs) > 0 {
			logs += newLine
		}
		logString, err := entry.String()
		if err != nil {
			logs += strings.TrimSpace(err.Error())
		} else {
			logs += strings.TrimSpace(logString)
		}
	}

	return logs
}

// NewNullLogger creates a discarding logger and installs the test hook.
func NewNullLogger() (*logrus.Logger, *Hook) {
	logger, hook := test.NewNullLogger()
	return logger, &Hook{hook}
}
