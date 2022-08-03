package feed

import (
	"github.com/RichardKnop/machinery/v1"
	"github.com/digitalmonsters/go-common/application"
	"github.com/digitalmonsters/go-common/error_codes"
	"github.com/digitalmonsters/go-common/router"
	"github.com/digitalmonsters/music/configs"
	"github.com/digitalmonsters/music/pkg/database"
	"github.com/digitalmonsters/music/pkg/feed/deduplicator"
	"github.com/digitalmonsters/music/pkg/feed/feed_converter"
	"gorm.io/gorm"
)

type Feed struct {
	deDuplicator  deduplicator.IDeDuplicator
	feedConverter *feed_converter.Service
	feedBuilder   *feedBuilder
}

func NewFeed(
	deduplicator deduplicator.IDeDuplicator,
	feedConverter *feed_converter.Service,
	machineryServer *machinery.Server,
	appConfig *application.Configurator[configs.AppConfig],
) *Feed {
	builder := newMusicFeedBuilder(machineryServer, database.GetDb(database.DbTypeMaster), appConfig)

	return &Feed{
		deDuplicator:  deduplicator,
		feedConverter: feedConverter,
		feedBuilder:   builder,
	}
}

func (f *Feed) GetFeed(db *gorm.DB, userId int64, count int, executionData router.MethodExecutionData) (*ContentFeedResponse, *error_codes.ErrorWithCode) {
	expirationData, idsToIgnore := f.deDuplicator.GetIdsToIgnore(userId, executionData.Context)

	var songs []*database.CreatorSong
	query := db.Model(songs).
		Where("short_song_url is not null").
		Where("full_song_url is not null").
		Where("reject_reason is null")

	query = query.Where("creator_songs.id not in (select song_id from listened_music "+
		" where listened_music.user_id = ?)", userId)

	if len(idsToIgnore) > 0 {
		query = query.Where("creator_songs.id not in ?", idsToIgnore)
	}

	query = query.Order("score desc")

	if err := query.Limit(count).Find(&songs).Error; err != nil {
		return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
	}

	go func() {
		f.deDuplicator.SetIdsToIgnore(songs, userId, expirationData, executionData.Context)
	}()

	return &ContentFeedResponse{
		Data:     f.feedConverter.ConvertToSongModel(songs, executionData.UserId, false, executionData.ApmTransaction, executionData.Context),
		FeedType: "music",
	}, nil
}
