package nsqlitedriver

import (
	"context"
	"database/sql/driver"
)

var (
	_ driver.Driver        = (*Driver)(nil)
	_ driver.DriverContext = (*Driver)(nil)
)

// Driver implements database/sql/driver.Driver for NSQLite.
type Driver struct{}

// Open creates a new connection using the provided connection string.
func (d *Driver) Open(connectionString string) (driver.Conn, error) {
	opts := NewConnectorOptions()
	opts.SetConnectionString(connectionString)

	conn := NewConnector(opts)
	return conn.Connect(context.Background())
}

// OpenConnector creates a new connector using the provided connection string.
func (d *Driver) OpenConnector(connectionString string) (driver.Connector, error) {
	opts := NewConnectorOptions()
	opts.SetConnectionString(connectionString)
	return NewConnector(opts), nil
}
