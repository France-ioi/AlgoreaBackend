package logging

import (
	"github.com/sirupsen/logrus/hooks/test"
)

// NewMockLogger creates a null/mock logger and return the logger and the hook
func NewMockLogger() (*Logger, *test.Hook) {
	l, h := test.NewNullLogger()
	return &Logger{l, nil}, h
}

// MockSharedLoggerHook set a null/mock logger as shared and return the hook and the function restoring the initial logger
func MockSharedLoggerHook() (*test.Hook, func()) {
	var h *test.Hook
	previousShared := SharedLogger
	SharedLogger, h = NewMockLogger()
	return h, func() {
		SharedLogger = previousShared
	}
}
