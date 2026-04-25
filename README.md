# go-adsbdb

Go client library for [adsbdb](https://www.adsbdb.com/), a public aircraft,
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

	aircraft, err := client.AircraftWithCallsign(context.Background(), "C0816E", "CJT620")
	if err != nil {
		if errors.Is(err, adsbdb.ErrNotFound) {
			log.Fatal("flight not found")
		}
		log.Fatal(err)
	}
	if aircraft.FlightRoute == nil {
		log.Fatal("route not found")
	}

	fmt.Printf("%s %s\n", aircraft.Aircraft.Registration, aircraft.FlightRoute.Destination.IATACode)
}
```

## Configuration

```go
client, err := adsbdb.NewClient(
	adsbdb.WithUserAgent("my-app/1.0"),
	adsbdb.WithBaseURL("https://api.adsbdb.com/v0/"),
)
```

Use `WithHTTPClient` to configure timeouts, transports, or test clients.

## Errors

Non-2xx API responses return `*adsbdb.APIError`. adsbdb 404 responses also
match `adsbdb.ErrNotFound`:

```go
if errors.Is(err, adsbdb.ErrNotFound) {
	// unknown aircraft, callsign, airline, or conversion input
}
```

## Development

Refresh recorded API fixtures with:

```sh
go generate ./...
```
