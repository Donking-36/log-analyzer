package main

import (
	"os"
	"path/filepath"
	"testing"
)

// TestWarningLogDoesNotCreateFileUntilFirstRecord verifies lazy file creation.
func TestWarningLogDoesNotCreateFileUntilFirstRecord(t *testing.T) {
	directory := t.TempDir()
	warnings := &warningLog{directory: directory}

	if err := warnings.Close(); err != nil {
		t.Fatal(err)
	}
	if path := warnings.Path(); path != "" {
		t.Fatalf("expected no warning path, got %q", path)
	}

	entries, err := os.ReadDir(directory)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Fatalf("expected no warning files, got %d", len(entries))
	}
}

// TestWarningLogRecordsMultipleDiagnostics verifies reuse of one temporary file.
func TestWarningLogRecordsMultipleDiagnostics(t *testing.T) {
	directory := t.TempDir()
	warnings := &warningLog{directory: directory}

	warnings.Record("first warning")
	warnings.Record("second warning")
	if err := warnings.Close(); err != nil {
		t.Fatal(err)
	}

	warningPath := warnings.Path()
	if filepath.Dir(warningPath) != directory {
		t.Fatalf("expected warning file in %q, got %q", directory, warningPath)
	}

	content, err := os.ReadFile(warningPath)
	if err != nil {
		t.Fatal(err)
	}
	const want = "first warning\nsecond warning\n"
	if got := string(content); got != want {
		t.Fatalf("expected warning content %q, got %q", want, got)
	}

	entries, err := os.ReadDir(directory)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected one warning file, got %d", len(entries))
	}
}
