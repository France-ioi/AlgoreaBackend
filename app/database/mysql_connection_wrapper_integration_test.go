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
	db, err := testhelpers.OpenRawDBConnection()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	ctx := context.Background()
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
		return driverConn.(driver.SessionResetter).ResetSession(context.Background())
	})
	require.NoError(t, err)

	err = conn.QueryRowContext(ctx, "SELECT @a").Scan(&result)
	require.NoError(t, err)
	assert.Nil(t, result)
}

func TestMysqlConnectionWrapper_ResetSession_FailsOnClosedConnection(t *testing.T) {
	db, err := testhelpers.OpenRawDBConnection()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	ctx := context.Background()
	conn, err := db.Conn(ctx)
	require.NoError(t, err)
	defer func() { _ = conn.Close() }()

	err = conn.Raw(func(driverConn any) error {
		require.NoError(t, driverConn.(driver.Conn).Close())
		return driverConn.(driver.SessionResetter).ResetSession(context.Background())
	})
	require.Equal(t, driver.ErrBadConn, err)
}

type resetter = interface {
	Reset(context.Context) error
}

func TestMysqlConnectionWrapper_Reset_FailsOnClosedConnection(t *testing.T) {
	monkey.PatchInstanceMethod(reflect.TypeOf(&viper.Viper{}), "GetBool",
		func(_ *viper.Viper, key string) bool { return false })
	defer monkey.UnpatchAll()

	db, err := testhelpers.OpenRawDBConnection()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	ctx := context.Background()
	conn, err := db.Conn(ctx)
	require.NoError(t, err)
	defer func() { _ = conn.Close() }()

	err = conn.Raw(func(driverConn any) error {
		require.NoError(t, driverConn.(driver.Conn).Close())
		return driverConn.(resetter).Reset(context.Background())
	})
	require.Equal(t, driver.ErrBadConn, err)
}

func TestMysqlConnectionWrapper_Reset_FailsWhenContextIsCancelled(t *testing.T) {
	monkey.PatchInstanceMethod(reflect.TypeOf(&viper.Viper{}), "GetBool",
		func(_ *viper.Viper, key string) bool { return false })
	defer monkey.UnpatchAll()

	db, err := testhelpers.OpenRawDBConnection()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	ctx, cancelFunc := context.WithCancel(context.Background())
	conn, err := db.Conn(ctx)
	require.NoError(t, err)
	defer func() { _ = conn.Close() }()

	err = conn.Raw(func(driverConn any) error {
		cancelFunc()
		return driverConn.(resetter).Reset(ctx)
	})
	require.Equal(t, context.Canceled, err)
}

func TestMysqlConnectionWrapper_Reset_FailsWhenWriteCommandPacketFails(t *testing.T) {
	monkey.PatchInstanceMethod(reflect.TypeOf(&viper.Viper{}), "GetBool",
		func(_ *viper.Viper, key string) bool { return false })
	defer monkey.UnpatchAll()

	db, err := testhelpers.OpenRawDBConnection()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	ctx := context.Background()
	conn, err := db.Conn(ctx)
	require.NoError(t, err)
	defer func() { _ = conn.Close() }()

	expectedError := errors.New("some error")
	err = conn.Raw(func(driverConn any) error {
		monkey.Patch(mysqlConnWriteCommandPacket, func(unsafe.Pointer, byte) error {
			return expectedError
		})
		defer monkey.UnpatchAll()
		return driverConn.(resetter).Reset(ctx)
	})
	require.Equal(t, expectedError, err)
}

//go:linkname mysqlConnWriteCommandPacket github.com/go-sql-driver/mysql.(*mysqlConn).writeCommandPacket
func mysqlConnWriteCommandPacket(unsafe.Pointer, byte) error
