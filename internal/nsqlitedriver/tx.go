package nsqlitedriver

import (
	"database/sql/driver"
	"fmt"

	"github.com/nsqlite/nsqlitego/nsqlitehttp"
)

var _ driver.Tx = (*Tx)(nil)

// Tx represents a transaction.
type Tx struct {
	// conn is the connection associated with the transaction.
	conn *Conn
}

// Commit commits the transaction.
func (t *Tx) Commit() error {
	defer t.resetConnectionTxId()

	resp, err := t.conn.client.Query(nsqlitehttp.Query{
		Query: "COMMIT",
		TxId:  t.conn.txId,
	})
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	if resp.Type == nsqlitehttp.QueryResponseError {
		return fmt.Errorf("failed to commit transaction: %s", resp.Error)
	}
	if resp.Type != nsqlitehttp.QueryResponseCommit {
		return fmt.Errorf("unexpected response type: %s", resp.Type)
	}
	return nil
}

// Rollback rolls back the transaction.
func (t *Tx) Rollback() error {
	defer t.resetConnectionTxId()

	resp, err := t.conn.client.Query(nsqlitehttp.Query{
		Query: "ROLLBACK",
		TxId:  t.conn.txId,
	})
	if err != nil {
		return fmt.Errorf("failed to rollback transaction: %w", err)
	}
	if resp.Type == nsqlitehttp.QueryResponseError {
		return fmt.Errorf("failed to rollback transaction: %s", resp.Error)
	}
	if resp.Type != nsqlitehttp.QueryResponseRollback {
		return fmt.Errorf("unexpected response type: %s", resp.Type)
	}
	return nil
}

// resetConnectionTxId resets the transaction ID for the connection.
func (t *Tx) resetConnectionTxId() {
	t.conn.txId = ""
}
