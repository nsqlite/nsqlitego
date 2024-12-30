package nsqlitehttp

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/nsqlite/nsqlitego/connstr"
)

// Client is an HTTP client for the NSQLite server.
type Client struct {
	httpc *httpClient
}

// NewClient creates a new NSQLite client.
func NewClient(connStr string) (*Client, error) {
	cStr, err := connstr.NewConnStrFromText(connStr)
	if err != nil {
		return nil, fmt.Errorf("invalid connection string: %v", err)
	}

	httpc := newHttpClient(cStr)
	return &Client{
		httpc: httpc,
	}, nil
}

// Ping sends a request to the server to check if it is alive. Returns an error
// if the server is not alive.
func (c *Client) Ping() error {
	req := c.httpc.NewRequest()
	req.SetPath("/health")

	res, err := req.Get()
	if err != nil {
		return fmt.Errorf("request failed: %v", err)
	}

	if strings.ToLower(res.Body) != "ok" {
		if len(res.Body) > 100 {
			res.Body = res.Body[:100] + "..."
		}
		return fmt.Errorf(
			`health check expected to return "OK" but got "%s"`, res.Body,
		)
	}

	if strings.ToLower(res.Headers.Get("x-server")) != "nsqlite" {
		return fmt.Errorf(
			`health check expected to return "NSQLite" in "X-Server" header but got "%s"`,
			res.Headers.Get("x-server"),
		)
	}

	return nil
}

// IsHealthy checks if the server is alive. Returns an error if the server is not
// healthy.
func (c *Client) IsHealthy() error {
	return c.Ping()
}

// Version returns the version of the NSQLite server.
func (c *Client) Version() (string, error) {
	req := c.httpc.NewRequest()
	req.SetPath("/version")

	res, err := req.Get()
	if err != nil {
		return "", fmt.Errorf("failed to get remote NSQLite server version: %v", err)
	}

	if res.Status == http.StatusUnauthorized {
		return "", fmt.Errorf("authentication failed, please check your credentials")
	}

	if res.Status != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", res.Status)
	}

	return res.Body, nil
}

// QueryResponseType represents the type of a query response.
type QueryResponseType string

const (
	QueryResponseError    QueryResponseType = "error"
	QueryResponseOK       QueryResponseType = "ok"
	QueryResponseBegin    QueryResponseType = "begin"
	QueryResponseCommit   QueryResponseType = "commit"
	QueryResponseRollback QueryResponseType = "rollback"
	QueryResponseWrite    QueryResponseType = "write"
	QueryResponseRead     QueryResponseType = "read"
)

// QueryResponse represents the response of a query sent to the remote NSQLite server.
type QueryResponse struct {
	Type QueryResponseType `json:"type"`
	Time float64           `json:"time"`

	// For read queries
	Columns []string `json:"columns"`
	Types   []string `json:"types"`
	Values  [][]any  `json:"values"`

	// For write queries
	LastInsertID int64 `json:"lastInsertId"`
	RowsAffected int64 `json:"rowsAffected"`

	// For begin, commit, and rollback
	TxId string `json:"txId"`

	// For errors
	Error string `json:"error"`
}

// Query represents the parameters to send a query to the remote server.
type Query struct {
	// Query is the SQL query to send (required).
	Query string `json:"query"`
	// Params are the parameters to send with a parameterized query (optional).
	Params []any `json:"params"`
	// TxId is used to send the query in the context of a transaction (optional).
	TxId string `json:"txId,omitempty"`
}

// Query sends a query to the remote server and returns the response.
func (c *Client) Query(q Query) (QueryResponse, error) {
	req := c.httpc.NewRequest()
	req.SetPath("/query")
	req.SetHeader("Content-Type", "application/json")

	httpRes, err := req.Post(q)
	if err != nil {
		return QueryResponse{}, fmt.Errorf("failed to send query: %v", err)
	}

	res := QueryResponse{}
	completeRes := struct {
		Results []QueryResponse `json:"results"`
	}{}

	if err := json.Unmarshal([]byte(httpRes.Body), &completeRes); err != nil {
		return res, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(completeRes.Results) == 0 {
		return res, fmt.Errorf("empty response")
	}

	return completeRes.Results[0], nil
}
