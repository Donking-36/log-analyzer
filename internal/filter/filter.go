// Package filter provides predicates for selecting parsed log entries.
package filter

import (
	"strings"

	"github.com/Donking-36/log-analyzer/internal/parser"
)

// MatchLevel reports whether entry matches level without regard to letter case.
// An empty level matches every entry.
func MatchLevel(entry parser.LogEntry, level string) bool {
	if level == "" {
		return true
	}

	return strings.EqualFold(entry.Level, level)
}
