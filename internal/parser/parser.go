package parser

import (
	"fmt"
	"strings"
)

type LogEntry struct {
	Level   string
	Message string
	Raw     string
}

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
