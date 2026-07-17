package filter

import (
	"testing"

	"github.com/Donking-36/log-analyzer/internal/parser"
)

func TestMatchLevelSameLevel(t *testing.T) {
	entry := parser.LogEntry{Level: "ERROR"}

	if !MatchLevel(entry, "ERROR") {
		t.Fatal("expected level to match")
	}
}

func TestMatchLevelIgnoreCase(t *testing.T) {
	entry := parser.LogEntry{Level: "ERROR"}

	if !MatchLevel(entry, "error") {
		t.Fatal("expected level to match ignoring case")
	}
}

func TestMatchLevelEmptyLevel(t *testing.T) {
	entry := parser.LogEntry{Level: "INFO"}

	if !MatchLevel(entry, "") {
		t.Fatal("expected empty filter level to match all")
	}
}

func TestMatchLevelDifferentLevel(t *testing.T) {
	entry := parser.LogEntry{Level: "INFO"}

	if MatchLevel(entry, "ERROR") {
		t.Fatal("expected different level not to match")
	}
}
