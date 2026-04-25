package adsbdb

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"strings"
)

func decodeResponse(r io.Reader, v any) error {
	var envelope struct {
		// adsbdb wraps every endpoint in "response", but that field can be an
		// object, array, or string depending on the route.
		Response json.RawMessage `json:"response"`
	}
	if err := json.NewDecoder(r).Decode(&envelope); err != nil {
		return fmt.Errorf("decode response envelope: %w", err)
	}
	if len(envelope.Response) == 0 {
		return errors.New("missing response field")
	}
	if err := json.Unmarshal(envelope.Response, v); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	return nil
}

func path(parts ...string) string {
	escaped := make([]string, 0, len(parts))
	for _, part := range parts {
		escaped = append(escaped, url.PathEscape(part))
	}
	return strings.Join(escaped, "/")
}
