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
	"github.com/digitalmonsters/music/utils"
	"github.com/pkg/errors"
	"github.com/samber/lo"
	"gorm.io/gorm"
)

type Feed struct {
	deDuplicator  deduplicator.IDeDuplicator
	feedConverter *feed_converter.Service
	feedBuilder   *feedBuilder
	appConfig     *application.Configurator[configs.AppConfig]
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
		appConfig:     appConfig,
	}
}

func (f *Feed) GetFeed(db *gorm.DB, userId int64, startContentsIds []int64, count int, executionData router.MethodExecutionData) (*ContentFeedResponse, *error_codes.ErrorWithCode) {
	var expirationData []deduplicator.SongExpiration
	var idsToIgnore []int64
	if f.appConfig.Values.MUSIC_FEATURE_FEED_IGNORE_IDS_ENABLED {
		expirationData, idsToIgnore = f.deDuplicator.GetIdsToIgnore(userId, executionData.Context)
	}

	var startContents []*database.CreatorSong
	var songs []*database.CreatorSong
	lenStartContentsIds := len(startContentsIds)

	if lenStartContentsIds > 0 {
		for _, id := range startContentsIds {
			if !lo.Contains(idsToIgnore, id) {
				idsToIgnore = append(idsToIgnore, id)
			}
		}

		idsToIgnore = append(idsToIgnore, startContentsIds...)

		if err := db.Where("id in ? and reject_reason is null", startContentsIds).Limit(count).Find(&startContents).Error; err != nil {
			utils.CaptureApmErrorFromTransaction(errors.WithStack(err), executionData.Context)
		}

		count = count - len(startContents)
	}

	if count != 0 {
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

		if f.appConfig.Values.MUSIC_FEATURE_FEED_IGNORE_IDS_ENABLED {
			go func() {
				f.deDuplicator.SetIdsToIgnore(songs, userId, expirationData, executionData.Context)
			}()
		}
	}

	songs = append(startContents, songs...)

	convertedSongs := f.feedConverter.ConvertToSongModel(songs, executionData.UserId, false, executionData.ApmTransaction, executionData.Context)

	finalRespItems := []MusicFeedItem{}
	for _, s := range convertedSongs {
		finalRespItems = append(finalRespItems, MusicFeedItem{
			Type: "music",
			Data: s,
		})
	}

	return &ContentFeedResponse{
		Data:     finalRespItems,
		FeedType: "music",
	}, nil
}
