package logfile

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadLinesSuccess(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "sample.log")

	content := "line one\nline two\nline three\n"
	err := os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		t.Fatal(err)
	}

	lines, err := ReadLines(filePath)
	if err != nil {
		t.Fatal(err)
	}

	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(lines))
	}

	if lines[0] != "line one" {
		t.Fatalf("expected first line to be line one, got %s", lines[0])
	}
}

func TestReadLinesFileNotExist(t *testing.T) {
	_, err := ReadLines("not-exist.log")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
