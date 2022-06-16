package utils

import "fmt"

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
