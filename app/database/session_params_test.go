package database

import (
	"context"
	"database/sql/driver"
	"errors"
	"reflect"
	"testing"

	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/France-ioi/AlgoreaBackend/v2/app/logging"
)

func TestSetSessionParams_GetSessionParams(t *testing.T) {
	defer func() { require.NoError(t, SetSessionParams(nil)) }()

	assert.Nil(t, getSessionParams())

	require.NoError(t, SetSessionParams(map[string]string{"innodb_lock_wait_timeout": "5"}))
	got := getSessionParams()
	assert.Equal(t, map[string]string{"innodb_lock_wait_timeout": "5"}, got)

	// Mutation of the returned map must not affect stored params.
	got["innodb_lock_wait_timeout"] = "99"
	assert.Equal(t, map[string]string{"innodb_lock_wait_timeout": "5"}, getSessionParams())

	require.NoError(t, SetSessionParams(nil))
	assert.Nil(t, getSessionParams())

	require.NoError(t, SetSessionParams(map[string]string{}))
	assert.Nil(t, getSessionParams())
}

func TestSetSessionParams_ValidatesBeforeStoring(t *testing.T) {
	defer func() { require.NoError(t, SetSessionParams(nil)) }()
	require.NoError(t, SetSessionParams(map[string]string{"innodb_lock_wait_timeout": "5"}))

	err := SetSessionParams(map[string]string{"bad-name": "1"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid database session param name")
	assert.Equal(t, map[string]string{"innodb_lock_wait_timeout": "5"}, getSessionParams())

	err = SetSessionParams(map[string]string{"innodb_lock_wait_timeout": "5;DROP"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid database session param value")
	assert.Equal(t, map[string]string{"innodb_lock_wait_timeout": "5"}, getSessionParams())
}

type execerConnMock struct {
	*connMock

	lastQuery string
	lastArgs  []driver.NamedValue
	execErr   error
}

func (c *execerConnMock) ExecContext(_ context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	c.lastQuery = query
	c.lastArgs = append([]driver.NamedValue(nil), args...)
	if c.execErr != nil {
		return nil, c.execErr
	}
	return resultMock{}, nil
}

type resultMock struct{}

func (resultMock) LastInsertId() (int64, error) { return 0, nil }
func (resultMock) RowsAffected() (int64, error) { return 0, nil }

var _ driver.ExecerContext = &execerConnMock{}

func Test_mysqlConnWrapper_applySessionParams(t *testing.T) {
	t.Run("no-op when empty", func(t *testing.T) {
		conn := &mysqlConnWrapper{conn: &connMock{}, sessionParams: nil}
		require.NoError(t, conn.applySessionParams(context.Background()))
	})

	t.Run("sets sorted session variables", func(t *testing.T) {
		inner := &execerConnMock{connMock: &connMock{}}
		conn := &mysqlConnWrapper{
			conn: inner,
			sessionParams: map[string]string{
				"max_execution_time":       "1000",
				"innodb_lock_wait_timeout": "5",
			},
		}
		require.NoError(t, conn.applySessionParams(context.Background()))
		assert.Equal(t,
			"SET @@SESSION.innodb_lock_wait_timeout=5, @@SESSION.max_execution_time=1000",
			inner.lastQuery)
		assert.Nil(t, inner.lastArgs)
	})

	t.Run("allows quoted string values", func(t *testing.T) {
		inner := &execerConnMock{connMock: &connMock{}}
		conn := &mysqlConnWrapper{
			conn:          inner,
			sessionParams: map[string]string{"sql_mode": "'STRICT_TRANS_TABLES'"},
		}
		require.NoError(t, conn.applySessionParams(context.Background()))
		assert.Equal(t, "SET @@SESSION.sql_mode='STRICT_TRANS_TABLES'", inner.lastQuery)
	})

	t.Run("rejects invalid names", func(t *testing.T) {
		conn := &mysqlConnWrapper{
			conn:          &execerConnMock{connMock: &connMock{}},
			sessionParams: map[string]string{"bad-name": "1"},
		}
		err := conn.applySessionParams(context.Background())
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid database session param name")
	})

	t.Run("rejects dangerous values", func(t *testing.T) {
		conn := &mysqlConnWrapper{
			conn:          &execerConnMock{connMock: &connMock{}},
			sessionParams: map[string]string{"innodb_lock_wait_timeout": "5;DROP TABLE x"},
		}
		err := conn.applySessionParams(context.Background())
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid database session param value")
	})

	t.Run("propagates exec error", func(t *testing.T) {
		expected := errors.New("exec failed")
		inner := &execerConnMock{connMock: &connMock{}, execErr: expected}
		conn := &mysqlConnWrapper{
			conn:          inner,
			sessionParams: map[string]string{"innodb_lock_wait_timeout": "5"},
		}
		assert.Equal(t, expected, conn.applySessionParams(context.Background()))
	})
}

type sessionResetterExecerConnMock struct {
	*execerConnMock

	resetSessionErr error
	resetCalls      int
}

func (c *sessionResetterExecerConnMock) ResetSession(context.Context) error {
	c.resetCalls++
	return c.resetSessionErr
}

var _ driver.SessionResetter = &sessionResetterExecerConnMock{}

func Test_mysqlConnWrapper_ResetSession_ReappliesSessionParams(t *testing.T) {
	ctx, _, _ := logging.NewContextWithNewMockLogger()
	inner := &sessionResetterExecerConnMock{execerConnMock: &execerConnMock{connMock: &connMock{}}}
	conn := &mysqlConnWrapper{
		conn:          inner,
		sessionParams: map[string]string{"innodb_lock_wait_timeout": "5"},
	}

	var resetCalled bool
	monkey.PatchInstanceMethod(reflect.TypeOf(conn), "Reset", func(_ *mysqlConnWrapper, _ context.Context) error {
		resetCalled = true
		return nil
	})
	defer monkey.UnpatchAll()

	require.NoError(t, conn.ResetSession(ctx))
	assert.True(t, resetCalled)
	assert.Equal(t, 1, inner.resetCalls)
	assert.Equal(t, "SET @@SESSION.innodb_lock_wait_timeout=5", inner.lastQuery)
	assert.Nil(t, inner.lastArgs)
}

func Test_mysqlConnWrapper_ResetSession_ReturnsErrBadConnOnResetError(t *testing.T) {
	ctx, _, logHook := logging.NewContextWithNewMockLogger()
	inner := &sessionResetterExecerConnMock{execerConnMock: &execerConnMock{connMock: &connMock{}}}
	conn := &mysqlConnWrapper{
		conn:          inner,
		sessionParams: map[string]string{"innodb_lock_wait_timeout": "5"},
	}

	expected := errors.New("reset failed")
	monkey.PatchInstanceMethod(reflect.TypeOf(conn), "Reset", func(_ *mysqlConnWrapper, _ context.Context) error {
		return expected
	})
	defer monkey.UnpatchAll()

	err := conn.ResetSession(ctx)
	require.Error(t, err)
	require.ErrorIs(t, err, driver.ErrBadConn)
	assert.Contains(t, err.Error(), "reset failed")
	assert.Empty(t, inner.lastQuery)
	require.NotEmpty(t, logHook.AllEntries())
	assert.Equal(t, "warning", logHook.LastEntry().Level.String())
}

func Test_mysqlConnWrapper_ResetSession_ReturnsErrBadConnOnApplyError(t *testing.T) {
	ctx, _, logHook := logging.NewContextWithNewMockLogger()
	expected := errors.New("apply failed")
	inner := &sessionResetterExecerConnMock{
		execerConnMock: &execerConnMock{connMock: &connMock{}, execErr: expected},
	}
	conn := &mysqlConnWrapper{
		conn:          inner,
		sessionParams: map[string]string{"innodb_lock_wait_timeout": "5"},
	}

	monkey.PatchInstanceMethod(reflect.TypeOf(conn), "Reset", func(_ *mysqlConnWrapper, _ context.Context) error {
		return nil
	})
	defer monkey.UnpatchAll()

	err := conn.ResetSession(ctx)
	require.Error(t, err)
	require.ErrorIs(t, err, driver.ErrBadConn)
	assert.Contains(t, err.Error(), "apply failed")
	require.NotEmpty(t, logHook.AllEntries())
	assert.Equal(t, "warning", logHook.LastEntry().Level.String())
}

func Test_discardConnectionAfterResetFailure_PreservesErrBadConn(t *testing.T) {
	ctx, _, logHook := logging.NewContextWithNewMockLogger()
	err := discardConnectionAfterResetFailure(ctx, "resetting session", driver.ErrBadConn)
	require.ErrorIs(t, err, driver.ErrBadConn)
	assert.Equal(t, driver.ErrBadConn, err)
	require.NotEmpty(t, logHook.AllEntries())
	assert.Equal(t, "warning", logHook.LastEntry().Level.String())
}

func Test_discardConnectionAfterResetFailure_LogsInfoOnCancelledContext(t *testing.T) {
	ctx, _, logHook := logging.NewContextWithNewMockLogger()
	err := discardConnectionAfterResetFailure(ctx, "resetting session", context.Canceled)
	require.ErrorIs(t, err, driver.ErrBadConn)
	require.NotEmpty(t, logHook.AllEntries())
	assert.Equal(t, "info", logHook.LastEntry().Level.String())
}

func Test_discardConnectionAfterResetFailure_LogsInfoOnDeadlineExceeded(t *testing.T) {
	ctx, _, logHook := logging.NewContextWithNewMockLogger()
	err := discardConnectionAfterResetFailure(ctx, "resetting session", context.DeadlineExceeded)
	require.ErrorIs(t, err, driver.ErrBadConn)
	require.NotEmpty(t, logHook.AllEntries())
	assert.Equal(t, "info", logHook.LastEntry().Level.String())
}

func Test_mysqlConnWrapper_ResetSession_PropagatesDriverResetSessionError(t *testing.T) {
	ctx, _, logHook := logging.NewContextWithNewMockLogger()
	expected := errors.New("driver reset failed")
	inner := &sessionResetterExecerConnMock{
		execerConnMock:  &execerConnMock{connMock: &connMock{}},
		resetSessionErr: expected,
	}
	conn := &mysqlConnWrapper{
		conn:          inner,
		sessionParams: map[string]string{"innodb_lock_wait_timeout": "5"},
	}
	err := conn.ResetSession(ctx)
	require.ErrorIs(t, err, driver.ErrBadConn)
	assert.Contains(t, err.Error(), "driver reset failed")
	require.NotEmpty(t, logHook.AllEntries())
	assert.Equal(t, "warning", logHook.LastEntry().Level.String())
}

func Test_mysqlConnectorWrapper_Connect_AppliesSessionParams(t *testing.T) {
	defer func() { require.NoError(t, SetSessionParams(nil)) }()
	require.NoError(t, SetSessionParams(map[string]string{"innodb_lock_wait_timeout": "5"}))

	innerConn := &closeTrackingExecerConn{
		execerConnMock: &execerConnMock{connMock: &connMock{}},
	}
	c := &mysqlConnectorWrapper{
		connector: &connectorMock{conn: innerConn},
	}
	got, err := c.Connect(context.Background())
	require.NoError(t, err)
	wrapped, ok := got.(*mysqlConnWrapper)
	require.True(t, ok)
	assert.Equal(t, map[string]string{"innodb_lock_wait_timeout": "5"}, wrapped.sessionParams)
	assert.Equal(t, "SET @@SESSION.innodb_lock_wait_timeout=5", innerConn.lastQuery)
}

func Test_mysqlConnectorWrapper_Connect_ClosesOnApplyError(t *testing.T) {
	defer func() { require.NoError(t, SetSessionParams(nil)) }()
	require.NoError(t, SetSessionParams(map[string]string{"innodb_lock_wait_timeout": "5"}))

	expected := errors.New("apply failed")
	innerConn := &closeTrackingExecerConn{
		execerConnMock: &execerConnMock{connMock: &connMock{}, execErr: expected},
	}
	c := &mysqlConnectorWrapper{
		connector: &connectorMock{conn: innerConn},
	}
	got, err := c.Connect(context.Background())
	assert.Equal(t, expected, err)
	assert.Nil(t, got)
	assert.True(t, innerConn.closed)
}

func Test_mysqlConnectorWrapper_Connect_PropagatesConnectError(t *testing.T) {
	expected := errors.New("connect failed")
	c := &mysqlConnectorWrapper{
		connector: &connectorMock{err: expected},
	}
	got, err := c.Connect(context.Background())
	assert.Equal(t, expected, err)
	assert.Nil(t, got)
}

type connectorMock struct {
	conn driver.Conn
	err  error
}

func (c *connectorMock) Connect(context.Context) (driver.Conn, error) {
	if c.err != nil {
		return nil, c.err
	}
	return c.conn, nil
}

func (c *connectorMock) Driver() driver.Driver { return nil }

var _ driver.Connector = &connectorMock{}

func Test_mysqlDriverWrapper_Open_AppliesSessionParams(t *testing.T) {
	defer func() { require.NoError(t, SetSessionParams(nil)) }()
	require.NoError(t, SetSessionParams(map[string]string{"innodb_lock_wait_timeout": "5"}))

	inner := &execerDriverMock{}
	d := &mysqlDriverWrapper{driver: inner}
	got, err := d.Open("dsn")
	require.NoError(t, err)
	wrapped, ok := got.(*mysqlConnWrapper)
	require.True(t, ok)
	assert.Equal(t, map[string]string{"innodb_lock_wait_timeout": "5"}, wrapped.sessionParams)
	assert.Equal(t, "SET @@SESSION.innodb_lock_wait_timeout=5", inner.conn.lastQuery)
}

func Test_mysqlDriverWrapper_Open_ClosesOnApplyError(t *testing.T) {
	defer func() { require.NoError(t, SetSessionParams(nil)) }()
	require.NoError(t, SetSessionParams(map[string]string{"innodb_lock_wait_timeout": "5"}))

	expected := errors.New("apply failed")
	inner := &execerDriverMock{execErr: expected}
	d := &mysqlDriverWrapper{driver: inner}
	got, err := d.Open("dsn")
	assert.Equal(t, expected, err)
	assert.Nil(t, got)
	assert.True(t, inner.conn.closed)
}

type execerDriverMock struct {
	conn    *closeTrackingExecerConn
	execErr error
}

func (d *execerDriverMock) Open(string) (driver.Conn, error) {
	d.conn = &closeTrackingExecerConn{
		execerConnMock: &execerConnMock{connMock: &connMock{}, execErr: d.execErr},
	}
	return d.conn, nil
}

var _ driver.Driver = &execerDriverMock{}

type closeTrackingExecerConn struct {
	*execerConnMock

	closed bool
}

func (c *closeTrackingExecerConn) Close() error {
	c.closed = true
	return nil
}
