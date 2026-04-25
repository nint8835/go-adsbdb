package adsbdb_test

//go:generate go run -tags fixturegen ./internal/fixturegen

import (
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	adsbdb "github.com/nint8835/go-adsbdb"
)

const (
	testBaseURL      = "https://example.test/v0/"
	defaultUserAgent = "go-adsbdb (github.com/nint8835/go-adsbdb)"
)

func TestAircraft(t *testing.T) {
	client := newClient(t, func(r *http.Request) *http.Response {
		assertRequest(t, r, "/v0/aircraft/C0816E")
		return jsonResponse(http.StatusOK, loadFixture(t, "aircraft_c0816e.json"))
	})

	aircraft, err := client.Aircraft(context.Background(), "C0816E")
	if err != nil {
		t.Fatal(err)
	}
	if aircraft.Registration != "C-GXAJ" {
		t.Fatalf("Registration = %q", aircraft.Registration)
	}
	if aircraft.RegisteredOwnerOperatorFlagCode == nil || *aircraft.RegisteredOwnerOperatorFlagCode != "CJT" {
		t.Fatalf("RegisteredOwnerOperatorFlagCode = %v", aircraft.RegisteredOwnerOperatorFlagCode)
	}
}

func TestAircraftWithCallsign(t *testing.T) {
	client := newClient(t, func(r *http.Request) *http.Response {
		assertRequest(t, r, "/v0/aircraft/C0816E")
		assertQuery(t, r, "callsign", "CJT620")
		return jsonResponse(http.StatusOK, loadFixture(t, "aircraft_c0816e_callsign_cjt620.json"))
	})

	result, err := client.AircraftWithCallsign(context.Background(), "C0816E", "CJT620")
	if err != nil {
		t.Fatal(err)
	}
	if result.Aircraft.ModeS == "" || result.FlightRoute == nil {
		t.Fatalf("expected aircraft and flight route: %#v", result)
	}
	if result.FlightRoute.Airline == nil || result.FlightRoute.Airline.ICAO != "CJT" {
		t.Fatalf("Airline = %#v", result.FlightRoute.Airline)
	}
	if result.FlightRoute.Destination.IATACode != "YYT" {
		t.Fatalf("Destination = %#v", result.FlightRoute.Destination)
	}
}

func TestCallsign(t *testing.T) {
	client := newClient(t, jsonHandler(t, "/v0/callsign/CJT620", loadFixture(t, "callsign_cjt620.json")))

	route, err := client.Callsign(context.Background(), "CJT620")
	if err != nil {
		t.Fatal(err)
	}
	if route.Callsign != "CJT620" || route.Destination.IATACode != "YYT" {
		t.Fatalf("route = %#v", route)
	}
}

func TestAirline(t *testing.T) {
	client := newClient(t, jsonHandler(t, "/v0/airline/CJT", loadFixture(t, "airline_cjt.json")))

	airlines, err := client.Airline(context.Background(), "CJT")
	if err != nil {
		t.Fatal(err)
	}
	if len(airlines) != 1 || airlines[0].ICAO != "CJT" {
		t.Fatalf("airlines = %#v", airlines)
	}
}

