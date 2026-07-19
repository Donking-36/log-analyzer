// Package main implements the log-tool command-line application.
package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/Donking-36/log-analyzer/internal/filter"
	"github.com/Donking-36/log-analyzer/internal/logfile"
	"github.com/Donking-36/log-analyzer/internal/parser"
)

// main connects process arguments and standard streams to the testable command runner.
func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, "错误:", err)
		os.Exit(1)
	}
}

// run executes one command invocation with injected output streams so its observable behavior can be tested without starting a subprocess.
func run(args []string, stdout, stderr io.Writer) error {
	flags := flag.NewFlagSet("log-tool", flag.ContinueOnError)
	flags.SetOutput(stderr)

	filePath := flags.String("file", "", "日志文件路径")
	level := flags.String("level", "", "日志级别，例如 INFO、WARN、ERROR")
	statMode := flags.Bool("stat", false, "启用日志统计模式")
	startDate := flags.String("start", "", "统计开始日期，格式为 YYYY-MM-DD")
	endDate := flags.String("end", "", "统计结束日期，格式为 YYYY-MM-DD")
	outputPath := flags.String("output", "report.csv", "CSV 报告输出路径")

	if err := flags.Parse(args); err != nil {
		return err
	}

	if *filePath == "" {
		return fmt.Errorf("请提供日志文件路径，例如 --file ./testdata/sample.log")
	}

	statisticsOptionSet := false
	flags.Visit(func(parsedFlag *flag.Flag) {
		switch parsedFlag.Name {
		case "start", "end", "output":
			statisticsOptionSet = true
		}
	})
	if !*statMode && statisticsOptionSet {
		return fmt.Errorf("--start、--end 和 --output 只能与 --stat 一起使用")
	}

	if *statMode {
		if *level != "" {
			return fmt.Errorf("--stat 不能与 --level 同时使用")
		}

		return runStatistics(statisticsOptions{
			filePath:   *filePath,
			startDate:  *startDate,
			endDate:    *endDate,
			outputPath: *outputPath,
		}, stdout, stderr)
	}

	// Process each record independently so one malformed line does not abort the command.
	scanErr := logfile.ScanLines(*filePath, func(line string) error {
		entry, err := parser.ParseLine(line)
		if err != nil {
			fmt.Fprintln(stderr, "跳过格式错误的日志:", line)
			return nil
		}

		if filter.MatchLevel(entry, *level) {
			fmt.Fprintln(stdout, entry.Raw)
		}

		return nil
	})
	if scanErr != nil {
		return fmt.Errorf("读取文件失败: %w", scanErr)
	}
	return nil
}
