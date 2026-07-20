package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/Donking-36/log-analyzer/internal/visualization"
)

const visualizationTimeout = 2 * time.Minute

type visualizationOptions struct {
	csvPath string
}

type commandDependencies struct {
	generateVisualization func(ctx context.Context, csvPath, htmlPath string) error
	openBrowser           func(filePath string) error
}

func productionCommandDependencies() commandDependencies {
	return commandDependencies{
		generateVisualization: func(ctx context.Context, csvPath, htmlPath string) error {
			return visualization.Generate(ctx, csvPath, htmlPath, visualization.ExecCommand)
		},
		openBrowser: openBrowser,
	}
}

// runVisualization generates an HTML report and then attempts to open it in the default browser.
func runVisualization(options visualizationOptions, stdout io.Writer, dependencies commandDependencies) error {
	absoluteCSVPath, err := filepath.Abs(options.csvPath)
	if err != nil {
		return visualization.ErrInvalidReport
	}
	htmlPath := visualization.HTMLPath(absoluteCSVPath)

	if dependencies.generateVisualization == nil {
		return fmt.Errorf("生成可视化报告失败: 未配置生成器")
	}

	ctx, cancel := context.WithTimeout(context.Background(), visualizationTimeout)
	defer cancel()
	if err := dependencies.generateVisualization(ctx, absoluteCSVPath, htmlPath); err != nil {
		if errors.Is(err, visualization.ErrInvalidReport) {
			return visualization.ErrInvalidReport
		}
		if errors.Is(err, visualization.ErrMissingPythonDependency) {
			return visualization.ErrMissingPythonDependency
		}
		return err
	}

	fmt.Fprintf(stdout, "已生成可视化报告: %s\n", htmlPath)
	if dependencies.openBrowser == nil {
		return fmt.Errorf("可视化报告已生成，但无法自动打开浏览器，请手动打开 %s: 未配置浏览器打开器", htmlPath)
	}
	if err := dependencies.openBrowser(htmlPath); err != nil {
		return fmt.Errorf("可视化报告已生成，但无法自动打开浏览器，请手动打开 %s: %w", htmlPath, err)
	}

	return nil
}

// openBrowser starts the operating system's default browser without invoking a command shell.
func openBrowser(filePath string) error {
	return openBrowserWithRunner(filePath, runBrowserCommand)
}

func openBrowserWithRunner(filePath string, run func(name string, args ...string) error) error {
	targetURL, err := pathToFileURL(filePath)
	if err != nil {
		return err
	}
	if run == nil {
		return fmt.Errorf("未配置浏览器命令执行器")
	}

	name, args := browserCommand(runtime.GOOS, targetURL)
	return run(name, args...)
}

func runBrowserCommand(name string, args ...string) error {
	return exec.Command(name, args...).Run()
}

func pathToFileURL(filePath string) (string, error) {
	absolutePath, err := filepath.Abs(filePath)
	if err != nil {
		return "", err
	}

	slashPath := filepath.ToSlash(absolutePath)
	if runtime.GOOS == "windows" && !strings.HasPrefix(slashPath, "/") {
		slashPath = "/" + slashPath
	}

	return (&url.URL{Scheme: "file", Path: slashPath}).String(), nil
}

func browserCommand(goos, targetURL string) (string, []string) {
	switch goos {
	case "windows":
		return "rundll32.exe", []string{"url.dll,FileProtocolHandler", targetURL}
	case "darwin":
		return "open", []string{targetURL}
	default:
		return "xdg-open", []string{targetURL}
	}
}
