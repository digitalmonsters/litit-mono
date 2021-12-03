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
					create table comment
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

					create index comment_idx_content_id
						on comment (content_id);

					create index idx_comment_content_id
						on comment (content_id);

					create index idx_comment_parent_id
						on comment (parent_id);
					
					create index idx_comment_profile_id
						on comment (profile_id);

					create table comment_vote
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
					)
				`
				return db.Exec(query).Error

			},
			Rollback: func(db *gorm.DB) error {
				return nil
			},
		},
	}
}
