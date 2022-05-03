package database

import (
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
	}
}
