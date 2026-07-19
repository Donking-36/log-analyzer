package visualization

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

type fakeExitError int

func (err fakeExitError) Error() string { return "process failed" }
func (err fakeExitError) ExitCode() int { return int(err) }

func TestGenerateRunsEmbeddedPythonScript(t *testing.T) {
	csvPath := writeCSVFixture(t, "level,count,percentage\nERROR,1,100.00\n")
	htmlPath := HTMLPath(csvPath)

	var gotName string
	var gotArgs []string
	var scriptPath string
	runner := func(_ context.Context, name string, args ...string) ([]byte, error) {
		gotName = name
		gotArgs = append([]string(nil), args...)
		for _, arg := range args {
			if strings.HasSuffix(arg, ".py") {
				scriptPath = arg
				content, err := os.ReadFile(arg)
				if err != nil {
					t.Fatal(err)
				}
				if !strings.Contains(string(content), "def main") {
					t.Fatal("expected embedded visualization script")
				}
			}
		}
		return nil, nil
	}

	if err := Generate(context.Background(), csvPath, htmlPath, runner); err != nil {
		t.Fatal(err)
	}

	wantCandidate := pythonCandidates(runtime.GOOS)[0]
	if gotName != wantCandidate.name {
		t.Fatalf("expected executable %q, got %q", wantCandidate.name, gotName)
	}
	if scriptPath == "" {
		t.Fatal("expected a temporary Python script argument")
	}
	if _, err := os.Stat(scriptPath); !os.IsNotExist(err) {
		t.Fatalf("expected temporary script removal, got stat error %v", err)
	}
	if !containsAdjacentArgs(gotArgs, "--csv", csvPath) {
		t.Fatalf("expected CSV path in arguments, got %v", gotArgs)
	}
	if !containsAdjacentArgs(gotArgs, "--output", htmlPath) {
		t.Fatalf("expected HTML path in arguments, got %v", gotArgs)
	}
}

func TestGenerateFallsBackToNextPythonCandidate(t *testing.T) {
	csvPath := writeCSVFixture(t, "level,count,percentage\n")
	calls := 0
	runner := func(_ context.Context, _ string, _ ...string) ([]byte, error) {
		calls++
		if calls == 1 {
			return nil, os.ErrNotExist
		}
		return nil, nil
	}

	if err := Generate(context.Background(), csvPath, HTMLPath(csvPath), runner); err != nil {
		t.Fatal(err)
	}
	if calls != 2 {
		t.Fatalf("expected two candidate attempts, got %d", calls)
	}
}

func TestGenerateMapsPythonExitCodes(t *testing.T) {
	tests := []struct {
		name       string
		exitCode   int
		wantTarget error
		wantText   string
		wantCalls  int
	}{
		{name: "invalid CSV", exitCode: 3, wantTarget: ErrInvalidReport, wantCalls: 1},
		{name: "missing matplotlib", exitCode: 4, wantTarget: ErrMissingPythonDependency, wantCalls: len(pythonCandidates(runtime.GOOS))},
		{name: "HTML write failure", exitCode: 5, wantText: "写入 HTML 报告失败", wantCalls: 1},
		{name: "internal renderer failure", exitCode: 6, wantText: "生成可视化报告失败", wantCalls: 1},
		{name: "argument failure", exitCode: 2, wantText: "生成可视化报告失败", wantCalls: 1},
		{name: "launcher failure", exitCode: 9, wantTarget: ErrMissingPythonDependency, wantCalls: len(pythonCandidates(runtime.GOOS))},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			csvPath := writeCSVFixture(t, "level,count,percentage\nERROR,1,100.00\n")
			calls := 0
			runner := func(_ context.Context, _ string, _ ...string) ([]byte, error) {
				calls++
				return []byte("renderer detail"), fakeExitError(tt.exitCode)
			}

			err := Generate(context.Background(), csvPath, HTMLPath(csvPath), runner)
			if tt.wantTarget != nil && !errors.Is(err, tt.wantTarget) {
				t.Fatalf("expected error %v, got %v", tt.wantTarget, err)
			}
			if tt.wantText != "" && !strings.Contains(err.Error(), tt.wantText) {
				t.Fatalf("expected error containing %q, got %v", tt.wantText, err)
			}
			if calls != tt.wantCalls {
				t.Fatalf("expected %d calls, got %d", tt.wantCalls, calls)
			}
		})
	}
}

func TestGenerateRejectsInvalidInputBeforeStartingPython(t *testing.T) {
	missingPath := filepath.Join(t.TempDir(), "missing.csv")
	called := false
	runner := func(_ context.Context, _ string, _ ...string) ([]byte, error) {
		called = true
		return nil, nil
	}

	err := Generate(context.Background(), missingPath, HTMLPath(missingPath), runner)
	if !errors.Is(err, ErrInvalidReport) {
		t.Fatalf("expected ErrInvalidReport, got %v", err)
	}
	if called {
		t.Fatal("expected Python not to start for invalid input")
	}
}

func TestGenerateRejectsInputAsOutput(t *testing.T) {
	csvPath := writeCSVFixture(t, "level,count,percentage\n")
	runner := func(_ context.Context, _ string, _ ...string) ([]byte, error) {
		t.Fatal("expected Python not to start when paths identify the same file")
		return nil, nil
	}

	err := Generate(context.Background(), csvPath, csvPath, runner)
	if err == nil || !strings.Contains(err.Error(), "HTML 输出路径不能与 CSV 报表路径相同") {
		t.Fatalf("expected identical-path error, got %v", err)
	}
}

func TestGenerateRejectsNilRunner(t *testing.T) {
	csvPath := writeCSVFixture(t, "level,count,percentage\n")
	err := Generate(context.Background(), csvPath, HTMLPath(csvPath), nil)
	if err == nil || !strings.Contains(err.Error(), "未配置子进程执行器") {
		t.Fatalf("expected missing runner error, got %v", err)
	}
}

func TestGenerateReturnsContextError(t *testing.T) {
	csvPath := writeCSVFixture(t, "level,count,percentage\n")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	runner := func(_ context.Context, _ string, _ ...string) ([]byte, error) {
		return nil, context.Canceled
	}

	err := Generate(ctx, csvPath, HTMLPath(csvPath), runner)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context cancellation, got %v", err)
	}
}

func TestPythonCandidates(t *testing.T) {
	windows := pythonCandidates("windows")
	if len(windows) != 3 || windows[0].name != "py" || len(windows[0].prefixArgs) != 1 || windows[0].prefixArgs[0] != "-3" {
		t.Fatalf("unexpected Windows candidates: %+v", windows)
	}

	linux := pythonCandidates("linux")
	if len(linux) != 2 || linux[0].name != "python3" || linux[1].name != "python" {
		t.Fatalf("unexpected Linux candidates: %+v", linux)
	}
}

func TestCommandDetailsTruncatesLongOutput(t *testing.T) {
	details := commandDetails([]byte(strings.Repeat("x", maxCommandOutput+10)))
	if len(details) != maxCommandOutput+5 || !strings.HasSuffix(details, "...") {
		t.Fatalf("expected truncated command details, got length %d", len(details))
	}
}

func containsAdjacentArgs(args []string, name, value string) bool {
	for index := 0; index+1 < len(args); index++ {
		if args[index] == name && args[index+1] == value {
			return true
		}
	}
	return false
}
