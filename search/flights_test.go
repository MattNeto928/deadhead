package search

import (
	"testing"
	"time"

	"github.com/mattneto928/deadhead/clients"
	"github.com/mattneto928/deadhead/models"
)

// ---- helpers ----

// makeFlight builds a clients.Flight with a single segment for use in tests.
func makeFlight(key, depAirport, arrAirport string, depTime time.Time) (string, clients.Flight) {
	f := clients.Flight{
		Count:    1,
		Duration: 300,
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
		}{
			{
				Airline:      "AA",
				FlightNumber: 100,
				Departure: struct {
					Time    time.Time `json:"time"`
					Airport string    `json:"airport"`
				}{Time: depTime, Airport: depAirport},
				Arrival: struct {
					Time    time.Time `json:"time"`
					Airport string    `json:"airport"`
				}{Time: depTime.Add(5 * time.Hour), Airport: arrAirport},
				Duration: 300,
			},
		},
	}
	return key, f
}

// makeRequest builds a basic one-way request with optional criteria applied.
func makeRequest(t *testing.T) *models.Request {
	t.Helper()
	req, err := models.NewRequest("NYC", "LAX", "2026-05-12", "", 1)
	if err != nil {
		t.Fatalf("makeRequest: %v", err)
	}
	return req
}

// baseDepTime is 10:00 AM on the departure date.
var baseDepTime = time.Date(2026, 5, 12, 10, 0, 0, 0, time.Local)

// ---- flightMeetsCriteria ----

