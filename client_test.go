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

func TestAircraft(t *testing.T) {
	client, err := adsbdb.NewClient(
		adsbdb.WithBaseURL("https://example.test/v0/"),
		adsbdb.WithHTTPClient(newTestHTTPClient(func(r *http.Request) *http.Response {
			if r.URL.Path != "/v0/aircraft/C0816E" {
				t.Fatalf("path = %q", r.URL.Path)
			}
			if got := r.Header.Get("Accept"); got != "application/json" {
				t.Fatalf("Accept = %q", got)
			}
			if got := r.Header.Get("User-Agent"); got != "go-adsbdb (github.com/nint8835/go-adsbdb)" {
				t.Fatalf("User-Agent = %q", got)
			}

			return jsonResponse(http.StatusOK, `{
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
			}`)
		})),
	)
	if err != nil {
		t.Fatal(err)
	}

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
	client, err := adsbdb.NewClient(
		adsbdb.WithBaseURL("https://example.test/v0/"),
		adsbdb.WithHTTPClient(newTestHTTPClient(func(r *http.Request) *http.Response {
			if r.URL.Path != "/v0/aircraft/C0816E" {
				t.Fatalf("path = %q", r.URL.Path)
			}
			if got := r.URL.Query().Get("callsign"); got != "CJT620" {
				t.Fatalf("callsign = %q", got)
			}

			return jsonResponse(http.StatusOK, `{
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
			}`)
		})),
	)
	if err != nil {
		t.Fatal(err)
	}

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

func TestNotFound(t *testing.T) {
	client, err := adsbdb.NewClient(
		adsbdb.WithBaseURL("https://example.test/v0/"),
		adsbdb.WithHTTPClient(newTestHTTPClient(func(r *http.Request) *http.Response {
			return jsonResponse(http.StatusNotFound, `{"response":"unknown aircraft"}`)
		})),
	)
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.Aircraft(context.Background(), "NOPE")
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

func TestAirline(t *testing.T) {
	client, err := adsbdb.NewClient(
		adsbdb.WithBaseURL("https://example.test/v0/"),
		adsbdb.WithHTTPClient(newTestHTTPClient(func(r *http.Request) *http.Response {
			if r.URL.Path != "/v0/airline/CJT" {
				t.Fatalf("path = %q", r.URL.Path)
			}
			return jsonResponse(http.StatusOK, `{
				"response": [
					{
						"name": "Cargojet Airways",
						"icao": "CJT",
						"iata": "W8",
						"country": "Canada",
						"country_iso": "CA",
						"callsign": "CARGOJET"
					}
				]
			}`)
		})),
	)
	if err != nil {
		t.Fatal(err)
	}

	airlines, err := client.Airline(context.Background(), "CJT")
	if err != nil {
		t.Fatal(err)
	}
	if len(airlines) != 1 || airlines[0].ICAO != "CJT" {
		t.Fatalf("airlines = %#v", airlines)
	}
}

func TestCallsign(t *testing.T) {
	client, err := adsbdb.NewClient(
		adsbdb.WithBaseURL("https://example.test/v0/"),
		adsbdb.WithHTTPClient(newTestHTTPClient(func(r *http.Request) *http.Response {
			if r.URL.Path != "/v0/callsign/CJT620" {
				t.Fatalf("path = %q", r.URL.Path)
			}
			return jsonResponse(http.StatusOK, `{
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
			}`)
		})),
	)
	if err != nil {
		t.Fatal(err)
	}

	route, err := client.Callsign(context.Background(), "CJT620")
	if err != nil {
		t.Fatal(err)
	}
	if route.Callsign != "CJT620" || route.Destination.IATACode != "YYT" {
		t.Fatalf("route = %#v", route)
	}
}

func TestStats(t *testing.T) {
	client, err := adsbdb.NewClient(
		adsbdb.WithBaseURL("https://example.test/v0/"),
		adsbdb.WithHTTPClient(newTestHTTPClient(func(r *http.Request) *http.Response {
			if r.URL.Path != "/v0/stats" {
				t.Fatalf("path = %q", r.URL.Path)
			}
			return jsonResponse(http.StatusOK, `{
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
			}`)
		})),
	)
	if err != nil {
		t.Fatal(err)
	}

	stats, err := client.Stats(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if stats.Daily.Aggregate != 2 || stats.Total.Aggregate != 3 {
		t.Fatalf("stats = %#v", stats)
	}
}

func TestModeSToNNumber(t *testing.T) {
	client, err := adsbdb.NewClient(
		adsbdb.WithBaseURL("https://example.test/v0/"),
		adsbdb.WithHTTPClient(newTestHTTPClient(func(r *http.Request) *http.Response {
			if r.URL.Path != "/v0/mode-s/A00001" {
				t.Fatalf("path = %q", r.URL.Path)
			}
			return jsonResponse(http.StatusOK, `{"response":"N1"}`)
		})),
	)
	if err != nil {
		t.Fatal(err)
	}

	nNumber, err := client.ModeSToNNumber(context.Background(), "A00001")
	if err != nil {
		t.Fatal(err)
	}
	if nNumber != "N1" {
		t.Fatalf("nNumber = %q", nNumber)
	}
}

func TestNNumberToModeS(t *testing.T) {
	client, err := adsbdb.NewClient(
		adsbdb.WithBaseURL("https://example.test/v0/"),
		adsbdb.WithHTTPClient(newTestHTTPClient(func(r *http.Request) *http.Response {
			if r.URL.Path != "/v0/n-number/N1" {
				t.Fatalf("path = %q", r.URL.Path)
			}
			return jsonResponse(http.StatusOK, `{"response":"A00001"}`)
		})),
	)
	if err != nil {
		t.Fatal(err)
	}

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
			client, err := adsbdb.NewClient(
				adsbdb.WithBaseURL("https://example.test/v0/"),
				adsbdb.WithHTTPClient(newTestHTTPClient(func(r *http.Request) *http.Response {
					if r.URL.Path != tt.path {
						t.Fatalf("path = %q", r.URL.Path)
					}
					return jsonResponse(http.StatusOK, tt.body)
				})),
			)
			if err != nil {
				t.Fatal(err)
			}
			if err := tt.call(client); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestOptions(t *testing.T) {
	tests := []struct {
		name string
		opt  adsbdb.Option
		want string
	}{
		{
			name: "nil http client",
			opt:  adsbdb.WithHTTPClient(nil),
			want: "nil http client",
		},
		{
			name: "invalid base url",
			opt:  adsbdb.WithBaseURL("%"),
			want: "parse base URL",
		},
		{
			name: "relative base url",
			opt:  adsbdb.WithBaseURL("/v0/"),
			want: "base URL must be absolute",
		},
		{
			name: "empty user agent",
			opt:  adsbdb.WithUserAgent(" \t "),
			want: "user agent cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := adsbdb.NewClient(tt.opt)
			if err == nil {
				t.Fatal("expected error")
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("err = %q, want substring %q", err, tt.want)
			}
		})
	}
}

func TestCustomUserAgentAndBaseURLTrailingSlash(t *testing.T) {
	client, err := adsbdb.NewClient(
		adsbdb.WithBaseURL("https://example.test/v0"),
		adsbdb.WithUserAgent("packages-to-the-island/1.0"),
		adsbdb.WithHTTPClient(newTestHTTPClient(func(r *http.Request) *http.Response {
			if r.URL.Path != "/v0/stats" {
				t.Fatalf("path = %q", r.URL.Path)
			}
			if got := r.Header.Get("User-Agent"); got != "packages-to-the-island/1.0" {
				t.Fatalf("User-Agent = %q", got)
			}
			return jsonResponse(http.StatusOK, `{"response":{"daily":{"aggregate":1},"total":{"aggregate":1}}}`)
		})),
	)
	if err != nil {
		t.Fatal(err)
	}

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
		{
			name: "malformed json",
			body: `{`,
			want: "decode response envelope",
		},
		{
			name: "missing response",
			body: `{}`,
			want: "missing response field",
		},
		{
			name: "wrong response shape",
			body: `{"response":{}}`,
			want: "decode response",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := adsbdb.NewClient(
				adsbdb.WithBaseURL("https://example.test/v0/"),
				adsbdb.WithHTTPClient(newTestHTTPClient(func(r *http.Request) *http.Response {
					return jsonResponse(http.StatusOK, tt.body)
				})),
			)
			if err != nil {
				t.Fatal(err)
			}

			_, err = client.ModeSToNNumber(context.Background(), "A00001")
			if err == nil {
				t.Fatal("expected error")
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("err = %q, want substring %q", err, tt.want)
			}
		})
	}
}

func TestMissingRequiredTopLevelFields(t *testing.T) {
	tests := []struct {
		name string
		call func(*adsbdb.Client) error
		body string
		want string
	}{
		{
			name: "aircraft",
			call: func(client *adsbdb.Client) error {
				_, err := client.Aircraft(context.Background(), "C0816E")
				return err
			},
			body: `{"response":{}}`,
			want: "missing aircraft field",
		},
		{
			name: "callsign",
			call: func(client *adsbdb.Client) error {
				_, err := client.Callsign(context.Background(), "CJT620")
				return err
			},
			body: `{"response":{}}`,
			want: "missing flightroute field",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := adsbdb.NewClient(
				adsbdb.WithBaseURL("https://example.test/v0/"),
				adsbdb.WithHTTPClient(newTestHTTPClient(func(r *http.Request) *http.Response {
					return jsonResponse(http.StatusOK, tt.body)
				})),
			)
			if err != nil {
				t.Fatal(err)
			}

			err = tt.call(client)
			if err == nil {
				t.Fatal("expected error")
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("err = %q, want substring %q", err, tt.want)
			}
		})
	}
}

func TestTransportError(t *testing.T) {
	wantErr := errors.New("ground stop")
	client, err := adsbdb.NewClient(
		adsbdb.WithBaseURL("https://example.test/v0/"),
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
	client, err := adsbdb.NewClient(
		adsbdb.WithBaseURL("https://example.test/v0/"),
		adsbdb.WithHTTPClient(newTestHTTPClient(func(r *http.Request) *http.Response {
			return jsonResponse(http.StatusTeapot, ``)
		})),
	)
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.Stats(context.Background())
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
