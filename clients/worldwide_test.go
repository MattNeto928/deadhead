package clients_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mattneto928/deadhead/clients"
	"github.com/mattneto928/deadhead/models"
)

// saveWorldwideState saves and restores CountryAPIBase and HTTPClient around a test.
func saveWorldwideState(t *testing.T) {
	t.Helper()
	origBase := clients.CountryAPIBase
	origClient := clients.HTTPClient
	t.Cleanup(func() {
		clients.CountryAPIBase = origBase
		clients.HTTPClient = origClient
	})
}

func makeMockCountryResponse() clients.CountryResponse {
	return clients.CountryResponse{
		Cities: map[string]clients.City{
			"LAX": {Name: "Los Angeles", Region: "CA", Airports: []string{"LAX"}},
			"ORD": {Name: "Chicago", Region: "IL", Airports: []string{"ORD"}},
		},
		Trips: []clients.Trip{
			{City: "LAX", Cost: 25000, HiddenCity: false},
			{City: "ORD", Cost: 18000, HiddenCity: true},
		},
	}
}

func TestGetWorldwideFlightsFromCity_Success(t *testing.T) {
	saveWorldwideState(t)

	mock := makeMockCountryResponse()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(mock)
	}))
	defer server.Close()

	clients.CountryAPIBase = server.URL
	clients.HTTPClient = server.Client()

	req, err := models.NewRequest("NYC", "", "2026-05-12", "2026-05-19", 1)
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}

	resp, err := clients.GetWorldwideFlightsFromCity(req)
	if err != nil {
		t.Fatalf("GetWorldwideFlightsFromCity: %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
	if len(resp.Trips) != 2 {
		t.Errorf("expected 2 trips, got %d", len(resp.Trips))
	}
	if _, ok := resp.Cities["LAX"]; !ok {
		t.Error("expected LAX in cities map")
	}
}

func TestGetWorldwideFlightsFromCity_RequestContainsQueryParams(t *testing.T) {
	saveWorldwideState(t)

	var capturedURL string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedURL = r.URL.String()
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(clients.CountryResponse{
			Cities: map[string]clients.City{},
			Trips:  nil,
		})
	}))
	defer server.Close()

	clients.CountryAPIBase = server.URL
	clients.HTTPClient = server.Client()

	req, _ := models.NewRequest("NYC", "", "2026-05-12", "2026-05-19", 1)
	_, _ = clients.GetWorldwideFlightsFromCity(req)

	checks := []string{"from=NYC", "depart=2026-05-12", "return=2026-05-19"}
	for _, check := range checks {
		if !containsSubstr(capturedURL, check) {
			t.Errorf("URL %q missing %q", capturedURL, check)
		}
	}
}

func TestGetWorldwideFlightsFromCity_HTTPError(t *testing.T) {
	saveWorldwideState(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	server.Close()

	clients.CountryAPIBase = server.URL
	clients.HTTPClient = server.Client()

	req, _ := models.NewRequest("NYC", "", "2026-05-12", "2026-05-19", 1)
	_, err := clients.GetWorldwideFlightsFromCity(req)
	if err == nil {
		t.Fatal("expected error for HTTP connection failure")
	}
}

func TestGetWorldwideFlightsFromCity_InvalidJSON(t *testing.T) {
	saveWorldwideState(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("{bad json"))
	}))
	defer server.Close()

	clients.CountryAPIBase = server.URL
	clients.HTTPClient = server.Client()

	req, _ := models.NewRequest("NYC", "", "2026-05-12", "2026-05-19", 1)
	_, err := clients.GetWorldwideFlightsFromCity(req)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestGetWorldwideFlightsFromCity_URLEncoding(t *testing.T) {
	saveWorldwideState(t)

	var capturedQuery string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(clients.CountryResponse{
			Cities: map[string]clients.City{},
			Trips:  nil,
		})
	}))
	defer server.Close()

	clients.CountryAPIBase = server.URL
	clients.HTTPClient = server.Client()

	req, _ := models.NewRequest("N Y C", "", "2026-05-12", "2026-05-19", 1)
	_, _ = clients.GetWorldwideFlightsFromCity(req)

	// Space in "N Y C" must be encoded — should not appear raw in query.
	if containsSubstr(capturedQuery, "from=N Y C") {
		t.Errorf("space in city code not URL-encoded, query: %s", capturedQuery)
	}
}

func TestGetWorldwideFlightsFromCity_HiddenCityTrips(t *testing.T) {
	saveWorldwideState(t)

	mock := clients.CountryResponse{
		Cities: map[string]clients.City{
			"BKK": {Name: "Bangkok", Region: "Thailand"},
		},
		Trips: []clients.Trip{
			{City: "BKK", Cost: 50000, HiddenCity: true, RegularCost: 80000},
		},
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(mock)
	}))
	defer server.Close()

	clients.CountryAPIBase = server.URL
	clients.HTTPClient = server.Client()

	req, _ := models.NewRequest("NYC", "", "2026-05-12", "2026-05-19", 1)
	resp, err := clients.GetWorldwideFlightsFromCity(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Trips) != 1 {
		t.Fatalf("expected 1 trip, got %d", len(resp.Trips))
	}
	if !resp.Trips[0].HiddenCity {
		t.Error("expected HiddenCity = true")
	}
	if resp.Trips[0].RegularCost != 80000 {
		t.Errorf("RegularCost = %d, want 80000", resp.Trips[0].RegularCost)
	}
}
