package logging

import (
	"context"

	"github.com/sirupsen/logrus/hooks/test"
)

// NewMockLogger creates a null/mock logger and returns the logger and the hook.
func NewMockLogger() (*Logger, *test.Hook) {
	l, h := test.NewNullLogger()
	return &Logger{l, nil}, h
}

// NewContextWithNewMockLogger creates a new context based on context.Background() with a new mock logger.
func NewContextWithNewMockLogger() (context.Context, *Logger, *test.Hook) {
	logger, hook := NewMockLogger()
	return ContextWithLogger(context.Background(), logger), logger, hook
}
