// Package nsqlitego provides a NSQLite driver compatible with the database/sql package.
package nsqlitego

import (
	"database/sql"
	"database/sql/driver"

	"github.com/nsqlite/nsqlitego/internal/nsqlitedriver"
)

// init registers the NSQLite driver for database/sql.
func init() {
	sql.Register("nsqlite", &Driver{})
}

// Driver implements database/sql/driver.Driver for NSQLite.
type Driver = nsqlitedriver.Driver

// NewConnectorOptions returns a new set of options to configure a NSQLite connector.
func NewConnectorOptions() *nsqlitedriver.ConnectorOptions {
	return nsqlitedriver.NewConnectorOptions()
}

// NewConnector returns a new NSQLite connector compatible with database/sql.OpenDB
//
// It accepts a number of options to configure the connector.
func NewConnector(opts *nsqlitedriver.ConnectorOptions) driver.Connector {
	return nsqlitedriver.NewConnector(opts)
}
