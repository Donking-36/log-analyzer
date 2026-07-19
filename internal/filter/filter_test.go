package filter

import (
	"testing"

	"github.com/Donking-36/log-analyzer/internal/parser"
)

// TestMatchLevelSameLevel verifies an exact severity match.
func TestMatchLevelSameLevel(t *testing.T) {
	entry := parser.LogEntry{Level: "ERROR"}

	if !MatchLevel(entry, "ERROR") {
		t.Fatal("expected level to match")
	}
}

// TestMatchLevelIgnoreCase preserves the command's case-insensitive filtering contract.
func TestMatchLevelIgnoreCase(t *testing.T) {
	entry := parser.LogEntry{Level: "ERROR"}

	if !MatchLevel(entry, "error") {
		t.Fatal("expected level to match ignoring case")
	}
}

// TestMatchLevelEmptyLevel verifies that an omitted filter accepts every entry.
func TestMatchLevelEmptyLevel(t *testing.T) {
	entry := parser.LogEntry{Level: "INFO"}

	if !MatchLevel(entry, "") {
		t.Fatal("expected empty filter level to match all")
	}
}

// TestMatchLevelDifferentLevel rejects entries with another severity.
func TestMatchLevelDifferentLevel(t *testing.T) {
	entry := parser.LogEntry{Level: "INFO"}

	if MatchLevel(entry, "ERROR") {
		t.Fatal("expected different level not to match")
	}
}
