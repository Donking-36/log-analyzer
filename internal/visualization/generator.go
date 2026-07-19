package visualization

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const maxCommandOutput = 2048

//go:embed visualize_report.py
var embeddedPythonScript []byte

// ErrMissingPythonDependency is returned when no usable Python 3 and matplotlib combination is available.
var ErrMissingPythonDependency = errors.New("请安装Python及matplotlib库后重试")

// CommandRunner executes one program without involving a command shell.
type CommandRunner func(ctx context.Context, name string, args ...string) ([]byte, error)

// ExecCommand runs a child process and returns its combined standard output and standard error.
func ExecCommand(ctx context.Context, name string, args ...string) ([]byte, error) {
	return exec.CommandContext(ctx, name, args...).CombinedOutput()
}

// Generate validates csvPath and runs the embedded Python renderer to create htmlPath.
func Generate(ctx context.Context, csvPath, htmlPath string, run CommandRunner) error {
	absoluteCSVPath, err := filepath.Abs(csvPath)
	if err != nil {
		return ErrInvalidReport
	}
	absoluteHTMLPath, err := filepath.Abs(htmlPath)
	if err != nil {
		return fmt.Errorf("解析 HTML 输出路径失败: %w", err)
	}

	if err := ValidateCSV(absoluteCSVPath); err != nil {
		return err
	}
	if pathsReferToSameFile(absoluteCSVPath, absoluteHTMLPath) {
		return fmt.Errorf("HTML 输出路径不能与 CSV 报表路径相同")
	}
	if run == nil {
		return fmt.Errorf("生成可视化报告失败: 未配置子进程执行器")
	}

	scriptFile, err := os.CreateTemp("", "log-tool-visualize-*.py")
	if err != nil {
		return fmt.Errorf("准备可视化脚本失败: %w", err)
	}
	scriptPath := scriptFile.Name()
	defer os.Remove(scriptPath)

	if _, err := scriptFile.Write(embeddedPythonScript); err != nil {
		_ = scriptFile.Close()
		return fmt.Errorf("准备可视化脚本失败: %w", err)
	}
	if err := scriptFile.Close(); err != nil {
		return fmt.Errorf("准备可视化脚本失败: %w", err)
	}

	var unexpectedExitSeen bool
	var unexpectedExitCode int
	var unexpectedOutput []byte

	for _, candidate := range pythonCandidates(runtime.GOOS) {
		args := append([]string{}, candidate.prefixArgs...)
		args = append(args,
			"-E",
			"-B",
			scriptPath,
			"--csv", absoluteCSVPath,
			"--output", absoluteHTMLPath,
		)

		output, runErr := run(ctx, candidate.name, args...)
		if runErr == nil {
			return nil
		}
		if ctxErr := ctx.Err(); ctxErr != nil {
			return fmt.Errorf("生成可视化报告超时或被取消: %w", ctxErr)
		}

		exitCode, ok := processExitCode(runErr)
		if !ok {
			continue
		}

		switch exitCode {
		case 3:
			return ErrInvalidReport
		case 4:
			continue
		case 5:
			return fmt.Errorf("写入 HTML 报告失败 %s%s", absoluteHTMLPath, commandDetails(output))
		case 2, 6:
			return fmt.Errorf("生成可视化报告失败%s", commandDetails(output))
		default:
			unexpectedExitSeen = true
			unexpectedExitCode = exitCode
			unexpectedOutput = append([]byte(nil), output...)
			continue
		}
	}

	if unexpectedExitSeen {
		return fmt.Errorf(
			"Python 可视化脚本异常退出（退出码 %d）%s",
			unexpectedExitCode,
			commandDetails(unexpectedOutput),
		)
	}
	return ErrMissingPythonDependency
}

type pythonCandidate struct {
	name       string
	prefixArgs []string
}

func pythonCandidates(goos string) []pythonCandidate {
	if goos == "windows" {
		return []pythonCandidate{
			{name: "py", prefixArgs: []string{"-3"}},
			{name: "python"},
			{name: "python3"},
		}
	}

	return []pythonCandidate{
		{name: "python3"},
		{name: "python"},
	}
}

func processExitCode(err error) (int, bool) {
	type exitCoder interface {
		ExitCode() int
	}

	var coded exitCoder
	if !errors.As(err, &coded) {
		return 0, false
	}
	return coded.ExitCode(), true
}

func commandDetails(output []byte) string {
	details := strings.TrimSpace(string(output))
	if details == "" {
		return ""
	}
	if len(details) > maxCommandOutput {
		details = details[:maxCommandOutput] + "..."
	}
	return ": " + details
}

func pathsReferToSameFile(firstPath, secondPath string) bool {
	firstInfo, err := os.Stat(firstPath)
	if err != nil {
		return false
	}
	secondInfo, err := os.Stat(secondPath)
	if err != nil {
		return false
	}
	return os.SameFile(firstInfo, secondInfo)
}
