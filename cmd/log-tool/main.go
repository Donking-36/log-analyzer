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

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, "错误:", err)
		os.Exit(1)
	}
}

func run(args []string, stdout, stderr io.Writer) error {
	flags := flag.NewFlagSet("log-tool", flag.ContinueOnError)
	flags.SetOutput(stderr)

	filePath := flags.String("file", "", "日志文件路径")
	level := flags.String("level", "", "日志级别，例如 INFO、WARN、ERROR")

	if err := flags.Parse(args); err != nil {
		return err
	}

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
			fmt.Fprintln(stderr, "跳过格式错误的日志:", line)
			continue
		}

		if filter.MatchLevel(entry, *level) {
			fmt.Fprintln(stdout, entry.Raw)
		}
	}

	return nil
}
