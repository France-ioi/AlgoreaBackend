//go:build !unit

package database_test

import (
	"database/sql/driver"
	"testing"

	"github.com/luna-duclos/instrumentedsql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/France-ioi/AlgoreaBackend/v2/app"
	"github.com/France-ioi/AlgoreaBackend/v2/app/appenv"
	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/testhelpers"
)

func Test_ConnectionOfWrappedDriverImplementsDriverSessionResetter(t *testing.T) {
	appenv.ForceTestEnv()

	ctx := testhelpers.CreateTestContext()
	// needs actual config for connection to DB
	config := testhelpers.GetConfigFromContext(ctx)
	dbConfig, _ := app.DBConfig(config)
	rawDB, err := database.OpenRawDBConnection(dbConfig.FormatDSN(), true)
	require.NoError(t, err)
	defer func() { assert.NoError(t, rawDB.Close()) }()

	assert.IsType(t, (*instrumentedsql.WrappedDriver)(nil), rawDB.Driver())
	connection, err := rawDB.Conn(ctx)
	require.NoError(t, err)
	defer func() { assert.NoError(t, connection.Close()) }()

	assert.NoError(t, connection.Raw(func(driverConn interface{}) error {
		assert.Implements(t, (*driver.SessionResetter)(nil), driverConn)
		return nil
	}))
}
