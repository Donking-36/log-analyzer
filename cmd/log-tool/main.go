package main

import (
	"flag"
	"fmt"

	"github.com/Donking-36/log-analyzer/internal/filter"
	"github.com/Donking-36/log-analyzer/internal/logfile"
	"github.com/Donking-36/log-analyzer/internal/parser"
)

func main() {
	filePath := flag.String("file", "", "日志文件路径")
	level := flag.String("level", "", "日志级别，例如 INFO、WARN、ERROR")

	flag.Parse()

	fmt.Println("日志文件路径:", *filePath)
	fmt.Println("日志级别:", *level)

	lines, err := logfile.ReadLines(*filePath)
	if err != nil {
		fmt.Println("读取文件失败:", err)
		return
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
}
