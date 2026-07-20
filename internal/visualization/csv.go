// Package visualization validates statistical CSV reports before rendering.
package visualization

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// ErrInvalidReport is returned when a CSV report cannot be read or violates the UC-002 schema.
var ErrInvalidReport = errors.New("报表文件无效或损坏，请重新生成")

// ValidateCSV verifies that filePath contains the UC-002 header and valid data rows.
// A header-only report is valid because UC-002 uses that representation for an empty result.
func ValidateCSV(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return invalidReport(err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1

	header, err := reader.Read()
	if err != nil {
		return invalidReport(err)
	}
	if len(header) > 0 {
		header[0] = strings.TrimPrefix(header[0], "\ufeff")
	}
	if !isExpectedHeader(header) {
		return ErrInvalidReport
	}

	for {
		record, err := reader.Read()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return invalidReport(err)
		}
		if !isValidRecord(record) {
			return ErrInvalidReport
		}
	}
}

// HTMLPath replaces a .csv extension with .html and otherwise appends .html.
func HTMLPath(csvPath string) string {
	directory, fileName := filepath.Split(csvPath)
	extension := filepath.Ext(fileName)
	if strings.EqualFold(extension, ".csv") {
		fileName = strings.TrimSuffix(fileName, extension)
	}

	return filepath.Join(directory, fileName+".html")
}

func isExpectedHeader(header []string) bool {
	return len(header) == 3 &&
		header[0] == "level" &&
		header[1] == "count" &&
		header[2] == "percentage"
}

func isValidRecord(record []string) bool {
	if len(record) != 3 || strings.TrimSpace(record[0]) == "" {
		return false
	}

	count, err := strconv.Atoi(strings.TrimSpace(record[1]))
	if err != nil || count < 0 {
		return false
	}

	percentage, err := strconv.ParseFloat(strings.TrimSpace(record[2]), 64)
	if err != nil || math.IsNaN(percentage) || math.IsInf(percentage, 0) {
		return false
	}

	return percentage >= 0 && percentage <= 100
}

func invalidReport(cause error) error {
	return fmt.Errorf("%w: %v", ErrInvalidReport, cause)
}
