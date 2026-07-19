package logfile

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"testing"
)


// TestScanLinesProcessesLinesInOrder verifies ordered delivery, including a final line without a trailing newline.
func TestScanLinesProcessesLinesInOrder(t *testing.T) {
	filePath := filepath.Join(t.TempDir(), "sample.log")
	content := "line one\nline two\nline three"

	if err := os.WriteFile(filePath, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	var got []string

	err := ScanLines(filePath, func(line string) error {
		got = append(got, line)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	want := []string{"line one", "line two", "line three"}
	if !slices.Equal(got, want) {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

// TestScanLinesStopsWhenHandlerReturnsError verifies early termination and preserves the handler error.
func TestScanLinesStopsWhenHandlerReturnsError(t *testing.T) {
	filePath := filepath.Join(t.TempDir(), "sample.log")
	content := "line one\nline two\nline three\n"

	if err := os.WriteFile(filePath, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	stopErr := errors.New("stop processing")
	calls := 0

	err := ScanLines(filePath, func(line string) error {
		calls++
		if line == "line two" {
			return stopErr
		}
		return nil
	})

	if !errors.Is(err, stopErr) {
		t.Fatalf("expected stop error, got %v", err)
	}

	if calls != 2 {
		t.Fatalf("expected handler to be called twice, got %d", calls)
	}
}

// TestScanLinesFileNotExist verifies that a missing path returns an error.
func TestScanLinesFileNotExist(t *testing.T) {
	filePath := filepath.Join(t.TempDir(), "missing.log")

	err := ScanLines(filePath, func(string) error {
		return nil
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
