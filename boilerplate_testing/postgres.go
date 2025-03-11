package boilerplate_testing

import (
	"bytes"
	"context"
	"fmt"
	"os"
	path2 "path"
	"strings"
	"testing"

	"github.com/digitalmonsters/go-common/boilerplate"
	"github.com/jackc/pgx/v4"
	"github.com/juju/fslock"
	"github.com/pkg/errors"
	"github.com/romanyx/polluter"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func GetPostgresConnection(config *boilerplate.DbConfig) (*gorm.DB, error) {
	if err := EnsurePostgresDbExists(*config); err != nil {
		return nil, err
	}

	if strings.ToLower(os.Getenv("ENVIRONMENT")) == "ci" {
		config.Db = randStringRunes()

		if err := EnsurePostgresDbExists(*config); err != nil {
			return nil, err
		}
	}

	str, _ := boilerplate.GetDbConnectionString(*config)

	return gorm.Open(postgres.Open(str), &gorm.Config{})
}

func getLock() *fslock.Lock {
	return fslock.New(path2.Join(os.TempDir(), "go_fs_lock"))
}

func EnsurePostgresDbExists(config boilerplate.DbConfig) error {
	lock := getLock()

	if err := lock.Lock(); err != nil {
		return err
	}

	defer func() {
		_ = lock.Unlock()
	}()

	oldDbName := config.Db

	config.Db = "postgres"
	rawStr, _ := boilerplate.GetDbConnectionString(config)

	conn, err := pgx.Connect(context.Background(), rawStr)
	config.Db = oldDbName

	if err != nil {
		return err
	}

	defer conn.Close(context.Background())

	r, err := conn.Query(context.Background(), fmt.Sprintf("SELECT * FROM pg_database WHERE datname='%v'", oldDbName))

	if err != nil {
		return err
	}

	r.Next()

	exists := true

	if len(r.RawValues()) == 0 {
		exists = false
	}

	if !exists {
		_, err = conn.Exec(context.Background(), fmt.Sprintf("CREATE DATABASE %v;", oldDbName))

		if err != nil {
			return err
		}
	}

	return nil
}

func DropPostgresCiDatabase(config *boilerplate.DbConfig, m *testing.M) error {
	oldDbName := config.Db

	config.Db = "postgres"
	rawStr, _ := boilerplate.GetDbConnectionString(*config)

	conn, err := pgx.Connect(context.Background(), rawStr)
	config.Db = oldDbName

	if err != nil {
		return err
	}

	defer conn.Close(context.Background())

	_, err = conn.Exec(context.Background(), fmt.Sprintf("SELECT pg_terminate_backend(pg_stat_activity.pid) FROM pg_stat_activity WHERE pg_stat_activity.datname = '%v'  AND pid <> pg_backend_pid();", oldDbName))

	if err != nil {
		return err
	}

	_, err = conn.Exec(context.Background(), fmt.Sprintf("DROP DATABASE %v", oldDbName))

	return err
}

func GetPostgresCiDatabaseName() string {
	return fmt.Sprintf("ci_%v", int64(os.Getpid()))
}

func randStringRunes() string {
	return fmt.Sprintf("ci_%v", boilerplate.GetGenerator().Generate().String())
}

func ExecutePostgresSql(db *gorm.DB, sql ...string) error {
	for _, s := range sql {
		if err := db.Exec(s).Error; err != nil {
			return err
		}
	}

	return nil
}

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
		data, err := os.ReadFile(f)
		if err != nil {
			return err
		}

		if err = seeder.Pollute(bytes.NewReader(data)); err != nil {
			return err
		}
	}

	return nil
}
