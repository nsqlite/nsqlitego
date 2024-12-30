package nsqlitehttp

import (
	"net/http"

	"github.com/nsqlite/nsqlitego/nsqlitedsn"
)

// httpClient is the underlying HTTP client used by the NSQLite client.
type httpClient struct {
	connStr *nsqlitedsn.ConnStr
	httpc   *http.Client
}

// newHttpClient creates a new HTTP client.
func newHttpClient(connStr *nsqlitedsn.ConnStr) *httpClient {
	return &httpClient{
		connStr: connStr,
		httpc:   http.DefaultClient,
	}
}

// NewRequest creates a new HTTP request.
func (hc *httpClient) NewRequest() *request {
	return newRequest(hc.connStr)
}