func TestStats(t *testing.T) {
	client := newClient(t, jsonHandler(t, "/v0/stats", loadFixture(t, "stats.json")))

	stats, err := client.Stats(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if stats.Daily.Aggregate == 0 || stats.Total.Aggregate == 0 {
		t.Fatalf("stats = %#v", stats)
	}
}

func TestModeSToNNumber(t *testing.T) {
	client := newClient(t, jsonHandler(t, "/v0/mode-s/A00001", loadFixture(t, "mode_s_a00001.json")))

	nNumber, err := client.ModeSToNNumber(context.Background(), "A00001")
	if err != nil {
		t.Fatal(err)
	}
	if nNumber != "N1" {
		t.Fatalf("nNumber = %q", nNumber)
	}
}

func TestNNumberToModeS(t *testing.T) {
	client := newClient(t, jsonHandler(t, "/v0/n-number/N1", loadFixture(t, "n_number_n1.json")))

	modeS, err := client.NNumberToModeS(context.Background(), "N1")
	if err != nil {
		t.Fatal(err)
	}
	if modeS != "A00001" {
		t.Fatalf("modeS = %q", modeS)
	}
}

func TestRandomHelpers(t *testing.T) {
	tests := []struct {
		name string
		call func(*adsbdb.Client) error
		path string
		body string
	}{
		{
			name: "aircraft",
			call: func(client *adsbdb.Client) error {
				_, err := client.RandomAircraft(context.Background())
				return err
			},
			path: "/v0/aircraft/random",
			body: `{"response":{"aircraft":{"mode_s":"C0816E"}}}`,
		},
		{
			name: "callsign",
			call: func(client *adsbdb.Client) error {
				_, err := client.RandomCallsign(context.Background())
				return err
			},
			path: "/v0/callsign/random",
			body: `{"response":{"flightroute":{"callsign":"CJT620"}}}`,
		},
		{
			name: "airline",
			call: func(client *adsbdb.Client) error {
				_, err := client.RandomAirline(context.Background())
				return err
			},
			path: "/v0/airline/random",
			body: `{"response":[{"name":"Cargojet Airways","icao":"CJT"}]}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := newClient(t, jsonHandler(t, tt.path, tt.body))
			if err := tt.call(client); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestNotFound(t *testing.T) {
	client := newClient(t, func(r *http.Request) *http.Response {
		return jsonResponse(http.StatusNotFound, `{"response":"unknown aircraft"}`)
	})

	_, err := client.Aircraft(context.Background(), "NOPE")
	if !errors.Is(err, adsbdb.ErrNotFound) {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}

	var apiErr *adsbdb.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("err = %T, want *APIError", err)
	}
	if apiErr.Message != "unknown aircraft" {
		t.Fatalf("Message = %q", apiErr.Message)
	}
	if got := apiErr.Error(); got != "api returned status 404: unknown aircraft" {
		t.Fatalf("Error() = %q", got)
	}
}

func TestOptions(t *testing.T) {
	tests := []struct {
		name string
		opt  adsbdb.Option
		want string
	}{
		{name: "nil http client", opt: adsbdb.WithHTTPClient(nil), want: "nil http client"},
		{name: "invalid base url", opt: adsbdb.WithBaseURL("%"), want: "parse base URL"},
		{name: "relative base url", opt: adsbdb.WithBaseURL("/v0/"), want: "base URL must be absolute"},
		{name: "empty user agent", opt: adsbdb.WithUserAgent(" \t "), want: "user agent cannot be empty"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := adsbdb.NewClient(tt.opt)
			assertErrorContains(t, err, tt.want)
		})
	}
}

func TestCustomUserAgentAndBaseURLTrailingSlash(t *testing.T) {
	client := newClient(
		t,
		func(r *http.Request) *http.Response {
			assertPath(t, r, "/v0/stats")
			if got := r.Header.Get("User-Agent"); got != "packages-to-the-island/1.0" {
				t.Fatalf("User-Agent = %q", got)
			}
			return jsonResponse(http.StatusOK, `{"response":{"daily":{"aggregate":1},"total":{"aggregate":1}}}`)
		},
		adsbdb.WithBaseURL("https://example.test/v0"),
		adsbdb.WithUserAgent("packages-to-the-island/1.0"),
	)

	if _, err := client.Stats(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestDecodeErrors(t *testing.T) {
	tests := []struct {
		name string
		body string
		want string
	}{
		{name: "malformed json", body: `{`, want: "decode response envelope"},
		{name: "missing response", body: `{}`, want: "missing response field"},
		{name: "wrong response shape", body: `{"response":{}}`, want: "decode response"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := newClient(t, jsonHandler(t, "/v0/mode-s/A00001", tt.body))
			_, err := client.ModeSToNNumber(context.Background(), "A00001")
			assertErrorContains(t, err, tt.want)
		})
	}
}

func TestMissingRequiredTopLevelFields(t *testing.T) {
	tests := []struct {
		name string
		call func(*adsbdb.Client) error
		want string
	}{
		{
			name: "aircraft",
			call: func(client *adsbdb.Client) error {
				_, err := client.Aircraft(context.Background(), "C0816E")
				return err
			},
			want: "missing aircraft field",
		},
		{
			name: "callsign",
			call: func(client *adsbdb.Client) error {
				_, err := client.Callsign(context.Background(), "CJT620")
				return err
			},
			want: "missing flightroute field",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := newClient(t, func(r *http.Request) *http.Response {
				return jsonResponse(http.StatusOK, `{"response":{}}`)
			})
			assertErrorContains(t, tt.call(client), tt.want)
		})
	}
}

func TestTransportError(t *testing.T) {
	wantErr := errors.New("ground stop")
	client, err := adsbdb.NewClient(
		adsbdb.WithBaseURL(testBaseURL),
		adsbdb.WithHTTPClient(&http.Client{
			Transport: errorRoundTripper{err: wantErr},
		}),
	)
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.Stats(context.Background())
	if !errors.Is(err, wantErr) {
		t.Fatalf("err = %v, want %v", err, wantErr)
	}
}

func TestAPIErrorWithoutMessage(t *testing.T) {
	client := newClient(t, func(r *http.Request) *http.Response {
		return jsonResponse(http.StatusTeapot, "")
	})

	_, err := client.Stats(context.Background())
	var apiErr *adsbdb.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("err = %T, want *APIError", err)
	}
	if apiErr.Error() != "api returned status 418" {
		t.Fatalf("Error() = %q", apiErr.Error())
	}
}

type roundTripFunc func(*http.Request) *http.Response

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r), nil
}

type errorRoundTripper struct {
	err error
}

func (rt errorRoundTripper) RoundTrip(*http.Request) (*http.Response, error) {
	return nil, rt.err
}

func newClient(t testing.TB, handler roundTripFunc, opts ...adsbdb.Option) *adsbdb.Client {
	t.Helper()

	allOpts := []adsbdb.Option{
		adsbdb.WithBaseURL(testBaseURL),
		adsbdb.WithHTTPClient(newTestHTTPClient(handler)),
	}
	allOpts = append(allOpts, opts...)

	client, err := adsbdb.NewClient(allOpts...)
	if err != nil {
		t.Fatal(err)
	}
	return client
}

func jsonHandler(t testing.TB, wantPath string, body string) roundTripFunc {
	t.Helper()

	return func(r *http.Request) *http.Response {
		assertRequest(t, r, wantPath)
		return jsonResponse(http.StatusOK, body)
	}
}

func assertRequest(t testing.TB, r *http.Request, wantPath string) {
	t.Helper()
	assertPath(t, r, wantPath)
	if got := r.Header.Get("Accept"); got != "application/json" {
		t.Fatalf("Accept = %q", got)
	}
	if got := r.Header.Get("User-Agent"); got != defaultUserAgent {
		t.Fatalf("User-Agent = %q", got)
	}
}

func assertPath(t testing.TB, r *http.Request, want string) {
	t.Helper()
	if r.URL.Path != want {
		t.Fatalf("path = %q", r.URL.Path)
	}
}

func assertQuery(t testing.TB, r *http.Request, key string, want string) {
	t.Helper()
	if got := r.URL.Query().Get(key); got != want {
		t.Fatalf("%s = %q", key, got)
	}
}

func assertErrorContains(t testing.TB, err error, want string) {
	t.Helper()
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), want) {
		t.Fatalf("err = %q, want substring %q", err, want)
	}
}

func loadFixture(t testing.TB, name string) string {
	t.Helper()

	body, err := os.ReadFile("testdata/" + name)
	if err != nil {
		t.Fatal(err)
	}
	return string(body)
}

func newTestHTTPClient(f roundTripFunc) *http.Client {
	return &http.Client{Transport: f}
}

func jsonResponse(statusCode int, body string) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}
