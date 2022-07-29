package database

import (
	"github.com/digitalmonsters/go-common/boilerplate"
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func getMigrations() []*gormigrate.Migration {
	return []*gormigrate.Migration{
		{
			ID: "re_migrate_in_prod_2022022",
			Migrate: func(db *gorm.DB) error {
				if boilerplate.GetCurrentEnvironment() == boilerplate.Prod {
					query := `drop table if exists migrations, playlist_song_relations, playlists, songs, favorites, music_storage cascade;
							create table migrations
							(
								id varchar(255) not null
									constraint migrations_pkey
										primary key
							)`
					return db.Exec(query).Error
				}

				return nil
			},
			Rollback: func(db *gorm.DB) error {
				return nil
			},
		},
		{
			ID: "feat_playlists_init_050120221603",
			Migrate: func(db *gorm.DB) error {
				query := `create table playlists
						(
							id          serial
								constraint playlists_pk
									primary key,
							name        text,
							color       varchar,
							is_active   boolean,
							songs_count integer,
							sort_order  integer,
							created_at  timestamp default CURRENT_TIMESTAMP,
							updated_at  timestamp default CURRENT_TIMESTAMP,
							deleted_at  timestamp
						);
						
						create unique index playlists_name_uindex
							on playlists (name);`
				return db.Exec(query).Error
			},
			Rollback: func(db *gorm.DB) error {
				return nil
			},
		},
		{
			ID: "feat_songs_init_050120221603",
			Migrate: func(db *gorm.DB) error {
				query := `create table songs
						(
							id serial
								constraint songs_pk
									primary key,
							source int,
							external_id text,
							title text,
							artist text,
							image_url text,
							genre text,
							duration numeric,
							listen_amount bigint,
							created_at timestamp default current_timestamp,
							updated_at timestamp default current_timestamp,
							deleted_at timestamp default null
						);
						
						create unique index songs_external_id_source_uindex
							on songs (external_id, source);
						
						`
				return db.Exec(query).Error
			},
			Rollback: func(db *gorm.DB) error {
				return nil
			},
		},
		{
			ID: "feat_playlist_song_relations_init_050120221603",
			Migrate: func(db *gorm.DB) error {
				query := `create table playlist_song_relations
						(
							playlist_id bigint not null
								constraint playlist_song_relations_playlists_id_fk
									references playlists,
							song_id     bigint not null
								constraint playlist_song_relations_songs_id_fk
									references songs,
							sort_order  integer,
							constraint playlist_song_relations_pk
								primary key (song_id, playlist_id)
						);`
				return db.Exec(query).Error
			},
			Rollback: func(db *gorm.DB) error {
				return nil
			},
		},
		{
			ID: "feat_triggers_050120221849",
			Migrate: func(db *gorm.DB) error {
				query := `CREATE OR REPLACE FUNCTION add_song_to_playlist()
						  RETURNS trigger AS
						$func$
						BEGIN
						   Update playlists set songs_count = songs_count + 1 where id = NEW.playlist_id;
						   RETURN NEW;
						END
						$func$  LANGUAGE plpgsql;
						
						CREATE TRIGGER add_song_to_playlist
							AFTER INSERT
							ON playlist_song_relations
							FOR EACH ROW
						EXECUTE PROCEDURE add_song_to_playlist();

						CREATE OR REPLACE FUNCTION delete_song_from_playlist()
							RETURNS trigger AS
						$func$
						BEGIN
							Update playlists set songs_count = songs_count - 1 where id = OLD.playlist_id;
							RETURN NEW;
						END
						$func$ LANGUAGE plpgsql;
						
						CREATE TRIGGER delete_song_from_playlist
							AFTER DELETE
							ON playlist_song_relations
							FOR EACH ROW
						EXECUTE PROCEDURE delete_song_from_playlist()`
				return db.Exec(query).Error
			},
			Rollback: func(db *gorm.DB) error {
				return nil
			},
		},
		{
			ID: "feat_favorites_init_060120221134",
			Migrate: func(db *gorm.DB) error {
				query := `
						create table if not exists favorites
					(
						user_id    bigint not null,
						song_id    bigint not null,
						created_at timestamp default CURRENT_TIMESTAMP,
						constraint favorites_pk
							primary key (user_id, song_id)
					);`
				return db.Exec(query).Error
			},
			Rollback: func(db *gorm.DB) error {
				return nil
			},
		},
		{
			ID: "feat_music_storage_init_180120221913",
			Migrate: func(db *gorm.DB) error {
				query := `
						create table if not exists music_storage
						(
							id serial
								constraint music_storage_pk
									primary key,
							title text,
							description text,
							image_url text,
							duration numeric,
							url text,
							genre text,
							artist text,
							hash text,
							created_at timestamp default current_timestamp,
							updated_at timestamp default current_timestamp,
							deleted_at timestamp default null
						);
				
						create unique index music_storage_hash_uindex
							on music_storage (hash);`
				return db.Exec(query).Error
			},
			Rollback: func(db *gorm.DB) error {
				return nil
			},
		},
		{
			ID: "feat_music_storage_index_190120221346",
			Migrate: func(db *gorm.DB) error {
				query := `
						drop index music_storage_hash_uindex;
						alter table music_storage drop column hash;
						create unique index music_storage_url_uindex on music_storage (url);`
				return db.Exec(query).Error
			},
			Rollback: func(db *gorm.DB) error {
				return nil
			},
		},
		{
			ID: "drop_index_260120221945",
			Migrate: func(db *gorm.DB) error {
				query := `drop index if exists music_storage_url_uindex;`
				return db.Exec(query).Error
			},
			Rollback: func(db *gorm.DB) error {
				return nil
			},
		},
		{
			ID: "playlists_name_uindex_03022022",
			Migrate: func(db *gorm.DB) error {
				query := `drop index playlists_name_uindex;
						  create unique index playlists_name_uindex on playlists (name) where deleted_at is null;`
				return db.Exec(query).Error
			},
			Rollback: func(db *gorm.DB) error {
				return nil
			},
		},
		{
			ID: "creator_init_080220221909",
			Migrate: func(db *gorm.DB) error {
				query := `create table if not exists creators
							(
								id serial
								constraint creators_pk
								primary key,
								user_id bigint,
								status int,
								reject_reason text,
								library_url text,
								created_at timestamp default current_timestamp,
								approved_at timestamp,
								deleted_at timestamp
							);`
				return db.Exec(query).Error
			},
			Rollback: func(db *gorm.DB) error {
				return nil
			},
		},
		{
			ID: "reject_reasons_init_10022022",
			Migrate: func(db *gorm.DB) error {
				query := `create table if not exists creator_reject_reasons
						(
							id serial
							constraint reject_reasons_pk
							primary key,
							reason text,
							created_at timestamp default current_timestamp,
							deleted_at timestamp
						);
					
							create unique index creator_reject_reasons_reason_uindex
							on creator_reject_reasons (reason);`
				return db.Exec(query).Error
			},
			Rollback: func(db *gorm.DB) error {
				return nil
			},
		},
		{
			ID: "new_fk_reject_reason_10022022",
			Migrate: func(db *gorm.DB) error {
				query := `alter table creators alter column reject_reason type int using reject_reason::int;
						  alter table creators add constraint creators_creator_reject_reasons_id_fk
						  foreign key (reject_reason) references creator_reject_reasons;`
				return db.Exec(query).Error
			},
			Rollback: func(db *gorm.DB) error {
				return nil
			},
		},
		{
			ID: "categories_init_11022022",
			Migrate: func(db *gorm.DB) error {
				query := `create table if not exists categories
						(
							id serial
							constraint categories_pk
							primary key,
							name text,
							sort_order int,
							songs_count int,
							created_at timestamp default current_timestamp,
							updated_at timestamp,
							deleted_at timestamp
						);
					
							create unique index categories_name_uindex
							on categories (name);`
				return db.Exec(query).Error
			},
			Rollback: func(db *gorm.DB) error {
				return nil
			},
		},
		{
			ID: "creator_songs_init_11022022",
			Migrate: func(db *gorm.DB) error {
				query := `create table if not exists creator_songs
						(
							id serial
								constraint creator_songs_pk
									primary key,
							user_id bigint,
							name text,
							status int,
							lyric_author text,
							music_author text,
							category_id bigint,
							full_song_url text,
							short_song_url text,
							image_url text,
							hashtags text[],
							short_listens int,
							full_listens int,
							likes int,
							comments int,
							used_in_video int,
							points_earned numeric,
							created_at timestamp default current_timestamp,
							updated_at timestamp,
							deleted_at timestamp
						);`
				return db.Exec(query).Error
			},
			Rollback: func(db *gorm.DB) error {
				return nil
			},
		},
		{
			ID: "add_reasons_28022022",
			Migrate: func(db *gorm.DB) error {
				query := `INSERT INTO creator_reject_reasons (id, reason, created_at, deleted_at) VALUES (DEFAULT, 'we have identified copyright issues'::text, DEFAULT, null::timestamp) on conflict do nothing;
						  INSERT INTO creator_reject_reasons (id, reason, created_at, deleted_at) VALUES (DEFAULT, 'the music wasn''t fully produced by you'::text, DEFAULT, null::timestamp) on conflict do nothing;
						  INSERT INTO creator_reject_reasons (id, reason, created_at, deleted_at) VALUES (DEFAULT, 'currently your music cannot be added to lit.it music library'::text, DEFAULT, null::timestamp) on conflict do nothing;`
				return db.Exec(query).Error
			},
			Rollback: func(db *gorm.DB) error {
				return nil
			},
		},
		{
			ID: "creator_songs_new_columns_160320222018",
			Migrate: func(db *gorm.DB) error {
				query := `alter table creator_songs add column if not exists full_song_duration numeric;
						  alter table creator_songs add column if not exists short_song_duration numeric;`
				return db.Exec(query).Error
			},
			Rollback: func(db *gorm.DB) error {
				return nil
			},
		},
		{
			ID: "moods_init_210320221903",
			Migrate: func(db *gorm.DB) error {
				query := `create table if not exists moods
						(
							id serial
							constraint moods_pk
							primary key,
							name text,
							sort_order int,
							songs_count int,
							is_active boolean,
							created_at timestamp default current_timestamp,
							updated_at timestamp,
							deleted_at timestamp
						);
					
							create unique index moods_name_uindex
							on moods (name);`
				return db.Exec(query).Error
			},
			Rollback: func(db *gorm.DB) error {
				return nil
			},
		},
		{
			ID: "categories_is_active_220320221246",
			Migrate: func(db *gorm.DB) error {
				query := `alter table categories add column if not exists is_active boolean;`
				return db.Exec(query).Error
			},
			Rollback: func(db *gorm.DB) error {
				return nil
			},
		},
		{
			ID: "feat_mood_field_220320221351",
			Migrate: func(db *gorm.DB) error {
				query := `alter table creator_songs add column if not exists mood_id int;`
				return db.Exec(query).Error
			},
			Rollback: func(db *gorm.DB) error {
				return nil
			},
		},
		{
			ID: "feat_songs_count_230320221405",
			Migrate: func(db *gorm.DB) error {
				query := `alter table creators add column if not exists songs_count int;`
				return db.Exec(query).Error
			},
			Rollback: func(db *gorm.DB) error {
				return nil
			},
		},
		{
			ID: "feat_songs_count_trigger_230320221412",
			Migrate: func(db *gorm.DB) error {
				query := `CREATE OR REPLACE FUNCTION creator_add_new_song()
						RETURNS trigger AS
						$func$
						BEGIN
						update creators set songs_count = creators.songs_count + 1 where user_id = NEW.user_id;
						update categories set songs_count = categories.songs_count + 1 where id = NEW.category_id;
						update moods set songs_count = moods.songs_count + 1 where id = NEW.mood_id;
						RETURN NEW;
						END
						$func$  LANGUAGE plpgsql;
				
						CREATE TRIGGER creator_add_new_song
						AFTER INSERT
						ON creator_songs
						FOR EACH ROW
						EXECUTE PROCEDURE creator_add_new_song();`
				return db.Exec(query).Error
			},
			Rollback: func(db *gorm.DB) error {
				return nil
			},
		},
		{
			ID: "feat_foreign_keys_230320221418",
			Migrate: func(db *gorm.DB) error {
				query := `alter table creator_songs
						add constraint creator_songs_categories_id_fk
						foreign key (category_id) references categories;

						alter table creator_songs
						add constraint creator_songs_moods_id_fk
						foreign key (mood_id) references moods;`
				return db.Exec(query).Error
			},
			Rollback: func(db *gorm.DB) error {
				return nil
			},
		},
		{
			ID: "creator_songs_reject_reason_300320221312",
			Migrate: func(db *gorm.DB) error {
				query := `alter table creator_songs add column if not exists reject_reason int;
						  alter table creator_reject_reasons add column if not exists type int;
						  update creator_reject_reasons set type = 1;`
				return db.Exec(query).Error
			},
			Rollback: func(db *gorm.DB) error {
				return nil
			},
		},
		{
			ID: "creators_add_user_info_210620221405",
			Migrate: func(db *gorm.DB) error {
				query := `alter table creators add column if not exists username text;
						  alter table creators add column if not exists firstname text;
						  alter table creators add column if not exists lastname text;
						  alter table creators add column if not exists email text;`
				return db.Exec(query).Error
			},
			Rollback: func(db *gorm.DB) error {
				return nil
			},
		},
		{
			ID: "feat_listened_music_21072022",
			Migrate: func(db *gorm.DB) error {
				query := `create table if not exists listened_music
						(
							user_id    bigint not null,
							song_id    bigint not null,
							primary key (user_id, song_id)
						);
				
						create index listened_by_user_inx
							on listened_music (user_id) include (song_id);`
				return db.Exec(query).Error
			},
			Rollback: func(db *gorm.DB) error {
				return nil
			},
		},
		{
			ID: "feat_music_reactions_220720221137",
			Migrate: func(db *gorm.DB) error {
				query := `alter table creator_songs add column if not exists dislikes integer default 0;
						  alter table creator_songs add column if not exists loves integer default 0;
						  alter table creator_songs add column if not exists shares integer default 0;
						  alter table creator_songs add column if not exists score integer default 0;`
				return db.Exec(query).Error
			},
		},
	}
}
