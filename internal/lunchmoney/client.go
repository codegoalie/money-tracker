// Package lunchmoney is a thin, dependency-free (stdlib only) client for the
// LunchMoney API (https://lunchmoney.dev).
//
// It performs the HTTP calls and JSON parsing needed to fetch account
// balances and recurring transactions from LunchMoney, and exposes them as
// plain Go structs. It does not perform any forecasting or cadence math
// itself; that is the responsibility of internal/forecast, which this
// package's callers feed with the data returned here.
package lunchmoney

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// Client is a LunchMoney API client. Construct one with New.
//
// A Client is safe for concurrent use by multiple goroutines, as it holds no
// mutable state after construction; all request-scoped state lives on the
// stack of each call.
type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

// Option configures a Client constructed via New.
type Option func(*Client)

// WithHTTPClient overrides the *http.Client used to make requests. If not
// supplied, New defaults to http.DefaultClient. Note that this client has no
// per-request timeout of its own; callers control timeouts via the
// context.Context passed to each method.
func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) {
		c.httpClient = hc
	}
}

// New constructs a Client for the LunchMoney API (or a fake/test server)
// reachable at baseURL, authenticating requests with token.
//
// baseURL should not have a trailing slash, but New tolerates one if present.
func New(baseURL, token string, opts ...Option) *Client {
	c := &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		token:      token,
		httpClient: http.DefaultClient,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// get issues an authenticated GET request against path (which must start
// with "/", e.g. "/v1/me") with the given raw query string (may be empty),
// and decodes a 200 JSON response body into out.
func (c *Client) get(path, rawQuery string, out any) error {
	u := c.baseURL + path
	if rawQuery != "" {
		u += "?" + rawQuery
	}

	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return fmt.Errorf("lunchmoney: building request for %s: %w", path, err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("lunchmoney: request to %s: %w", path, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return newAPIError(resp)
	}

	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("lunchmoney: decoding response from %s: %w", path, err)
	}
	return nil
}
