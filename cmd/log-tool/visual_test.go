package main

import (
	"bytes"
	"context"
	"errors"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Donking-36/log-analyzer/internal/visualization"
)

func TestRunVisualizesCSVWithoutLogFileArgument(t *testing.T) {
	csvPath := writeVisualizationCSV(t, "level,count,percentage\nERROR,2,100.00\n")
	wantHTMLPath := visualization.HTMLPath(csvPath)

	var generatedCSVPath string
	var generatedHTMLPath string
	var openedPath string
	dependencies := commandDependencies{
		generateVisualization: func(_ context.Context, gotCSVPath, gotHTMLPath string) error {
			generatedCSVPath = gotCSVPath
			generatedHTMLPath = gotHTMLPath
			return os.WriteFile(gotHTMLPath, []byte("<html></html>"), 0o600)
		},
		openBrowser: func(filePath string) error {
			openedPath = filePath
			return nil
		},
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	err := runWithDependencies(
		[]string{"--visual", "--csv", csvPath},
		&stdout,
		&stderr,
		dependencies,
	)
	if err != nil {
		t.Fatal(err)
	}

	if generatedCSVPath != csvPath {
		t.Fatalf("expected generator CSV path %q, got %q", csvPath, generatedCSVPath)
	}
	if generatedHTMLPath != wantHTMLPath {
		t.Fatalf("expected generator HTML path %q, got %q", wantHTMLPath, generatedHTMLPath)
	}
	if openedPath != wantHTMLPath {
		t.Fatalf("expected opened path %q, got %q", wantHTMLPath, openedPath)
	}
	wantStdout := "已生成可视化报告: " + wantHTMLPath + "\n"
	if got := stdout.String(); got != wantStdout {
		t.Fatalf("expected stdout %q, got %q", wantStdout, got)
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected empty stderr, got %q", stderr.String())
	}
}

func TestRunVisualizationValidatesArguments(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{name: "missing CSV", args: []string{"--visual"}, wantErr: "请提供 CSV 统计报表路径"},
		{name: "CSV without visual mode", args: []string{"--csv", "report.csv"}, wantErr: "--csv 只能与 --visual 一起使用"},
		{name: "file conflict", args: []string{"--visual", "--csv", "report.csv", "--file", "app.log"}, wantErr: "--visual 不能与 --file 同时使用"},
		{name: "level conflict", args: []string{"--visual", "--csv", "report.csv", "--level", "ERROR"}, wantErr: "--visual 不能与 --level 同时使用"},
		{name: "stat conflict", args: []string{"--visual", "--csv", "report.csv", "--stat"}, wantErr: "--visual 不能与 --stat 同时使用"},
		{name: "start conflict", args: []string{"--visual", "--csv", "report.csv", "--start", "2026-03-01"}, wantErr: "--visual 不能与 --start 同时使用"},
		{name: "end conflict", args: []string{"--visual", "--csv", "report.csv", "--end", "2026-03-02"}, wantErr: "--visual 不能与 --end 同时使用"},
		{name: "output conflict", args: []string{"--visual", "--csv", "report.csv", "--output", "other.csv"}, wantErr: "--visual 不能与 --output 同时使用"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout bytes.Buffer
			var stderr bytes.Buffer
			err := runWithDependencies(tt.args, &stdout, &stderr, commandDependencies{})
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("expected error containing %q, got %v", tt.wantErr, err)
			}
			if stdout.Len() != 0 || stderr.Len() != 0 {
				t.Fatalf("expected no output, got stdout %q and stderr %q", stdout.String(), stderr.String())
			}
		})
	}
}

func TestRunVisualizationMapsStableErrors(t *testing.T) {
	tests := []struct {
		name       string
		generation error
		want       error
	}{
		{name: "invalid report", generation: fmtWrap(visualization.ErrInvalidReport), want: visualization.ErrInvalidReport},
		{name: "missing dependency", generation: fmtWrap(visualization.ErrMissingPythonDependency), want: visualization.ErrMissingPythonDependency},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			csvPath := writeVisualizationCSV(t, "level,count,percentage\n")
			opened := false
			dependencies := commandDependencies{
				generateVisualization: func(_ context.Context, _, _ string) error {
					return tt.generation
				},
				openBrowser: func(string) error {
					opened = true
					return nil
				},
			}

			var stdout bytes.Buffer
			var stderr bytes.Buffer
			err := runWithDependencies([]string{"--visual", "--csv", csvPath}, &stdout, &stderr, dependencies)
			if !errors.Is(err, tt.want) || err.Error() != tt.want.Error() {
				t.Fatalf("expected stable error %q, got %v", tt.want, err)
			}
			if opened {
				t.Fatal("expected browser not to open after generation failure")
			}
		})
	}
}

