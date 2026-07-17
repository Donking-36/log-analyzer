package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
)

func main() {
	filePath := flag.String("file", "", "日志文件路径")
	level := flag.String("level", "", "日志级别，例如 INFO、WARN、ERROR")

	flag.Parse()

	fmt.Println("日志文件路径:", *filePath)
	fmt.Println("日志级别:", *level)

	file, err := os.Open(*filePath)
	if err != nil {
		fmt.Println("打开文件失败:", err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		fmt.Println(line)
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("读取文件失败:", err)
	}
}