package database

import "database/sql/driver"

type mysqlDriverWrapper struct {
	driver driver.Driver
}

func newMySQLDriverWrapper(driverToWrap driver.Driver) driver.Driver {
	return &mysqlDriverWrapper{driver: driverToWrap}
}

func (d *mysqlDriverWrapper) Open(name string) (driver.Conn, error) {
	conn, err := d.driver.Open(name)
	if err != nil {
		return nil, err
	}
	return &mysqlConnWrapper{conn: conn}, nil
}

var _ driver.Driver = &mysqlDriverWrapper{}

func (d *mysqlDriverWrapper) OpenConnector(name string) (driver.Connector, error) {
	connector, err := d.driver.(driver.DriverContext).OpenConnector(name)
	if err != nil {
		return nil, err
	}
	return &mysqlConnectorWrapper{connector: connector}, nil
}

var _ driver.DriverContext = &mysqlDriverWrapper{}