func TestRunVisualizationPreservesHTMLWhenBrowserFails(t *testing.T) {
	csvPath := writeVisualizationCSV(t, "level,count,percentage\n")
	htmlPath := visualization.HTMLPath(csvPath)
	dependencies := commandDependencies{
		generateVisualization: func(_ context.Context, _, gotHTMLPath string) error {
			return os.WriteFile(gotHTMLPath, []byte("<html></html>"), 0o600)
		},
		openBrowser: func(string) error {
			return errors.New("browser unavailable")
		},
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	err := runWithDependencies([]string{"--visual", "--csv", csvPath}, &stdout, &stderr, dependencies)
	if err == nil || !strings.Contains(err.Error(), "请手动打开 "+htmlPath) {
		t.Fatalf("expected manual-open error, got %v", err)
	}
	if _, statErr := os.Stat(htmlPath); statErr != nil {
		t.Fatalf("expected generated HTML to remain, got %v", statErr)
	}
}

func TestBrowserCommand(t *testing.T) {
	tests := []struct {
		goos     string
		wantName string
		wantArgs []string
	}{
		{goos: "windows", wantName: "rundll32.exe", wantArgs: []string{"url.dll,FileProtocolHandler", "file:///report.html"}},
		{goos: "darwin", wantName: "open", wantArgs: []string{"file:///report.html"}},
		{goos: "linux", wantName: "xdg-open", wantArgs: []string{"file:///report.html"}},
	}

	for _, tt := range tests {
		t.Run(tt.goos, func(t *testing.T) {
			name, args := browserCommand(tt.goos, "file:///report.html")
			if name != tt.wantName || strings.Join(args, "|") != strings.Join(tt.wantArgs, "|") {
				t.Fatalf("expected %s %v, got %s %v", tt.wantName, tt.wantArgs, name, args)
			}
		})
	}
}

func TestOpenBrowserReturnsLauncherFailure(t *testing.T) {
	filePath := filepath.Join(t.TempDir(), "report.html")
	wantErr := errors.New("launcher failed")
	called := false

	err := openBrowserWithRunner(filePath, func(name string, args ...string) error {
		called = true
		if name == "" {
			t.Fatal("expected browser launcher name")
		}
		if len(args) == 0 || !strings.HasPrefix(args[len(args)-1], "file:") {
			t.Fatalf("expected file URL as final argument, got %v", args)
		}
		return wantErr
	})
	if !called {
		t.Fatal("expected browser launcher to run")
	}
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected launcher failure, got %v", err)
	}
}

func TestOpenBrowserRejectsMissingRunner(t *testing.T) {
	err := openBrowserWithRunner(filepath.Join(t.TempDir(), "report.html"), nil)
	if err == nil || !strings.Contains(err.Error(), "未配置浏览器命令执行器") {
		t.Fatalf("expected missing browser command runner error, got %v", err)
	}
}

func TestPathToFileURLPreservesSpecialCharacters(t *testing.T) {
	filePath := filepath.Join(t.TempDir(), "统计 report #1.html")
	target, err := pathToFileURL(filePath)
	if err != nil {
		t.Fatal(err)
	}

	parsed, err := url.Parse(target)
	if err != nil {
		t.Fatal(err)
	}
	if parsed.Scheme != "file" {
		t.Fatalf("expected file URL, got %q", target)
	}
	if !strings.Contains(parsed.Path, "统计 report #1.html") {
		t.Fatalf("expected decoded path in URL, got %q", parsed.Path)
	}
	if strings.Contains(target, " ") || strings.Contains(target, "#1") {
		t.Fatalf("expected escaped URL, got %q", target)
	}
}

func writeVisualizationCSV(t *testing.T, content string) string {
	t.Helper()

	filePath := filepath.Join(t.TempDir(), "report.csv")
	if err := os.WriteFile(filePath, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return filePath
}

func fmtWrap(target error) error {
	return &wrappedTestError{target: target}
}

type wrappedTestError struct {
	target error
}

func (err *wrappedTestError) Error() string { return "wrapped: " + err.target.Error() }
func (err *wrappedTestError) Unwrap() error { return err.target }

func TestProductionVisualizationDependencyValidatesBeforePython(t *testing.T) {
	dependencies := productionCommandDependencies()
	missingPath := filepath.Join(t.TempDir(), "missing.csv")
	err := dependencies.generateVisualization(
		context.Background(),
		missingPath,
		visualization.HTMLPath(missingPath),
	)
	if !errors.Is(err, visualization.ErrInvalidReport) {
		t.Fatalf("expected ErrInvalidReport, got %v", err)
	}
}

func TestRunVisualizationReturnsGenerationError(t *testing.T) {
	csvPath := writeVisualizationCSV(t, "level,count,percentage\n")
	wantErr := errors.New("renderer failed")
	dependencies := commandDependencies{
		generateVisualization: func(context.Context, string, string) error {
			return wantErr
		},
	}

	var stdout bytes.Buffer
	err := runWithDependencies(
		[]string{"--visual", "--csv", csvPath},
		&stdout,
		&bytes.Buffer{},
		dependencies,
	)
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected renderer error, got %v", err)
	}
	if stdout.Len() != 0 {
		t.Fatalf("expected no success output, got %q", stdout.String())
	}
}

func TestRunVisualizationRejectsMissingGenerator(t *testing.T) {
	csvPath := writeVisualizationCSV(t, "level,count,percentage\n")
	var stdout bytes.Buffer
	err := runWithDependencies(
		[]string{"--visual", "--csv", csvPath},
		&stdout,
		&bytes.Buffer{},
		commandDependencies{},
	)
	if err == nil || !strings.Contains(err.Error(), "未配置生成器") {
		t.Fatalf("expected missing generator error, got %v", err)
	}
}

func TestRunVisualizationRejectsMissingBrowserOpenerAfterGeneration(t *testing.T) {
	csvPath := writeVisualizationCSV(t, "level,count,percentage\n")
	htmlPath := visualization.HTMLPath(csvPath)
	dependencies := commandDependencies{
		generateVisualization: func(_ context.Context, _, gotHTMLPath string) error {
			return os.WriteFile(gotHTMLPath, []byte("<html></html>"), 0o600)
		},
	}

	var stdout bytes.Buffer
	err := runWithDependencies(
		[]string{"--visual", "--csv", csvPath},
		&stdout,
		&bytes.Buffer{},
		dependencies,
	)
	if err == nil || !strings.Contains(err.Error(), "未配置浏览器打开器") {
		t.Fatalf("expected missing browser opener error, got %v", err)
	}
	if _, statErr := os.Stat(htmlPath); statErr != nil {
		t.Fatalf("expected generated HTML to remain, got %v", statErr)
	}
}
