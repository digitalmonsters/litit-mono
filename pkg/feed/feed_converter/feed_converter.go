package feed_converter

import (
	"context"
	"fmt"
	"github.com/digitalmonsters/go-common/apm_helper"
	frontend2 "github.com/digitalmonsters/go-common/frontend"
	"github.com/digitalmonsters/go-common/wrappers/follow"
	"github.com/digitalmonsters/go-common/wrappers/go_tokenomics"
	"github.com/digitalmonsters/go-common/wrappers/like"
	"github.com/digitalmonsters/go-common/wrappers/user_go"
	"github.com/digitalmonsters/music/pkg/database"
	"github.com/digitalmonsters/music/pkg/frontend"
	"github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/thoas/go-funk"
	"go.elastic.co/apm"
	"gopkg.in/guregu/null.v4"
	"gorm.io/gorm"
	"sync"
	"time"
)

type Service struct {
	categoryCache       *cache.Cache
	moodCache           *cache.Cache
	rejectReasonCache   *cache.Cache
	ctx                 context.Context
	db                  *gorm.DB
	userWrapper         user_go.IUserGoWrapper
	followWrapper       follow.IFollowWrapper
	likeWrapper         like.ILikeWrapper
	goTokenomicsWrapper go_tokenomics.IGoTokenomicsWrapper
}

func NewFeedConverter(
	userWrapper user_go.IUserGoWrapper,
	followWrapper follow.IFollowWrapper,
	likeWrapper like.ILikeWrapper,
	goTokenomicsWrapper go_tokenomics.IGoTokenomicsWrapper,
	ctx context.Context,
) *Service {
	s := &Service{
		db:                  database.GetDb(database.DbTypeReadonly),
		userWrapper:         userWrapper,
		followWrapper:       followWrapper,
		likeWrapper:         likeWrapper,
		goTokenomicsWrapper: goTokenomicsWrapper,
		categoryCache:       cache.New(10*time.Minute, 12*time.Minute),
		moodCache:           cache.New(10*time.Minute, 12*time.Minute),
		rejectReasonCache:   cache.New(10*time.Minute, 12*time.Minute),
		ctx:                 ctx,
	}

	s.runJobs()

	return s
}

func (s *Service) runJobs() {
	go func() {
		sleepTime := 5 * time.Minute

		for s.ctx.Err() == nil {
			if err := s.updateCache(); err != nil {
				log.Error().Str("service", "feed").Err(errors.WithStack(err)).Send()
				time.Sleep(sleepTime)
				continue
			}

			time.Sleep(sleepTime)
			continue
		}
	}()
}

func (s *Service) updateCache() error {
	var categories []database.Category
	var moods []database.Mood
	var rejectReasons []database.CreatorRejectReasons

	var wg sync.WaitGroup
	wg.Add(3)

	go func() {
		if err := s.db.Find(&categories).Error; err != nil {
			apm_helper.LogError(err, s.ctx)
		}
		wg.Done()
	}()

	go func() {
		if err := s.db.Find(&moods).Error; err != nil {
			apm_helper.LogError(err, s.ctx)
		}
		wg.Done()
	}()

	go func() {
		if err := s.db.Find(&rejectReasons).Error; err != nil {
			apm_helper.LogError(err, s.ctx)
		}
		wg.Done()
	}()

	wg.Wait()

	for _, c := range categories {
		s.categoryCache.Set(fmt.Sprint(c.Id), c, cache.DefaultExpiration)
	}

	for _, m := range moods {
		s.moodCache.Set(fmt.Sprint(m.Id), m, cache.DefaultExpiration)
	}

	for _, r := range rejectReasons {
		s.rejectReasonCache.Set(fmt.Sprint(r.Id), r, cache.DefaultExpiration)
	}

	return nil
}

func (s *Service) getCategory(categoryId int64) *frontend.Category {
	if v, ok := s.categoryCache.Get(fmt.Sprint(categoryId)); ok {
		cat := v.(database.Category)
		return &frontend.Category{
			Id:   cat.Id,
			Name: cat.Name,
		}
	}

	return nil
}

