package database

import (
	"context"
	"database/sql/driver"
	"time"
)

type mysqlDriverWrapper struct {
	driver driver.Driver
}

func newMySQLDriverWrapper(driverToWrap driver.Driver) driver.Driver {
	return &mysqlDriverWrapper{driver: driverToWrap}
}

const applySessionParamsTimeout = 5 * time.Second

func (d *mysqlDriverWrapper) Open(name string) (driver.Conn, error) {
	conn, err := d.driver.Open(name)
	if err != nil {
		return nil, err
	}
	wrapped := &mysqlConnWrapper{conn: conn, sessionParams: getSessionParams()}
	// Open has no context; bound the SET so a hung server cannot block forever on this legacy path.
	ctx, cancel := context.WithTimeout(context.Background(), applySessionParamsTimeout)
	defer cancel()
	if err := wrapped.applySessionParams(ctx); err != nil {
		_ = wrapped.Close()
		return nil, err
	}
	return wrapped, nil
}

var _ driver.Driver = &mysqlDriverWrapper{}

func (d *mysqlDriverWrapper) OpenConnector(name string) (driver.Connector, error) {
	//nolint:forcetypeassert // panic if d.driver does not implement driver.DriverContext
	connector, err := d.driver.(driver.DriverContext).OpenConnector(name)
	if err != nil {
		return nil, err
	}
	return &mysqlConnectorWrapper{connector: connector}, nil
}

var _ driver.DriverContext = &mysqlDriverWrapper{}
