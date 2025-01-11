package nsqlitedriver

import (
	"context"
	"database/sql/driver"
	"fmt"
	"io"
	"strings"

	"github.com/nsqlite/nsqlitego/nsqlitehttp"
)

var (
	_ driver.Stmt                           = (*Stmt)(nil)
	_ driver.StmtExecContext                = (*Stmt)(nil)
	_ driver.StmtQueryContext               = (*Stmt)(nil)
	_ driver.RowsColumnTypeDatabaseTypeName = (*QueryRows)(nil)
)

// Stmt represents a prepared statement.
type Stmt struct {
	// conn is the connection associated with the statement.
	conn *Conn
	// query is the query string to be executed with NSQLite.
	query string
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

// ExecContext executes a query without returning rows (e.g., INSERT, UPDATE).
func (s *Stmt) ExecContext(ctx context.Context, args []driver.NamedValue) (driver.Result, error) {
	params := convertNamedValueToAnyArray(args)
	resp, err := s.conn.client.SendQuery(ctx, nsqlitehttp.Query{
		Query:  s.query,
		Params: params,
		TxId:   s.conn.txId,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	if resp.Type == nsqlitehttp.QueryResponseError {
		return nil, fmt.Errorf("failed to execute query: %s", resp.Error)
	}
	return &ExecResult{
		lastInsertId: resp.LastInsertID,
		rowsAffected: resp.RowsAffected,
	}, nil
}

// Exec executes a query without returning rows (e.g., INSERT, UPDATE, DELETE).
func (s *Stmt) Exec(args []driver.Value) (driver.Result, error) {
	return s.ExecContext(context.Background(), convertValueToNamedValue(args))
}

// QueryRows represents a set of query results.
type QueryRows struct {
	columns   []string
	types     []string
	values    [][]any
	valuesLen int
	rowIdx    int
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
	if r.rowIdx >= r.valuesLen {
		return io.EOF
	}

	row := r.values[r.rowIdx]
	for i, val := range row {
		dest[i] = val
	}

	r.rowIdx++
	return nil
}

// ColumnTypeDatabaseTypeName returns the database type name for the column.
func (r *QueryRows) ColumnTypeDatabaseTypeName(index int) string {
	if index < 0 {
		return ""
	}
	if index >= len(r.types) {
		return ""
	}
	return strings.ToUpper(r.types[index])
}

// QueryContext executes a query that returns rows (e.g., SELECT).
func (s *Stmt) QueryContext(ctx context.Context, args []driver.NamedValue) (driver.Rows, error) {
	params := convertNamedValueToAnyArray(args)
	resp, err := s.conn.client.SendQuery(ctx, nsqlitehttp.Query{
		Query:  s.query,
		Params: params,
		TxId:   s.conn.txId,
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
		columns:   resp.Columns,
		types:     resp.Types,
		values:    resp.Values,
		valuesLen: len(resp.Values),
		rowIdx:    0,
	}, nil
}

// Query executes a query that returns rows (e.g., SELECT).
func (s *Stmt) Query(args []driver.Value) (driver.Rows, error) {
	return s.QueryContext(context.Background(), convertValueToNamedValue(args))
}

// convertNamedValueToAnyArray converts driver.NamedValue arguments to []any.
func convertNamedValueToAnyArray(args []driver.NamedValue) []any {
	converted := make([]any, len(args))
	for i, arg := range args {
		converted[i] = arg.Value
	}
	return converted
}

// convertValueToNamedValue converts driver.Value arguments to
// []driver.NamedValue.
func convertValueToNamedValue(args []driver.Value) []driver.NamedValue {
	if len(args) == 0 {
		return nil
	}
	namedArgs := make([]driver.NamedValue, len(args))
	for i, arg := range args {
		namedArgs[i] = driver.NamedValue{
			Ordinal: i + 1,
			Value:   arg,
		}
	}
	return namedArgs
}
