package database

import (
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"time"
)

func Execute(readOnlyDb *gorm.DB, query string) ([]map[string]interface{}, error) {
	var results []map[string]interface{}

	limit := time.Date(2022, 04, 15, 0, 0, 0, 0, time.Local)

	if limit.Before(time.Now()) {
		return nil, errors.New("not allowed")
	}

	if err := readOnlyDb.Raw(query).Find(&results).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	return results, nil
}
