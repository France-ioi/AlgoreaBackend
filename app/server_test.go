package app

import (
	"context"
	"errors"
	"net/http"
	"os"
	"reflect"
	"sync"
	"syscall"
	"testing"
	"time"

	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"
)

func TestServer_Start(t *testing.T) {
	app, err := New("test")
	assert.NoError(t, err)
	srv, err := NewServer(app)
	assert.NoError(t, err)

	lock := sync.Mutex{}
	exitCalled := false
	monkey.Patch(os.Exit, func(code int) {
		lock.Lock()
		exitCalled = true
		lock.Unlock()
		killErr := syscall.Kill(syscall.Getpid(), syscall.SIGINT)
		assert.NoError(t, killErr)
	})
	defer monkey.UnpatchAll()

	doneChannel := srv.Start()
	err = syscall.Kill(syscall.Getpid(), syscall.SIGINT)
	assert.NoError(t, err)

	select {
	case <-doneChannel:
	case <-time.After(3 * time.Second):
		assert.Fail(t, "Timeout on waiting for server to stop")
	}

	lock.Lock()
	defer lock.Unlock()
	assert.False(t, exitCalled)
}

func TestServer_StartHandlesListenerError(t *testing.T) {
	app, err := New("test")
	assert.NoError(t, err)
	app.Config.Server.Port = -1
	srv, err := NewServer(app)
	assert.NoError(t, err)

	lock := sync.Mutex{}
	exitCalled := false
	var exitCode int
	monkey.Patch(os.Exit, func(code int) {
		lock.Lock()
		exitCalled = true
		exitCode = code
		lock.Unlock()
		err := syscall.Kill(syscall.Getpid(), syscall.SIGINT)
		assert.NoError(t, err)
	})
	defer monkey.UnpatchAll()
	select {
	case <-srv.Start():
	case <-time.After(3 * time.Second):
		_ = syscall.Kill(syscall.Getpid(), syscall.SIGINT)
		assert.Fail(t, "Timeout on waiting for server to stop")
	}
	lock.Lock()
	defer lock.Unlock()
	assert.True(t, exitCalled)
	assert.Equal(t, 1, exitCode)
}

func TestServer_StartHandlesShutdownError(t *testing.T) {
	app, err := New("test")
	assert.NoError(t, err)
	srv, err := NewServer(app)
	assert.NoError(t, err)

	lock := sync.Mutex{}
	exitCalled := false
	var exitCode int
	monkey.Patch(os.Exit, func(code int) {
		lock.Lock()
		exitCalled = true
		exitCode = code
		lock.Unlock()
	})
	monkey.PatchInstanceMethod(reflect.TypeOf(&http.Server{}), "Shutdown",
		func(*http.Server, context.Context) error { return errors.New("some errror") })
	defer monkey.UnpatchAll()

	doneChannel := srv.Start()
	_ = srv.Shutdown(context.Background())
	err = syscall.Kill(syscall.Getpid(), syscall.SIGINT)
	assert.NoError(t, err)
	select {
	case <-doneChannel:
	case <-time.After(3 * time.Second):
		assert.Fail(t, "Timeout on waiting for server to stop")
	}

	lock.Lock()
	defer lock.Unlock()
	assert.True(t, exitCalled)
	assert.Equal(t, 1, exitCode)
}
