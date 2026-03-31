package formatters

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"

	"github.com/mattneto928/deadhead/models"
	"github.com/mattneto928/deadhead/search"
)

func TestToJSON_Basic(t *testing.T) {
	req := &models.Request{
		HomeCity:  "NYC",
		TripCity:  "LAX",
		Travelers: 1,
	}
	dept := time.Date(2026, 5, 12, 15, 0, 0, 0, time.UTC)
	arr := time.Date(2026, 5, 12, 18, 0, 0, 0, time.UTC)
	summaries := []*search.CitySummary{
		{
			Name:              "LAX",
			FullName:          "Los Angeles, CA",
			MinRoundTripPrice: 250,
			MinLeavingPrice:   250,
			Leaving: []*models.Flight{
				{
					Price:   250,
					Airline: "Delta",
					Departure: struct {
						Time    time.Time `json:"time"`
						Airport string    `json:"airport"`
					}{Time: dept, Airport: "JFK"},
					Arrival: struct {
						Time    time.Time `json:"time"`
						Airport string    `json:"airport"`
					}{Time: arr, Airport: "LAX"},
				},
			},
		},
	}

	var buf bytes.Buffer
	if err := ToJSON(&buf, req, summaries); err != nil {
		t.Fatalf("ToJSON error: %v", err)
	}

	// Validate it's well-formed JSON.
	var out struct {
		Request *models.Request           `json:"request"`
		Data    []*search.CitySummary `json:"data"`
	}
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("output is not valid JSON: %v\nraw: %s", err, buf.String())
	}

	if out.Request.HomeCity != "NYC" {
		t.Errorf("request.home_city = %q, want NYC", out.Request.HomeCity)
	}
	if len(out.Data) != 1 {
		t.Fatalf("data length = %d, want 1", len(out.Data))
	}
	if out.Data[0].FullName != "Los Angeles, CA" {
		t.Errorf("data[0].FullName = %q, want %q", out.Data[0].FullName, "Los Angeles, CA")
	}
}

func TestToJSON_EmptySummaries(t *testing.T) {
	req := &models.Request{HomeCity: "NYC"}
	var buf bytes.Buffer
	if err := ToJSON(&buf, req, []*search.CitySummary{}); err != nil {
		t.Fatalf("ToJSON error: %v", err)
	}
	var out struct {
		Data []interface{} `json:"data"`
	}
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	if len(out.Data) != 0 {
		t.Errorf("data length = %d, want 0", len(out.Data))
	}
}

func TestToJSON_NilRequest(t *testing.T) {
	var buf bytes.Buffer
	if err := ToJSON(&buf, nil, []*search.CitySummary{}); err != nil {
		t.Fatalf("ToJSON with nil request error: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("expected non-empty output")
	}
}

func TestToJSON_IndentedOutput(t *testing.T) {
	var buf bytes.Buffer
	_ = ToJSON(&buf, &models.Request{}, []*search.CitySummary{})
	output := buf.String()
	// Indented JSON uses tabs.
	if !bytes.Contains(buf.Bytes(), []byte("\t")) {
		t.Errorf("expected indented output with tabs, got: %s", output)
	}
}

func TestToJSON_WriterError(t *testing.T) {
	// Use a writer that always fails.
	w := &failWriter{}
	err := ToJSON(w, &models.Request{}, []*search.CitySummary{})
	if err == nil {
		t.Error("expected error when writer fails")
	}
}

type failWriter struct{}

func (f *failWriter) Write(p []byte) (int, error) {
	return 0, bytes.ErrTooLarge
}
