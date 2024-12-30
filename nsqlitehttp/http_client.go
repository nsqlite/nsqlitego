package nsqlitehttp

import (
	"net/http"

	"github.com/nsqlite/nsqlitego/connstr"
)

// httpClient is the underlying HTTP client used by the NSQLite client.
type httpClient struct {
	connStr *connstr.ConnStr
	httpc   *http.Client
}

// newHttpClient creates a new HTTP client.
func newHttpClient(connStr *connstr.ConnStr) *httpClient {
	return &httpClient{
		connStr: connStr,
		httpc:   http.DefaultClient,
	}
}

// NewRequest creates a new HTTP request.
func (hc *httpClient) NewRequest() *request {
	return newRequest(hc.connStr)
}
