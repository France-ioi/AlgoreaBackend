package database

import (
	"context"
	"database/sql/driver"
)

type mysqlConnectorWrapper struct {
	connector driver.Connector
	driver    *mysqlDriverWrapper
}

func (c *mysqlConnectorWrapper) Connect(ctx context.Context) (driver.Conn, error) {
	conn, err := c.connector.Connect(ctx)
	if err != nil {
		return nil, err
	}
	wrapped := &mysqlConnWrapper{conn: conn, sessionParams: getSessionParams()}
	if err := wrapped.applySessionParams(ctx); err != nil {
		_ = wrapped.Close()
		return nil, err
	}
	return wrapped, nil
}

func (c *mysqlConnectorWrapper) Driver() driver.Driver {
	return c.driver
}

var _ driver.Connector = &mysqlConnectorWrapper{}
