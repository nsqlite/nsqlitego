package nsqlitedriver

import (
	"context"
	"database/sql/driver"
)

var _ driver.Tx = (*Tx)(nil)

// Tx represents a transaction.
type Tx struct {
	// conn is the connection associated with the transaction.
	conn *Conn
}

// Commit commits the transaction.
func (t *Tx) Commit() error {
	return t.conn.CommitTx(context.Background())
}

// Rollback rolls back the transaction.
func (t *Tx) Rollback() error {
	return t.conn.RollbackTx(context.Background())
}
