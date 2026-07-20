package main

import (
	"fmt"
	"io"
	"os"
)

// warningLog lazily persists malformed-record diagnostics in one temporary file.
type warningLog struct {
	directory string
	file      *os.File
	path      string
	err       error
}

// Record appends one diagnostic, creating the temporary file on first use.
func (log *warningLog) Record(diagnostic string) {
	if log.err != nil {
		return
	}

	if log.file == nil {
		file, err := os.CreateTemp(log.directory, "log-tool-warnings-*.log")
		if err != nil {
			log.err = fmt.Errorf("创建告警日志失败: %w", err)
			return
		}

		log.file = file
		log.path = file.Name()
	}

	if _, err := fmt.Fprintln(log.file, diagnostic); err != nil {
		log.err = fmt.Errorf("写入告警日志失败: %w", err)
	}
}

// Close closes the temporary file and returns any recorded storage error.
func (log *warningLog) Close() error {
	if log.file != nil {
		if err := log.file.Close(); err != nil && log.err == nil {
			log.err = fmt.Errorf("关闭告警日志失败: %w", err)
		}
		log.file = nil
	}

	return log.err
}

// Path returns the temporary warning file path, or an empty string if unused.
func (log *warningLog) Path() string {
	return log.path
}

// reportMalformedLine writes a diagnostic to stderr and the temporary warning log.
func reportMalformedLine(stderr io.Writer, warnings *warningLog, line string) {
	diagnostic := fmt.Sprintf("跳过格式错误的日志: %s", line)
	fmt.Fprintln(stderr, diagnostic)
	warnings.Record(diagnostic)
}

// finalizeWarningLog closes the warning file and reports its location.
// Warning persistence is best-effort and must not interrupt normal log processing.
func finalizeWarningLog(stderr io.Writer, warnings *warningLog) {
	if err := warnings.Close(); err != nil {
		fmt.Fprintf(stderr, "无法保存告警日志: %v\n", err)
		return
	}

	if path := warnings.Path(); path != "" {
		fmt.Fprintf(stderr, "告警日志已记录到: %s\n", path)
	}
}
