package cmd

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"os"
	"reflect"
	"syscall"
	"testing"
	"time"

	"bou.ke/monkey"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/France-ioi/AlgoreaBackend/v2/app"
	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/app/logging"
	"github.com/France-ioi/AlgoreaBackend/v2/testhelpers/testoutput"
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

func TestEnablePropagationCommandDebugLogging(t *testing.T) {
	logger, logHook := logging.NewMockLogger()
	require.False(t, logger.IsDebugEnabled())
	ctx := logging.ContextWithLogger(context.Background(), logger)

	db, mock := database.NewDBMock(ctx)
	defer func() { _ = db.Close() }()

	enablePropagationCommandDebugLogging(&app.Application{Database: db})
	assert.True(t, logger.IsDebugEnabled())
	logs := ""
	for _, entry := range logHook.AllEntries() {
		logs += entry.Message
	}
	assert.Contains(t, logs, "logging level raised to debug")
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRunPropagationCommand_SuccessEnablesDebug(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	logger, _ := logging.NewMockLogger()
	require.False(t, logger.IsDebugEnabled())
	ctx := logging.ContextWithLogger(context.Background(), logger)
	db, mock := database.NewDBMock(ctx)
	defer func() { _ = db.Close() }()

	application := &app.Application{Database: db}
	monkey.Patch(app.New, func(...*logging.Logger) (*app.Application, error) {
		return application, nil
	})
	monkey.PatchInstanceMethod(reflect.TypeOf(&database.DataStore{}), "WithNamedLock",
		func(s *database.DataStore, _ string, _ time.Duration, fn func(*database.DataStore) error) error {
			return fn(s)
		})
	monkey.PatchInstanceMethod(reflect.TypeOf(&database.DataStore{}), "InTransaction",
		func(s *database.DataStore, fn func(*database.DataStore) error, _ ...*sql.TxOptions) error {
			return fn(s)
		})
	monkey.PatchInstanceMethod(reflect.TypeOf(&database.DataStore{}), "SchedulePermissionsPropagation",
		func(*database.DataStore) {})
	monkey.PatchInstanceMethod(reflect.TypeOf(&database.DataStore{}), "ScheduleResultsPropagation",
		func(*database.DataStore) {})
	defer monkey.UnpatchAll()

	maxDuration := time.Duration(0)
	cmd := &cobra.Command{}
	var out bytes.Buffer
	cmd.SetOut(&out)

	require.NoError(t, runPropagationCommand(&maxDuration)(cmd, []string{"test"}))
	assert.True(t, logger.IsDebugEnabled())
	assert.Contains(t, out.String(), "Propagation done.")
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRunPropagationCommand_WithMaxDuration(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	logger, _ := logging.NewMockLogger()
	ctx := logging.ContextWithLogger(context.Background(), logger)
	db, mock := database.NewDBMock(ctx)
	defer func() { _ = db.Close() }()

	application := &app.Application{Database: db}
	monkey.Patch(app.New, func(...*logging.Logger) (*app.Application, error) {
		return application, nil
	})
	monkey.PatchInstanceMethod(reflect.TypeOf(&database.DataStore{}), "WithNamedLock",
		func(s *database.DataStore, _ string, _ time.Duration, fn func(*database.DataStore) error) error {
			return fn(s)
		})
	monkey.PatchInstanceMethod(reflect.TypeOf(&database.DataStore{}), "InTransaction",
		func(s *database.DataStore, fn func(*database.DataStore) error, _ ...*sql.TxOptions) error {
			return fn(s)
		})
	monkey.PatchInstanceMethod(reflect.TypeOf(&database.DataStore{}), "SchedulePermissionsPropagation",
		func(*database.DataStore) {})
	monkey.PatchInstanceMethod(reflect.TypeOf(&database.DataStore{}), "ScheduleResultsPropagation",
		func(*database.DataStore) {})
	defer monkey.UnpatchAll()

	maxDuration := propagationShutdownMargin + time.Minute
	cmd := &cobra.Command{}
	var out bytes.Buffer
	cmd.SetOut(&out)

	require.NoError(t, runPropagationCommand(&maxDuration)(cmd, nil))
	assert.Contains(t, out.String(), "Propagation done.")
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRunPropagationCommand_ValidationError(t *testing.T) {
	maxDuration := propagationShutdownMargin
	err := runPropagationCommand(&maxDuration)(&cobra.Command{}, nil)
	require.Error(t, err)
}

func TestRunPropagationCommand_AppNewError(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	expected := errors.New("boot failed")
	monkey.Patch(app.New, func(...*logging.Logger) (*app.Application, error) {
		return nil, expected
	})
	defer monkey.UnpatchAll()

	maxDuration := time.Duration(0)
	assert.Equal(t, expected, runPropagationCommand(&maxDuration)(&cobra.Command{}, nil))
}

func TestRunPropagationCommand_NamedLockTimeoutWithinBudget(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	logger, _ := logging.NewMockLogger()
	ctx := logging.ContextWithLogger(context.Background(), logger)
	db, mock := database.NewDBMock(ctx)
	defer func() { _ = db.Close() }()

	application := &app.Application{Database: db}
	monkey.Patch(app.New, func(...*logging.Logger) (*app.Application, error) {
		return application, nil
	})
	monkey.PatchInstanceMethod(reflect.TypeOf(&database.DataStore{}), "WithNamedLock",
		func(_ *database.DataStore, _ string, _ time.Duration, _ func(*database.DataStore) error) error {
			return database.ErrNamedLockWaitTimeoutExceeded
		})
	defer monkey.UnpatchAll()

	maxDuration := propagationShutdownMargin + time.Minute
	cmd := &cobra.Command{}
	var out bytes.Buffer
	cmd.SetOut(&out)

	require.NoError(t, runPropagationCommand(&maxDuration)(cmd, nil))
	assert.Contains(t, out.String(), "Propagation skipped")
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRunPropagationCommand_OtherLockError(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	logger, _ := logging.NewMockLogger()
	ctx := logging.ContextWithLogger(context.Background(), logger)
	db, mock := database.NewDBMock(ctx)
	defer func() { _ = db.Close() }()

	application := &app.Application{Database: db}
	monkey.Patch(app.New, func(...*logging.Logger) (*app.Application, error) {
		return application, nil
	})
	monkey.PatchInstanceMethod(reflect.TypeOf(&database.DataStore{}), "WithNamedLock",
		func(_ *database.DataStore, _ string, _ time.Duration, _ func(*database.DataStore) error) error {
			return errors.New("lock exploded")
		})
	defer monkey.UnpatchAll()

	maxDuration := time.Duration(0)
	err := runPropagationCommand(&maxDuration)(&cobra.Command{}, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "lock exploded")
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestConfigurePropagationSoftDeadline(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	logger, logHook := logging.NewMockLogger()
	ctx := logging.ContextWithLogger(context.Background(), logger)
	db, mock := database.NewDBMock(ctx)
	defer func() { _ = db.Close() }()

	application := &app.Application{Database: db}
	lockTimeout := configurePropagationSoftDeadline(application, time.Now(), propagationShutdownMargin+time.Minute)
	assert.Positive(t, lockTimeout)

	_ = configurePropagationSoftDeadline(application, time.Now().Add(-time.Hour), propagationShutdownMargin+time.Second)
	logs := ""
	for _, e := range logHook.AllEntries() {
		logs += e.Message
	}
	assert.Contains(t, logs, "soft deadline")
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestConfigurePropagationSoftDeadline_NoLimit(t *testing.T) {
	application := &app.Application{}
	assert.Equal(t, propagationCommandLockTimeout,
		configurePropagationSoftDeadline(application, time.Now(), 0))
}

func TestHandlePropagationLockError(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	logger, _ := logging.NewMockLogger()
	ctx := logging.ContextWithLogger(context.Background(), logger)
	db, mock := database.NewDBMock(ctx)
	defer func() { _ = db.Close() }()
	application := &app.Application{Database: db}
	cmd := &cobra.Command{}
	var out bytes.Buffer
	cmd.SetOut(&out)

	require.NoError(t, handlePropagationLockError(cmd, application, time.Minute, database.ErrNamedLockWaitTimeoutExceeded))
	assert.Contains(t, out.String(), "Propagation skipped")

	err := handlePropagationLockError(cmd, application, 0, errors.New("other"))
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPrintPropagationCompletion(t *testing.T) {
	logger, _ := logging.NewMockLogger()
	ctx := logging.ContextWithLogger(context.Background(), logger)
	db, mock := database.NewDBMock(ctx)
	defer func() { _ = db.Close() }()

	cmd := &cobra.Command{}
	var out bytes.Buffer
	cmd.SetOut(&out)
	printPropagationCompletion(cmd, db)
	assert.Contains(t, out.String(), "Propagation done.")

	database.SetPropagationSoftDeadline(db, time.Now().Add(-time.Second))
	out.Reset()
	printPropagationCompletion(cmd, db)
	assert.Contains(t, out.String(), "stopped early")
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestStartPropagationSIGTERMHandler(t *testing.T) {
	logger, logHook := logging.NewMockLogger()
	ctx := logging.ContextWithLogger(context.Background(), logger)
	db, mock := database.NewDBMock(ctx)
	defer func() { _ = db.Close() }()

	application := &app.Application{Database: db}
	database.SetPropagationSoftDeadline(db, time.Now().Add(time.Hour))
	stop := startPropagationSIGTERMHandler(application)
	defer stop()

	// Ordering is load-bearing: Notify must run before Kill. signal.Notify redirects SIGTERM to
	// our channel (does not terminate the process); if Kill raced ahead of Notify, the default
	// disposition would kill the test binary.
	require.NoError(t, syscall.Kill(os.Getpid(), syscall.SIGTERM))
	require.Eventually(t, func() bool {
		return database.SoftDeadlineExceeded(db)
	}, time.Second, 10*time.Millisecond)

	logs := ""
	for _, entry := range logHook.AllEntries() {
		logs += entry.Message
	}
	assert.Contains(t, logs, "Got SIGTERM")
	require.NoError(t, mock.ExpectationsWereMet())
}
