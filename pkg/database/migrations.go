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
		{
			ID: "feat_index_change_17022022",
			Migrate: func(db *gorm.DB) error {
				query := `drop index messages_title_uindex;
						  create unique index messages_title_uindex on messages (title) where deleted_at is null;`
				return db.Exec(query).Error
			},
			Rollback: func(db *gorm.DB) error {
				return nil
			},
		},
		{
			ID: "feat_messages_type_220520221310",
			Migrate: func(db *gorm.DB) error {
				query := `	alter table messages add column if not exists type int;
							update messages set type = 1;
							drop index messages_title_uindex;
						  	create unique index messages_title_uindex on messages (title, type) where deleted_at is null;`
				return db.Exec(query).Error
			},
			Rollback: func(db *gorm.DB) error {
				return nil
			},
		},
		{
			ID: "ad_campaigns_2507221240",
			Migrate: func(db *gorm.DB) error {
				return db.Exec(`create table if not exists ad_campaigns
					(
						id serial
							constraint ad_campaigns_pk
								primary key,
						user_id bigint,
						name text not null,
						ad_type smallint not null default 1,
						status smallint not null default 0,
						content_id bigint not null,
						link text,
						link_button_id int,
						country varchar(255),
						created_at timestamp with time zone default current_timestamp,
						started_at timestamp with time zone,
						ended_at timestamp with time zone,
						duration_min int not null default 0,
						budget int not null default 0,
						gender varchar(255) default NULL::character varying,
						age_from int,
						age_to int
					)
				`).Error
			},
			Rollback: func(db *gorm.DB) error {
				return nil
			},
		},
		{
			ID: "dictionary_tables_create_2607221623",
			Migrate: func(db *gorm.DB) error {
				return db.Exec(
					`create table if not exists
    reject_reasons
(
    id         bigserial primary key,
    reason     text,
    created_at timestamp with time zone default current_timestamp,
    deleted_at timestamp with time zone
);

create unique index if not exists reject_reasons_reason_uq on reject_reasons(reason) where deleted_at is null;

create table if not exists
    action_buttons
(
    id         bigserial primary key,
    name     text,
	type smallint,
    created_at timestamp with time zone default current_timestamp,
    deleted_at timestamp with time zone
);

create unique index if not exists action_buttons_name_uq on action_buttons(name) where deleted_at is null;`).Error
			},
			Rollback: func(db *gorm.DB) error {
				return nil
			},
		},
		{
			ID: "ads_campaign_reject_reason_220520221310",
			Migrate: func(db *gorm.DB) error {
				query := `alter table ad_campaigns add column if not exists reject_reason_id bigint;`
				return db.Exec(query).Error
			},
			Rollback: func(db *gorm.DB) error {
				return nil
			},
		},
		{
			ID: "action_buttons_insert_220520221310",
			Migrate: func(db *gorm.DB) error {
				query := `insert into action_buttons (name, type) values
                                               ('Learn more', 1),
                                               ('Shop now', 1),
                                               ('Sign up', 1),
                                               ('Contact us', 1),
                                               ('Apply now', 1),
                                               ('Book now', 1);`
				return db.Exec(query).Error
			},
			Rollback: func(db *gorm.DB) error {
				return nil
			},
		},
	}
}
