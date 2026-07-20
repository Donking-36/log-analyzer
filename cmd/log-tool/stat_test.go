package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestRunStatisticsWritesDefaultCSVReport verifies the complete statistics command flow.
func TestRunStatisticsWritesDefaultCSVReport(t *testing.T) {
	workingDir := t.TempDir()
	t.Chdir(workingDir)

	filePath := writeLogFile(t, strings.Join([]string{
		"2026-02-28 23:59:59 INFO before range",
		"bad line",
		"2026-03-01 10:00:00 INFO service started",
		"2026-03-02 10:01:00 ERROR database failed",
		"2026-03-02 10:02:00 error payment failed",
		"2026-03-03 00:00:00 WARN after range",
	}, "\n")+"\n")

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	err := run([]string{
		"--file", filePath,
		"--stat",
		"--start", "2026-03-01",
		"--end", "2026-03-02",
	}, &stdout, &stderr)
	if err != nil {
		t.Fatal(err)
	}

	gotReport, err := os.ReadFile(filepath.Join(workingDir, "report.csv"))
	if err != nil {
		t.Fatal(err)
	}

	wantReport := strings.Join([]string{
		"level,count,percentage",
		"ERROR,2,66.67",
		"INFO,1,33.33",
	}, "\n") + "\n"
	if got := string(gotReport); got != wantReport {
		t.Fatalf("expected report %q, got %q", wantReport, got)
	}

	if got := stdout.String(); got != "" {
		t.Fatalf("expected empty stdout, got %q", got)
	}

	wantStderr := "跳过格式错误的日志: bad line\n"
	if got := stderr.String(); got != wantStderr {
		t.Fatalf("expected stderr %q, got %q", wantStderr, got)
	}
}

// TestRunStatisticsValidatesArguments defines the errors for invalid statistics options.
func TestRunStatisticsValidatesArguments(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{
			name:    "missing start date",
			args:    []string{"--file", "input.log", "--stat", "--end", "2026-03-02"},
			wantErr: "请提供开始日期",
		},
		{
			name:    "missing end date",
			args:    []string{"--file", "input.log", "--stat", "--start", "2026-03-01"},
			wantErr: "请提供结束日期",
		},
		{
			name:    "invalid start date",
			args:    []string{"--file", "input.log", "--stat", "--start", "2026/03/01", "--end", "2026-03-02"},
			wantErr: "开始日期格式错误",
		},
		{
			name:    "invalid end date",
			args:    []string{"--file", "input.log", "--stat", "--start", "2026-03-01", "--end", "2026/03/02"},
			wantErr: "结束日期格式错误",
		},
		{
			name:    "reversed range",
			args:    []string{"--file", "input.log", "--stat", "--start", "2026-03-03", "--end", "2026-03-02"},
			wantErr: "开始日期不能晚于结束日期",
		},
		{
			name: "level conflicts with statistics",
			args: []string{
				"--file", "input.log",
				"--stat",
				"--level", "ERROR",
				"--start", "2026-03-01",
				"--end", "2026-03-02",
			},
			wantErr: "--stat 不能与 --level 同时使用",
		},
		{
			name: "statistics options require statistics mode",
			args: []string{
				"--file", "input.log",
				"--output", "report.csv",
			},
			wantErr: "--start、--end 和 --output 只能与 --stat 一起使用",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout bytes.Buffer
			var stderr bytes.Buffer

			err := run(tt.args, &stdout, &stderr)
			if err == nil {
				t.Fatalf("expected error containing %q", tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("expected error containing %q, got %q", tt.wantErr, err)
			}
			if stdout.Len() != 0 || stderr.Len() != 0 {
				t.Fatalf("expected no output, got stdout %q and stderr %q", stdout.String(), stderr.String())
			}
		})
	}
}

// TestRunStatisticsWritesHeaderForEmptyRange verifies the empty-report convention.
func TestRunStatisticsWritesHeaderForEmptyRange(t *testing.T) {
	filePath := writeLogFile(t, "2026-03-03 10:00:00 INFO outside range\n")
	outputPath := filepath.Join(t.TempDir(), "empty.csv")

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	err := run([]string{
		"--file", filePath,
		"--stat",
		"--start", "2026-03-01",
		"--end", "2026-03-02",
		"--output", outputPath,
	}, &stdout, &stderr)
	if err != nil {
		t.Fatal(err)
	}

	gotReport, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(gotReport), "level,count,percentage\n"; got != want {
		t.Fatalf("expected report %q, got %q", want, got)
	}

	wantStdout := fmt.Sprintf("未找到符合日期范围的日志，已生成空报告: %s\n", outputPath)
	if got := stdout.String(); got != wantStdout {
		t.Fatalf("expected stdout %q, got %q", wantStdout, got)
	}
	if got := stderr.String(); got != "" {
		t.Fatalf("expected empty stderr, got %q", got)
	}
}

// TestRunStatisticsReturnsInputReadError verifies that no report is created when scanning fails.
func TestRunStatisticsReturnsInputReadError(t *testing.T) {
	outputPath := filepath.Join(t.TempDir(), "report.csv")
	missingPath := filepath.Join(t.TempDir(), "missing.log")

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	err := run([]string{
		"--file", missingPath,
		"--stat",
		"--start", "2026-03-01",
		"--end", "2026-03-02",
		"--output", outputPath,
	}, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected missing file to return an error")
	}

	const wantErr = "文件路径无效，请检查路径后重试"
	if got := err.Error(); got != wantErr {
		t.Fatalf("expected error %q, got %q", wantErr, got)
	}
	if _, statErr := os.Stat(outputPath); !os.IsNotExist(statErr) {
		t.Fatalf("expected no report file, got stat error %v", statErr)
	}
}

// TestRunStatisticsReturnsOutputCreateError verifies report-path failures are returned.
func TestRunStatisticsReturnsOutputCreateError(t *testing.T) {
	filePath := writeLogFile(t, "2026-03-01 10:00:00 INFO service started\n")
	outputPath := filepath.Join(t.TempDir(), "missing", "report.csv")

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	err := run([]string{
		"--file", filePath,
		"--stat",
		"--start", "2026-03-01",
		"--end", "2026-03-02",
		"--output", outputPath,
	}, &stdout, &stderr)
	if err == nil || !strings.Contains(err.Error(), "创建报告文件失败") {
		t.Fatalf("expected report creation error, got %v", err)
	}
}

// TestRunStatisticsRejectsInputAsOutput verifies that reporting cannot truncate the source log.
func TestRunStatisticsRejectsInputAsOutput(t *testing.T) {
	content := "2026-03-01 10:00:00 INFO service started\n"
	filePath := writeLogFile(t, content)

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	err := run([]string{
		"--file", filePath,
		"--stat",
		"--start", "2026-03-01",
		"--end", "2026-03-02",
		"--output", filePath,
	}, &stdout, &stderr)
	if err == nil || !strings.Contains(err.Error(), "CSV 输出路径不能与日志文件路径相同") {
		t.Fatalf("expected identical-path error, got %v", err)
	}

	gotContent, readErr := os.ReadFile(filePath)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if got := string(gotContent); got != content {
		t.Fatalf("expected source log to remain %q, got %q", content, got)
	}
}
