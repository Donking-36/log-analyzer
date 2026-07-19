package visualization

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestValidateCSVAcceptsValidReports(t *testing.T) {
	tests := []struct {
		name    string
		content string
	}{
		{
			name: "report with data",
			content: "level,count,percentage\n" +
				"ERROR,2,66.67\n" +
				"INFO,1,33.33\n",
		},
		{
			name:    "header only",
			content: "level,count,percentage\n",
		},
		{
			name:    "UTF-8 BOM",
			content: "\ufefflevel,count,percentage\nWARN,0,0.00\n",
		},
		{
			name:    "numeric whitespace",
			content: "level,count,percentage\nINFO, 2 , 100.00 \n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath := writeCSVFixture(t, tt.content)
			if err := ValidateCSV(filePath); err != nil {
				t.Fatalf("expected valid report, got %v", err)
			}
		})
	}
}

func TestValidateCSVRejectsInvalidReports(t *testing.T) {
	tests := []struct {
		name    string
		content string
	}{
		{name: "empty file", content: ""},
		{name: "wrong header", content: "severity,total,ratio\n"},
		{name: "missing field", content: "level,count,percentage\nERROR,2\n"},
		{name: "extra field", content: "level,count,percentage\nERROR,2,100.00,extra\n"},
		{name: "empty level", content: "level,count,percentage\n,2,100.00\n"},
		{name: "non-integer count", content: "level,count,percentage\nERROR,two,100.00\n"},
		{name: "negative count", content: "level,count,percentage\nERROR,-1,100.00\n"},
		{name: "non-numeric percentage", content: "level,count,percentage\nERROR,1,all\n"},
		{name: "NaN percentage", content: "level,count,percentage\nERROR,1,NaN\n"},
		{name: "infinite percentage", content: "level,count,percentage\nERROR,1,+Inf\n"},
		{name: "negative percentage", content: "level,count,percentage\nERROR,1,-0.01\n"},
		{name: "percentage above 100", content: "level,count,percentage\nERROR,1,100.01\n"},
		{name: "malformed quoting", content: "level,count,percentage\n\"ERROR,1,100.00\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath := writeCSVFixture(t, tt.content)
			err := ValidateCSV(filePath)
			if !errors.Is(err, ErrInvalidReport) {
				t.Fatalf("expected ErrInvalidReport, got %v", err)
			}
		})
	}
}

func TestValidateCSVRejectsMissingFile(t *testing.T) {
	missingPath := filepath.Join(t.TempDir(), "missing.csv")
	if err := ValidateCSV(missingPath); !errors.Is(err, ErrInvalidReport) {
		t.Fatalf("expected ErrInvalidReport, got %v", err)
	}
}

func TestHTMLPath(t *testing.T) {
	directory := t.TempDir()
	tests := []struct {
		name     string
		input    string
		wantName string
	}{
		{name: "CSV extension", input: filepath.Join(directory, "report.csv"), wantName: "report.html"},
		{name: "uppercase CSV extension", input: filepath.Join(directory, "REPORT.CSV"), wantName: "REPORT.html"},
		{name: "no extension", input: filepath.Join(directory, "report"), wantName: "report.html"},
		{name: "other extension", input: filepath.Join(directory, "report.data"), wantName: "report.data.html"},
		{name: "hidden file", input: filepath.Join(directory, ".report"), wantName: ".report.html"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			want := filepath.Join(directory, tt.wantName)
			if got := HTMLPath(tt.input); got != want {
				t.Fatalf("expected %q, got %q", want, got)
			}
		})
	}
}

func writeCSVFixture(t *testing.T, content string) string {
	t.Helper()

	filePath := filepath.Join(t.TempDir(), "report.csv")
	if err := os.WriteFile(filePath, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return filePath
}
