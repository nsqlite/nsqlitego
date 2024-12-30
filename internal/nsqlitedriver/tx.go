package nsqlitedriver

import (
	"database/sql/driver"
	"fmt"

	"github.com/nsqlite/nsqlitego/nsqlitehttp"
)

var _ driver.Tx = (*Tx)(nil)

// Tx represents a transaction.
type Tx struct {
	client *nsqlitehttp.Client
	txId   string
}

// Commit commits the transaction.
func (t *Tx) Commit() error {
	resp, err := t.client.Query(nsqlitehttp.Query{
		Query: "COMMIT",
		TxId:  t.txId,
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
	resp, err := t.client.Query(nsqlitehttp.Query{
		Query: "ROLLBACK",
		TxId:  t.txId,
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
