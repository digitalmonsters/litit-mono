package common

import (
	"regexp"
	"strings"
)

var stringRg = regexp.MustCompile(`\s+`)
var alpha = regexp.MustCompile("[^a-zA-Z0-9 ]+")

func ToTsQuery(query string) string {
	if len(query) == 0 {
		return query
	}

	query = strings.ToLower(query)

	sp := strings.Split(stringRg.ReplaceAllString(alpha.ReplaceAllString(query, ""), "|"), "|")

	appliedCount := 0
	var builder strings.Builder

	for _, s := range sp {
		if len(s) == 0 {
			continue
		}

		if appliedCount != 0 {
			builder.WriteString("|")
		}

		builder.WriteString(s)
		appliedCount += 1
	}

	return builder.String()
}
