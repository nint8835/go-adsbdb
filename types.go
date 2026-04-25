package adsbdb

type aircraftResponse struct {
	Aircraft *Aircraft `json:"aircraft"`
}

type callsignResponse struct {
	FlightRoute *FlightRoute `json:"flightroute"`
}

// Aircraft is an aircraft record returned by ADSBDB.
type Aircraft struct {
	Type                            string  `json:"type"`
	ICAOType                        string  `json:"icao_type"`
	Manufacturer                    string  `json:"manufacturer"`
	ModeS                           string  `json:"mode_s"`
	Registration                    string  `json:"registration"`
	RegisteredOwnerCountryISOName   string  `json:"registered_owner_country_iso_name"`
	RegisteredOwnerCountryName      string  `json:"registered_owner_country_name"`
	RegisteredOwnerOperatorFlagCode *string `json:"registered_owner_operator_flag_code"`
	RegisteredOwner                 string  `json:"registered_owner"`
	URLPhoto                        *string `json:"url_photo"`
	URLPhotoThumbnail               *string `json:"url_photo_thumbnail"`
}

// AircraftAndFlightRoute is returned by the aircraft endpoint when a callsign
// query parameter is supplied.
type AircraftAndFlightRoute struct {
	Aircraft    Aircraft     `json:"aircraft"`
	FlightRoute *FlightRoute `json:"flightroute,omitempty"`
}

// FlightRoute is a route associated with a callsign.
type FlightRoute struct {
	Callsign     string   `json:"callsign"`
	CallsignICAO *string  `json:"callsign_icao"`
	CallsignIATA *string  `json:"callsign_iata"`
	Airline      *Airline `json:"airline"`
	Origin       Airport  `json:"origin"`
	Midpoint     *Airport `json:"midpoint,omitempty"`
	Destination  Airport  `json:"destination"`
}

// Airline is an airline record returned by ADSBDB.
type Airline struct {
	Name       string  `json:"name"`
	ICAO       string  `json:"icao"`
	IATA       *string `json:"iata"`
	Country    string  `json:"country"`
	CountryISO string  `json:"country_iso"`
	Callsign   *string `json:"callsign"`
}

// Airport is an airport record embedded in a flight route.
type Airport struct {
	CountryISOName string  `json:"country_iso_name"`
	CountryName    string  `json:"country_name"`
	Elevation      float64 `json:"elevation"`
	IATACode       string  `json:"iata_code"`
	ICAOCode       string  `json:"icao_code"`
	Latitude       float64 `json:"latitude"`
	Longitude      float64 `json:"longitude"`
	Municipality   string  `json:"municipality"`
	Name           string  `json:"name"`
}

// Stats contains API usage statistics.
type Stats struct {
	Daily StatsPeriod `json:"daily"`
	Total StatsPeriod `json:"total"`
}

// StatsPeriod contains request counts for a period.
type StatsPeriod struct {
	Aircraft  []StatsEntry `json:"aircraft"`
	Airline   []StatsEntry `json:"airline"`
	Callsign  []StatsEntry `json:"callsign"`
	ModeS     []StatsEntry `json:"mode_s"`
	NNumber   []StatsEntry `json:"n_number"`
	Online    []StatsEntry `json:"online"`
	Stats     []StatsEntry `json:"stats"`
	Aggregate int          `json:"aggregate"`
}

// StatsEntry is a request count for an API path.
type StatsEntry struct {
	URL   string `json:"url"`
	Count int    `json:"count"`
}
