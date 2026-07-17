package parser

import "testing"

func TestParseLineSuccess(t *testing.T) {
	line := "2026-03-01 10:02:00 ERROR database connection failed"

	entry, err := ParseLine(line)
	if err != nil {
		t.Fatal(err)
	}

	if entry.Level != "ERROR" {
		t.Fatalf("expected level ERROR, got %s", entry.Level)
	}

	if entry.Message != "database connection failed" {
		t.Fatalf("expected message database connection failed, got %s", entry.Message)
	}

	if entry.Raw != line {
		t.Fatalf("expected raw line unchanged")
	}
}

func TestParseLineInvalidFormat(t *testing.T) {
	_, err := ParseLine("bad line")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
