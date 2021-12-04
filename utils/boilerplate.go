package utils

import (
	"bytes"
	"github.com/digitalmonsters/go-common/boilerplate"
	"github.com/pkg/errors"
	"github.com/romanyx/polluter"
	"gorm.io/gorm"
	"io/ioutil"
)

func PollutePostgresDatabase(gormDb *gorm.DB, filePaths ...string) error {
	var found []string

	for _, fileToFind := range filePaths {
		if path, err := boilerplate.RecursiveFindFile(fileToFind, "./", 30); err != nil {
			return errors.WithStack(err)
		} else {
			found = append(found, path)
		}
	}

	db, err := gormDb.DB()
	if err != nil {
		return err
	}

	seeder := polluter.
		New(polluter.JSONParser, polluter.PostgresEngine(db))

	for _, f := range found {
		data, err := ioutil.ReadFile(f)
		if err != nil {
			return err
		}

		if err = seeder.Pollute(bytes.NewReader(data)); err != nil {
			return err
		}
	}

	return nil
}
