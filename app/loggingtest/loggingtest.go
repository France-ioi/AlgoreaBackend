package loggingtest

import (
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
)

type Hook struct {
	*test.Hook
}

// GetAllLogs returns all the logs collected by the hook as a string
func (hook *Hook) GetAllLogs() string {
	logs := ""
	for _, entry := range hook.AllEntries() {
		if len(logs) > 0 {
			logs += "\n"
		}
		logs = logs + entry.Message
	}
	return logs
}

func (hook *Hook) GetAllStructuredLogs() string {
	logs := ""
	for _, entry := range hook.AllEntries() {
		if len(logs) > 0 {
			logs += "\n"
		}
		logString, _ := entry.String()
		logs = logs + logString
	}
	return logs
}

func NewNullLogger() (*logrus.Logger, *Hook) {
	logger, hook := test.NewNullLogger()
	return logger, &Hook{hook}
}
