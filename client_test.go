package adsbdb

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestAircraft(t *testing.T) {
	client, err := NewClient(
		WithBaseURL("https://example.test/v0/"),
		WithHTTPClient(newTestHTTPClient(func(r *http.Request) *http.Response {
			if r.URL.Path != "/v0/aircraft/4CA645" {
				t.Fatalf("path = %q", r.URL.Path)
			}
			if got := r.Header.Get("Accept"); got != "application/json" {
				t.Fatalf("Accept = %q", got)
			}
			if got := r.Header.Get("User-Agent"); got != defaultUserAgent {
				t.Fatalf("User-Agent = %q", got)
			}

			return jsonResponse(http.StatusOK, `{
				"response": {
					"aircraft": {
						"type": "737NG 8AS/W",
						"icao_type": "B738",
						"manufacturer": "Boeing",
						"mode_s": "4CA645",
						"registration": "EI-DYC",
						"registered_owner_country_iso_name": "IE",
						"registered_owner_country_name": "Ireland",
						"registered_owner_operator_flag_code": "RYR",
						"registered_owner": "Ryanair",
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

	aircraft, err := client.Aircraft(context.Background(), "4CA645")
	if err != nil {
		t.Fatal(err)
	}
	if aircraft.Registration != "EI-DYC" {
		t.Fatalf("Registration = %q", aircraft.Registration)
	}
	if aircraft.RegisteredOwnerOperatorFlagCode == nil || *aircraft.RegisteredOwnerOperatorFlagCode != "RYR" {
		t.Fatalf("RegisteredOwnerOperatorFlagCode = %v", aircraft.RegisteredOwnerOperatorFlagCode)
	}
}

func TestAircraftWithCallsign(t *testing.T) {
	client, err := NewClient(
		WithBaseURL("https://example.test/v0/"),
		WithHTTPClient(newTestHTTPClient(func(r *http.Request) *http.Response {
			if r.URL.Path != "/v0/aircraft/4CA645" {
				t.Fatalf("path = %q", r.URL.Path)
			}
			if got := r.URL.Query().Get("callsign"); got != "RYR9KH" {
				t.Fatalf("callsign = %q", got)
			}

			return jsonResponse(http.StatusOK, `{
				"response": {
					"aircraft": {
						"type": "737NG 8AS/W",
						"icao_type": "B738",
						"manufacturer": "Boeing",
						"mode_s": "4CA645",
						"registration": "EI-DYC",
						"registered_owner_country_iso_name": "IE",
						"registered_owner_country_name": "Ireland",
						"registered_owner_operator_flag_code": "RYR",
						"registered_owner": "Ryanair",
						"url_photo": null,
						"url_photo_thumbnail": null
					},
					"flightroute": {
						"callsign": "RYR9KH",
						"callsign_icao": "RYR9KH",
						"callsign_iata": null,
						"airline": {
							"name": "Ryanair",
							"icao": "RYR",
							"iata": "FR",
							"country": "Ireland",
							"country_iso": "IE",
							"callsign": "RYANAIR"
						},
						"origin": {
							"country_iso_name": "GB",
							"country_name": "United Kingdom",
							"elevation": 622,
							"iata_code": "STN",
							"icao_code": "EGSS",
							"latitude": 51.8849983215,
							"longitude": 0.2349999994,
							"municipality": "London",
							"name": "London Stansted Airport"
						},
						"destination": {
							"country_iso_name": "IE",
							"country_name": "Ireland",
							"elevation": 242,
							"iata_code": "DUB",
							"icao_code": "EIDW",
							"latitude": 53.421299,
							"longitude": -6.27007,
							"municipality": "Dublin",
							"name": "Dublin Airport"
						}
					}
				}
			}`)
		})),
	)
	if err != nil {
		t.Fatal(err)
	}

	result, err := client.AircraftWithCallsign(context.Background(), "4CA645", "RYR9KH")
	if err != nil {
		t.Fatal(err)
	}
	if result.Aircraft.ModeS == "" || result.FlightRoute == nil {
		t.Fatalf("expected aircraft and flight route: %#v", result)
	}
	if result.FlightRoute.Airline == nil || result.FlightRoute.Airline.ICAO != "RYR" {
		t.Fatalf("Airline = %#v", result.FlightRoute.Airline)
	}
}

func TestNotFound(t *testing.T) {
	client, err := NewClient(
		WithBaseURL("https://example.test/v0/"),
		WithHTTPClient(newTestHTTPClient(func(r *http.Request) *http.Response {
			return jsonResponse(http.StatusNotFound, `{"response":"unknown aircraft"}`)
		})),
	)
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.Aircraft(context.Background(), "NOPE")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}

	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("err = %T, want *APIError", err)
	}
	if apiErr.Message != "unknown aircraft" {
		t.Fatalf("Message = %q", apiErr.Message)
	}
}

func TestAirline(t *testing.T) {
	client, err := NewClient(
		WithBaseURL("https://example.test/v0/"),
		WithHTTPClient(newTestHTTPClient(func(r *http.Request) *http.Response {
			if r.URL.Path != "/v0/airline/RYR" {
				t.Fatalf("path = %q", r.URL.Path)
			}
			return jsonResponse(http.StatusOK, `{
				"response": [
					{
						"name": "Ryanair",
						"icao": "RYR",
						"iata": "FR",
						"country": "Ireland",
						"country_iso": "IE",
						"callsign": "RYANAIR"
					}
				]
			}`)
		})),
	)
	if err != nil {
		t.Fatal(err)
	}

	airlines, err := client.Airline(context.Background(), "RYR")
	if err != nil {
		t.Fatal(err)
	}
	if len(airlines) != 1 || airlines[0].ICAO != "RYR" {
		t.Fatalf("airlines = %#v", airlines)
	}
}

func TestModeSToNNumber(t *testing.T) {
	client, err := NewClient(
		WithBaseURL("https://example.test/v0/"),
		WithHTTPClient(newTestHTTPClient(func(r *http.Request) *http.Response {
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

type roundTripFunc func(*http.Request) *http.Response

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r), nil
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
