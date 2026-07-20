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

// run executes one command invocation with production dependencies.
func run(args []string, stdout, stderr io.Writer) error {
	return runWithDependencies(args, stdout, stderr, productionCommandDependencies())
}

// runWithDependencies injects external process and browser boundaries for deterministic tests.
func runWithDependencies(args []string, stdout, stderr io.Writer, dependencies commandDependencies) error {
	flags := flag.NewFlagSet("log-tool", flag.ContinueOnError)
	flags.SetOutput(stderr)

	filePath := flags.String("file", "", "日志文件路径")
	level := flags.String("level", "", "日志级别，例如 INFO、WARN、ERROR")
	statMode := flags.Bool("stat", false, "启用日志统计模式")
	startDate := flags.String("start", "", "统计开始日期，格式为 YYYY-MM-DD")
	endDate := flags.String("end", "", "统计结束日期，格式为 YYYY-MM-DD")
	outputPath := flags.String("output", "report.csv", "CSV 报告输出路径")
	visualMode := flags.Bool("visual", false, "启用统计结果可视化模式")
	csvPath := flags.String("csv", "", "CSV 统计报表路径")

	if err := flags.Parse(args); err != nil {
		return err
	}

	setFlags := make(map[string]bool)
	flags.Visit(func(parsedFlag *flag.Flag) {
		setFlags[parsedFlag.Name] = true
	})

	if *visualMode {
		for _, conflictingFlag := range []string{"file", "level", "stat", "start", "end", "output"} {
			if setFlags[conflictingFlag] {
				return fmt.Errorf("--visual 不能与 --%s 同时使用", conflictingFlag)
			}
		}
		if *csvPath == "" {
			return fmt.Errorf("请提供 CSV 统计报表路径，例如 --csv ./report.csv")
		}

		return runVisualization(visualizationOptions{csvPath: *csvPath}, stdout, dependencies)
	}

	if setFlags["csv"] {
		return fmt.Errorf("--csv 只能与 --visual 一起使用")
	}
	if *filePath == "" {
		return fmt.Errorf("请提供日志文件路径，例如 --file ./testdata/sample.log")
	}

	statisticsOptionSet := setFlags["start"] || setFlags["end"] || setFlags["output"]
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
	warnings := &warningLog{}
	scanErr := logfile.ScanLines(*filePath, func(line string) error {
		entry, err := parser.ParseLine(line)
		if err != nil {
			reportMalformedLine(stderr, warnings, line)
			return nil
		}

		if filter.MatchLevel(entry, *level) {
			fmt.Fprintln(stdout, entry.Raw)
		}

		return nil
	})
	finalizeWarningLog(stderr, warnings)

	if scanErr != nil {
		return logFileReadError(scanErr)
	}
	return nil
}
