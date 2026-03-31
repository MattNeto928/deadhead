//go:build integration

// Package integration contains end-to-end tests that run against the real
// Skiplagged site. They require Google Chrome and a network connection.
//
// Run with:
//
//	go test -v -tags integration ./integration/
package integration_test

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/mattneto928/deadhead/clients"
	"github.com/mattneto928/deadhead/models"
	"github.com/mattneto928/deadhead/search"
)

// TestMain initialises the shared browser session once for the entire suite.
func TestMain(m *testing.M) {
	fmt.Println("Launching browser to clear Cloudflare challenge...")
	if err := clients.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "browser init failed: %v\n", err)
		os.Exit(1)
	}
	os.Exit(m.Run())
}

// departDate returns a date one month from today, formatted yyyy-MM-dd.
func departDate() string {
	return time.Now().AddDate(0, 1, 0).Format("2006-01-02")
}

// returnDate returns a date one month + N days from today.
func returnDate(extraDays int) string {
	return time.Now().AddDate(0, 1, extraDays).Format("2006-01-02")
}

// TestOneWayToCity verifies a one-way search to a specific city returns flights.
func TestOneWayToCity(t *testing.T) {
	req, err := models.NewRequest("NYC", "LAX", departDate(), "", 1)
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}

	summary, err := search.GetFlightSummaryToCity(req)
	if err != nil {
		t.Fatalf("GetFlightSummaryToCity: %v", err)
	}
	if len(summary.Leaving) == 0 {
		t.Error("expected at least one outbound flight")
	}

	t.Logf("%s — %d outbound flights, min $%d", summary.FullName, len(summary.Leaving), summary.MinLeavingPrice)
	for _, f := range summary.Leaving {
		hidden := ""
		if f.IsHiddenCity {
			hidden = fmt.Sprintf(" [HIDDEN-CITY to %s]", f.HiddenDestination)
		}
		t.Logf("  %s → %s  $%d  %-22s  %s → %s%s",
			f.Departure.Airport, f.Arrival.Airport, f.Price, f.Airline,
			f.Departure.Time.Format("3:04 PM"), f.Arrival.Time.Format("3:04 PM"),
			hidden,
		)
	}
}

// TestRoundTripToCity verifies a round-trip search returns outbound flights and
// logs any return flights found. The API omits the return date parameter to
// force one-way pricing, so Inbound results are not guaranteed for city pairs.
func TestRoundTripToCity(t *testing.T) {
	req, err := models.NewRequest("NYC", "LAX", departDate(), returnDate(7), 1)
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}

	summary, err := search.GetFlightSummaryToCity(req)
	if err != nil {
		t.Fatalf("GetFlightSummaryToCity: %v", err)
	}

	t.Logf("%s — %d outbound, %d return, min round-trip $%d",
		summary.FullName, len(summary.Leaving), len(summary.Returning), summary.MinRoundTripPrice)

	if len(summary.Leaving) == 0 {
		t.Error("expected at least one outbound flight")
	}
}

// TestWorldwideSearch verifies that a worldwide search returns destination cities.
func TestWorldwideSearch(t *testing.T) {
	req, err := models.NewRequest("NYC", "", departDate(), returnDate(7), 1)
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}

	summaries, err := search.GetCitySummaryLeavingCity(req)
	if err != nil {
		t.Fatalf("GetCitySummaryLeavingCity: %v", err)
	}
	if len(summaries) == 0 {
		t.Error("expected at least one destination")
	}

	t.Logf("found %d destinations (showing first 5):", len(summaries))
	for i, s := range summaries {
		if i >= 5 {
			break
		}
		t.Logf("  %-30s $%d", s.FullName, s.MinRoundTripPrice)
	}
}

// TestPriceFilter verifies that the max-price filter excludes results above the limit.
func TestPriceFilter(t *testing.T) {
	const maxPrice = 100
	req, err := models.NewRequest("NYC", "LAX", departDate(), "", 1)
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}
	req.WithMaxPrice(maxPrice)

	summary, err := search.GetFlightSummaryToCity(req)
	if err != nil {
		t.Fatalf("GetFlightSummaryToCity: %v", err)
	}

	for _, f := range summary.Leaving {
		if f.Price > maxPrice {
			t.Errorf("flight price $%d exceeds max-price $%d", f.Price, maxPrice)
		}
	}
	t.Logf("price filter $%d: %d flights passed", maxPrice, len(summary.Leaving))
}
