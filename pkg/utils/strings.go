package utils

import (
	"fmt"
	"time"
)

func JoinStringsForInStatement(values []string) string {
	result := ""

	for i, value := range values {
		if i == 0 {
			result = fmt.Sprintf("'%v'", value)
			continue
		}

		result = fmt.Sprintf("%v,'%v'", result, value)
	}

	return result
}

func FormatToScyllaDateTime(dateTime time.Time) string {
	return dateTime.Format("2006-01-02 15:04:05.000")
}

func JoinDatesForInStatement(values []time.Time) string {
	result := ""

	for i, value := range values {
		formatted := FormatToScyllaDateTime(value)

		if i == 0 {
			result = fmt.Sprintf("'%v'", formatted)
			continue
		}

		result = fmt.Sprintf("%v,'%v'", result, formatted)
	}

	return result
}
