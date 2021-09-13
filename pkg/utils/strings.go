package utils

import (
	"encoding/csv"
	"strings"
)

func SplitSpaceDelimitedString(s string) []string {
	reader := csv.NewReader(strings.NewReader(s))
	reader.Comma = ' '
	reader.ReuseRecord = true
	reader.TrimLeadingSpace = true
	reader.LazyQuotes = true

	all, _ := reader.ReadAll()
	parts := []string{}

	for _, group := range all {
		parts = append(parts, group...)
	}

	return parts
}
