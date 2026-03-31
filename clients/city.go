package clients

import (
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/mattneto928/deadhead/models"
)

// CityAPIBase is the base URL for the city flight search endpoint.
// Override in tests to point at a mock server.
var CityAPIBase = "https://skiplagged.com/api/search.php"

// Flight represents a single flight (which may have multiple segments) returned by the API.
type Flight struct {
	Segments []struct {
		Airline      string `json:"airline"`
		FlightNumber int    `json:"flight_number"`
		Departure    struct {
			Time    time.Time `json:"time"`
			Airport string    `json:"airport"`
		} `json:"departure"`
		Arrival struct {
			Time    time.Time `json:"time"`
			Airport string    `json:"airport"`
		} `json:"arrival"`
		Duration int `json:"duration"`
	} `json:"segments"`
	Duration int    `json:"duration"`
	Count    int    `json:"count"`
	Data     string `json:"data"`
}

// InOutBoundFlight is an outbound or inbound itinerary entry that references a Flight by key.
type InOutBoundFlight struct {
	Data              string `json:"data"`
	Flight            string `json:"flight"`
	MinRoundTripPrice int    `json:"min_round_trip_price,omitempty"`
	OneWayPrice       int    `json:"one_way_price,omitempty"`
}

// CityResponse represents the full API response for a city-pair flight search.
type CityResponse struct {
	Airlines map[string]struct {
		Name string `json:"name"`
	} `json:"airlines"`
	Cities map[string]struct {
		Name string `json:"name"`
	} `json:"cities"`
	Airports map[string]struct {
		Name string `json:"name"`
	} `json:"airports"`
	Flights     map[string]Flight `json:"flights"`
	Itineraries struct {
		Outbound []InOutBoundFlight `json:"outbound"`
		Inbound  []InOutBoundFlight `json:"inbound"`
	} `json:"itineraries"`
	Info struct {
		From Location `json:"from"`
		To   Location `json:"to"`
	} `json:"info"`
	Duration float64 `json:"duration"`
}

// GetFlightsToCity fetches one-way flight options for a city pair.
// The return parameter is intentionally left blank to force one-way pricing;
// passing a return date causes the API to price the trip as round-trip.
func GetFlightsToCity(req *models.Request) (*CityResponse, error) {
	rawURL := fmt.Sprintf(
		"%s?from=%s&to=%s&depart=%s&return=&poll=true&format=v3&counts[adults]=%d&counts[children]=0",
		CityAPIBase,
		url.QueryEscape(req.HomeCity),
		url.QueryEscape(req.TripCity),
		req.LeavingDay.Format("2006-01-02"),
		req.Travelers,
	)
	res, err := HTTPClient.Get(rawURL)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var payload CityResponse
	if err = json.NewDecoder(res.Body).Decode(&payload); err != nil {
		return nil, err
	}

	return &payload, nil
}
