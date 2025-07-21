package logging

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewContextWithNewMockLogger(t *testing.T) {
	ctx, logger, hook := NewContextWithNewMockLogger()
	require.NotNil(t, logger)
	assert.Equal(t, logger, LoggerFromContext(ctx))
	logger.WithContext(ctx).Info("Hello World")
	require.NotNil(t, hook)
	require.NotNil(t, hook.LastEntry())
	assert.Equal(t, "Hello World", hook.LastEntry().Message)
}
