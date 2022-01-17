package database

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func getMigrations() []*gormigrate.Migration {
	return []*gormigrate.Migration{
		{
			ID: "feat_playlists_init_050120221603",
			Migrate: func(db *gorm.DB) error {
				query := `create table playlists
						(
							id serial
								constraint playlists_pk
									primary key,
							name text,
							color varchar,
							sort_order int,
							songs_count int default 0,
							created_at timestamp default current_timestamp,
							deleted_at timestamp default null
						);`
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
							id text
								constraint songs_pk
									primary key,
							title text,
							artist text,
							url text,
							image_url text,
							created_at timestamp default current_timestamp,
							updated_at timestamp default current_timestamp
						);`
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
						playlist_id bigint,
						song_id text,
						sort_order int,
						constraint playlist_song_relations_pk
						primary key (playlist_id, song_id)
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
			ID: "feat_playlist_song_relations_fk_050120221859",
			Migrate: func(db *gorm.DB) error {
				query := `alter table playlist_song_relations add constraint playlist_song_relations_playlists_id_fk foreign key (playlist_id) references playlists;
						  alter table playlist_song_relations add constraint playlist_song_relations_songs_id_fk foreign key (song_id) references songs;`
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
							user_id bigint,
							song_id text
								constraint favorites_songs_id_fk
									references songs,
							created_at timestamp default current_timestamp,
							constraint favorites_pk
								primary key (user_id, song_id)
						);
				`
				return db.Exec(query).Error
			},
			Rollback: func(db *gorm.DB) error {
				return nil
			},
		},
		{
			ID: "feat_listen_amount_060120221741",
			Migrate: func(db *gorm.DB) error {
				query := `alter table songs add listen_amount int default 0;`
				return db.Exec(query).Error
			},
			Rollback: func(db *gorm.DB) error {
				return nil
			},
		},
		{
			ID: "alter_songs_170120221139",
			Migrate: func(db *gorm.DB) error {
				query := `alter table songs add genre text;` +
				`alter table songs add duration numeric;`
				return db.Exec(query).Error
			},
			Rollback: func(db *gorm.DB) error {
				return nil
			},
		},
	}
}

/*

 */
