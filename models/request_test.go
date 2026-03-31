package models

import (
	"testing"
	"time"
)

// parseDate is a test helper that parses a yyyy-MM-dd string in local time.
func parseDate(t *testing.T, s string) time.Time {
	t.Helper()
	tm, err := time.ParseInLocation("2006-01-02", s, time.Local)
	if err != nil {
		t.Fatalf("parseDate(%q): %v", s, err)
	}
	return tm
}

// ---- NewRequest ----

func TestNewRequest_ValidOneWay(t *testing.T) {
	req, err := NewRequest("NYC", "LAX", "2026-05-12", "", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.HomeCity != "NYC" {
		t.Errorf("HomeCity = %q, want %q", req.HomeCity, "NYC")
	}
	if req.TripCity != "LAX" {
		t.Errorf("TripCity = %q, want %q", req.TripCity, "LAX")
	}
	if req.Travelers != 1 {
		t.Errorf("Travelers = %d, want 1", req.Travelers)
	}
	want := parseDate(t, "2026-05-12")
	if !req.LeavingDay.Equal(want) {
		t.Errorf("LeavingDay = %v, want %v", req.LeavingDay, want)
	}
	if !req.ReturningDay.IsZero() {
		t.Errorf("ReturningDay = %v, want zero", req.ReturningDay)
	}
}

func TestNewRequest_ValidRoundTrip(t *testing.T) {
	req, err := NewRequest("JFK", "LHR", "2026-06-01", "2026-06-15", 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	wantLeave := parseDate(t, "2026-06-01")
	wantReturn := parseDate(t, "2026-06-15")
	if !req.LeavingDay.Equal(wantLeave) {
		t.Errorf("LeavingDay = %v, want %v", req.LeavingDay, wantLeave)
	}
	if !req.ReturningDay.Equal(wantReturn) {
		t.Errorf("ReturningDay = %v, want %v", req.ReturningDay, wantReturn)
	}
	if req.Travelers != 2 {
		t.Errorf("Travelers = %d, want 2", req.Travelers)
	}
}

func TestNewRequest_InvalidDepartDate(t *testing.T) {
	_, err := NewRequest("NYC", "LAX", "not-a-date", "", 1)
	if err == nil {
		t.Fatal("expected error for invalid depart date, got nil")
	}
}

func TestNewRequest_InvalidReturnDate(t *testing.T) {
	_, err := NewRequest("NYC", "LAX", "2026-05-12", "bad-date", 1)
	if err == nil {
		t.Fatal("expected error for invalid return date, got nil")
	}
}

func TestNewRequest_EmptyToCity(t *testing.T) {
	req, err := NewRequest("NYC", "", "2026-05-12", "", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.TripCity != "" {
		t.Errorf("TripCity = %q, want empty", req.TripCity)
	}
}

func TestNewRequest_CriteriaInitializedEmpty(t *testing.T) {
	req, err := NewRequest("NYC", "LAX", "2026-05-12", "", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.Criteria.MaxPrice != 0 {
		t.Errorf("default MaxPrice = %d, want 0", req.Criteria.MaxPrice)
	}
	if req.Criteria.Leave != nil {
		t.Errorf("default Leave should be nil")
	}
	if req.Criteria.Return != nil {
		t.Errorf("default Return should be nil")
	}
	if len(req.Criteria.ExcludeAirports) != 0 {
		t.Errorf("default ExcludeAirports should be empty")
	}
}

// ---- WithMaxPrice ----

func TestWithMaxPrice(t *testing.T) {
	req, _ := NewRequest("NYC", "LAX", "2026-05-12", "", 1)
	result := req.WithMaxPrice(500)
	if result != req {
		t.Error("WithMaxPrice should return the same *Request (chaining)")
	}
	if req.Criteria.MaxPrice != 500 {
		t.Errorf("MaxPrice = %d, want 500", req.Criteria.MaxPrice)
	}
}

func TestWithMaxPrice_Zero(t *testing.T) {
	req, _ := NewRequest("NYC", "LAX", "2026-05-12", "", 1)
	req.WithMaxPrice(0)
	if req.Criteria.MaxPrice != 0 {
		t.Errorf("MaxPrice = %d, want 0", req.Criteria.MaxPrice)
	}
}

// ---- WithLeavingCriteria ----

func TestWithLeavingCriteria_BothSet(t *testing.T) {
	req, _ := NewRequest("NYC", "LAX", "2026-05-12", "", 1)
	result := req.WithLeavingCriteria(8, 20)
	if result != req {
		t.Error("WithLeavingCriteria should return the same *Request (chaining)")
	}
	if req.Criteria.Leave == nil {
		t.Fatal("Leave criteria should not be nil")
	}
	wantAfter := parseDate(t, "2026-05-12").Add(8 * time.Hour)
	wantBefore := parseDate(t, "2026-05-12").Add(20 * time.Hour)
	if !req.Criteria.Leave.After.Equal(wantAfter) {
		t.Errorf("Leave.After = %v, want %v", req.Criteria.Leave.After, wantAfter)
	}
	if !req.Criteria.Leave.Before.Equal(wantBefore) {
		t.Errorf("Leave.Before = %v, want %v", req.Criteria.Leave.Before, wantBefore)
	}
}

func TestWithLeavingCriteria_ZeroValues(t *testing.T) {
	req, _ := NewRequest("NYC", "LAX", "2026-05-12", "", 1)
	req.WithLeavingCriteria(0, 0)
	if req.Criteria.Leave == nil {
		t.Fatal("Leave criteria should not be nil even with zero args")
	}
	if !req.Criteria.Leave.After.IsZero() {
		t.Errorf("Leave.After should be zero when afterHour=0")
	}
	if !req.Criteria.Leave.Before.IsZero() {
		t.Errorf("Leave.Before should be zero when beforeHour=0")
	}
}

func TestWithLeavingCriteria_OnlyAfter(t *testing.T) {
	req, _ := NewRequest("NYC", "LAX", "2026-05-12", "", 1)
	req.WithLeavingCriteria(9, 0)
	if req.Criteria.Leave.After.IsZero() {
		t.Error("Leave.After should not be zero")
	}
	if !req.Criteria.Leave.Before.IsZero() {
		t.Error("Leave.Before should be zero when beforeHour=0")
	}
}

// ---- WithReturningCriteria ----

func TestWithReturningCriteria_BothSet(t *testing.T) {
	req, _ := NewRequest("NYC", "LAX", "2026-05-12", "2026-05-19", 1)
	result := req.WithReturningCriteria(10, 22)
	if result != req {
		t.Error("WithReturningCriteria should return the same *Request (chaining)")
	}
	if req.Criteria.Return == nil {
		t.Fatal("Return criteria should not be nil")
	}
	wantAfter := parseDate(t, "2026-05-19").Add(10 * time.Hour)
	wantBefore := parseDate(t, "2026-05-19").Add(22 * time.Hour)
	if !req.Criteria.Return.After.Equal(wantAfter) {
		t.Errorf("Return.After = %v, want %v", req.Criteria.Return.After, wantAfter)
	}
	if !req.Criteria.Return.Before.Equal(wantBefore) {
		t.Errorf("Return.Before = %v, want %v", req.Criteria.Return.Before, wantBefore)
	}
}

func TestWithReturningCriteria_ZeroValues(t *testing.T) {
	req, _ := NewRequest("NYC", "LAX", "2026-05-12", "2026-05-19", 1)
	req.WithReturningCriteria(0, 0)
	if req.Criteria.Return == nil {
		t.Fatal("Return criteria should not be nil even with zero args")
	}
	if !req.Criteria.Return.After.IsZero() {
		t.Error("Return.After should be zero when afterHour=0")
	}
	if !req.Criteria.Return.Before.IsZero() {
		t.Error("Return.Before should be zero when beforeHour=0")
	}
}

// ---- WithExcludeAirportsCriteria ----

func TestWithExcludeAirportsCriteria(t *testing.T) {
	req, _ := NewRequest("NYC", "LAX", "2026-05-12", "", 1)
	result := req.WithExcludeAirportsCriteria([]string{"JFK", "EWR"})
	if result != req {
		t.Error("WithExcludeAirportsCriteria should return the same *Request (chaining)")
	}
	if len(req.Criteria.ExcludeAirports) != 2 {
		t.Fatalf("ExcludeAirports length = %d, want 2", len(req.Criteria.ExcludeAirports))
	}
	if req.Criteria.ExcludeAirports[0] != "JFK" || req.Criteria.ExcludeAirports[1] != "EWR" {
		t.Errorf("ExcludeAirports = %v, want [JFK EWR]", req.Criteria.ExcludeAirports)
	}
}

func TestWithExcludeAirportsCriteria_SkipsEmpty(t *testing.T) {
	req, _ := NewRequest("NYC", "LAX", "2026-05-12", "", 1)
	req.WithExcludeAirportsCriteria([]string{"JFK", "", "EWR", ""})
	if len(req.Criteria.ExcludeAirports) != 2 {
		t.Errorf("ExcludeAirports length = %d, want 2 (empty strings should be skipped)", len(req.Criteria.ExcludeAirports))
	}
}

func TestWithExcludeAirportsCriteria_EmptySlice(t *testing.T) {
	req, _ := NewRequest("NYC", "LAX", "2026-05-12", "", 1)
	req.WithExcludeAirportsCriteria([]string{})
	if len(req.Criteria.ExcludeAirports) != 0 {
		t.Errorf("ExcludeAirports should be empty, got %v", req.Criteria.ExcludeAirports)
	}
}

func TestWithExcludeAirportsCriteria_Accumulates(t *testing.T) {
	req, _ := NewRequest("NYC", "LAX", "2026-05-12", "", 1)
	req.WithExcludeAirportsCriteria([]string{"JFK"})
	req.WithExcludeAirportsCriteria([]string{"EWR"})
	if len(req.Criteria.ExcludeAirports) != 2 {
		t.Errorf("ExcludeAirports length = %d, want 2 after two calls", len(req.Criteria.ExcludeAirports))
	}
}

// ---- Method chaining ----

func TestMethodChaining(t *testing.T) {
	req, err := NewRequest("NYC", "LAX", "2026-05-12", "2026-05-19", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// All methods should return the same pointer and be composable.
	result := req.
		WithMaxPrice(300).
		WithLeavingCriteria(8, 20).
		WithReturningCriteria(10, 22).
		WithExcludeAirportsCriteria([]string{"EWR"})

	if result != req {
		t.Error("chained calls should return the same *Request")
	}
	if req.Criteria.MaxPrice != 300 {
		t.Errorf("MaxPrice = %d, want 300", req.Criteria.MaxPrice)
	}
	if req.Criteria.Leave == nil {
		t.Error("Leave criteria should be set")
	}
	if req.Criteria.Return == nil {
		t.Error("Return criteria should be set")
	}
	if len(req.Criteria.ExcludeAirports) != 1 {
		t.Errorf("ExcludeAirports length = %d, want 1", len(req.Criteria.ExcludeAirports))
	}
}
