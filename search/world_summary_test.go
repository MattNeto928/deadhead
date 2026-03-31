package search

import (
	"testing"

	"github.com/mattneto928/deadhead/clients"
	"github.com/mattneto928/deadhead/models"
)

func makeCountryResponse(trips []clients.Trip, cities map[string]clients.City) *clients.CountryResponse {
	r := &clients.CountryResponse{
		Cities: cities,
		Trips:  trips,
	}
	return r
}

func TestBuildWorldSummaries_Basic(t *testing.T) {
	req, _ := models.NewRequest("NYC", "", "2026-05-12", "2026-05-19", 1)
	cities := map[string]clients.City{
		"LAX": {Name: "Los Angeles", Region: "CA"},
		"ORD": {Name: "Chicago", Region: "IL"},
	}
	trips := []clients.Trip{
		{City: "LAX", Cost: 25000}, // $250
		{City: "ORD", Cost: 15000}, // $150
	}
	manifest := makeCountryResponse(trips, cities)

	summaries := buildWorldSummaries(req, manifest)
	if len(summaries) != 2 {
		t.Fatalf("expected 2 summaries, got %d", len(summaries))
	}
}

func TestBuildWorldSummaries_PriceFilter(t *testing.T) {
	req, _ := models.NewRequest("NYC", "", "2026-05-12", "2026-05-19", 1)
	req.WithMaxPrice(200)
	cities := map[string]clients.City{
		"LAX": {Name: "Los Angeles", Region: "CA"},
		"ORD": {Name: "Chicago", Region: "IL"},
	}
	trips := []clients.Trip{
		{City: "LAX", Cost: 30000}, // $300 — excluded
		{City: "ORD", Cost: 15000}, // $150 — included
	}
	manifest := makeCountryResponse(trips, cities)

	summaries := buildWorldSummaries(req, manifest)
	if len(summaries) != 1 {
		t.Fatalf("expected 1 summary after price filter, got %d", len(summaries))
	}
	if summaries[0].Name != "ORD" {
		t.Errorf("Name = %q, want ORD", summaries[0].Name)
	}
}

func TestBuildWorldSummaries_DuplicateCityKeepsCheapest(t *testing.T) {
	req, _ := models.NewRequest("NYC", "", "2026-05-12", "2026-05-19", 1)
	cities := map[string]clients.City{
		"LAX": {Name: "Los Angeles", Region: "CA"},
	}
	// Two trips to the same city with different prices.
	trips := []clients.Trip{
		{City: "LAX", Cost: 30000},
		{City: "LAX", Cost: 20000}, // cheaper
		{City: "LAX", Cost: 25000},
	}
	manifest := makeCountryResponse(trips, cities)

	summaries := buildWorldSummaries(req, manifest)
	if len(summaries) != 1 {
		t.Fatalf("expected 1 summary (deduped by city name), got %d", len(summaries))
	}
	if summaries[0].MinRoundTripPrice != 200 {
		t.Errorf("MinRoundTripPrice = %d, want 200 (cheapest of the three)", summaries[0].MinRoundTripPrice)
	}
}

func TestBuildWorldSummaries_EmptyTrips(t *testing.T) {
	req, _ := models.NewRequest("NYC", "", "2026-05-12", "2026-05-19", 1)
	manifest := makeCountryResponse(nil, map[string]clients.City{})

	summaries := buildWorldSummaries(req, manifest)
	if len(summaries) != 0 {
		t.Errorf("expected 0 summaries for empty trips, got %d", len(summaries))
	}
}

func TestBuildWorldSummaries_FullName(t *testing.T) {
	req, _ := models.NewRequest("NYC", "", "2026-05-12", "2026-05-19", 1)
	cities := map[string]clients.City{
		"CDG": {Name: "Paris", Region: "France"},
	}
	trips := []clients.Trip{
		{City: "CDG", Cost: 50000},
	}
	manifest := makeCountryResponse(trips, cities)

	summaries := buildWorldSummaries(req, manifest)
	if len(summaries) != 1 {
		t.Fatalf("expected 1 summary, got %d", len(summaries))
	}
	if summaries[0].FullName != "Paris, France" {
		t.Errorf("FullName = %q, want %q", summaries[0].FullName, "Paris, France")
	}
}

func TestBuildWorldSummaries_NoPriceFilter_AllIncluded(t *testing.T) {
	req, _ := models.NewRequest("NYC", "", "2026-05-12", "2026-05-19", 1)
	// MaxPrice = 0 means no limit
	cities := map[string]clients.City{
		"LAX": {Name: "Los Angeles", Region: "CA"},
		"NRT": {Name: "Tokyo", Region: "Japan"},
		"CDG": {Name: "Paris", Region: "France"},
	}
	trips := []clients.Trip{
		{City: "LAX", Cost: 20000},
		{City: "NRT", Cost: 100000},
		{City: "CDG", Cost: 80000},
	}
	manifest := makeCountryResponse(trips, cities)

	summaries := buildWorldSummaries(req, manifest)
	if len(summaries) != 3 {
		t.Errorf("expected 3 summaries with no price filter, got %d", len(summaries))
	}
}
