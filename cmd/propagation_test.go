package cmd

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPropagationLockTimeout(t *testing.T) {
	assert.Equal(t, propagationCommandLockTimeout, propagationLockTimeout(time.Hour))
	assert.Equal(t, 30*time.Second, propagationLockTimeout(30*time.Second))
	assert.Equal(t, time.Duration(0), propagationLockTimeout(-time.Second))
	assert.Equal(t, time.Duration(0), propagationLockTimeout(0))
}

func TestValidatePropagationMaxDuration(t *testing.T) {
	require.NoError(t, validatePropagationMaxDuration(0))
	require.NoError(t, validatePropagationMaxDuration(propagationShutdownMargin+time.Second))
	require.Error(t, validatePropagationMaxDuration(propagationShutdownMargin))
	require.Error(t, validatePropagationMaxDuration(time.Second))
}
