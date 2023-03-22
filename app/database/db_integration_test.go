// +build !unit

package database_test

import (
	"context"
	"database/sql/driver"
	"testing"

	"github.com/luna-duclos/instrumentedsql"
	"github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/testhelpers"
)

func Test_ConnectionOfWrappedDriverImplementsDriverSessionResetter(t *testing.T) {
	rawDB, err := testhelpers.OpenRawDBConnection()
	assert.NoError(t, err)
	if err == nil {
		defer func() { assert.NoError(t, rawDB.Close()) }()
	}

	assert.IsType(t, (*instrumentedsql.WrappedDriver)(nil), rawDB.Driver())
	connection, err := rawDB.Conn(context.Background())
	assert.NoError(t, err)
	if err == nil {
		defer func() { assert.NoError(t, connection.Close()) }()
	}
	assert.NoError(t, connection.Raw(func(driverConn interface{}) error {
		assert.Implements(t, (*driver.SessionResetter)(nil), driverConn)
		return nil
	}))
}
