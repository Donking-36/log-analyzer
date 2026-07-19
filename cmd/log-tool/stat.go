package main

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/Donking-36/log-analyzer/internal/logfile"
	"github.com/Donking-36/log-analyzer/internal/parser"
	"github.com/Donking-36/log-analyzer/internal/report"
	"github.com/Donking-36/log-analyzer/internal/stats"
)

const dateLayout = "2006-01-02"

type statisticsOptions struct {
	filePath   string
	startDate  string
	endDate    string
	outputPath string
}

// runStatistics scans the input once, aggregates matching entries, and writes the CSV report.
func runStatistics(options statisticsOptions, stdout, stderr io.Writer) error {
	start, end, err := parseDateRange(options.startDate, options.endDate)
	if err != nil {
		return err
	}

	counter, err := stats.NewCounter(start, end)
	if err != nil {
		return err
	}

	if pathsReferToSameFile(options.filePath, options.outputPath) {
		return fmt.Errorf("CSV 输出路径不能与日志文件路径相同")
	}

	scanErr := logfile.ScanLines(options.filePath, func(line string) error {
		entry, err := parser.ParseLine(line)
		if err != nil {
			fmt.Fprintln(stderr, "跳过格式错误的日志:", line)
			return nil
		}

		counter.Add(entry)
		return nil
	})
	if scanErr != nil {
		return fmt.Errorf("读取文件失败: %w", scanErr)
	}

	summary := counter.Summary()
	outputFile, err := os.Create(options.outputPath)
	if err != nil {
		return fmt.Errorf("创建报告文件失败: %w", err)
	}

	if err := report.WriteCSV(outputFile, summary); err != nil {
		_ = outputFile.Close()
		return fmt.Errorf("写入报告失败: %w", err)
	}
	if err := outputFile.Close(); err != nil {
		return fmt.Errorf("关闭报告文件失败: %w", err)
	}

	if summary.Total == 0 {
		fmt.Fprintf(stdout, "未找到符合日期范围的日志，已生成空报告: %s\n", options.outputPath)
	}

	return nil
}

// parseDateRange validates required YYYY-MM-DD values and returns an inclusive range.
func parseDateRange(startText, endText string) (time.Time, time.Time, error) {
	if startText == "" {
		return time.Time{}, time.Time{}, fmt.Errorf("请提供开始日期，例如 --start 2026-03-01")
	}
	if endText == "" {
		return time.Time{}, time.Time{}, fmt.Errorf("请提供结束日期，例如 --end 2026-03-02")
	}

	start, err := time.Parse(dateLayout, startText)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("开始日期格式错误，必须使用 YYYY-MM-DD: %w", err)
	}

	end, err := time.Parse(dateLayout, endText)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("结束日期格式错误，必须使用 YYYY-MM-DD: %w", err)
	}

	if start.After(end) {
		return time.Time{}, time.Time{}, fmt.Errorf("开始日期不能晚于结束日期")
	}

	return start, end, nil
}

// pathsReferToSameFile detects path aliases, symbolic links, and hard links.
func pathsReferToSameFile(inputPath, outputPath string) bool {
	inputInfo, err := os.Stat(inputPath)
	if err != nil {
		return false
	}

	outputInfo, err := os.Stat(outputPath)
	if err != nil {
		return false
	}

	return os.SameFile(inputInfo, outputInfo)
}
