package search

import (
	"fmt"
	"testing"
	"time"

	"github.com/mattneto928/deadhead/clients"
	"github.com/mattneto928/deadhead/models"
)

// buildManifest creates a minimal CityResponse for tests.
// depTime is the departure time for the outbound flight.
// pricecents is the one-way price in cents (e.g. 25000 = $250).
func buildManifest(depAirport, arrAirport string, depTime time.Time, pricesCents ...int) *clients.CityResponse {
	m := &clients.CityResponse{}
	m.Info.To = clients.Location{City: "Los Angeles", State: "CA"}
	m.Info.From = clients.Location{City: "New York", State: "NY"}
	m.Airlines = map[string]struct{ Name string `json:"name"` }{
		"AA": {Name: "American Airlines"},
	}
	m.Flights = map[string]clients.Flight{}
	m.Itineraries.Outbound = nil

	for i, cents := range pricesCents {
		key := fmt.Sprintf("f%d", i+1)
		f := clients.Flight{
			Count:    1,
			Duration: 300,
			Segments: makeSegments(depAirport, arrAirport, depTime),
		}
		m.Flights[key] = f
		m.Itineraries.Outbound = append(m.Itineraries.Outbound, clients.InOutBoundFlight{
			Flight:      key,
			OneWayPrice: cents,
		})
	}
	return m
}

func makeSegments(dep, arr string, depTime time.Time) []struct {
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
} {
	return []struct {
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
	}{
		{
			Airline:      "AA",
			FlightNumber: 1,
			Departure: struct {
				Time    time.Time `json:"time"`
				Airport string    `json:"airport"`
			}{Time: depTime, Airport: dep},
			Arrival: struct {
				Time    time.Time `json:"time"`
				Airport string    `json:"airport"`
			}{Time: depTime.Add(5 * time.Hour), Airport: arr},
			Duration: 300,
		},
	}
}

var testDepTime = time.Date(2026, 5, 12, 10, 0, 0, 0, time.Local)

func TestBuildCitySummary_NoFlights(t *testing.T) {
	req, _ := models.NewRequest("NYC", "LAX", "2026-05-12", "", 1)
	manifest := &clients.CityResponse{}
	manifest.Info.To = clients.Location{City: "Los Angeles", State: "CA"}
	manifest.Airlines = map[string]struct{ Name string `json:"name"` }{}
	manifest.Flights = map[string]clients.Flight{}

	s := buildCitySummary(req, manifest)
	if len(s.Leaving) != 0 {
		t.Errorf("expected 0 leaving flights, got %d", len(s.Leaving))
	}
}

func TestBuildCitySummary_OutboundOnly(t *testing.T) {
	req, _ := models.NewRequest("NYC", "LAX", "2026-05-12", "", 1)
	manifest := buildManifest("JFK", "LAX", testDepTime, 25000) // $250

	s := buildCitySummary(req, manifest)
	if len(s.Leaving) != 1 {
		t.Fatalf("expected 1 leaving flight, got %d", len(s.Leaving))
	}
	if s.Leaving[0].Price != 250 {
		t.Errorf("Price = %d, want 250", s.Leaving[0].Price)
	}
	if s.MinLeavingPrice != 250 {
		t.Errorf("MinLeavingPrice = %d, want 250", s.MinLeavingPrice)
	}
	if s.FullName != "Los Angeles, CA" {
		t.Errorf("FullName = %q, want %q", s.FullName, "Los Angeles, CA")
	}
}

func TestBuildCitySummary_MultipleOutbound_TracksMinPrice(t *testing.T) {
	req, _ := models.NewRequest("NYC", "LAX", "2026-05-12", "", 1)
	manifest := buildManifest("JFK", "LAX", testDepTime, 35000, 20000, 28000) // $350, $200, $280

	s := buildCitySummary(req, manifest)
	if len(s.Leaving) != 3 {
		t.Fatalf("expected 3 leaving flights, got %d", len(s.Leaving))
	}
	if s.MinLeavingPrice != 200 {
		t.Errorf("MinLeavingPrice = %d, want 200", s.MinLeavingPrice)
	}
}

func TestBuildCitySummary_PriceFilter_ExcludesExpensive(t *testing.T) {
	req, _ := models.NewRequest("NYC", "LAX", "2026-05-12", "", 1)
	req.WithMaxPrice(200)
	manifest := buildManifest("JFK", "LAX", testDepTime, 30000, 15000) // $300 excluded, $150 ok

	s := buildCitySummary(req, manifest)
	if len(s.Leaving) != 1 {
		t.Errorf("expected 1 leaving flight after price filter, got %d", len(s.Leaving))
	}
	if s.Leaving[0].Price != 150 {
		t.Errorf("Price = %d, want 150", s.Leaving[0].Price)
	}
}

func TestBuildCitySummary_WithReturn_NilReturnDay_NoReturnFlights(t *testing.T) {
	// ReturningDay is zero (one-way) — inbound flights should be ignored.
	req, _ := models.NewRequest("NYC", "LAX", "2026-05-12", "", 1)
	manifest := buildManifest("JFK", "LAX", testDepTime, 25000)

	// Add an inbound itinerary to the manifest.
	manifest.Flights["r1"] = clients.Flight{
		Count:    1,
		Duration: 300,
		Segments: makeSegments("LAX", "JFK", testDepTime.Add(7*24*time.Hour)),
	}
	manifest.Itineraries.Inbound = append(manifest.Itineraries.Inbound, clients.InOutBoundFlight{
		Flight:      "r1",
		OneWayPrice: 20000,
	})

	s := buildCitySummary(req, manifest)
	if len(s.Returning) != 0 {
		t.Errorf("expected 0 returning flights for one-way trip, got %d", len(s.Returning))
	}
}

