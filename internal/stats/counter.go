// Package stats aggregates parsed log entries over a date range.
package stats

import (
	"fmt"
	"time"

	"github.com/Donking-36/log-analyzer/internal/parser"
)

// Counter counts log levels within an inclusive calendar-date range.
type Counter struct {
	start        time.Time
	endExclusive time.Time
	counts       map[string]int
	total        int
}

// Summary is a detached snapshot of the current aggregate values.
type Summary struct {
	Counts map[string]int
	Total  int
}

// NewCounter creates a counter whose start and end dates are both inclusive.
func NewCounter(start, end time.Time) (*Counter, error) {
	startDay := startOfDay(start)
	endDay := startOfDay(end)

	if endDay.Before(startDay) {
		return nil, fmt.Errorf("开始日期不能晚于结束日期")
	}

	return &Counter{
		start:        startDay,
		endExclusive: endDay.AddDate(0, 0, 1),
		counts:       make(map[string]int),
	}, nil
}

// Add includes entry when its timestamp falls inside the configured date range.
func (c *Counter) Add(entry parser.LogEntry) {
	if entry.Timestamp.Before(c.start) ||
		!entry.Timestamp.Before(c.endExclusive) {
		return
	}

	c.counts[entry.Level]++
	c.total++
}

// Summary returns a copy so callers cannot mutate the counter's internal map.
func (c *Counter) Summary() Summary {
	counts := make(map[string]int, len(c.counts))
	for level, count := range c.counts {
		counts[level] = count
	}

	return Summary{
		Counts: counts,
		Total:  c.total,
	}
}

func startOfDay(value time.Time) time.Time {
	year, month, day := value.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, value.Location())
}
