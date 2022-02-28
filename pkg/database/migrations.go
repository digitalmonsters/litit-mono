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
						create table favorites
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
						create table music_storage
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
				query := `create table creators
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
				query := `create table creator_reject_reasons
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
				query := `create table categories
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
				query := `create table creator_songs
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
	}
}