func TestFlightMeetsCriteria_NoConstraints(t *testing.T) {
	req := makeRequest(t)
	key, flight := makeFlight("f1", "JFK", "LAX", baseDepTime)
	flights := map[string]clients.Flight{key: flight}
	bound := clients.InOutBoundFlight{Flight: key, OneWayPrice: 25000} // $250

	got, err := flightMeetsCriteria(flights, bound, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil {
		t.Fatal("expected non-nil flight")
	}
}

func TestFlightMeetsCriteria_PriceTooExpensive(t *testing.T) {
	req := makeRequest(t)
	req.WithMaxPrice(200)
	key, flight := makeFlight("f1", "JFK", "LAX", baseDepTime)
	flights := map[string]clients.Flight{key: flight}
	// OneWayPrice is in cents: $300 = 30000
	bound := clients.InOutBoundFlight{Flight: key, OneWayPrice: 30000}

	_, err := flightMeetsCriteria(flights, bound, req)
	if err == nil {
		t.Fatal("expected error for price too expensive")
	}
}

func TestFlightMeetsCriteria_ZeroOneWayPrice_Rejected(t *testing.T) {
	req := makeRequest(t)
	req.WithMaxPrice(500)
	key, flight := makeFlight("f1", "JFK", "LAX", baseDepTime)
	flights := map[string]clients.Flight{key: flight}
	bound := clients.InOutBoundFlight{Flight: key, OneWayPrice: 0} // no price

	_, err := flightMeetsCriteria(flights, bound, req)
	if err == nil {
		t.Fatal("expected error when OneWayPrice is 0 and max price is set")
	}
}

func TestFlightMeetsCriteria_RoundTripTooExpensive(t *testing.T) {
	req := makeRequest(t)
	req.WithMaxPrice(200)
	key, flight := makeFlight("f1", "JFK", "LAX", baseDepTime)
	flights := map[string]clients.Flight{key: flight}
	// OneWayPrice $150 (ok), but MinRoundTripPrice $350 (too high)
	bound := clients.InOutBoundFlight{Flight: key, OneWayPrice: 15000, MinRoundTripPrice: 35000}

	_, err := flightMeetsCriteria(flights, bound, req)
	if err == nil {
		t.Fatal("expected error for round-trip price too expensive")
	}
}

func TestFlightMeetsCriteria_FlightNotFound(t *testing.T) {
	req := makeRequest(t)
	flights := map[string]clients.Flight{}
	bound := clients.InOutBoundFlight{Flight: "missing-key", OneWayPrice: 10000}

	_, err := flightMeetsCriteria(flights, bound, req)
	if err == nil {
		t.Fatal("expected error for flight not found")
	}
}

func TestFlightMeetsCriteria_MultiLegRejected(t *testing.T) {
	req := makeRequest(t)
	key, flight := makeFlight("f1", "JFK", "LAX", baseDepTime)
	flight.Count = 2 // multi-leg
	flights := map[string]clients.Flight{key: flight}
	bound := clients.InOutBoundFlight{Flight: key, OneWayPrice: 10000}

	_, err := flightMeetsCriteria(flights, bound, req)
	if err == nil {
		t.Fatal("expected error for multi-leg flight")
	}
}

func TestFlightMeetsCriteria_NoPriceConstraint_AnyPriceAllowed(t *testing.T) {
	req := makeRequest(t) // MaxPrice defaults to 0 = no limit
	key, flight := makeFlight("f1", "JFK", "LAX", baseDepTime)
	flights := map[string]clients.Flight{key: flight}
	bound := clients.InOutBoundFlight{Flight: key, OneWayPrice: 1000000} // very expensive

	got, err := flightMeetsCriteria(flights, bound, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil {
		t.Fatal("expected non-nil flight when no price constraint")
	}
}

// ---- flightMeetsLeavingCriteria ----

func TestFlightMeetsLeavingCriteria_Passes(t *testing.T) {
	req := makeRequest(t)
	req.WithLeavingCriteria(8, 20) // depart between 8am and 8pm
	key, flight := makeFlight("f1", "JFK", "LAX", baseDepTime) // 10am
	flights := map[string]clients.Flight{key: flight}
	bound := clients.InOutBoundFlight{Flight: key, OneWayPrice: 10000}

	got, err := flightMeetsLeavingCriteria(flights, bound, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil {
		t.Fatal("expected non-nil flight")
	}
}

func TestFlightMeetsLeavingCriteria_TooEarly(t *testing.T) {
	req := makeRequest(t)
	req.WithLeavingCriteria(12, 22) // must depart after noon
	key, flight := makeFlight("f1", "JFK", "LAX", baseDepTime) // 10am — too early
	flights := map[string]clients.Flight{key: flight}
	bound := clients.InOutBoundFlight{Flight: key, OneWayPrice: 10000}

	_, err := flightMeetsLeavingCriteria(flights, bound, req)
	if err == nil {
		t.Fatal("expected error for too-early departure")
	}
}

func TestFlightMeetsLeavingCriteria_TooLate(t *testing.T) {
	req := makeRequest(t)
	req.WithLeavingCriteria(6, 9) // must depart before 9am
	key, flight := makeFlight("f1", "JFK", "LAX", baseDepTime) // 10am — too late
	flights := map[string]clients.Flight{key: flight}
	bound := clients.InOutBoundFlight{Flight: key, OneWayPrice: 10000}

	_, err := flightMeetsLeavingCriteria(flights, bound, req)
	if err == nil {
		t.Fatal("expected error for too-late departure")
	}
}

func TestFlightMeetsLeavingCriteria_ExcludedDepartureAirport(t *testing.T) {
	req := makeRequest(t)
	req.WithExcludeAirportsCriteria([]string{"JFK"})
	key, flight := makeFlight("f1", "JFK", "LAX", baseDepTime)
	flights := map[string]clients.Flight{key: flight}
	bound := clients.InOutBoundFlight{Flight: key, OneWayPrice: 10000}

	_, err := flightMeetsLeavingCriteria(flights, bound, req)
	if err == nil {
		t.Fatal("expected error for excluded departure airport")
	}
}

func TestFlightMeetsLeavingCriteria_NonExcludedAirportPasses(t *testing.T) {
	req := makeRequest(t)
	req.WithExcludeAirportsCriteria([]string{"EWR"}) // exclude EWR, not JFK
	key, flight := makeFlight("f1", "JFK", "LAX", baseDepTime)
	flights := map[string]clients.Flight{key: flight}
	bound := clients.InOutBoundFlight{Flight: key, OneWayPrice: 10000}

	got, err := flightMeetsLeavingCriteria(flights, bound, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil {
		t.Fatal("expected non-nil flight")
	}
}

func TestFlightMeetsLeavingCriteria_NilLeaveCriteria_NoPanic(t *testing.T) {
	req := makeRequest(t)
	// Do NOT call WithLeavingCriteria — Leave should be nil
	if req.Criteria.Leave != nil {
		t.Fatal("precondition: Leave should be nil")
	}
	key, flight := makeFlight("f1", "JFK", "LAX", baseDepTime)
	flights := map[string]clients.Flight{key: flight}
	bound := clients.InOutBoundFlight{Flight: key, OneWayPrice: 10000}

	got, err := flightMeetsLeavingCriteria(flights, bound, req)
	if err != nil {
		t.Fatalf("unexpected error with nil Leave criteria: %v", err)
	}
	if got == nil {
		t.Fatal("expected non-nil flight")
	}
}

// ---- flightMeetsReturningCriteria ----

func TestFlightMeetsReturningCriteria_Passes(t *testing.T) {
	// Use a round-trip request so that WithReturningCriteria bases its times on ReturningDay.
	req, _ := models.NewRequest("NYC", "LAX", "2026-05-12", "2026-05-19", 1)
	req.WithReturningCriteria(8, 20) // depart between 8am and 8pm on May 19
	// Flight departs LAX at 10am May 19 — within window.
	depTime := time.Date(2026, 5, 19, 10, 0, 0, 0, time.Local)
	key, flight := makeFlight("r1", "LAX", "JFK", depTime)
	flights := map[string]clients.Flight{key: flight}
	bound := clients.InOutBoundFlight{Flight: key, OneWayPrice: 10000}

	got, err := flightMeetsReturningCriteria(flights, bound, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil {
		t.Fatal("expected non-nil flight")
	}
}

func TestFlightMeetsReturningCriteria_DepartureTooEarly(t *testing.T) {
	// Use a round-trip request so ReturningDay is set.
	req2, _ := models.NewRequest("NYC", "LAX", "2026-05-12", "2026-05-19", 1)
	req2.WithReturningCriteria(18, 23) // depart after 6pm

	// Flight departs LAX at 10am — too early.
	depTime := time.Date(2026, 5, 19, 10, 0, 0, 0, time.Local)
	key, flight := makeFlight("r1", "LAX", "JFK", depTime)
	flights := map[string]clients.Flight{key: flight}
	bound := clients.InOutBoundFlight{Flight: key, OneWayPrice: 10000}

	_, err := flightMeetsReturningCriteria(flights, bound, req2)
	if err == nil {
		t.Fatal("expected error for departure too early")
	}
}

func TestFlightMeetsReturningCriteria_DepartureTooLate(t *testing.T) {
	req2, _ := models.NewRequest("NYC", "LAX", "2026-05-12", "2026-05-19", 1)
	req2.WithReturningCriteria(6, 14) // depart before 2pm

	// Flight departs LAX at 6pm — too late.
	depTime := time.Date(2026, 5, 19, 18, 0, 0, 0, time.Local)
	key, flight := makeFlight("r1", "LAX", "JFK", depTime)
	flights := map[string]clients.Flight{key: flight}
	bound := clients.InOutBoundFlight{Flight: key, OneWayPrice: 10000}

	_, err := flightMeetsReturningCriteria(flights, bound, req2)
	if err == nil {
		t.Fatal("expected error for departure too late")
	}
}

func TestFlightMeetsReturningCriteria_ExcludedArrivalAirport(t *testing.T) {
	req := makeRequest(t)
	req.WithExcludeAirportsCriteria([]string{"JFK"})
	key, flight := makeFlight("r1", "LAX", "JFK", baseDepTime)
	flights := map[string]clients.Flight{key: flight}
	bound := clients.InOutBoundFlight{Flight: key, OneWayPrice: 10000}

	_, err := flightMeetsReturningCriteria(flights, bound, req)
	if err == nil {
		t.Fatal("expected error for excluded arrival airport")
	}
}

func TestFlightMeetsReturningCriteria_NilReturnCriteria_NoPanic(t *testing.T) {
	req := makeRequest(t)
	if req.Criteria.Return != nil {
		t.Fatal("precondition: Return should be nil")
	}
	key, flight := makeFlight("r1", "LAX", "JFK", baseDepTime)
	flights := map[string]clients.Flight{key: flight}
	bound := clients.InOutBoundFlight{Flight: key, OneWayPrice: 10000}

	got, err := flightMeetsReturningCriteria(flights, bound, req)
	if err != nil {
		t.Fatalf("unexpected error with nil Return criteria: %v", err)
	}
	if got == nil {
		t.Fatal("expected non-nil flight")
	}
}
