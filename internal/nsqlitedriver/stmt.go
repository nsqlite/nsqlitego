package nsqlitedriver

import (
	"database/sql/driver"
	"fmt"

	"github.com/nsqlite/nsqlitego/nsqlitehttp"
)

var _ driver.Stmt = (*Stmt)(nil)

// Stmt represents a prepared statement.
type Stmt struct {
	client *nsqlitehttp.Client
	query  string
	txId   string // Optional transaction ID
}

// Close releases resources associated with the statement.
func (s *Stmt) Close() error {
	return nil
}

// NumInput returns the number of placeholder parameters for the statement.
// -1 indicates that the number is unknown or dynamic.
func (s *Stmt) NumInput() int {
	return -1
}

// ExecResult represents the result of a query.
type ExecResult struct {
	lastInsertId int64
	rowsAffected int64
}

// LastInsertId returns the ID of the last inserted row.
func (r *ExecResult) LastInsertId() (int64, error) {
	return r.lastInsertId, nil
}

// RowsAffected returns the number of rows affected by the query.
func (r *ExecResult) RowsAffected() (int64, error) {
	return r.rowsAffected, nil
}

// Exec executes a query without returning rows (e.g., INSERT, UPDATE, DELETE).
func (s *Stmt) Exec(args []driver.Value) (driver.Result, error) {
	params := convertArgs(args)
	resp, err := s.client.Query(nsqlitehttp.Query{
		Query:  s.query,
		Params: params,
		TxId:   s.txId,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	if resp.Type == nsqlitehttp.QueryResponseError {
		return nil, fmt.Errorf("failed to execute query: %s", resp.Error)
	}
	if resp.Type != nsqlitehttp.QueryResponseWrite {
		return nil, fmt.Errorf("unexpected response type: %s", resp.Type)
	}
	return &ExecResult{
		lastInsertId: resp.LastInsertID,
		rowsAffected: resp.RowsAffected,
	}, nil
}

// QueryRows represents a set of query results.
type QueryRows struct {
	columns []string
	values  [][]any
	rowIdx  int
}

// Columns returns the column names.
func (r *QueryRows) Columns() []string {
	return r.columns
}

// Close releases resources associated with the rows.
func (r *QueryRows) Close() error {
	return nil
}

// Next prepares the next row for reading.
func (r *QueryRows) Next(dest []driver.Value) error {
	if r.rowIdx >= len(r.values) {
		return driver.ErrSkip
	}

	row := r.values[r.rowIdx]
	for i, val := range row {
		dest[i] = val
	}
	r.rowIdx++
	return nil
}

// Query executes a query that returns rows (e.g., SELECT).
func (s *Stmt) Query(args []driver.Value) (driver.Rows, error) {
	params := convertArgs(args)
	resp, err := s.client.Query(nsqlitehttp.Query{
		Query:  s.query,
		Params: params,
		TxId:   s.txId,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	if resp.Type == nsqlitehttp.QueryResponseError {
		return nil, fmt.Errorf("failed to execute query: %s", resp.Error)
	}
	if resp.Type != nsqlitehttp.QueryResponseRead {
		return nil, fmt.Errorf("unexpected response type: %s", resp.Type)
	}
	return &QueryRows{
		columns: resp.Columns,
		values:  resp.Values,
	}, nil
}

// Helper function to convert driver.Value arguments to []any.
func convertArgs(args []driver.Value) []any {
	converted := make([]any, len(args))
	for i, arg := range args {
		converted[i] = arg
	}
	return converted
}
