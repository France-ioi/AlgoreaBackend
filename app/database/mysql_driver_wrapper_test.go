package database

import (
	"database/sql/driver"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type connMockWithName struct {
	*connMock
	name string
}

type driverMock struct{}

func (d *driverMock) Open(name string) (driver.Conn, error) {
	return &connMockWithName{&connMock{}, name}, nil
}

var _ driver.Driver = &driverMock{}

const expectedDriverName = "name"

func Test_mysqlDriverWrapper_Open(t *testing.T) {
	d := &mysqlDriverWrapper{
		driver: &driverMock{},
	}
	got, err := d.Open(expectedDriverName)
	require.NoError(t, err)
	assert.Equal(t, &mysqlConnWrapper{&connMockWithName{&connMock{}, expectedDriverName}}, got)
}

type driverMockWithError struct{}

func (d *driverMockWithError) Open(name string) (driver.Conn, error) {
	//nolint:nilnil // we return the value for test purposes
	return &connMockWithName{&connMock{}, name}, fmt.Errorf("error for %s", name)
}

var _ driver.Driver = &driverMockWithError{}

func Test_mysqlDriverWrapper_Open_Error(t *testing.T) {
	d := &mysqlDriverWrapper{
		driver: &driverMockWithError{},
	}
	got, err := d.Open(expectedDriverName)
	assert.Equal(t, fmt.Errorf("error for %s", expectedDriverName), err)
	assert.Nil(t, got)
}

type driverConnectorMockWithError struct{}

func (d *driverConnectorMockWithError) Open(string) (driver.Conn, error) { return nil, nil } //nolint:nilnil // It's just a mock.

var _ driver.Driver = &driverConnectorMockWithError{}

func (d *driverConnectorMockWithError) OpenConnector(name string) (driver.Connector, error) {
	//nolint:nilnil // we return the value for test purposes
	return &mysqlConnectorWrapper{}, fmt.Errorf("error for %s", name)
}

var _ driver.DriverContext = &driverConnectorMockWithError{}

func Test_mysqlDriverWrapper_OpenConnector_Error(t *testing.T) {
	d := &mysqlDriverWrapper{
		driver: &driverConnectorMockWithError{},
	}
	got, err := d.OpenConnector(expectedDriverName)
	assert.Equal(t, fmt.Errorf("error for %s", expectedDriverName), err)
	assert.Nil(t, got)
}
