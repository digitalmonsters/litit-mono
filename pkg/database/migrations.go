package database

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func getMigrations() []*gormigrate.Migration {
	return []*gormigrate.Migration{
		{
			ID: "feat_comment_init_031220211717",
			Migrate: func(db *gorm.DB) error {
				query := `
					create table if not exists comment
					(
						id            serial
							constraint comment_pkey
								primary key,
						author_id     integer,
						content_id    integer
							constraint comment_content_id_foreign
								references content
								on delete cascade,
						profile_id    integer
							constraint comment_profile_id_foreign
								references profile
								on delete cascade,
						parent_id     integer
							constraint comment_parent_id_foreign
								references comment
								on delete cascade,
						comment       text,
						num_replies   integer                  default 0,
						num_upvotes   integer                  default 0,
						num_downvotes integer                  default 0,
						flagged       boolean                  default false,
						active        boolean                  default true,
						created_at    timestamp with time zone default CURRENT_TIMESTAMP not null,
						updated_at    timestamp with time zone default CURRENT_TIMESTAMP not null
					);

					create index if not exists comment_idx_content_id
						on comment (content_id);

					create index if not exists idx_comment_content_id
						on comment (content_id);

					create index if not exists idx_comment_parent_id
						on comment (parent_id);
					
					create index if not exists idx_comment_profile_id
						on comment (profile_id);

					alter table comment
						drop constraint if exists comment_author_id_foreign;

					alter table content
						drop constraint if exists content_user_id_foreign;

					create table if not exists comment_vote
					(
						user_id    integer                                            not null,
						comment_id integer                                            not null
							constraint comment_vote_comment_id_foreign
								references comment
								on delete cascade,
						vote_up    boolean,
						created_at timestamp with time zone default CURRENT_TIMESTAMP not null,
						updated_at timestamp with time zone default CURRENT_TIMESTAMP not null,
						constraint comment_vote_pkey
							primary key (user_id, comment_id)
					);

					create table if not exists user_stats_content
					(
						id                integer                                            not null
							constraint user_stats_content_pkey
								primary key,
						uploads           integer                  default 0,
						viral_power       integer                  default 0,
						watch_time        integer                  default 0,
						views             integer                  default 0,
						shares            integer                  default 0,
						likes             integer                  default 0,
						up_score          integer                  default 0,
						dislikes          integer                  default 0,
						down_score        integer                  default 0,
						comments          integer                  default 0,
						created_at        timestamp with time zone default CURRENT_TIMESTAMP not null,
						updated_at        timestamp with time zone default CURRENT_TIMESTAMP not null,
						paid_views        integer                  default 0,
						paid_guest_views  integer                  default 0,
						paid_shares       integer                  default 0,
						paid_guest_shares integer                  default 0
					);

					create table if not exists user_stats_action
					(
						id             integer                                            not null
							constraint user_stats_action_pkey
								primary key,
						viral_power    integer                  default 0,
						watch_time     integer                  default 0,
						views          integer                  default 0,
						shares         integer                  default 0,
						likes          integer                  default 0,
						up_score       integer                  default 0,
						dislikes       integer                  default 0,
						down_score     integer                  default 0,
						comments       integer                  default 0,
						created_at     timestamp with time zone default CURRENT_TIMESTAMP not null,
						updated_at     timestamp with time zone default CURRENT_TIMESTAMP not null,
						paid_views     integer                  default 0,
						paid_shares    integer                  default 0,
						profile_shares integer                  default 0
					);

					create table if not exists report
					(
						id          integer                  default nextval('report_id_seq'::regclass) not null
							constraint report_pkey
								primary key,
						content_id  integer
							constraint report_content_id_foreign
								references content
								on delete cascade,
						user_id     integer,
						type        varchar(255),
						detail      text,
						created_at  timestamp with time zone default CURRENT_TIMESTAMP                  not null,
						updated_at  timestamp with time zone default CURRENT_TIMESTAMP                  not null,
						reporter_id integer,
						report_type varchar(255)             default 'content'::character varying,
						comment_id  integer
							constraint report_comment_id_foreign
								references comment
								on delete cascade,
						resolved    boolean                  default false
					);
				`
				return db.Exec(query).Error

			},
			Rollback: func(db *gorm.DB) error {
				return nil
			},
		},
	}
}
