// Package gofish provides a minimal Go client for HPE iLO 7 Redfish APIs,
// focused on the "Core compute" and "Chassis & hardware" resource groups.
package gofish

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// ODataID represents a Redfish "@odata.id" navigation link.
type ODataID struct {
	ODataID string `json:"@odata.id"`
}

// ClientConfig holds the parameters needed to connect to an iLO Redfish
// service endpoint.
type ClientConfig struct {
	// Endpoint is the base URL of the Redfish service, e.g. "https://ilo.example.com".
	Endpoint string
	// Username and Password are used for HTTP Basic authentication.
	Username string
	Password string
	// InsecureSkipVerify disables TLS certificate verification. iLO
	// deployments frequently use self-signed certificates.
	InsecureSkipVerify bool
	// HTTPClient allows callers to supply a pre-configured *http.Client.
	// When nil, a client with a sane default timeout is created.
	HTTPClient *http.Client
	// Timeout is used only when HTTPClient is nil. Defaults to 30s.
	Timeout time.Duration
}

// Client is a lightweight Redfish HTTP client bound to a single iLO
// service endpoint.
type Client struct {
	endpoint   string
	username   string
	password   string
	httpClient *http.Client
}

// NewClient creates a Client from the given configuration.
func NewClient(cfg ClientConfig) (*Client, error) {
	if cfg.Endpoint == "" {
		return nil, fmt.Errorf("gofish: endpoint must not be empty")
	}

	endpoint := strings.TrimRight(cfg.Endpoint, "/")

	httpClient := cfg.HTTPClient
	if httpClient == nil {
		timeout := cfg.Timeout
		if timeout == 0 {
			timeout = 30 * time.Second
		}
		httpClient = &http.Client{
			Timeout: timeout,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: cfg.InsecureSkipVerify}, //nolint:gosec // opt-in for self-signed iLO certs
			},
		}
	}

	return &Client{
		endpoint:   endpoint,
		username:   cfg.Username,
		password:   cfg.Password,
		httpClient: httpClient,
	}, nil
}

// Endpoint returns the base URL this client is bound to.
func (c *Client) Endpoint() string {
	return c.endpoint
}

// resolveURL turns a path (which may be an absolute "@odata.id" style path
// such as "/redfish/v1/Systems/1") into a full URL against this client's
// endpoint. Already-absolute URLs are returned unchanged.
func (c *Client) resolveURL(path string) string {
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return path
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return c.endpoint + path
}

// HTTPError is returned when the Redfish service responds with a non-2xx
// status code.
type HTTPError struct {
	StatusCode int
	Method     string
	URL        string
	Body       []byte
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("gofish: %s %s: unexpected status %d: %s", e.Method, e.URL, e.StatusCode, string(e.Body))
}

// doRequest issues an HTTP request against the Redfish service and decodes
// a JSON response body into out (when out is non-nil).
func (c *Client) doRequest(ctx context.Context, method, path string, body interface{}, out interface{}) error {
	url := c.resolveURL(path)

	var reqBody io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("gofish: marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return fmt.Errorf("gofish: build request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.username != "" || c.password != "" {
		req.SetBasicAuth(c.username, c.password)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("gofish: %s %s: %w", method, url, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("gofish: read response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return &HTTPError{StatusCode: resp.StatusCode, Method: method, URL: url, Body: respBody}
	}

	if out != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, out); err != nil {
			return fmt.Errorf("gofish: decode response from %s: %w", url, err)
		}
	}
	return nil
}

// Get performs a GET request against path and decodes the JSON response into out.
func (c *Client) Get(ctx context.Context, path string, out interface{}) error {
	return c.doRequest(ctx, http.MethodGet, path, nil, out)
}

// Post performs a POST request against path with the given JSON body.
func (c *Client) Post(ctx context.Context, path string, body interface{}, out interface{}) error {
	return c.doRequest(ctx, http.MethodPost, path, body, out)
}

// Patch performs a PATCH request against path with the given JSON body.
func (c *Client) Patch(ctx context.Context, path string, body interface{}, out interface{}) error {
	return c.doRequest(ctx, http.MethodPatch, path, body, out)
}

// Delete performs a DELETE request against path.
func (c *Client) Delete(ctx context.Context, path string) error {
	return c.doRequest(ctx, http.MethodDelete, path, nil, nil)
}
