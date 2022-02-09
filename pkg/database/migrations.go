package database

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func getMigrations() []*gormigrate.Migration {
	return []*gormigrate.Migration{
		{
			ID: "feat_messages_init01022022",
			Migrate: func(db *gorm.DB) error {
				query := `create table messages
						(
							id serial
								constraint messages_pk
									primary key,
							title text,
							description text,
							countries text[],
							verification_status smallint,
							age_from smallint default null,
							age_to smallint default null,
							points_from numeric default null,
							points_to numeric default null,
							is_active   boolean default false,
							created_at timestamp default current_timestamp,
							updated_at timestamp default current_timestamp,
							deleted_at timestamp default null,
							deactivated_at timestamp default null
						);`
				return db.Exec(query).Error
			},
			Rollback: func(db *gorm.DB) error {
				return nil
			},
		},
		{
			ID: "feat_title_ui_09022022",
			Migrate: func(db *gorm.DB) error {
				query := `create unique index messages_title_uindex on messages (title);`
				return db.Exec(query).Error
			},
			Rollback: func(db *gorm.DB) error {
				return nil
			},
		},
	}
}
