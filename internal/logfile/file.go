// Package logfile provides line-oriented access to local log files.
package logfile

import (
	"bufio"
	"os"
)


// ScanLines reads filePath sequentially and calls handle for each successfully scanned line.
// It uses bufio.Scanner's default maximum token size and returns the first open,
// scan, or handler error.
func ScanLines(filePath string, handle func(string) error) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if err := handle(scanner.Text()); err != nil {
			return err
		}
	}
	return scanner.Err()
}
