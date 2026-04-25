package adsbdb

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const (
	defaultBaseURL   = "https://api.adsbdb.com/v0/"
	defaultUserAgent = "go-adsbdb (github.com/nint8835/go-adsbdb)"
)

// Client is an ADSBDB API client.
type Client struct {
	baseURL    *url.URL
	httpClient *http.Client
	userAgent  string
}

// Option configures a Client.
type Option func(*Client) error

// NewClient creates a Client for the public ADSBDB API.
func NewClient(opts ...Option) (*Client, error) {
	baseURL, err := url.Parse(defaultBaseURL)
	if err != nil {
		return nil, err
	}

	c := &Client{
		baseURL:    baseURL,
		httpClient: http.DefaultClient,
		userAgent:  defaultUserAgent,
	}

	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, err
		}
	}

	return c, nil
}

// WithHTTPClient configures the HTTP client used to make requests.
func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) error {
		if httpClient == nil {
			return errors.New("nil http client")
		}
		c.httpClient = httpClient
		return nil
	}
}

// WithBaseURL configures the ADSBDB API base URL.
//
// This is mainly useful for tests or self-hosted ADSBDB deployments.
func WithBaseURL(baseURL string) Option {
	return func(c *Client) error {
		parsed, err := url.Parse(baseURL)
		if err != nil {
			return fmt.Errorf("parse base URL: %w", err)
		}
		if parsed.Scheme == "" || parsed.Host == "" {
			return errors.New("base URL must be absolute")
		}
		if !strings.HasSuffix(parsed.Path, "/") {
			parsed.Path += "/"
		}
		c.baseURL = parsed
		return nil
	}
}

// WithUserAgent configures the User-Agent header sent by the client.
func WithUserAgent(userAgent string) Option {
	return func(c *Client) error {
		userAgent = strings.TrimSpace(userAgent)
		if userAgent == "" {
			return errors.New("user agent cannot be empty")
		}
		c.userAgent = userAgent
		return nil
	}
}

// Aircraft returns aircraft data for a Mode-S hex code or registration.
func (c *Client) Aircraft(ctx context.Context, identifier string) (Aircraft, error) {
	var out aircraftResponse
	if err := c.get(ctx, path("aircraft", identifier), nil, &out); err != nil {
		return Aircraft{}, err
	}
	if out.Aircraft == nil {
		return Aircraft{}, errors.New("missing aircraft field")
	}
	return *out.Aircraft, nil
}

// RandomAircraft returns data for a random aircraft.
func (c *Client) RandomAircraft(ctx context.Context) (Aircraft, error) {
	return c.Aircraft(ctx, "random")
}

// AircraftWithCallsign returns aircraft data and, when the callsign is known,
// route data for the supplied callsign.
func (c *Client) AircraftWithCallsign(ctx context.Context, identifier, callsign string) (AircraftAndFlightRoute, error) {
	var out AircraftAndFlightRoute
	query := url.Values{"callsign": []string{callsign}}
	if err := c.get(ctx, path("aircraft", identifier), query, &out); err != nil {
		return AircraftAndFlightRoute{}, err
	}
	return out, nil
}

// Callsign returns flight route data for a callsign.
func (c *Client) Callsign(ctx context.Context, callsign string) (FlightRoute, error) {
	var out callsignResponse
	if err := c.get(ctx, path("callsign", callsign), nil, &out); err != nil {
		return FlightRoute{}, err
	}
	if out.FlightRoute == nil {
		return FlightRoute{}, errors.New("missing flightroute field")
	}
	return *out.FlightRoute, nil
}

// RandomCallsign returns flight route data for a random callsign.
func (c *Client) RandomCallsign(ctx context.Context) (FlightRoute, error) {
	return c.Callsign(ctx, "random")
}

// Airline returns airline records for an ICAO or IATA airline code.
func (c *Client) Airline(ctx context.Context, code string) ([]Airline, error) {
	var out []Airline
	if err := c.get(ctx, path("airline", code), nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// RandomAirline returns random airline records.
func (c *Client) RandomAirline(ctx context.Context) ([]Airline, error) {
	return c.Airline(ctx, "random")
}

// Stats returns API usage statistics.
func (c *Client) Stats(ctx context.Context) (Stats, error) {
	var out Stats
	if err := c.get(ctx, "stats", nil, &out); err != nil {
		return Stats{}, err
	}
	return out, nil
}

// ModeSToNNumber converts a Mode-S hex code to a US N-Number.
func (c *Client) ModeSToNNumber(ctx context.Context, modeS string) (string, error) {
	var out string
	if err := c.get(ctx, path("mode-s", modeS), nil, &out); err != nil {
		return "", err
	}
	return out, nil
}

// NNumberToModeS converts a US N-Number to a Mode-S hex code.
func (c *Client) NNumberToModeS(ctx context.Context, nNumber string) (string, error) {
	var out string
	if err := c.get(ctx, path("n-number", nNumber), nil, &out); err != nil {
		return "", err
	}
	return out, nil
}

func (c *Client) get(ctx context.Context, endpoint string, query url.Values, v any) error {
	u := c.baseURL.ResolveReference(&url.URL{Path: strings.TrimPrefix(endpoint, "/")})
	if len(query) > 0 {
		u.RawQuery = query.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.userAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return newAPIError(resp.StatusCode, body)
	}

	return decodeResponse(resp.Body, v)
}

func decodeResponse(r io.Reader, v any) error {
	var envelope struct {
		// ADSBDB wraps every endpoint in "response", but that field can be an
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
