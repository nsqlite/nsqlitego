package nsqlitedriver

import (
	"context"
	"database/sql/driver"
)

var _ driver.Driver = (*Driver)(nil)

// Driver implements database/sql/driver.Driver for NSQLite.
type Driver struct{}

// Open creates a new connection using the provided DSN (Data Source Name).
func (d *Driver) Open(connectionString string) (driver.Conn, error) {
	opts := NewConnectorOptions()
	opts.SetConnectionString(connectionString)

	conn := NewConnector(opts)
	return conn.Connect(context.Background())
}
