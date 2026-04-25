package adsbdb

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

// ErrNotFound is matched by errors.Is for 404 responses from the adsbdb API.
var ErrNotFound = errors.New("not found")

// APIError describes a non-2xx response from the adsbdb API.
type APIError struct {
	StatusCode int
	Message    string
	Body       []byte
}

func newAPIError(statusCode int, body []byte) error {
	var envelope struct {
		Response string `json:"response"`
	}
	message := strings.TrimSpace(string(body))
	if err := json.Unmarshal(body, &envelope); err == nil && envelope.Response != "" {
		message = envelope.Response
	}
	return &APIError{
		StatusCode: statusCode,
		Message:    message,
		Body:       body,
	}
}

func (e *APIError) Error() string {
	if e.Message == "" {
		return fmt.Sprintf("api returned status %d", e.StatusCode)
	}
	return fmt.Sprintf("api returned status %d: %s", e.StatusCode, e.Message)
}

func (e *APIError) Is(target error) bool {
	return target == ErrNotFound && e.StatusCode == http.StatusNotFound
}