func (s *Service) getMood(moodId int64) *frontend.Mood {
	if v, ok := s.moodCache.Get(fmt.Sprint(moodId)); ok {
		mood := v.(database.Mood)
		return &frontend.Mood{
			Id:   mood.Id,
			Name: mood.Name,
		}
	}

	return nil
}

func (s *Service) getRejectReason(rejectReasonId int64) *frontend.RejectReason {
	if v, ok := s.rejectReasonCache.Get(fmt.Sprint(rejectReasonId)); ok {
		reason := v.(database.CreatorRejectReasons)
		return &frontend.RejectReason{
			Id:     reason.Id,
			Reason: reason.Reason,
		}
	}

	return nil
}

func (s *Service) ConvertToSongModel(songs []*database.CreatorSong, currentUserId int64, withPrivateInfo bool, apmTransaction *apm.Transaction, ctx context.Context) []frontend.CreatorSongModel {
	if len(songs) == 0 {
		return []frontend.CreatorSongModel{}
	}

	var authorIds, songIds []int64

	feedSongsMap := make(map[int64]*frontend.CreatorSongModel)
	feedSongsArr := make([]*frontend.CreatorSongModel, 0)

	for _, song := range songs {
		if !funk.ContainsInt64(authorIds, song.UserId) {
			authorIds = append(authorIds, song.UserId)
		}

		if !funk.ContainsInt64(songIds, song.Id) {
			songIds = append(songIds, song.Id)
		}

		model := &frontend.CreatorSongModel{
			Id:                song.Id,
			UserId:            song.UserId,
			Name:              song.Name,
			Status:            song.Status,
			LyricAuthor:       song.LyricAuthor,
			MusicAuthor:       song.MusicAuthor,
			CategoryId:        song.CategoryId,
			MoodId:            song.MoodId,
			FullSongUrl:       song.FullSongUrl,
			FullSongDuration:  song.FullSongDuration,
			ShortSongUrl:      song.ShortSongUrl,
			ShortSongDuration: song.ShortSongDuration,
			ImageUrl:          song.ImageUrl,
			Hashtags:          song.Hashtags,
			Shares:            song.Shares,
			ShortListens:      song.ShortListens,
			FullListens:       song.FullListens,
			Likes:             song.Likes,
			Dislikes:          song.Dislikes,
			Loves:             song.Loves,
			Comments:          song.Comments,
			UsedInVideo:       song.UsedInVideo,
			CreatedAt:         song.CreatedAt,
			CreatedAtTs:       song.CreatedAt.Unix(),
		}

		if model.CategoryId > 0 {
			model.Category = s.getCategory(model.CategoryId)
		}

		if model.MoodId > 0 {
			model.Mood = s.getMood(model.MoodId)
		}

		if song.RejectReason.Valid {
			model.RejectReason = null.StringFrom(s.getRejectReason(song.RejectReason.Int64).Reason)
		}

		feedSongsMap[song.Id] = model
		feedSongsArr = append(feedSongsArr, model)
	}

	routines := []chan error{
		s.fillFollowingData(feedSongsMap, currentUserId, authorIds, apmTransaction),
		s.fillUsersAndApplyUserPrivacySettings(feedSongsMap, ctx),
		s.fillMyReactions(feedSongsMap, songIds, currentUserId, apmTransaction),
	}

	if withPrivateInfo {
		routines = append(routines, s.fillPointsCount(feedSongsMap, apmTransaction))
	}

	for _, c := range routines {
		if err := <-c; err != nil {
			apm_helper.LogError(err, ctx)
		}
	}

	finalResp := make([]frontend.CreatorSongModel, 0)
	for _, v := range feedSongsArr {
		if v.User.Id == 0 {
			continue
		}

		finalResp = append(finalResp, *v)
	}

	return finalResp
}