func TestBuildCitySummary_RoundTrip_MinRoundTripPrice(t *testing.T) {
	req, _ := models.NewRequest("NYC", "LAX", "2026-05-12", "2026-05-19", 1)
	manifest := buildManifest("JFK", "LAX", testDepTime, 25000) // $250 outbound

	retTime := time.Date(2026, 5, 19, 10, 0, 0, 0, time.Local)
	manifest.Flights["r1"] = clients.Flight{
		Count:    1,
		Duration: 300,
		Segments: makeSegments("LAX", "JFK", retTime),
	}
	manifest.Itineraries.Inbound = append(manifest.Itineraries.Inbound, clients.InOutBoundFlight{
		Flight:      "r1",
		OneWayPrice: 20000, // $200 return
	})

	s := buildCitySummary(req, manifest)
	if len(s.Leaving) != 1 {
		t.Fatalf("expected 1 leaving flight, got %d", len(s.Leaving))
	}
	if len(s.Returning) != 1 {
		t.Fatalf("expected 1 returning flight, got %d", len(s.Returning))
	}
	if s.MinRoundTripPrice != 450 {
		t.Errorf("MinRoundTripPrice = %d, want 450 (250+200)", s.MinRoundTripPrice)
	}
}

func TestBuildCitySummary_RoundTripPriceExceedsMax_ClearsFlights(t *testing.T) {
	req, _ := models.NewRequest("NYC", "LAX", "2026-05-12", "2026-05-19", 1)
	req.WithMaxPrice(400)
	manifest := buildManifest("JFK", "LAX", testDepTime, 25000) // $250

	retTime := time.Date(2026, 5, 19, 10, 0, 0, 0, time.Local)
	manifest.Flights["r1"] = clients.Flight{
		Count:    1,
		Duration: 300,
		Segments: makeSegments("LAX", "JFK", retTime),
	}
	manifest.Itineraries.Inbound = append(manifest.Itineraries.Inbound, clients.InOutBoundFlight{
		Flight:      "r1",
		OneWayPrice: 20000, // $200, round-trip = $450 > $400
	})

	s := buildCitySummary(req, manifest)
	// Round trip total ($450) exceeds max ($400) — both lists should be cleared.
	if len(s.Leaving) != 0 {
		t.Errorf("expected Leaving cleared when round-trip exceeds max, got %d", len(s.Leaving))
	}
	if len(s.Returning) != 0 {
		t.Errorf("expected Returning cleared when round-trip exceeds max, got %d", len(s.Returning))
	}
}

func TestBuildCitySummary_HiddenCityDetection(t *testing.T) {
	req, _ := models.NewRequest("NYC", "LAX", "2026-05-12", "", 1)
	manifest := &clients.CityResponse{}
	manifest.Info.To = clients.Location{City: "Los Angeles", State: "CA"}
	manifest.Airlines = map[string]struct{ Name string `json:"name"` }{
		"TK": {Name: "Turkish Airlines"},
	}

	// A 2-segment flight where final arrival is SFO, not LAX (hidden-city).
	seg1dep := testDepTime
	seg1 := struct {
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
	}{
		Airline:      "TK",
		FlightNumber: 1,
		Departure: struct {
			Time    time.Time `json:"time"`
			Airport string    `json:"airport"`
		}{Time: seg1dep, Airport: "JFK"},
		Arrival: struct {
			Time    time.Time `json:"time"`
			Airport string    `json:"airport"`
		}{Time: seg1dep.Add(5 * time.Hour), Airport: "LAX"},
		Duration: 300,
	}
	seg2 := seg1
	seg2.Departure = seg1.Arrival
	seg2.Arrival = struct {
		Time    time.Time `json:"time"`
		Airport string    `json:"airport"`
	}{Time: seg1dep.Add(7 * time.Hour), Airport: "SFO"}

	hiddenFlight := clients.Flight{
		Count:    1, // non-stop per-leg (hidden-city via layover)
		Duration: 420,
		Segments: []struct {
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
		}{seg1, seg2},
	}
	manifest.Flights = map[string]clients.Flight{"hc1": hiddenFlight}
	manifest.Itineraries.Outbound = []clients.InOutBoundFlight{
		{Flight: "hc1", OneWayPrice: 30000},
	}

	s := buildCitySummary(req, manifest)
	if len(s.Leaving) != 1 {
		t.Fatalf("expected 1 leaving flight, got %d", len(s.Leaving))
	}
	f := s.Leaving[0]
	if !f.IsHiddenCity {
		t.Error("expected IsHiddenCity = true")
	}
	if f.HiddenDestination != "SFO" {
		t.Errorf("HiddenDestination = %q, want SFO", f.HiddenDestination)
	}
	if f.Layovers != 1 {
		t.Errorf("Layovers = %d, want 1", f.Layovers)
	}
}

func TestBuildCitySummary_DirectFlight_NotHiddenCity(t *testing.T) {
	req, _ := models.NewRequest("NYC", "LAX", "2026-05-12", "", 1)
	manifest := buildManifest("JFK", "LAX", testDepTime, 25000)

	s := buildCitySummary(req, manifest)
	if len(s.Leaving) != 1 {
		t.Fatalf("expected 1 leaving flight")
	}
	if s.Leaving[0].IsHiddenCity {
		t.Error("direct flight should not be flagged as hidden-city")
	}
	if s.Leaving[0].Layovers != 0 {
		t.Errorf("Layovers = %d, want 0 for direct flight", s.Leaving[0].Layovers)
	}
}
