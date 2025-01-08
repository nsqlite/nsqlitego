// Package nsqlitego provides a NSQLite driver compatible with the database/sql
// package.
package nsqlitego

import (
	"database/sql"
	"database/sql/driver"

	"github.com/nsqlite/nsqlitego/internal/nsqlitedriver"
	"github.com/nsqlite/nsqlitego/nsqlitehttp"
)

// init registers the NSQLite driver for database/sql.
func init() {
	sql.Register("nsqlite", &Driver{})
}

// Driver implements database/sql/driver.Driver for NSQLite.
type Driver = nsqlitedriver.Driver

// NewConnector returns a new NSQLite connector compatible with
// database/sql.OpenDB
func NewConnector(nsqliteHTTPClient *nsqlitehttp.Client) driver.Connector {
	return nsqlitedriver.NewConnector(nsqliteHTTPClient)
}
