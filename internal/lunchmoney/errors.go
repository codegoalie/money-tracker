package lunchmoney

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

// ErrUnauthorized is the sentinel error wrapped by an *APIError whenever the
// LunchMoney API responds with HTTP 401. Callers should prefer
// errors.Is(err, ErrUnauthorized) over inspecting APIError.StatusCode
// directly.
var ErrUnauthorized = errors.New("lunchmoney: unauthorized (invalid or missing access token)")

// APIError represents a non-2xx response from the LunchMoney API.
//
// Name and Message come from the API's JSON error body ({"name": ...,
// "message": ...}) when present; for non-JSON or unparsable bodies, Message
// holds a truncated (first ~500 bytes) copy of the raw response body so the
// failure is still debuggable.
type APIError struct {
	StatusCode int
	Name       string
	Message    string
}

// Error implements the error interface.
func (e *APIError) Error() string {
	return fmt.Sprintf("lunchmoney: api error (status %d, name %q): %s", e.StatusCode, e.Name, e.Message)
}

// Unwrap allows errors.Is(err, ErrUnauthorized) to succeed for 401 responses.
// For any other status code, Unwrap returns nil.
func (e *APIError) Unwrap() error {
	if e.StatusCode == http.StatusUnauthorized {
		return ErrUnauthorized
	}
	return nil
}

// maxErrorBodyBytes caps how much of a non-2xx response body is read and
// retained on an APIError, so a huge or malicious error body can't blow up
// memory.
const maxErrorBodyBytes = 500

// newAPIError builds an *APIError from a non-2xx HTTP response. It attempts
// to parse the body as the API's standard {"name", "message"} error shape;
// if that fails (e.g. non-JSON body), it falls back to using the raw
// (truncated) body text as the Message.
func newAPIError(resp *http.Response) error {
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxErrorBodyBytes))
	if err != nil {
		return &APIError{
			StatusCode: resp.StatusCode,
			Name:       "HTTPError",
			Message:    fmt.Sprintf("failed to read response body: %v", err),
		}
	}

	var parsed struct {
		Name    string `json:"name"`
		Message string `json:"message"`
	}
	if jsonErr := json.Unmarshal(body, &parsed); jsonErr == nil && (parsed.Name != "" || parsed.Message != "") {
		return &APIError{StatusCode: resp.StatusCode, Name: parsed.Name, Message: parsed.Message}
	}

	return &APIError{StatusCode: resp.StatusCode, Name: "HTTPError", Message: string(body)}
}
