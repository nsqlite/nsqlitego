package nsclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/nsqlite/nsqlitego/connstr"
)

// request represents an HTTP request.
type request struct {
	connStr *connstr.ConnStr
	method  string
	path    string
	header  http.Header
	body    any
}

// newRequest creates a new HTTP request using the following defaults:
//
//   - method: GET
//   - baseUrl: provided by the connection string
//   - path: /
//   - header: Content-Type: application/json; Authorization: <token from conn string>
func newRequest(connStr *connstr.ConnStr) *request {
	header := http.Header{
		"Content-Type":  []string{"application/json"},
		"Authorization": []string{connStr.AuthToken},
	}

	return &request{
		connStr: connStr,
		method:  http.MethodGet,
		path:    "/",
		header:  header,
	}
}

// makeRequest creates an HTTP request based on the request parameters.
func (r *request) makeRequest() (*http.Request, error) {
	url, err := r.connStr.CreateUrl(r.path)
	if err != nil {
		return nil, fmt.Errorf("failed to create request URL: %w", err)
	}

	req := &http.Request{
		Method: r.method,
		URL:    url,
	}

	return req, nil
}

// SetPath sets the URL path.
func (r *request) SetPath(path string) *request {
	r.path = path
	return r
}

// SetHeader sets the HTTP header. If the header already exists, it will be overwritten.
func (r *request) SetHeader(key, value string) *request {
	r.header.Set(key, value)
	return r
}

// DelHeader deletes the HTTP header.
func (r *request) DelHeader(key string) *request {
	r.header.Del(key)
	return r
}

// Response represents an HTTP response.
type Response struct {
	IsJson       bool
	Body         string
	Status       int
	StatusText   string
	Headers      http.Header
	HttpResponse *http.Response
}

// Get fires a GET request with the current request settings.
func (r *request) Get() (*Response, error) {
	r.method = http.MethodGet
	return r.do()
}

// Post fires a POST request with the current request settings.
func (r *request) Post(body any) (*Response, error) {
	r.method = http.MethodPost
	r.body = body
	return r.do()
}

// do sends the HTTP request.
func (r *request) do() (*Response, error) {
	url, err := r.connStr.CreateUrl(r.path)
	if err != nil {
		return nil, fmt.Errorf("failed to create request URL: %w", err)
	}

	// Marshal the body if it exists, use JSON when possible.
	var body []byte = nil
	if r.body != nil {
		switch v := r.body.(type) {
		case []byte:
			body = v
		case string:
			body = []byte(v)
		case map[string]string:
			b, err := json.Marshal(v)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal body: %w", err)
			}
			body = b
		default:
			b, err := json.Marshal(v)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal body: %w", err)
			}
			body = b
		}
	}

	httpReq := &http.Request{
		Method: r.method,
		URL:    url,
		Header: r.header,
		Body:   http.NoBody,
	}

	if body != nil && len(body) > 0 {
		httpReq.Body = io.NopCloser(bytes.NewReader(body))
	}

	httpResp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send HTTP %s request: %w", r.method, err)
	}

	bodyb, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed reading response body: %w", err)
	}

	isJson := strings.Contains(httpResp.Header.Get("Content-Type"), "application/json")
	return &Response{
		IsJson:       isJson,
		Body:         string(bodyb),
		Status:       httpResp.StatusCode,
		StatusText:   httpResp.Status,
		Headers:      httpResp.Header,
		HttpResponse: httpResp,
	}, nil
}
