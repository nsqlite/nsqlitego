package nsqlitedsn

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
)

// ConnStr holds the NSQLite connection string divided into its parts.
type ConnStr struct {
	// Protocol can be either "http" or "https" (default is "http").
	Protocol string
	// Host is the IP address or domain name of the server without protocol and
	// port (default is "localhost").
	Host string
	// Port is the port number of the server (default is 9876).
	Port string
	// AuthToken is the authentication token sent to the server on every request
	// (optional).
	AuthToken string
}

// setDefaultsIfEmpty sets the default values for the connection string if they
// are empty.
func (c *ConnStr) setDefaultsIfEmpty() {
	if c.Protocol == "" {
		c.Protocol = "http"
	}

	if c.Host == "" {
		c.Host = "localhost"
	}

	if c.Port == "" {
		c.Port = "9876"
	}
}

// NewConnStrFromStr creates a new ConnStr from a connection string.
//
// The connection string must be in the format
// "protocol://host:port?authToken=token".
//
//   - The protocol must be either "http" or "https".
//   - The host is the IP address or domain name of the server.
//   - The port is the port number of the server (default is 9876).
//   - The authToken is the optional authentication token sent to the server on
//     every request.
//
// If the connection string is invalid, an error is returned.
func NewConnStrFromText(connStrText string) (*ConnStr, error) {
	parsedURL, err := url.Parse(connStrText)
	if err != nil {
		return &ConnStr{}, err
	}

	protocol := parsedURL.Scheme
	if protocol != "http" && protocol != "https" {
		return &ConnStr{}, errors.New("invalid protocol, must be http or https")
	}

	host := parsedURL.Hostname()
	if host == "" {
		return &ConnStr{}, errors.New("host is required")
	}

	port := parsedURL.Port()
	if port == "" {
		port = "9876"
	}

	return &ConnStr{
		Protocol:  protocol,
		Host:      host,
		Port:      port,
		AuthToken: parsedURL.Query().Get("authToken"),
	}, nil
}

// String returns the string representation of the connection string without
// the auth token.
func (c *ConnStr) String() string {
	c.setDefaultsIfEmpty()

	if c.AuthToken == "" {
		return c.Protocol + "://" + c.Host + ":" + c.Port
	}

	return c.Protocol + "://" + c.Host + ":" + c.Port + "?authToken=****"
}

// BaseUrlStr returns the full URL of the connection string without the auth
// token.
func (c *ConnStr) BaseUrlStr() string {
	c.setDefaultsIfEmpty()
	return c.Protocol + "://" + c.Host + ":" + c.Port
}

// CreateUrlStr returns a string URL from the connection string and the
// provided path.
//
// This does not include the auth token in the URL.
func (c *ConnStr) CreateUrlStr(path string) (string, error) {
	c.setDefaultsIfEmpty()

	parts := strings.Split(path, "?")
	query := ""
	if len(parts) > 1 {
		path = parts[0]
		query = parts[1]
	}

	joined, err := url.JoinPath(c.BaseUrlStr(), path)
	if err != nil {
		return "", fmt.Errorf("failed to join URL path: %w", err)
	}

	if query != "" {
		joined += "?" + query
	}

	return joined, nil
}

// CreateUrl returns an *url.URL from the connection string and the
// provided path.
//
// This does not include the auth token in the URL.
func (c *ConnStr) CreateUrl(path string) (*url.URL, error) {
	c.setDefaultsIfEmpty()

	joined, err := c.CreateUrlStr(path)
	if err != nil {
		return nil, err
	}

	parsed, err := url.Parse(joined)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	return parsed, nil
}
