package main

import (
	"errors"
	"os"
	"strings"
	"testing"
)

// TestLogFileReadError defines friendly and wrapped file-read error behavior.
func TestLogFileReadError(t *testing.T) {
	t.Run("maps missing file to friendly error", func(t *testing.T) {
		inputErr := &os.PathError{
			Op:   "open",
			Path: "missing.log",
			Err:  os.ErrNotExist,
		}

		got := logFileReadError(inputErr)

		if !errors.Is(got, errInvalidLogFilePath) {
			t.Fatalf("expected invalid-path error, got %v", got)
		}
		if got.Error() != "文件路径无效，请检查路径后重试" {
			t.Fatalf("unexpected error message %q", got)
		}
	})

	t.Run("preserves other file errors", func(t *testing.T) {
		inputErr := &os.PathError{
			Op:   "open",
			Path: "protected.log",
			Err:  os.ErrPermission,
		}

		got := logFileReadError(inputErr)

		if !strings.Contains(got.Error(), "读取文件失败") {
			t.Fatalf("expected file-read context, got %q", got)
		}
		if !errors.Is(got, os.ErrPermission) {
			t.Fatalf("expected permission error to remain discoverable, got %v", got)
		}
	})
}
