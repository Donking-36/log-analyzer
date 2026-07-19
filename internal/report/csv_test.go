package report

import (
	"bytes"
	"errors"
	"testing"

	"github.com/Donking-36/log-analyzer/internal/stats"
)

// TestWriteCSVWritesSortedRowsAndPercentages verifies deterministic CSV output.
func TestWriteCSVWritesSortedRowsAndPercentages(t *testing.T) {
	summary := stats.Summary{
		Counts: map[string]int{
			"WARN":  1,
			"ERROR": 2,
		},
		Total: 3,
	}

	var output bytes.Buffer

	if err := WriteCSV(&output, summary); err != nil {
		t.Fatal(err)
	}

	want := "" +
		"level,count,percentage\n" +
		"ERROR,2,66.67\n" +
		"WARN,1,33.33\n"

	if got := output.String(); got != want {
		t.Fatalf("expected CSV %q, got %q", want, got)
	}
}

// TestWriteCSVWritesOnlyHeaderForEmptySummary verifies the empty-report contract.
func TestWriteCSVWritesOnlyHeaderForEmptySummary(t *testing.T) {
	var output bytes.Buffer

	if err := WriteCSV(&output, stats.Summary{}); err != nil {
		t.Fatal(err)
	}

	want := "level,count,percentage\n"
	if got := output.String(); got != want {
		t.Fatalf("expected CSV %q, got %q", want, got)
	}
}

// TestWriteCSVReturnsWriterError verifies output failures are propagated.
func TestWriteCSVReturnsWriterError(t *testing.T) {
	writeErr := errors.New("write failed")

	err := WriteCSV(failingWriter{err: writeErr}, stats.Summary{})
	if !errors.Is(err, writeErr) {
		t.Fatalf("expected writer error, got %v", err)
	}
}

type failingWriter struct {
	err error
}

func (w failingWriter) Write(_ []byte) (int, error) {
	return 0, w.err
}
