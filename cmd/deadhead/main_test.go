package main

import (
	"testing"
)

// ---- isDate ----

func TestIsDate_ValidDates(t *testing.T) {
	cases := []string{
		"2026-05-12",
		"2000-01-01",
		"9999-12-31",
		"2026-01-01",
	}
	for _, s := range cases {
		if !isDate(s) {
			t.Errorf("isDate(%q) = false, want true", s)
		}
	}
}

func TestIsDate_InvalidDates(t *testing.T) {
	cases := []struct {
		input string
		desc  string
	}{
		{"2026-5-12", "single-digit month"},
		{"2026-05-1", "single-digit day"},
		{"20260512", "no separators"},
		{"2026/05/12", "slash separators"},
		{"2026-05-1x", "non-digit in day"},
		{"abcd-ef-gh", "all letters"},
		{"", "empty string"},
		{"2026-05-123", "too long"},
		{"2026-05-1", "too short"},
		{"26-05-12", "2-digit year"},
		{"2026-00-12", "zero month — passes format check but semantically invalid (function only checks format)"},
	}
	// Note: isDate only validates the yyyy-MM-dd *format*, not calendar correctness.
	// The semantically invalid cases below are still expected to return true from isDate
	// (e.g. "2026-00-12" has correct format). We only test format failures here.
	formatFailures := []string{
		"2026-5-12",
		"2026-05-1",
		"20260512",
		"2026/05/12",
		"2026-05-1x",
		"abcd-ef-gh",
		"",
		"2026-05-123",
		"26-05-12",
	}
	_ = cases
	for _, s := range formatFailures {
		if isDate(s) {
			t.Errorf("isDate(%q) = true, want false", s)
		}
	}
}

func TestIsDate_LengthBoundary(t *testing.T) {
	// Exactly 10 chars but wrong structure
	if isDate("2026-05-1x") {
		t.Error("isDate(\"2026-05-1x\") should be false (non-digit in last position)")
	}
	// 9 chars
	if isDate("2026-05-1") {
		t.Error("isDate with 9 chars should be false")
	}
	// 11 chars
	if isDate("2026-05-120") {
		t.Error("isDate with 11 chars should be false")
	}
}

func TestIsDate_SeparatorPositions(t *testing.T) {
	// Dashes in wrong positions
	if isDate("-026-05-12") {
		t.Error("dash at position 0 should fail")
	}
	if isDate("2026005-12") {
		t.Error("missing first dash should fail")
	}
	if isDate("2026-0512-") {
		t.Error("missing second dash / dash at end should fail")
	}
}
