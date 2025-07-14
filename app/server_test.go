package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"syscall"
	"testing"
	"time"

	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServer_Start(t *testing.T) {
	app, err := New()
	require.NoError(t, err)
	srv, err := NewServer(app)
	require.NoError(t, err)

	// check defaults are applied correctly
	assert.True(t, strings.HasSuffix(srv.Addr, ":8088"))
	assert.Equal(t, time.Duration(60000000000), srv.ReadTimeout)
	assert.Equal(t, time.Duration(60000000000), srv.WriteTimeout)

	doneChannel := srv.Start()
	defer close(doneChannel)

	err = syscall.Kill(syscall.Getpid(), syscall.SIGINT)
	require.NoError(t, err)

	select {
	case err = <-doneChannel:
		require.NoError(t, err)
	case <-time.After(3 * time.Second):
		assert.Fail(t, "Timeout on waiting for server to stop")
	}
}

func TestServer_Start_HandlesListenerError(t *testing.T) {
	app, err := New()
	require.NoError(t, err)
	app.Config.Set("server.port", -1)
	srv, err := NewServer(app)
	require.NoError(t, err)

	doneChannel := srv.Start()
	defer close(doneChannel)

	select {
	case err = <-doneChannel:
		require.EqualError(t, err, "server returned an error: listen tcp: address -1: invalid port")
	case <-time.After(3 * time.Second):
		_ = syscall.Kill(syscall.Getpid(), syscall.SIGINT)
		assert.Fail(t, "Timeout on waiting for server to stop")
	}
}

func TestServer_Start_HandlesKillingAfterListenerError(t *testing.T) {
	app, err := New()
	require.NoError(t, err)
	srv, err := NewServer(app)
	require.NoError(t, err)

	expectedError := errors.New("some error")

	shutdownCalledCh := make(chan struct{})

	monkey.PatchInstanceMethod(
		reflect.TypeOf(&http.Server{}), //nolint:gosec // the instance of http.Server will never be used
		"ListenAndServe", func(_ *http.Server) error {
			err = syscall.Kill(syscall.Getpid(), syscall.SIGINT)
			assert.NoError(t, err)
			select {
			case <-shutdownCalledCh:
			case <-time.After(3 * time.Second):
				assert.Fail(t, "Timeout on waiting for server to call Shutdown()")
			}
			return expectedError
		})
	var shutdownGuard *monkey.PatchGuard
	shutdownGuard = monkey.PatchInstanceMethod(
		reflect.TypeOf(&http.Server{}), //nolint:gosec // the instance of http.Server will never be used
		"Shutdown", func(srv *http.Server, ctx context.Context) error {
			close(shutdownCalledCh)
			shutdownGuard.Unpatch()
			defer shutdownGuard.Restore()
			return srv.Shutdown(ctx)
		})
	defer monkey.UnpatchAll()

	doneChannel := srv.Start()
	defer close(doneChannel)

	select {
	case err = <-doneChannel:
		require.EqualError(t, err, "server returned an error: some error")
	case <-time.After(3 * time.Second):
		_ = syscall.Kill(syscall.Getpid(), syscall.SIGINT)
		assert.Fail(t, "Timeout on waiting for server to stop")
	}
}

func TestServer_Start_CanBeStoppedByShutdown(t *testing.T) {
	app, err := New()
	require.NoError(t, err)
	srv, err := NewServer(app)
	require.NoError(t, err)

	doneChannel := srv.Start()
	defer close(doneChannel)

	_ = srv.Shutdown(context.Background())

	select {
	case err := <-doneChannel:
		require.NoError(t, err)
	case <-time.After(3 * time.Second):
		assert.Fail(t, "Timeout on waiting for server to stop")
	}
}

func TestServer_Start_HandlesShutdownError_OnKilling(t *testing.T) {
	app, err := New()
	require.NoError(t, err)
	srv, err := NewServer(app)
	require.NoError(t, err)

	expectedError := errors.New("some error")
	var patchGuard *monkey.PatchGuard
	patchGuard = monkey.PatchInstanceMethod(
		reflect.TypeOf(&http.Server{}), //nolint:gosec // the instance of http.Server will never be used
		"Shutdown",
		func(server *http.Server, ctx context.Context) error {
			patchGuard.Unpatch()
			defer patchGuard.Restore()
			_ = server.Shutdown(ctx)
			return expectedError
		})
	defer monkey.UnpatchAll()

	doneChannel := srv.Start()
	defer close(doneChannel)

	err = syscall.Kill(syscall.Getpid(), syscall.SIGINT)
	require.NoError(t, err)

	select {
	case err := <-doneChannel:
		assert.Equal(t, fmt.Errorf("can't shut down the server: %w", expectedError), err)
	case <-time.After(3 * time.Second):
		assert.Fail(t, "Timeout on waiting for server to stop")
	}
}
