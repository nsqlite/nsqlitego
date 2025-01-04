package nsqlitehttp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/nsqlite/nsqlitego/nsqlitedsn"
)

// Client is an HTTP client for the NSQLite server.
type Client struct {
	connStr *nsqlitedsn.ConnStr
	httpc   *http.Client
}

// NewClient creates a new NSQLite client.
func NewClient(connStr string) (*Client, error) {
	cStr, err := nsqlitedsn.NewConnStrFromText(connStr)
	if err != nil {
		return nil, fmt.Errorf("NewClient: invalid connection string: %v", err)
	}

	return &Client{
		connStr: cStr,
		httpc:   http.DefaultClient,
	}, nil
}

// newRequest creates a new HTTP request with the NSQLite URL and authentication
func (c *Client) newRequest(ctx context.Context, method string, path string, body io.Reader) (*http.Request, error) {
	url, err := c.connStr.CreateUrlStr(path)
	if err != nil {
		return nil, fmt.Errorf("newRequest: failed to create URL: %w", err)
	}

	request, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("newRequest: failed to create request: %w", err)
	}
	request.Header.Set("Content-Type", "application/json")

	if c.connStr.AuthToken != "" {
		request.Header.Set("Authorization", c.connStr.AuthToken)
	}

	return request, nil
}

// Ping sends a request to the server to check if it is alive. Returns an error
// if the server is not alive.
func (c *Client) Ping(ctx context.Context) error {
	request, err := c.newRequest(ctx, http.MethodGet, "/health", nil)
	if err != nil {
		return fmt.Errorf("Ping: failed to create request: %w", err)
	}

	response, err := c.httpc.Do(request)
	if err != nil {
		return fmt.Errorf("Ping: failed to send request: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("Ping: unwanted response status %s", response.Status)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("Ping: failed to read response body: %w", err)
	}
	bodyStr := string(body)

	if strings.ToLower(bodyStr) != "ok" {
		if len(bodyStr) > 100 {
			bodyStr = bodyStr[:100] + "..."
		}
		return fmt.Errorf(
			`Ping: health check expected to return "OK" but got "%s"`, bodyStr,
		)
	}

	if strings.ToLower(response.Header.Get("X-Server")) != "nsqlite" {
		return fmt.Errorf(
			`Ping: health check expected to return NSQLite in X-Server header but got "%s"`,
			response.Header.Get("X-Server"),
		)
	}

	return nil
}

// IsHealthy checks if the server is alive. Returns an error if the server is
// not healthy.
func (c *Client) IsHealthy(ctx context.Context) error {
	return c.Ping(ctx)
}

// Version returns the version of the NSQLite server.
func (c *Client) Version(ctx context.Context) (string, error) {
	request, err := c.newRequest(ctx, http.MethodGet, "/version", nil)
	if err != nil {
		return "", fmt.Errorf("Version: failed to create request: %w", err)
	}

	response, err := c.httpc.Do(request)
	if err != nil {
		return "", fmt.Errorf("Version: failed to send request: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode == http.StatusUnauthorized {
		return "", fmt.Errorf("Version: authentication failed, please check your credentials")
	}

	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Version: unwanted response status: %s", response.Status)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return "", fmt.Errorf("Version: failed to read response body: %w", err)
	}

	return string(body), nil
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

// QueryResponse represents the response of a query sent to the remote NSQLite
// server.
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
func (c *Client) Query(ctx context.Context, q Query) (QueryResponse, error) {
	requestBody, err := json.Marshal(q)
	if err != nil {
		return QueryResponse{}, fmt.Errorf("Query: failed to marshal request body: %w", err)
	}

	request, err := c.newRequest(ctx, http.MethodPost, "/query", bytes.NewReader(requestBody))
	if err != nil {
		return QueryResponse{}, fmt.Errorf("Query: failed to create request: %w", err)
	}

	response, err := c.httpc.Do(request)
	if err != nil {
		return QueryResponse{}, fmt.Errorf("Query: failed to send request: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode == http.StatusUnauthorized {
		return QueryResponse{}, fmt.Errorf("Query: authentication failed, please check your credentials")
	}

	if response.StatusCode != http.StatusOK {
		return QueryResponse{}, fmt.Errorf("Query: unwanted response status: %s", response.Status)
	}

	result := struct {
		Results []QueryResponse `json:"results"`
	}{}
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		return QueryResponse{}, fmt.Errorf("Query: failed to decode response: %w", err)
	}

	if len(result.Results) == 0 {
		return QueryResponse{}, fmt.Errorf("Query: empty response")
	}

	return result.Results[0], nil
}
