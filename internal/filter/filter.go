package filter

import (
	"strings"

	"github.com/Donking-36/log-analyzer/internal/parser"
)

func MatchLevel(entry parser.LogEntry, level string) bool {
	if level == "" {
		return true
	}

	return strings.EqualFold(entry.Level, level)
}
