package clients_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/mattneto928/deadhead/clients"
	"github.com/mattneto928/deadhead/models"
)

// saveCityState saves and restores CityAPIBase and HTTPClient around a test.
func saveCityState(t *testing.T) {
	t.Helper()
	origBase := clients.CityAPIBase
	origClient := clients.HTTPClient
	t.Cleanup(func() {
		clients.CityAPIBase = origBase
		clients.HTTPClient = origClient
	})
}

func makeMockCityResponse() clients.CityResponse {
	dep := time.Date(2026, 5, 12, 15, 0, 0, 0, time.UTC)
	arr := dep.Add(5 * time.Hour)
	return clients.CityResponse{
		Airlines: map[string]struct {
			Name string `json:"name"`
		}{
			"AA": {Name: "American Airlines"},
		},
		Flights: map[string]clients.Flight{
			"f1": {
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
						}{Time: dep, Airport: "JFK"},
						Arrival: struct {
							Time    time.Time `json:"time"`
							Airport string    `json:"airport"`
						}{Time: arr, Airport: "LAX"},
						Duration: 300,
					},
				},
			},
		},
	}
}

func TestGetFlightsToCity_Success(t *testing.T) {
	saveCityState(t)

	mock := makeMockCityResponse()
	mock.Info.To = clients.Location{City: "Los Angeles", State: "CA"}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(mock)
	}))
	defer server.Close()

	clients.CityAPIBase = server.URL
	clients.HTTPClient = server.Client()

	req, err := models.NewRequest("NYC", "LAX", "2026-05-12", "", 1)
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}

	resp, err := clients.GetFlightsToCity(req)
	if err != nil {
		t.Fatalf("GetFlightsToCity: %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
	if resp.Info.To.City != "Los Angeles" {
		t.Errorf("Info.To.City = %q, want %q", resp.Info.To.City, "Los Angeles")
	}
	if _, ok := resp.Flights["f1"]; !ok {
		t.Error("expected flight f1 in response")
	}
}

func TestGetFlightsToCity_RequestContainsQueryParams(t *testing.T) {
	saveCityState(t)

	var capturedURL string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedURL = r.URL.String()
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(clients.CityResponse{
			Airlines: map[string]struct {
				Name string `json:"name"`
			}{},
			Flights: map[string]clients.Flight{},
		})
	}))
	defer server.Close()

	clients.CityAPIBase = server.URL
	clients.HTTPClient = server.Client()

	req, _ := models.NewRequest("NYC", "LAX", "2026-05-12", "", 2)
	_, _ = clients.GetFlightsToCity(req)

	// Verify the request URL contained the expected parameters.
	checks := []string{"from=NYC", "to=LAX", "depart=2026-05-12", "adults"}
	for _, check := range checks {
		found := false
		for _, part := range splitQuery(capturedURL) {
			if part == check || containsSubstr(part, check) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("URL %q missing parameter containing %q", capturedURL, check)
		}
	}
}

func TestGetFlightsToCity_HTTPError(t *testing.T) {
	saveCityState(t)

	// Point at a closed server to force connection error.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	server.Close() // close immediately

	clients.CityAPIBase = server.URL
	clients.HTTPClient = server.Client()

	req, _ := models.NewRequest("NYC", "LAX", "2026-05-12", "", 1)
	_, err := clients.GetFlightsToCity(req)
	if err == nil {
		t.Fatal("expected error for HTTP connection failure")
	}
}

func TestGetFlightsToCity_InvalidJSON(t *testing.T) {
	saveCityState(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("not json at all"))
	}))
	defer server.Close()

	clients.CityAPIBase = server.URL
	clients.HTTPClient = server.Client()

	req, _ := models.NewRequest("NYC", "LAX", "2026-05-12", "", 1)
	_, err := clients.GetFlightsToCity(req)
	if err == nil {
		t.Fatal("expected error for invalid JSON response")
	}
}

func TestGetFlightsToCity_URLEncoding(t *testing.T) {
	saveCityState(t)

	var capturedQuery string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(clients.CityResponse{
			Airlines: map[string]struct{ Name string `json:"name"` }{},
			Flights:  map[string]clients.Flight{},
		})
	}))
	defer server.Close()

	clients.CityAPIBase = server.URL
	clients.HTTPClient = server.Client()

	// A city code with a special character to verify URL encoding.
	req, _ := models.NewRequest("NYC", "L A X", "2026-05-12", "", 1)
	_, _ = clients.GetFlightsToCity(req)

	// The space in "L A X" should be encoded in the query string.
	if containsSubstr(capturedQuery, "to=L A X") {
		t.Errorf("expected space in 'L A X' to be URL-encoded, but query was: %s", capturedQuery)
	}
}

// splitQuery is a test helper that checks if a query string contains a substring.
func splitQuery(url string) []string {
	for i, c := range url {
		if c == '?' {
			return []string{url[i+1:]}
		}
	}
	return []string{url}
}

func containsSubstr(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && containsSubstrHelper(s, sub))
}

func containsSubstrHelper(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
