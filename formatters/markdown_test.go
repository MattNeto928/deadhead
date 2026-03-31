package formatters

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/mattneto928/deadhead/models"
	"github.com/mattneto928/deadhead/search"
)

func makeTestSummary() *search.CitySummary {
	dept := time.Date(2026, 5, 12, 15, 30, 0, 0, time.UTC)
	arr := time.Date(2026, 5, 12, 18, 45, 0, 0, time.UTC)
	f := &models.Flight{
		Price:   279,
		Airline: "Norse Atlantic UK",
		Departure: struct {
			Time    time.Time `json:"time"`
			Airport string    `json:"airport"`
		}{Time: dept, Airport: "JFK"},
		Arrival: struct {
			Time    time.Time `json:"time"`
			Airport string    `json:"airport"`
		}{Time: arr, Airport: "LGW"},
	}
	return &search.CitySummary{
		Name:              "LGW",
		FullName:          "London, England",
		MinRoundTripPrice: 279,
		MinLeavingPrice:   279,
		Leaving:           []*models.Flight{f},
		Returning:         []*models.Flight{},
	}
}

func TestToMarkdown_Basic(t *testing.T) {
	summaries := []*search.CitySummary{makeTestSummary()}
	var buf bytes.Buffer
	if err := ToMarkdown(&buf, summaries); err != nil {
		t.Fatalf("ToMarkdown error: %v", err)
	}
	out := buf.String()

	checks := []string{
		"# London, England",
		"$279",
		"Norse Atlantic UK",
		"JFK",
		"LGW",
	}
	for _, want := range checks {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q\nfull output:\n%s", want, out)
		}
	}
}

func TestToMarkdown_Empty(t *testing.T) {
	var buf bytes.Buffer
	if err := ToMarkdown(&buf, []*search.CitySummary{}); err != nil {
		t.Fatalf("ToMarkdown with empty slice error: %v", err)
	}
	// Template should still produce valid (blank) output without errors.
}

func TestToMarkdown_MultipleDestinations(t *testing.T) {
	s1 := makeTestSummary()
	s2 := &search.CitySummary{
		Name:              "CDG",
		FullName:          "Paris, France",
		MinRoundTripPrice: 320,
		MinLeavingPrice:   320,
		Leaving:           []*models.Flight{},
		Returning:         []*models.Flight{},
	}
	var buf bytes.Buffer
	if err := ToMarkdown(&buf, []*search.CitySummary{s1, s2}); err != nil {
		t.Fatalf("ToMarkdown error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "London, England") {
		t.Errorf("output missing London")
	}
	if !strings.Contains(out, "Paris, France") {
		t.Errorf("output missing Paris")
	}
}

func TestToMarkdown_WithReturnFlights(t *testing.T) {
	s := makeTestSummary()
	ret := &models.Flight{
		Price:   199,
		Airline: "British Airways",
		Departure: struct {
			Time    time.Time `json:"time"`
			Airport string    `json:"airport"`
		}{Time: time.Date(2026, 5, 19, 10, 0, 0, 0, time.UTC), Airport: "LGW"},
		Arrival: struct {
			Time    time.Time `json:"time"`
			Airport string    `json:"airport"`
		}{Time: time.Date(2026, 5, 19, 13, 0, 0, 0, time.UTC), Airport: "JFK"},
	}
	s.Returning = append(s.Returning, ret)

	var buf bytes.Buffer
	if err := ToMarkdown(&buf, []*search.CitySummary{s}); err != nil {
		t.Fatalf("ToMarkdown error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "British Airways") {
		t.Errorf("output missing return airline\nfull output:\n%s", out)
	}
}

func TestToMarkdown_WriterError(t *testing.T) {
	w := &failWriter{}
	// The template executes against the writer; a write failure should surface as an error.
	err := ToMarkdown(w, []*search.CitySummary{makeTestSummary()})
	if err == nil {
		t.Error("expected error when writer fails")
	}
}
