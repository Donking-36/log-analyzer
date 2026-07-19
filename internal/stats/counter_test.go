package stats

import (
	"maps"
	"testing"
	"time"

	"github.com/Donking-36/log-analyzer/internal/parser"
)

// TestCounterCountsEntriesInInclusiveDateRange verifies both date boundaries.
func TestCounterCountsEntriesInInclusiveDateRange(t *testing.T) {
	start := time.Date(2026, time.March, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, time.March, 2, 0, 0, 0, 0, time.UTC)

	counter, err := NewCounter(start, end)
	if err != nil {
		t.Fatal(err)
	}

	entries := []parser.LogEntry{
		{
			Timestamp: time.Date(2026, time.February, 28, 23, 59, 59, 0, time.UTC),
			Level:     "WARN",
		},
		{
			Timestamp: start,
			Level:     "INFO",
		},
		{
			Timestamp: time.Date(2026, time.March, 2, 23, 59, 59, 0, time.UTC),
			Level:     "ERROR",
		},
		{
			Timestamp: time.Date(2026, time.March, 3, 0, 0, 0, 0, time.UTC),
			Level:     "WARN",
		},
	}

	for _, entry := range entries {
		counter.Add(entry)
	}

	got := counter.Summary()
	wantCounts := map[string]int{
		"INFO":  1,
		"ERROR": 1,
	}

	if got.Total != 2 {
		t.Fatalf("expected total 2, got %d", got.Total)
	}

	if !maps.Equal(got.Counts, wantCounts) {
		t.Fatalf("expected counts %v, got %v", wantCounts, got.Counts)
	}
}

// TestNewCounterRejectsReversedDateRange verifies start cannot be after end.
func TestNewCounterRejectsReversedDateRange(t *testing.T) {
	start := time.Date(2026, time.March, 3, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, time.March, 2, 0, 0, 0, 0, time.UTC)

	counter, err := NewCounter(start, end)
	if err == nil {
		t.Fatal("expected reversed date range to return an error")
	}

	if counter != nil {
		t.Fatal("expected nil counter when date range is invalid")
	}
}

// TestCounterMergesLevelsIgnoringCase verifies canonical uppercase grouping.
func TestCounterMergesLevelsIgnoringCase(t *testing.T) {
	date := time.Date(2026, time.March, 1, 0, 0, 0, 0, time.UTC)

	counter, err := NewCounter(date, date)
	if err != nil {
		t.Fatal(err)
	}

	for _, level := range []string{"error", "ERROR", "Error"} {
		counter.Add(parser.LogEntry{
			Timestamp: date,
			Level:     level,
		})
	}

	got := counter.Summary()
	wantCounts := map[string]int{
		"ERROR": 3,
	}

	if got.Total != 3 {
		t.Fatalf("expected total 3, got %d", got.Total)
	}

	if !maps.Equal(got.Counts, wantCounts) {
		t.Fatalf("expected counts %v, got %v", wantCounts, got.Counts)
	}
}
