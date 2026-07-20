package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestRunFiltersByLevelIgnoringCase verifies filtering and the separation of standard output from diagnostics.
func TestRunFiltersByLevelIgnoringCase(t *testing.T) {
	filePath := writeLogFile(t, strings.Join([]string{
		"2026-03-01 10:00:00 INFO service started",
		"2026-03-01 10:01:00 ERROR database connection failed",
		"2026-03-01 10:02:00 WARN cache miss rate high",
		"2026-03-01 10:03:00 ERROR payment service timeout",
	}, "\n")+"\n")

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	err := run([]string{"--file", filePath, "--level", "error"}, &stdout, &stderr)
	if err != nil {
		t.Fatal(err)
	}

	want := strings.Join([]string{
		"2026-03-01 10:01:00 ERROR database connection failed",
		"2026-03-01 10:03:00 ERROR payment service timeout",
	}, "\n") + "\n"

	if got := stdout.String(); got != want {
		t.Fatalf("expected stdout %q, got %q", want, got)
	}

	if got := stderr.String(); got != "" {
		t.Fatalf("expected empty stderr, got %q", got)
	}
}

// TestRunWithoutLevelOutputsAllLines verifies that omitting the optional filter preserves every valid record.
func TestRunWithoutLevelOutputsAllLines(t *testing.T) {
	content := strings.Join([]string{
		"2026-03-01 10:00:00 INFO service started",
		"2026-03-01 10:01:00 WARN cache miss rate high",
	}, "\n") + "\n"
	filePath := writeLogFile(t, content)

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	err := run([]string{"--file", filePath}, &stdout, &stderr)
	if err != nil {
		t.Fatal(err)
	}

	if got := stdout.String(); got != content {
		t.Fatalf("expected stdout %q, got %q", content, got)
	}

	if got := stderr.String(); got != "" {
		t.Fatalf("expected empty stderr, got %q", got)
	}
}

// TestRunRequiresFile defines the validation error for a missing required file argument.
func TestRunRequiresFile(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	err := run(nil, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected missing file argument to return an error")
	}

	if !strings.Contains(err.Error(), "请提供日志文件路径") {
		t.Fatalf("expected missing file error, got %q", err)
	}

	if stdout.Len() != 0 || stderr.Len() != 0 {
		t.Fatalf("expected no output, got stdout %q and stderr %q", stdout.String(), stderr.String())
	}
}

// TestRunReturnsFileReadError verifies that file-system failures are returned to the caller.
func TestRunReturnsFileReadError(t *testing.T) {
	missingPath := filepath.Join(t.TempDir(), "missing.log")

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	err := run([]string{"--file", missingPath}, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected missing file to return an error")
	}

	const wantErr = "文件路径无效，请检查路径后重试"
	if got := err.Error(); got != wantErr {
		t.Fatalf("expected error %q, got %q", wantErr, got)
	}
}

// TestRunSkipsMalformedLines verifies that a bad record is diagnosed without hiding later valid records.
func TestRunSkipsMalformedLines(t *testing.T) {
	filePath := writeLogFile(t, strings.Join([]string{
		"bad line",
		"2026-03-01 10:01:00 ERROR database connection failed",
	}, "\n")+"\n")

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	err := run([]string{"--file", filePath}, &stdout, &stderr)
	if err != nil {
		t.Fatal(err)
	}

	wantStdout := "2026-03-01 10:01:00 ERROR database connection failed\n"
	if got := stdout.String(); got != wantStdout {
		t.Fatalf("expected stdout %q, got %q", wantStdout, got)
	}

	wantStderr := "跳过格式错误的日志: bad line\n"
	if got := stderr.String(); got != wantStderr {
		t.Fatalf("expected stderr %q, got %q", wantStderr, got)
	}
}

// writeLogFile creates an isolated log fixture and returns its path to the command test.
func writeLogFile(t *testing.T, content string) string {
	t.Helper()

	filePath := filepath.Join(t.TempDir(), "sample.log")
	if err := os.WriteFile(filePath, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	return filePath
}
