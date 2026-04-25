package adsbdb_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	adsbdb "github.com/nint8835/go-adsbdb"
)

const (
	testBaseURL      = "https://example.test/v0/"
	defaultUserAgent = "go-adsbdb (github.com/nint8835/go-adsbdb)"

	aircraftResponse = `{
		"response": {
			"aircraft": {
				"type": "767-323ERSF",
				"icao_type": "B763",
				"manufacturer": "Boeing",
				"mode_s": "C0816E",
				"registration": "C-GXAJ",
				"registered_owner_country_iso_name": "CA",
				"registered_owner_country_name": "Canada",
				"registered_owner_operator_flag_code": "CJT",
				"registered_owner": "Cargojet Airways Ltd",
				"url_photo": null,
				"url_photo_thumbnail": null
			}
		}
	}`

	aircraftWithRouteResponse = `{
		"response": {
			"aircraft": {
				"type": "767-323ERSF",
				"icao_type": "B763",
				"manufacturer": "Boeing",
				"mode_s": "C0816E",
				"registration": "C-GXAJ",
				"registered_owner_country_iso_name": "CA",
				"registered_owner_country_name": "Canada",
				"registered_owner_operator_flag_code": "CJT",
				"registered_owner": "Cargojet Airways Ltd",
				"url_photo": null,
				"url_photo_thumbnail": null
			},
			"flightroute": {
				"callsign": "CJT620",
				"callsign_icao": "CJT620",
				"callsign_iata": "W8620",
				"airline": {
					"name": "Cargojet Airways",
					"icao": "CJT",
					"iata": "W8",
					"country": "Canada",
					"country_iso": "CA",
					"callsign": "CARGOJET"
				},
				"origin": {
					"country_iso_name": "CA",
					"country_name": "Canada",
					"elevation": 780,
					"iata_code": "YHM",
					"icao_code": "CYHM",
					"latitude": 43.1735992432,
					"longitude": -79.9349975586,
					"municipality": "Hamilton",
					"name": "John C. Munro Hamilton International Airport"
				},
				"midpoint": {
					"country_iso_name": "CA",
					"country_name": "Canada",
					"elevation": 232,
					"iata_code": "YQM",
					"icao_code": "CYQM",
					"latitude": 46.1122016907,
					"longitude": -64.6785964966,
					"municipality": "Moncton",
					"name": "Greater Moncton Romeo LeBlanc International Airport"
				},
				"destination": {
					"country_iso_name": "CA",
					"country_name": "Canada",
					"elevation": 461,
					"iata_code": "YYT",
					"icao_code": "CYYT",
					"latitude": 47.618598938,
					"longitude": -52.7518997192,
					"municipality": "St. John's",
					"name": "St. John's International Airport"
				}
			}
		}
	}`

	callsignResponse = `{
		"response": {
			"flightroute": {
				"callsign": "CJT620",
				"callsign_icao": "CJT620",
				"callsign_iata": "W8620",
				"airline": {
					"name": "Cargojet Airways",
					"icao": "CJT",
					"iata": "W8",
					"country": "Canada",
					"country_iso": "CA",
					"callsign": "CARGOJET"
				},
				"origin": {
					"country_iso_name": "CA",
					"country_name": "Canada",
					"elevation": 780,
					"iata_code": "YHM",
					"icao_code": "CYHM",
					"latitude": 43.1735992432,
					"longitude": -79.9349975586,
					"municipality": "Hamilton",
					"name": "John C. Munro Hamilton International Airport"
				},
				"destination": {
					"country_iso_name": "CA",
					"country_name": "Canada",
					"elevation": 461,
					"iata_code": "YYT",
					"icao_code": "CYYT",
					"latitude": 47.618598938,
					"longitude": -52.7518997192,
					"municipality": "St. John's",
					"name": "St. John's International Airport"
				}
			}
		}
	}`

	airlineResponse = `{
		"response": [{
			"name": "Cargojet Airways",
			"icao": "CJT",
			"iata": "W8",
			"country": "Canada",
			"country_iso": "CA",
			"callsign": "CARGOJET"
		}]
	}`

	statsResponse = `{
		"response": {
			"daily": {
				"aircraft": [{"url": "/v0/aircraft/C0816E", "count": 2}],
				"airline": [],
				"callsign": [],
				"mode_s": [],
				"n_number": [],
				"online": [],
				"stats": [],
				"aggregate": 2
			},
			"total": {
				"aircraft": [],
				"airline": [],
				"callsign": [],
				"mode_s": [],
				"n_number": [],
				"online": [],
				"stats": [{"url": "/v0/stats", "count": 1}],
				"aggregate": 3
			}
		}
	}`
)

func TestAircraft(t *testing.T) {
	client := newClient(t, func(r *http.Request) *http.Response {
		assertRequest(t, r, "/v0/aircraft/C0816E")
		return jsonResponse(http.StatusOK, aircraftResponse)
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
		return jsonResponse(http.StatusOK, aircraftWithRouteResponse)
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
	if result.FlightRoute.Midpoint == nil || result.FlightRoute.Midpoint.IATACode != "YQM" {
		t.Fatalf("Midpoint = %#v", result.FlightRoute.Midpoint)
	}
}

func TestCallsign(t *testing.T) {
	client := newClient(t, jsonHandler(t, "/v0/callsign/CJT620", callsignResponse))

	route, err := client.Callsign(context.Background(), "CJT620")
	if err != nil {
		t.Fatal(err)
	}
	if route.Callsign != "CJT620" || route.Destination.IATACode != "YYT" {
		t.Fatalf("route = %#v", route)
	}
}

func TestAirline(t *testing.T) {
	client := newClient(t, jsonHandler(t, "/v0/airline/CJT", airlineResponse))

	airlines, err := client.Airline(context.Background(), "CJT")
	if err != nil {
		t.Fatal(err)
	}
	if len(airlines) != 1 || airlines[0].ICAO != "CJT" {
		t.Fatalf("airlines = %#v", airlines)
	}
}

func TestStats(t *testing.T) {
	client := newClient(t, jsonHandler(t, "/v0/stats", statsResponse))

	stats, err := client.Stats(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if stats.Daily.Aggregate != 2 || stats.Total.Aggregate != 3 {
		t.Fatalf("stats = %#v", stats)
	}
}

func TestModeSToNNumber(t *testing.T) {
	client := newClient(t, jsonHandler(t, "/v0/mode-s/A00001", `{"response":"N1"}`))

	nNumber, err := client.ModeSToNNumber(context.Background(), "A00001")
	if err != nil {
		t.Fatal(err)
	}
	if nNumber != "N1" {
		t.Fatalf("nNumber = %q", nNumber)
	}
}

func TestNNumberToModeS(t *testing.T) {
	client := newClient(t, jsonHandler(t, "/v0/n-number/N1", `{"response":"A00001"}`))

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