func (s *Service) fillFollowingData(contentModels map[int64]*frontend.CreatorSongModel, currentUserId int64,
	creatorIds []int64, apmTransaction *apm.Transaction) chan error {
	ch := make(chan error, 2)

	if currentUserId == 0 || len(creatorIds) == 0 {
		close(ch)

		return ch
	}

	go func() {
		defer func() {
			close(ch)
		}()

		data := <-s.followWrapper.GetUserFollowingRelationBulk(currentUserId, creatorIds, apmTransaction, false)
		if data.Error != nil {
			ch <- data.Error.ToError()
			return
		}

		for _, c := range contentModels {
			if v, ok := data.Data[c.UserId]; ok {
				c.IsCreatorFollowing = v.IsFollower
				c.IsFollowing = v.IsFollowing
			}
		}

		ch <- nil
	}()

	return ch
}

func (s *Service) fillUsersAndApplyUserPrivacySettings(
	contentModels map[int64]*frontend.CreatorSongModel,
	ctx context.Context,
) chan error {
	ch := make(chan error, 2)

	if len(contentModels) == 0 {
		close(ch)

		return ch
	}

	go func() {
		defer func() {
			close(ch)
		}()

		var userIds []int64

		for _, c := range contentModels {
			if c.UserId <= 0 {
				continue
			}

			hasUserId := false

			for _, userId := range userIds {
				if userId == c.UserId {
					hasUserId = true
					break
				}
			}

			if hasUserId {
				continue
			}

			userIds = append(userIds, c.UserId)
		}

		resp := <-s.userWrapper.GetUsers(userIds, ctx, false)

		if resp.Error != nil {
			ch <- errors.Wrap(errors.New(resp.Error.Message), "fill users failed")
		}

		for _, c := range contentModels {
			userResp, hasUser := resp.Response[c.UserId]

			if !hasUser {
				continue
			}

			c.User = frontend2.VideoUserModel{
				Avatar:            userResp.Avatar.ValueOrZero(),
				FirstName:         userResp.Firstname,
				Id:                userResp.UserId,
				LastName:          userResp.Lastname,
				UserName:          userResp.Username,
				Verified:          userResp.Verified,
				IsTipEnabled:      userResp.IsTipEnabled,
				NamePrivacyStatus: userResp.NamePrivacyStatus,
			}
		}
	}()

	return ch
}

func (s *Service) fillMyReactions(contentModels map[int64]*frontend.CreatorSongModel, contentIds []int64, currentUserId int64,
	apmTransaction *apm.Transaction) chan error {
	ch := make(chan error, 2)

	if len(contentModels) == 0 || currentUserId == 0 || len(contentIds) == 0 {
		close(ch)

		return ch
	}

	go func() {
		defer func() {
			close(ch)
		}()

		reactionsData := <-s.likeWrapper.GetInternalSpotReactionsByUser(contentIds, currentUserId, apmTransaction, false)
		if reactionsData.Error != nil {
			ch <- reactionsData.Error.ToError()
			return
		}

		for contentId, reaction := range reactionsData.Data {
			if v, ok := contentModels[contentId]; ok {
				v.LikedByMe = reaction.Like
				v.DislikedByMe = reaction.Dislike
				v.LovedByMe = reaction.Love
			}
		}

		ch <- nil
	}()

	return ch
}

func (s *Service) fillPointsCount(contentModels map[int64]*frontend.CreatorSongModel, apmTransaction *apm.Transaction) chan error {
	ch := make(chan error, 2)

	go func() {
		defer func() {
			close(ch)
		}()

		if len(contentModels) == 0 {
			return
		}

		var contentIds []int64

		for _, content := range contentModels {
			if content.Id <= 0 {
				continue
			}

			hasContentId := false

			for _, contentId := range contentIds {
				if contentId == content.Id {
					hasContentId = true
					break
				}
			}

			if hasContentId {
				continue
			}

			contentIds = append(contentIds, content.Id)
		}

		resp := <-s.goTokenomicsWrapper.GetContentEarningsTotalByContentIds(contentIds, apmTransaction, false)

		if resp.Error != nil {
			ch <- errors.Wrap(errors.New(resp.Error.Message), "fill points count failed")
		}

		for _, content := range contentModels {
			pointsCountResp, hasPointsCount := resp.Items[content.Id]

			if !hasPointsCount {
				continue
			}

			content.PointsEarned, _ = pointsCountResp.Float64()
		}
	}()

	return ch
}
