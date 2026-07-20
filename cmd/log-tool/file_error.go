package main

import (
	"errors"
	"fmt"
	"os"
)

// errInvalidLogFilePath is the stable user-facing error for a missing input file.
var errInvalidLogFilePath = errors.New("文件路径无效，请检查路径后重试")

// logFileReadError maps missing files to a friendly message while preserving other failures.
func logFileReadError(err error) error {
	if errors.Is(err, os.ErrNotExist) {
		return errInvalidLogFilePath
	}

	return fmt.Errorf("读取文件失败: %w", err)
}
