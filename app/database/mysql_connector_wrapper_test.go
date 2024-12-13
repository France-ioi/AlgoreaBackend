package database

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_mysqlConnectorWrapper_Driver(t *testing.T) {
	expectedDriver := &mysqlDriverWrapper{}
	c := &mysqlConnectorWrapper{
		connector: nil,
		driver:    expectedDriver,
	}
	assert.Equal(t, expectedDriver, c.Driver())
}
