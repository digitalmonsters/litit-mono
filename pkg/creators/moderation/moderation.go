package moderation

import (
	"context"
	"fmt"
	"github.com/digitalmonsters/go-common/error_codes"
	"github.com/digitalmonsters/go-common/wrappers/user_go"
	"github.com/digitalmonsters/music/pkg/database"
	"github.com/pkg/errors"
	"gopkg.in/guregu/null.v4"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func RejectMusic(req RejectMusicRequest, db *gorm.DB) error {
	if req.SongId == 0 {
		return errors.New("wrong song_id")
	}

	if req.RejectReason == 0 {
		return errors.New("reject_reason is required")
	}

	tx := db.Begin()
	defer tx.Rollback()

	var song database.CreatorSong
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&song, req.SongId).Error; err != nil {
		return errors.WithStack(err)
	}

	updateMap := map[string]interface{}{
		"status":        database.CreatorSongStatusRejected,
		"reject_reason": null.IntFrom(req.RejectReason),
	}

	if err := tx.Model(&song).Updates(updateMap).Error; err != nil {
		return errors.WithStack(err)
	}

	if err := tx.Commit().Error; err != nil {
		return errors.WithStack(err)
	}

	/*
		if contentNotifier != nil {
			contentNotifier.Enqueue(content, eventsourcing.ChangeEventTypeUpdated, "rejected")
		}
	*/

	return nil
}

func ApproveMusic(req ApproveMusicRequest, db *gorm.DB) error {
	tx := db.Begin()
	defer tx.Rollback()

	var song database.CreatorSong
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&song, req.SongId).Error; err != nil {
		return errors.WithStack(err)
	}

	song.Status = database.CreatorSongStatusApproved
	song.RejectReason = null.Int{}

	if err := tx.Save(&song).Error; err != nil {
		return errors.WithStack(err)
	}

	if err := tx.Commit().Error; err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func List(req ListRequest, db *gorm.DB, userGoWrapper user_go.IUserGoWrapper, ctx context.Context) (*ListResponse, *error_codes.ErrorWithCode) {
	var records []database.CreatorSong
	query := db.Model(records).Preload("Category").Preload("Mood")

	if req.Keyword.Valid {
		query = query.Where("name ilike ?", fmt.Sprintf("%%%v%%", req.Keyword.String)).
			Or("lyric_author ilike ?", fmt.Sprintf("%%%v%%", req.Keyword.String)).
			Or("music_author ilike ?", fmt.Sprintf("%%%v%%", req.Keyword.String))
	}

	if req.UserId.Valid {
		query = query.Where("user_id = ?", req.UserId.Int64)
	}

	if req.CategoryId.Valid {
		query = query.Where("category_id = ?", req.CategoryId.Int64)
	}

	if req.MoodId.Valid {
		query = query.Where("mood_id = ?", req.CategoryId.Int64)
	}

	if len(req.Status) > 0 {
		query = query.Where("status in ?", req.Status)
	}

	var totalCount int64
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
	}

	if err := query.Limit(req.Limit).Offset(req.Offset).Order("id desc").Find(&records).Error; err != nil {
		return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
	}

	var userIds []int64
	for _, song := range records {
		userIds = append(userIds, song.UserId)
	}

	userResp := <-userGoWrapper.GetUsers(userIds, ctx, false)
	if userResp.Error != nil {
		return nil, error_codes.NewErrorWithCodeRef(userResp.Error.ToError(), error_codes.GenericServerError)
	}

	var listItems []listItem
	for _, song := range records {
		item := listItem{
			SongId:            song.Id,
			SongName:          song.Name,
			Status:            song.Status,
			LyricAuthor:       song.LyricAuthor,
			MusicAuthor:       song.MusicAuthor,
			CategoryId:        song.CategoryId,
			MoodId:            song.MoodId,
			FullSongUrl:       song.FullSongUrl,
			FullSongDuration:  song.ShortSongDuration,
			ShortSongUrl:      song.ShortSongUrl,
			ShortSongDuration: song.ShortSongDuration,
			ImageUrl:          song.ImageUrl,
			UserId:            song.UserId,
		}

		if song.Category != nil {
			item.Category = song.Category.Name
		}

		if song.Mood != nil {
			item.Mood = song.Mood.Name
		}

		if user, ok := userResp.Response[song.UserId]; ok {
			item.Username = user.Username
		}

		listItems = append(listItems, item)
	}

	return &ListResponse{
		Items:      listItems,
		TotalCount: totalCount,
	}, nil
}
