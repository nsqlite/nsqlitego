package nsqlitedriver

import (
	"context"
	"database/sql/driver"
	"fmt"

	"github.com/nsqlite/nsqlitego/nsqlitehttp"
)

// ConnectorOptions represents a set of options to configure a NSQLite
// connector.
type ConnectorOptions struct {
	connectionString string
}

// NewConnectorOptions returns a new set of options to configure a NSQLite
// connector.
func NewConnectorOptions() *ConnectorOptions {
	return &ConnectorOptions{}
}

// SetConnectionString sets the connection string to use when connecting to the
// database.
func (co *ConnectorOptions) SetConnectionString(
	connectionString string,
) *ConnectorOptions {
	co.connectionString = connectionString
	return co
}

// NewConnector returns a new NSQLite connector compatible with
// database/sql.OpenDB
//
// It accepts a number of options to configure the connector.
func NewConnector(opts *ConnectorOptions) driver.Connector {
	connector := &nsqliteConnector{
		connString: opts.connectionString,
	}
	return connector
}

var _ driver.Connector = (*nsqliteConnector)(nil)

// nsqliteConnector represents a driver in a fixed configuration and can create
// any number of equivalent Conns for use by multiple goroutines.
type nsqliteConnector struct {
	connString string
}

// Connect returns a connection to the database.
func (c *nsqliteConnector) Connect(_ context.Context) (driver.Conn, error) {
	client, err := nsqlitehttp.NewClient(c.connString)
	if err != nil {
		return nil, fmt.Errorf("error creating NSQLite client: %w", err)
	}
	return &Conn{client: client}, nil
}

// Driver returns the underlying Driver of the Connector
func (c *nsqliteConnector) Driver() driver.Driver {
	return &Driver{}
}
