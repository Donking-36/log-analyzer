// Package parser converts raw log lines into structured entries.
package parser

import (
	"fmt"
	"strings"
	"time"
)

const timestampLayout = "2006-01-02 15:04:05"

// LogEntry contains the parsed timestamp, level, and message while retaining the source record.
type LogEntry struct {
	// Timestamp is the timezone-free log date and time represented in UTC.
	Timestamp time.Time
	// Level is the record's severity token.
	Level string
	// Message contains the fields after the level, joined with single spaces.
	Message string
	// Raw preserves the original, unmodified log line.
	Raw string
}

// ParseLine expects a record in the form "<date> <time> <level> <message>".
// It parses the first two fields as Timestamp, uses the third as Level,
// joins the remaining fields as Message, and preserves the original input in Raw.
func ParseLine(line string) (LogEntry, error) {
	parts := strings.Fields(line)
	if len(parts) < 4 {
		return LogEntry{}, fmt.Errorf("日志格式错误: %s", line)
	}

	timestampText := parts[0] + " " + parts[1]
	timestamp, err := time.Parse(timestampLayout, timestampText)
	if err != nil {
		return LogEntry{}, fmt.Errorf("日志时间格式错误 %q: %w", timestampText, err)
	}

	entry := LogEntry{
		Timestamp: timestamp,
		Level:     parts[2],
		Message:   strings.Join(parts[3:], " "),
		Raw:       line,
	}

	return entry, nil
}
