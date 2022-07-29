package feed

import (
	"context"
	"fmt"
	"github.com/RichardKnop/machinery/v1"
	"github.com/RichardKnop/machinery/v1/backends/result"
	"github.com/RichardKnop/machinery/v1/tasks"
	"github.com/digitalmonsters/go-common/apm_helper"
	"github.com/digitalmonsters/go-common/application"
	"github.com/digitalmonsters/go-common/boilerplate"
	"github.com/digitalmonsters/music/configs"
	"github.com/digitalmonsters/music/pkg/database"
	"github.com/digitalmonsters/music/utils"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
	"gorm.io/gorm"
	"math"
	time2 "time"
)

type feedBuilder struct {
	machineryServer *machinery.Server
	appConfig       *application.Configurator[configs.AppConfig]
	db              *gorm.DB
}

func newMusicFeedBuilder(
	machineryServer *machinery.Server,
	db *gorm.DB,
	appConfig *application.Configurator[configs.AppConfig],
) *feedBuilder {
	builder := &feedBuilder{
		machineryServer: machineryServer,
		db:              db,
		appConfig:       appConfig,
	}

	if boilerplate.GetCurrentEnvironment() != boilerplate.Ci {
		if err := builder.registerTask(); err != nil {
			log.Fatal().Err(err).Send()
		}
	}

	return builder
}

//nolint
func (b *feedBuilder) updateMusicFeed(db *gorm.DB, ctx context.Context) error {
	musicRecords, err := b.findCreatorSongs(db.WithContext(ctx))
	if err != nil {
		return err
	}

	groups := lo.GroupBy[database.CreatorSong, int](musicRecords, func(t database.CreatorSong) int {
		return t.Score
	})

	for score, contents := range groups {
		ids := lo.Map(contents, func(t database.CreatorSong, i int) int64 {
			return t.Id
		})

		if err = db.Exec("update creator_songs set score = ? where id in ?;", score, ids).Error; err != nil {
			apm_helper.LogError(err, ctx)
		}
	}

	return nil
}

//nolint
func (b *feedBuilder) findCreatorSongs(tx *gorm.DB) ([]database.CreatorSong, error) {
	var records []database.CreatorSong

	var finalRecords []database.CreatorSong

	if err := tx.Where("status != ?", database.CreatorSongStatusRejected).
		Where("reject_reason is null").
		Limit(b.appConfig.Values.MUSIC_FEED_LIMIT).
		Order("id desc").
		Find(&records).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	if b.appConfig == nil {
		return records, nil
	}

	for _, r := range records {
		score := (r.Loves * b.appConfig.Values.MUSIC_CALCULATION_LOVE_COUNT_WEIGHT) +
			(r.Likes * b.appConfig.Values.MUSIC_CALCULATION_LIKE_COUNT_WEIGHT) +
			(r.ShortListens * b.appConfig.Values.MUSIC_CALCULATION_SHORT_LISTEN_COUNT_WEIGHT) -
			(r.Dislikes * b.appConfig.Values.MUSIC_CALCULATION_DISLIKE_COUNT_WEIGHT)

		timeNow := time2.Now().UTC()

		if b.appConfig.Values.MUSIC_CALCULATION_TIMING_START_CONF > 0 && b.appConfig.Values.MUSIC_CALCULATION_TIMING_DELIMITER > 0 {
			val := int(math.Round(float64(timeNow.Unix()-r.CreatedAt.Unix()) / float64(b.appConfig.Values.MUSIC_CALCULATION_TIMING_DELIMITER)))

			if val > 0 {
				score += b.appConfig.Values.MUSIC_CALCULATION_TIMING_START_CONF / val
			}
		}

		r.Score = score

		finalRecords = append(finalRecords, r)
	}

	return finalRecords, nil
}

const taskName = "music:feed:generate"

//nolint
func (b *feedBuilder) registerTask() error {
	if err := b.machineryServer.RegisterTask(taskName, func() error {
		var apmTransaction = apm_helper.StartNewApmTransaction(taskName, "task", nil, nil)
		defer apmTransaction.End()
		ctx := boilerplate.CreateCustomContext(context.Background(), apmTransaction, log.Logger)

		if err := b.updateMusicFeed(b.db, ctx); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return err
	}

	every := fmt.Sprintf("@every %vm", b.appConfig.Values.MUSIC_FEED_UPDATE_SCORE_FREQUENCY_MINUTES)

	return b.machineryServer.RegisterPeriodicTask(every, taskName, &tasks.Signature{
		Name: taskName,
	})
}

func (b *feedBuilder) LaunchTask() (*result.AsyncResult, error) {
	res, err := utils.SendTask(b.machineryServer, taskName, []tasks.Arg{}, true)

	if err != nil {
		log.Error().Str("service", "music feed").Err(errors.WithStack(err))
	}

	return res, err
}
