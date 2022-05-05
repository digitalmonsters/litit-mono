package database

import (
	"github.com/digitalmonsters/go-common/boilerplate_testing"
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func getMigrations() []*gormigrate.Migration {
	return []*gormigrate.Migration{
		{
			ID: "init_20220425",
			Migrate: func(db *gorm.DB) error {
				var sqls = []string{
					`create table if not exists feature_toggles (id bigserial primary key  not null, key text, value jsonb,
created_at timestamp with time zone default current_timestamp, updated_at timestamp with time zone default current_timestamp,
 deleted_at timestamp with time zone default null);`,

					`create unique index feature_toggles_uq on feature_toggles(key) where deleted_at is null;`,
				}
				for _, s := range sqls {
					if err := db.Exec(s).Error; err != nil {
						return err
					}
				}
				return nil
			},
		},
		{
			ID: "feature_events_table_20220428",
			Migrate: func(db *gorm.DB) error {
				var sqls = []string{
					`create table if not exists feature_toggle_events (id bigserial primary key  not null, feature_toggle_event jsonb,
created_at timestamp with time zone default current_timestamp);`,
				}
				for _, s := range sqls {
					if err := db.Exec(s).Error; err != nil {
						return err
					}
				}
				return nil
			},
		},
		{
			ID: "create_configs_table_20220503",
			Migrate: func(db *gorm.DB) error {
				return boilerplate_testing.ExecutePostgresSql(db,
					"create table if not exists configs (     key         varchar(255)  not null         primary key,     value       varchar(255),     type  varchar(255),     description varchar(255),     admin_only  boolean,     created_at  timestamp with time zone default CURRENT_TIMESTAMP not null,     updated_at  timestamp with time zone default CURRENT_TIMESTAMP not null,     category    varchar(255) );")
			},
		},
		{
			ID: "create_config_logs_table_20220504",
			Migrate: func(db *gorm.DB) error {
				return boilerplate_testing.ExecutePostgresSql(db, `create table if not exists config_logs
(
    id         bigserial
        primary key,
    key        varchar(255),
    value      varchar(255),
    related_user_id      bigint,
    created_at timestamp with time zone default CURRENT_TIMESTAMP not null,
    updated_at timestamp with time zone default CURRENT_TIMESTAMP not null
);`)
			},
		},
		{
			ID: "cleanup_20220504",
			Migrate: func(db *gorm.DB) error {
				return boilerplate_testing.ExecutePostgresSql(db, `
drop table if exists feature_toggle_events;
drop table if exists  feature_toggles;`)
			},
		},
	}
}
