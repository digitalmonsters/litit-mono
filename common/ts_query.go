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

	return stringRg.ReplaceAllString(alpha.ReplaceAllString(query, ""), "|")
}
