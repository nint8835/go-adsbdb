# go-adsbdb

Go client library for [ADSBDB](https://www.adsbdb.com/), a public aircraft,
airline, and flight route API.

## Installation

```sh
go get github.com/nint8835/go-adsbdb
```

## Usage

```go
package main

import (
	"context"
	"errors"
	"fmt"
	"log"

	adsbdb "github.com/nint8835/go-adsbdb"
)

func main() {
	client, err := adsbdb.NewClient()
	if err != nil {
		log.Fatal(err)
	}

	aircraft, err := client.Aircraft(context.Background(), "4CA645")
	if err != nil {
		if errors.Is(err, adsbdb.ErrNotFound) {
			log.Fatal("aircraft not found")
		}
		log.Fatal(err)
	}

	fmt.Printf("%s %s\n", aircraft.Registration, aircraft.RegisteredOwner)
}
```

## API

The client supports the public read-only ADSBDB endpoints:

- `Aircraft(ctx, identifier)` and `RandomAircraft(ctx)`
- `AircraftWithCallsign(ctx, identifier, callsign)`
- `Callsign(ctx, callsign)` and `RandomCallsign(ctx)`
- `Airline(ctx, code)` and `RandomAirline(ctx)`
- `Stats(ctx)`
- `ModeSToNNumber(ctx, modeS)`
- `NNumberToModeS(ctx, nNumber)`

`identifier` can be a Mode-S hex code or an aircraft registration. Airline
codes can be ICAO or IATA codes.

## Configuration

```go
client, err := adsbdb.NewClient(
	adsbdb.WithUserAgent("my-app/1.0"),
	adsbdb.WithBaseURL("https://api.adsbdb.com/v0/"),
)
```

Use `WithHTTPClient` to configure timeouts, transports, or test clients.

## Errors

Non-2xx API responses return `*adsbdb.APIError`. ADSBDB 404 responses also
match `adsbdb.ErrNotFound`:

```go
if errors.Is(err, adsbdb.ErrNotFound) {
	// unknown aircraft, callsign, airline, or conversion input
}
```
