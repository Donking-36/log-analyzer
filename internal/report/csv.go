// Package report writes statistical reports.
package report

import (
	"encoding/csv"
	"fmt"
	"io"
	"sort"
	"strconv"

	"github.com/Donking-36/log-analyzer/internal/stats"
)

// WriteCSV writes a deterministic CSV representation of summary.
func WriteCSV(output io.Writer, summary stats.Summary) error {
	writer := csv.NewWriter(output)

	if err := writer.Write([]string{"level", "count", "percentage"}); err != nil {
		return fmt.Errorf("写入 CSV 表头失败: %w", err)
	}

	levels := make([]string, 0, len(summary.Counts))
	for level := range summary.Counts {
		levels = append(levels, level)
	}
	sort.Strings(levels)

	for _, level := range levels {
		count := summary.Counts[level]
		percentage := float64(count) / float64(summary.Total) * 100

		record := []string{
			level,
			strconv.Itoa(count),
			fmt.Sprintf("%.2f", percentage),
		}

		if err := writer.Write(record); err != nil {
			return fmt.Errorf("写入 CSV 数据失败: %w", err)
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return fmt.Errorf("刷新 CSV 输出失败: %w", err)
	}

	return nil
}
