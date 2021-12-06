package database

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func getMigrations() []*gormigrate.Migration {
	return []*gormigrate.Migration{
		{
			ID: "feat_comment_init_061220211736",
			Migrate: func(db *gorm.DB) error {
				query := `
					create table if not exists application
					(
						id             serial
							constraint application_pkey
								primary key,
						user_id        integer,
						type           integer,
						data           text,
						approved       boolean,
						reviewed_by    integer,
						reviewed_at    date,
						reason         varchar(255),
						upload_credits integer,
						ohw_used       boolean                  default false,
						created_at     timestamp with time zone default CURRENT_TIMESTAMP not null,
						updated_at     timestamp with time zone default CURRENT_TIMESTAMP not null,
						viewed         boolean                  default false
					);
					
					create table if not exists category
					(
						id          serial
							constraint category_pkey
								primary key,
						name        varchar(255),
						parent_id   integer
							constraint category_parent_id_foreign
								references category,
						status      integer,
						sort_order  integer,
						created_at  timestamp with time zone default CURRENT_TIMESTAMP not null,
						updated_at  timestamp with time zone default CURRENT_TIMESTAMP not null,
						emojis      varchar(255)             default NULL::character varying,
						views_count integer                  default 0,
						selected    integer                  default 0
					);
					
					create index if not exists idx_category_parent_id
						on category (parent_id);
					
					create table if not exists content
					(
						id                 serial
							constraint content_pkey
								primary key,
						user_id            integer,
						video_id           varchar(255),
						page_url           varchar(255),
						title              varchar(255),
						artist             varchar(255),
						description        text,
						tags               varchar(255),
						category_id        integer
							constraint content_category_id_foreign
								references category,
						subcategory_id     integer,
						duration           numeric,
						age_restricted     boolean                  default false,
						points             integer                  default 0,
						approved           boolean,
						reason             varchar(255),
						flagged            boolean                  default false,
						unlisted           boolean                  default false,
						live_at            timestamp with time zone,
						created_at         timestamp with time zone default CURRENT_TIMESTAMP not null,
						updated_at         timestamp with time zone default CURRENT_TIMESTAMP not null,
						ohw_application_id integer
							constraint content_ohw_application_id_foreign
								references application,
						whitelisted        boolean                  default false,
						whitelisted_by_id  integer,
						whitelisted_at     timestamp with time zone,
						approved_by_id     integer,
						suspended          boolean                  default false,
						suspended_by_id    integer,
						suspended_at       timestamp with time zone,
						deleted            boolean                  default false,
						deleted_by_id      integer,
						deleted_at         timestamp with time zone,
						hashtags           jsonb,
						allow_comments     boolean                  default false,
						video_share_link   varchar(255)             default NULL::character varying,
						draft              boolean                  default true,
						width              integer,
						height             integer,
						not_to_repeat      boolean                  default false,
						upload_status      integer                  default 0,
						by_admin           boolean                  default false,
						moderator_rate     integer,
						allow_download     boolean                  default true,
						fps                varchar(255),
						bitrate            varchar(255),
						size               varchar(255),
						likes_count        bigint,
						watch_count        bigint,
						hashtags_array     text[],
						shares_count       bigint                   default 0                 not null,
						comments_count     bigint                   default 0                 not null
					);
					
					create index if not exists idx_hashtags
						on content (hashtags);
					
					create index if not exists idx_user_otp_id
						on content (user_id);
					
					create index if not exists upload_status_idx
						on content (upload_status);
					
					create index if not exists gin_content_hashtags_idx
						on content (hashtags_array);
					
					alter table content
						add COLUMN IF NOT EXISTS comments_count bigint default 0 not null;
					
					create table if not exists country
					(
						code      char(2) not null
							constraint country_pkey
								primary key,
						name      varchar(255),
						min_age   integer default 13,
						adult_age integer default 18
					);
					
					create table if not exists profile
					(
						id                     integer                                            not null
							constraint profile_pkey
								primary key,
						bio                    text,
						phone                  varchar(255),
						address_city           varchar(255),
						address_street         varchar(255),
						address_postal_code    varchar(255),
						company_name           varchar(255),
						company_phone          varchar(255),
						company_tax            varchar(255),
						company_address        varchar(255),
						affiliation_agent      varchar(255),
						affiliation_management varchar(255),
						affiliation_publisher  varchar(255),
						affiliation_partners   varchar(255),
						social_facebook        varchar(255),
						social_instagram       varchar(255),
						social_twitter         varchar(255),
						social_website         varchar(255),
						social_youtube         varchar(255),
						created_at             timestamp with time zone default CURRENT_TIMESTAMP not null,
						updated_at             timestamp with time zone default CURRENT_TIMESTAMP not null,
						address_country_code   char(2)
							constraint profile_address_country_code_foreign
								references country,
						address_state          varchar(255),
						profile_share_link     varchar(255)             default NULL::character varying,
						social_tiktok          varchar(255)             default NULL::character varying,
						social_clubhouse       varchar(255)             default NULL::character varying,
						social_twitch          varchar(255)             default NULL::character varying,
						guardian_id            integer,
						social_reddit          varchar(255)             default ''::character varying,
						social_quora           varchar(255)             default ''::character varying,
						social_medium          varchar(255)             default ''::character varying,
						social_linkedin        varchar(255)             default ''::character varying,
						social_discord         varchar(255)             default ''::character varying,
						social_telegram        varchar(255)             default ''::character varying,
						social_viber           varchar(255)             default ''::character varying,
						social_whatsapp        varchar(255)             default ''::character varying,
						country_code           varchar(255)             default ''::character varying
					);
					
					
					
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
					
					create sequence if not exists report_id_seq;
					
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
					
					alter sequence if exists report_id_seq owned by report.id;
				`
				return db.Exec(query).Error

			},
			Rollback: func(db *gorm.DB) error {
				return nil
			},
		},
	}
}
