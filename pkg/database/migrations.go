package database

import (
	"github.com/digitalmonsters/go-common/boilerplate_testing"
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func getMigrations() []*gormigrate.Migration {
	return []*gormigrate.Migration{
		{
			ID: "initial_node_tables_20220310",
			Migrate: func(db *gorm.DB) error {
				return boilerplate_testing.ExecutePostgresSql(db,
					"create table if not exists \"Devices\"\n(\n    id          uuid                     default gen_random_uuid() not null\n        primary key,\n    \"userId\"    integer                                            not null,\n    \"deviceId\"  varchar(255)                                       not null,\n    \"pushToken\" varchar(255)                                       not null,\n    platform    varchar(255)                                       not null,\n    \"createdAt\" timestamp with time zone default CURRENT_TIMESTAMP not null,\n    \"updatedAt\" timestamp with time zone,\n    \"deletedAt\" timestamp with time zone\n);\n\ncreate index if not exists \"Devices_user_idx\"\n    on \"Devices\" (\"userId\");\n\n",
					"create table if not exists notifications\n(\n    id                     uuid                     default gen_random_uuid() not null\n        primary key,\n    user_id                integer                                            not null,\n    type                   varchar(255)                                       not null,\n    title                  varchar(255)                                       not null,\n    message                varchar(255)                                       not null,\n    related_user_id        integer,\n    comment_id             integer,\n    comment                jsonb,\n    content_id             integer,\n    content                jsonb,\n    question_id            integer,\n    created_at             timestamp with time zone default CURRENT_TIMESTAMP not null,\n    kyc_reason             varchar(255),\n    kyc_status             varchar(255),\n    content_creator_status integer\n);\n\ncreate index if not exists notifications_search_idx\n    on notifications (user_id, type, created_at);\n\ncreate index if not exists notifications_user_idx\n    on notifications (user_id);\n\n",
					"create table if not exists \"SequelizeMeta\"\n(\n    name varchar(255) not null\n        primary key\n);",
					"create table if not exists \"Tasks\"\n(\n    id                 uuid                     default gen_random_uuid() not null\n        primary key,\n    priority           integer                                            not null,\n    status             integer                                            not null,\n    \"notificationType\" varchar(255)                                       not null,\n    metadata           jsonb,\n    \"createdAt\"        timestamp with time zone default CURRENT_TIMESTAMP not null,\n    \"updatedAt\"        timestamp with time zone\n);\n\ncomment on column \"Tasks\".priority is '0 - low, 1 - medium, 2 - high';\n\ncomment on column \"Tasks\".status is '0 - todo, 1 - in progress, 2 - success, 3 - failed, 4 - background';\n\ncomment on column \"Tasks\".\"notificationType\" is 'the type of notification which is uniq for each one';\n\ncreate index if not exists \"Tasks_createdAt_idx\"\n    on \"Tasks\" (\"createdAt\");\n\ncreate index if not exists \"Tasks_notificationType_idx\"\n    on \"Tasks\" (\"notificationType\");\n\ncreate index if not exists \"Tasks_priority_idx\"\n    on \"Tasks\" (priority);\n\ncreate index if not exists \"Tasks_status_idx\"\n    on \"Tasks\" (status);\n\n",
					"create table if not exists user_notifications\n(\n    user_id      integer           not null\n        primary key,\n    unread_count integer default 0 not null\n);\n\ncreate index if not exists user_notifications_unread_count_idx\n    on user_notifications (unread_count);\n\n",
				)
			},
		},
		{
			ID: "additional_tables_20220310",
			Migrate: func(db *gorm.DB) error {
				return boilerplate_testing.ExecutePostgresSql(db,
					"create table if not exists render_templates\n(\n    id         text      not null\n        constraint render_templates_pk\n            primary key,\n    title      text,\n    body       text,\n    created_at timestamp not null,\n    updated_at timestamp not null\n);",
					"INSERT INTO public.render_templates (id, title, body, created_at, updated_at)\nVALUES ('first_x_paid_views', 'You earned your first {{.total_earned}} LIT points for watching video ',\n        'You earn {{.view_bonus}} LIT point for each video you watched for {{.watch_threshold}}% of the time or more',\n        '2022-03-10 12:32:42.000000', '2022-03-10 12:32:42.000000');",
					"INSERT INTO public.render_templates (id, title, body, created_at, updated_at)\nVALUES ('signup_bonus', 'You earned your first {{.reward_amount}} LIT points for joining Lit.it',\n        'You can invite friends and earn {{.invite_reward_amount}} LIT points for each friend who will join. ',\n        '2022-03-10 12:32:42.000000', '2022-03-10 12:32:42.000000');",
				)
			},
		},
	}
}
