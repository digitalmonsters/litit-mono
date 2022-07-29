package utils

import (
	"fmt"
	"github.com/jinzhu/inflection"
	"gorm.io/gorm"
	"strings"
)

func AddSearchQuery(query *gorm.DB, keywords []string, fieldNames []string) *gorm.DB {
	if len(keywords) == 0 || len(fieldNames) == 0 {
		return query
	}

	for _, fieldName := range fieldNames {
		if len(fieldName) == 0 {
			continue
		}

		for _, keyword := range keywords {
			if len(keyword) == 0 {
				continue
			}

			keywordList := GetSingleAndPluralValues(keyword)

			for _, word := range keywordList {
				query = query.Or(fmt.Sprintf("%v ILIKE '%%' || ? || '%%'", fieldName), word)
			}
		}
	}

	return query
}

func GetSingleAndPluralValues(keywordStr string) []string {
	keywordList := strings.Fields(strings.ToLower(keywordStr))
	var res []string

	for _, keyword := range keywordList {
		res = append(res, keyword)

		pluralKeyword := inflection.Plural(keyword)

		if pluralKeyword != keyword {
			res = append(res, pluralKeyword)
		}

		singularKeyword := inflection.Singular(keyword)

		if singularKeyword != keyword {
			res = append(res, singularKeyword)
		}
	}

	return res
}
