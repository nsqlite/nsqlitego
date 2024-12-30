package nsqlitedriver

import (
	"context"
	"database/sql/driver"
	"fmt"

	"github.com/nsqlite/nsqlitego/nsqlitehttp"
)

var (
	_ driver.Conn               = (*Conn)(nil)
	_ driver.ConnBeginTx        = (*Conn)(nil)
	_ driver.ConnPrepareContext = (*Conn)(nil)
	_ driver.Pinger             = (*Conn)(nil)
	_ driver.SessionResetter    = (*Conn)(nil)
	_ driver.Validator          = (*Conn)(nil)
)

// Conn represents a connection to the NSQLite server.
type Conn struct {
	client *nsqlitehttp.Client
}

// Prepare creates a prepared statement with the given query.
func (c *Conn) Prepare(query string) (driver.Stmt, error) {
	return c.PrepareContext(context.Background(), query)
}

// PrepareContext creates a prepared statement with the given query and context.
func (c *Conn) PrepareContext(_ context.Context, query string) (driver.Stmt, error) {
	return &Stmt{
		client: c.client,
		query:  query,
	}, nil
}

// Close closes the connection (no-op for an API-based client).
func (c *Conn) Close() error {
	// TODO: rollback any open transactions
	return nil
}

// Begin starts a new transaction.
func (c *Conn) Begin() (driver.Tx, error) {
	return c.BeginTx(context.Background(), driver.TxOptions{})
}

// BeginTx starts a new transaction with the provided context.
func (c *Conn) BeginTx(_ context.Context, opts driver.TxOptions) (driver.Tx, error) {
	resp, err := c.client.Query(nsqlitehttp.Query{
		Query: "BEGIN;",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	if resp.Type == nsqlitehttp.QueryResponseError {
		return nil, fmt.Errorf("failed to begin transaction: %s", resp.Error)
	}
	if resp.Type != nsqlitehttp.QueryResponseBegin {
		return nil, fmt.Errorf("unexpected response type: %s", resp.Type)
	}
	return &Tx{
		client: c.client,
		txId:   resp.TxId,
	}, nil
}

// Ping verifies that the connection is still alive.
func (c *Conn) Ping(_ context.Context) error {
	return c.client.Ping()
}

// ResetSession resets the session state used when the connection was used before and needs
// to be reused.
func (c *Conn) ResetSession(_ context.Context) error {
	// TODO: Rollback transaction if any
	return nil
}

// IsValid is called prior to placing the connection into the connection pool. The connection
// will be discarded if false is returned.
func (c *Conn) IsValid() bool {
	return c.client.IsHealthy() == nil
}
