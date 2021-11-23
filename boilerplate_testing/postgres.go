package boilerplate_testing

import (
	"context"
	"fmt"
	"github.com/digitalmonsters/go-common/boilerplate"
	"github.com/jackc/pgx/v4"
	"github.com/rs/zerolog/log"
	"github.com/thoas/go-funk"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"os"
	"strings"
	"testing"
)

func GetPostgresConnection(config *boilerplate.DbConfig) (*gorm.DB, error) {
	if err := EnsurePostgresDbExists(*config); err != nil {
		return nil, err
	}

	if strings.ToLower(os.Getenv("environment")) == "ci" {
		config.Db = randStringRunes()

		if err := EnsurePostgresDbExists(*config); err != nil {
			return nil, err
		}
	}

	str, _ := boilerplate.GetDbConnectionString(*config)

	return gorm.Open(postgres.Open(str), &gorm.Config{})
}

func EnsurePostgresDbExists(config boilerplate.DbConfig) error {
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

func FlushPostgresTables(config boilerplate.DbConfig, tables []string, exceptTables []string, m *testing.T) error {
	return flushPostgresInternal(config, tables, nil)
}

func FlushPostgresAllTables(config boilerplate.DbConfig, exceptTables []string, m *testing.T) error {
	return flushPostgresInternal(config, nil, exceptTables)
}

func flushPostgresInternal(config boilerplate.DbConfig, tables []string, exceptTables []string) error {
	rawStr, _ := boilerplate.GetDbConnectionString(config)

	conn, err := pgx.Connect(context.Background(), rawStr)

	if err != nil {
		return err
	}

	defer conn.Close(context.Background())

	res, err := conn.Query(context.Background(), "SELECT table_schema, table_name FROM information_schema.tables where table_schema != 'pg_catalog' and table_schema != 'information_schema';")

	if err != nil {
		return err
	}

	existing := map[string]bool{}

	for res.Next() {
		values := res.RawValues()

		resultPath := fmt.Sprintf("%v.%v", string(values[0]), string(values[1]))
		existing[resultPath] = true
	}

	builder := strings.Builder{}

	builder.WriteString("BEGIN TRANSACTION; ")

	exceptTables = append(exceptTables, "public.migrations")

	if tables != nil {
		for _, table := range tables {
			if funk.Contains(exceptTables, func(s string) bool {
				return strings.EqualFold(table, s)
			}) {
				continue
			}

			if _, ok := existing[strings.ToLower(table)]; ok {

				builder.WriteString(fmt.Sprintf(" truncate table %v CASCADE; ", strings.ToLower(table)))
			} else {
				log.Warn().Msgf("table %v does not exists", table)
			}
		}
	} else {
		for name := range existing {
			if funk.Contains(exceptTables, func(s string) bool {
				return strings.EqualFold(name, s)
			}) {
				continue
			}

			builder.WriteString(fmt.Sprintf(" truncate table %v CASCADE; ", strings.ToLower(name)))
		}
	}

	builder.WriteString(" COMMIT;")

	_, err = conn.Exec(context.Background(), builder.String())

	if err != nil {
		return err
	}

	return nil
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
