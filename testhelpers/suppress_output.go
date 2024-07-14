package testhelpers

import (
	"testing"

	"github.com/zenovich/flowmingo"
)

// SuppressOutputIfPasses immediately mutes output (to both STDOUT and STDERR) of the test
// so that the output will only be shown if the test fails.
// The cleanup part is scheduled to be run automatically,
// no need to clean up manually.
//
// Note: This function will only work when the test is not run in parallel within the test binary,
// i.e. when the test (or its parents) haven't called t.Parallel(),
// otherwise it will panic.
// After the output is suppressed, t.Parallel() for the test will panic.
//
// Note: This function does nothing if the test is run in verbose mode.
func SuppressOutputIfPasses(t *testing.T) {
	if testing.Verbose() {
		return
	}

	// panics if t.Parallel() has been called before, prevents future calls to t.Parallel()
	t.Setenv("OUTPUT_CAPTURED", "1")

	restoreFunc := flowmingo.CaptureStdoutAndStderr()
	t.Cleanup(func() {
		restoreFunc(t.Failed())
	})
}
