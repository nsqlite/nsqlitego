package nsqlitehttp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/nsqlite/nsqlitego/nsqlitedsn"
)

// Client is an HTTP client for the NSQLite server.
type Client struct {
	connStr *nsqlitedsn.ConnStr
	httpc   *http.Client
}

// ClientOption is a function that configures a Client.
type ClientOption func(*Client) error

// WithHTTPTimeout sets the timeout for the default NSQLite HTTP client. Default is 30 seconds.
func WithHTTPTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) error {
		c.httpc.Timeout = timeout
		return nil
	}
}

// WithHTTPTransport sets the transport for the default NSQLite HTTP client. The default is
// http.DefaultTransport with MaxIdleConns, MaxConnsPerHost, and MaxIdleConnsPerHost set to 100.
func WithHTTPTransport(transport *http.Transport) ClientOption {
	return func(c *Client) error {
		c.httpc.Transport = transport
		return nil
	}
}

// WithHTTPClient entirely replaces the default NSQLite HTTP client with a custom one.
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *Client) error {
		c.httpc = httpClient
		return nil
	}
}

// NewClient creates a new NSQLite client.
func NewClient(connectionString string, options ...ClientOption) (*Client, error) {
	connStr, err := nsqlitedsn.NewConnStrFromText(connectionString)
	if err != nil {
		return nil, fmt.Errorf("invalid connection string: %v", err)
	}

	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.MaxIdleConns = 100
	transport.MaxIdleConnsPerHost = 100

	httpClient := &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
	}

	client := &Client{
		connStr: connStr,
		httpc:   httpClient,
	}

	for idx, opt := range options {
		if err := opt(client); err != nil {
			return nil, fmt.Errorf("failed to apply option %d: %w", idx+1, err)
		}
	}

	return client, nil
}

// newRequest creates a new HTTP request with the NSQLite URL and authentication
func (c *Client) newRequest(ctx context.Context, method string, path string, body io.Reader) (*http.Request, error) {
	url, err := c.connStr.CreateUrlStr(path)
	if err != nil {
		return nil, fmt.Errorf("failed to create URL: %w", err)
	}

	request, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	request.Header.Set("Content-Type", "application/json")

	if c.connStr.AuthToken != "" {
		request.Header.Set("Authorization", c.connStr.AuthToken)
	}

	return request, nil
}

// SendPing sends a request to the server to check if it is alive. Returns an error
// if the server is not alive.
func (c *Client) SendPing(ctx context.Context) error {
	request, err := c.newRequest(ctx, http.MethodGet, "/health", nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	response, err := c.httpc.Do(request)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("unwanted response status %s", response.Status)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}
	bodyStr := string(body)

	if strings.ToLower(bodyStr) != "ok" {
		if len(bodyStr) > 100 {
			bodyStr = bodyStr[:100] + "..."
		}
		return fmt.Errorf(
			`health check expected to return "OK" but got "%s"`, bodyStr,
		)
	}

	if strings.ToLower(response.Header.Get("X-Server")) != "nsqlite" {
		return fmt.Errorf(
			`health check expected to return NSQLite in X-Server header but got "%s"`,
			response.Header.Get("X-Server"),
		)
	}

	return nil
}

// IsHealthy checks if the server is alive. Returns an error if the server is
// not healthy.
func (c *Client) IsHealthy(ctx context.Context) error {
	return c.SendPing(ctx)
}

// GetVersion returns the version of the NSQLite server.
func (c *Client) GetVersion(ctx context.Context) (string, error) {
	request, err := c.newRequest(ctx, http.MethodGet, "/version", nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	response, err := c.httpc.Do(request)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode == http.StatusUnauthorized {
		return "", fmt.Errorf("authentication failed, please check your credentials")
	}

	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unwanted response status: %s", response.Status)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
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

// SendQueries sends one or more queries to the remote server and returns the responses in same order.
func (c *Client) SendQueries(ctx context.Context, queries []Query) ([]QueryResponse, error) {
	requestBody, err := json.Marshal(queries)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	request, err := c.newRequest(ctx, http.MethodPost, "/query", bytes.NewReader(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	response, err := c.httpc.Do(request)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode == http.StatusUnauthorized {
		return nil, fmt.Errorf("authentication failed, please check your credentials")
	}

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unwanted response status: %s", response.Status)
	}

	result := struct {
		Results []QueryResponse `json:"results"`
	}{}

	decoder := json.NewDecoder(response.Body)
	decoder.UseNumber()
	if err := decoder.Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(result.Results) == 0 {
		return nil, fmt.Errorf("empty response")
	}

	return result.Results, nil
}

// SendQuery sends a single query to the remote server and returns the response.
func (c *Client) SendQuery(ctx context.Context, query Query) (QueryResponse, error) {
	responses, err := c.SendQueries(ctx, []Query{query})
	if err != nil {
		return QueryResponse{}, err
	}

	return responses[0], nil
}

// Stats represents the database stats returned by the server.
type Stats struct {
	StartedAt          string      `json:"startedAt"`
	Uptime             string      `json:"uptime"`
	QueuedWrites       int64       `json:"queuedWrites"`
	QueuedHTTPRequests int64       `json:"queuedHttpRequests"`
	Totals             StatsTotals `json:"totals"`
	Stats              []StatsStat `json:"stats"`
}

type StatsTotals struct {
	Reads        int64 `json:"reads"`
	Writes       int64 `json:"writes"`
	Begins       int64 `json:"begins"`
	Commits      int64 `json:"commits"`
	Rollbacks    int64 `json:"rollbacks"`
	Errors       int64 `json:"errors"`
	HTTPRequests int64 `json:"httpRequests"`
}

type StatsStat struct {
	Minute       string `json:"minute"`
	Reads        int64  `json:"reads"`
	Writes       int64  `json:"writes"`
	Begins       int64  `json:"begins"`
	Commits      int64  `json:"commits"`
	Rollbacks    int64  `json:"rollbacks"`
	Errors       int64  `json:"errors"`
	HTTPRequests int64  `json:"httpRequests"`
}

// GetStats returns the database stats from the server.
func (c *Client) GetStats(ctx context.Context) (Stats, error) {
	request, err := c.newRequest(ctx, http.MethodGet, "/stats", nil)
	if err != nil {
		return Stats{}, fmt.Errorf("failed to create request: %w", err)
	}

	response, err := c.httpc.Do(request)
	if err != nil {
		return Stats{}, fmt.Errorf("failed to send request: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode == http.StatusUnauthorized {
		return Stats{}, fmt.Errorf("authentication failed, please check your credentials")
	}

	if response.StatusCode != http.StatusOK {
		return Stats{}, fmt.Errorf("unwanted response status: %s", response.Status)
	}

	result := Stats{}
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		return Stats{}, fmt.Errorf("failed to decode response: %w", err)
	}

	return result, nil
}
