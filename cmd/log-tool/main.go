package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/Donking-36/log-analyzer/internal/filter"
	"github.com/Donking-36/log-analyzer/internal/logfile"
	"github.com/Donking-36/log-analyzer/internal/parser"
)

func main() {
	if err := run(); err != nil {
		fmt.Println("错误:", err)
		os.Exit(1)
	}
}

func run() error {
	filePath := flag.String("file", "", "日志文件路径")
	level := flag.String("level", "", "日志级别，例如 INFO、WARN、ERROR")

	flag.Parse()

	if *filePath == "" {
		return fmt.Errorf("请提供日志文件路径，例如 --file ./testdata/sample.log")
	}

	lines, err := logfile.ReadLines(*filePath)
	if err != nil {
		return fmt.Errorf("读取文件失败: %w", err)
	}

	for _, line := range lines {
		entry, err := parser.ParseLine(line)
		if err != nil {
			fmt.Println("跳过格式错误的日志:", line)
			continue
		}

		if filter.MatchLevel(entry, *level) {
			fmt.Println(entry.Raw)
		}
	}

	return nil
}
