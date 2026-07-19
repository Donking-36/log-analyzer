// Package parser converts raw log lines into structured entries.
package parser

import (
	"fmt"
	"strings"
)

// LogEntry contains the parsed level and message while retaining the source record.
type LogEntry struct {
	// Level is the record's severity token.
	Level string
	// Message contains the fields after the level, joined with single spaces.
	Message string
	// Raw preserves the original, unmodified log line.
	Raw string
}

// ParseLine expects at least four whitespace-separated fields.
// It uses the third field as Level, joins the remaining fields as Message,
// and retains the original input in Raw; the first two fields are not validated.
func ParseLine(line string) (LogEntry, error) {
	parts := strings.Fields(line)
	if len(parts) < 4 {
		return LogEntry{}, fmt.Errorf("日志格式错误: %s", line)
	}

	entry := LogEntry{
		Level:   parts[2],
		Message: strings.Join(parts[3:], " "),
		Raw:     line,
	}

	return entry, nil
}
