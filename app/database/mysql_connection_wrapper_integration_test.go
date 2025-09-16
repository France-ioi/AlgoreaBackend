//go:build !unit

package database_test

import (
	"context"
	"database/sql/driver"
	"errors"
	"reflect"
	"testing"
	"unsafe"

	"bou.ke/monkey"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/France-ioi/AlgoreaBackend/v2/testhelpers"
)

func TestMysqlConnectionWrapper_ResetSession(t *testing.T) {
	ctx := testhelpers.CreateTestContext()
	db := testhelpers.OpenRawDBConnection(ctx)
	defer func() { _ = db.Close() }()

	conn, err := db.Conn(ctx)
	require.NoError(t, err)
	defer func() { _ = conn.Close() }()

	var result *int64
	_, err = conn.ExecContext(ctx, "SET @a := 1")
	require.NoError(t, err)
	err = conn.QueryRowContext(ctx, "SELECT @a").Scan(&result)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, int64(1), *result)

	err = conn.Raw(func(driverConn any) error {
		sessionResetter, ok := driverConn.(driver.SessionResetter)
		require.True(t, ok)
		return sessionResetter.ResetSession(ctx)
	})
	require.NoError(t, err)

	err = conn.QueryRowContext(ctx, "SELECT @a").Scan(&result)
	require.NoError(t, err)
	assert.Nil(t, result)
}

func TestMysqlConnectionWrapper_ResetSession_ResetsForeignKeyChecks(t *testing.T) {
	ctx := testhelpers.CreateTestContext()
	db := testhelpers.OpenRawDBConnection(ctx)
	defer func() { _ = db.Close() }()

	conn, err := db.Conn(ctx)
	require.NoError(t, err)
	defer func() { _ = conn.Close() }()

	var result *int64
	_, err = conn.ExecContext(ctx, "SET FOREIGN_KEY_CHECKS = 0")
	require.NoError(t, err)
	err = conn.QueryRowContext(ctx, "SELECT @@FOREIGN_KEY_CHECKS").Scan(&result)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, int64(0), *result)

	err = conn.Raw(func(driverConn any) error {
		sessionResetter, ok := driverConn.(driver.SessionResetter)
		require.True(t, ok)
		return sessionResetter.ResetSession(ctx)
	})
	require.NoError(t, err)

	err = conn.QueryRowContext(ctx, "SELECT @@FOREIGN_KEY_CHECKS").Scan(&result)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, int64(1), *result)
}

func TestMysqlConnectionWrapper_ResetSession_FailsOnClosedConnection(t *testing.T) {
	ctx := testhelpers.CreateTestContext()
	db := testhelpers.OpenRawDBConnection(ctx)
	defer func() { _ = db.Close() }()

	conn, err := db.Conn(ctx)
	require.NoError(t, err)
	defer func() { _ = conn.Close() }()

	err = conn.Raw(func(driverConn any) error {
		driverConnObject, ok := driverConn.(driver.Conn)
		require.True(t, ok)
		require.NoError(t, driverConnObject.Close())
		sessionResetter, ok := driverConn.(driver.SessionResetter)
		require.True(t, ok)
		return sessionResetter.ResetSession(ctx)
	})
	require.Equal(t, driver.ErrBadConn, err)
}

type resetter = interface {
	Reset(ctx context.Context) error
}

func TestMysqlConnectionWrapper_Reset_FailsOnClosedConnection(t *testing.T) {
	monkey.PatchInstanceMethod(reflect.TypeOf(&viper.Viper{}), "GetBool",
		func(_ *viper.Viper, _ string) bool { return false })
	defer monkey.UnpatchAll()

	ctx := testhelpers.CreateTestContext()
	db := testhelpers.OpenRawDBConnection(ctx)
	defer func() { _ = db.Close() }()

	conn, err := db.Conn(ctx)
	require.NoError(t, err)
	defer func() { _ = conn.Close() }()

	err = conn.Raw(func(driverConn any) error {
		driverConnObject, ok := driverConn.(driver.Conn)
		require.True(t, ok)
		require.NoError(t, driverConnObject.Close())
		resetterObject, ok := driverConn.(resetter)
		require.True(t, ok)
		return resetterObject.Reset(ctx)
	})
	require.Equal(t, driver.ErrBadConn, err)
}

func TestMysqlConnectionWrapper_Reset_FailsWhenContextIsCancelled(t *testing.T) {
	monkey.PatchInstanceMethod(reflect.TypeOf(&viper.Viper{}), "GetBool",
		func(_ *viper.Viper, _ string) bool { return false })
	defer monkey.UnpatchAll()

	ctx := testhelpers.CreateTestContext()
	db := testhelpers.OpenRawDBConnection(ctx)
	defer func() { _ = db.Close() }()

	var cancelFunc context.CancelFunc
	ctx, cancelFunc = context.WithCancel(ctx)
	conn, err := db.Conn(ctx)
	require.NoError(t, err)
	defer func() { _ = conn.Close() }()

	err = conn.Raw(func(driverConn any) error {
		cancelFunc()
		resetterObject, ok := driverConn.(resetter)
		require.True(t, ok)
		return resetterObject.Reset(ctx)
	})
	require.Equal(t, context.Canceled, err)
}

func TestMysqlConnectionWrapper_Reset_FailsWhenWriteCommandPacketFails(t *testing.T) {
	monkey.PatchInstanceMethod(reflect.TypeOf(&viper.Viper{}), "GetBool",
		func(_ *viper.Viper, _ string) bool { return false })
	defer monkey.UnpatchAll()

	ctx := testhelpers.CreateTestContext()
	db := testhelpers.OpenRawDBConnection(ctx)
	defer func() { _ = db.Close() }()

	conn, err := db.Conn(ctx)
	require.NoError(t, err)
	defer func() { _ = conn.Close() }()

	expectedError := errors.New("some error")
	err = conn.Raw(func(driverConn any) error {
		monkey.Patch(mysqlConnWriteCommandPacket, func(unsafe.Pointer, byte) error {
			return expectedError
		})
		defer monkey.UnpatchAll()
		resetterObject, ok := driverConn.(resetter)
		require.True(t, ok)
		return resetterObject.Reset(ctx)
	})
	require.Equal(t, expectedError, err)
}

//go:linkname mysqlConnWriteCommandPacket github.com/go-sql-driver/mysql.(*mysqlConn).writeCommandPacket
func mysqlConnWriteCommandPacket(unsafe.Pointer, byte) error
