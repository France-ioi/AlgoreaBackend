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
)

func TestServer_Start(t *testing.T) {
	app, err := New()
	assert.NoError(t, err)
	srv, err := NewServer(app)
	assert.NoError(t, err)

	// check defaults are applied correctly
	assert.True(t, strings.HasSuffix(srv.Addr, ":8088"))
	assert.Equal(t, time.Duration(60000000000), srv.ReadTimeout)
	assert.Equal(t, time.Duration(60000000000), srv.WriteTimeout)

	doneChannel := srv.Start()
	defer close(doneChannel)

	err = syscall.Kill(syscall.Getpid(), syscall.SIGINT)
	assert.NoError(t, err)

	select {
	case err = <-doneChannel:
		assert.NoError(t, err)
	case <-time.After(3 * time.Second):
		assert.Fail(t, "Timeout on waiting for server to stop")
	}
}

func TestServer_StartHandlesListenerError(t *testing.T) {
	app, err := New()
	assert.NoError(t, err)
	app.Config.Set("server.port", -1)
	srv, err := NewServer(app)
	assert.NoError(t, err)

	doneChannel := srv.Start()
	defer close(doneChannel)

	select {
	case err = <-doneChannel:
		assert.Error(t, err)
	case <-time.After(3 * time.Second):
		_ = syscall.Kill(syscall.Getpid(), syscall.SIGINT)
		assert.Fail(t, "Timeout on waiting for server to stop")
	}
}

func TestServer_StartCanBeStoppedByShutdown(t *testing.T) {
	app, err := New()
	assert.NoError(t, err)
	srv, err := NewServer(app)
	assert.NoError(t, err)

	doneChannel := srv.Start()
	defer close(doneChannel)

	_ = srv.Shutdown(context.Background())

	select {
	case err := <-doneChannel:
		assert.NoError(t, err)
	case <-time.After(3 * time.Second):
		assert.Fail(t, "Timeout on waiting for server to stop")
	}
}

func TestServer_StartHandlesShutdownError_OnKilling(t *testing.T) {
	app, err := New()
	assert.NoError(t, err)
	srv, err := NewServer(app)
	assert.NoError(t, err)

	expectedError := errors.New("some error")
	var patchGuard *monkey.PatchGuard
	patchGuard = monkey.PatchInstanceMethod(reflect.TypeOf(&http.Server{}), "Shutdown",
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
	assert.NoError(t, err)

	select {
	case err := <-doneChannel:
		assert.Equal(t, fmt.Errorf("can't shut down the server: %v", expectedError), err)
	case <-time.After(3 * time.Second):
		assert.Fail(t, "Timeout on waiting for server to stop")
	}
}
