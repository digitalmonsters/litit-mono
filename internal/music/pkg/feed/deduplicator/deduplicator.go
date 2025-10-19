package deduplicator

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/digitalmonsters/go-common/boilerplate"
	"github.com/digitalmonsters/music/pkg/database"
	"github.com/digitalmonsters/music/utils"
	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
	"time"
)

type IDeDuplicator interface {
	SetIdsToIgnore(content []*database.CreatorSong, userId int64, expirationData []SongExpiration, ctx context.Context)
	GetIdsToIgnore(userId int64, ctx context.Context) ([]SongExpiration, []int64)
}

type DeDuplicator struct {
	redisClient *redis.Client
}

func NewDeDuplicator(redisClient *redis.Client) IDeDuplicator {
	return &DeDuplicator{redisClient: redisClient}
}

func (v *DeDuplicator) SetIdsToIgnore(songs []*database.CreatorSong, userId int64, expirationData []SongExpiration, ctx context.Context) {
	if len(songs) == 0 || userId == 0 {
		return
	}

	exp := time.Now().UTC().Add(defaultTTL).Unix()
	for _, s := range songs {
		expirationData = append(expirationData, SongExpiration{SongId: s.Id, ExpiresAt: exp})
	}

	if len(expirationData) == 0 {
		return
	}

	if d, err := json.Marshal(expirationData); err != nil {
		utils.CaptureApmErrorFromTransaction(err, ctx)
	} else {
		if c := v.redisClient.Set(ctx, v.buildKey(userId), string(d),
			defaultTTL+(1*time.Minute)); c.Err() != nil {
			utils.CaptureApmErrorFromTransaction(c.Err(), ctx)
		}
	}
}

func (v *DeDuplicator) GetIdsToIgnore(userId int64, ctx context.Context) ([]SongExpiration, []int64) {
	var newExpirationData []SongExpiration
	var redisKey string
	var idsToIgnore []int64

	if userId <= 0 {
		return newExpirationData, idsToIgnore
	}

	var expirationData []SongExpiration
	redisKey = v.buildKey(userId)

	if cmd := v.redisClient.Get(ctx, redisKey); cmd.Err() != nil && !errors.Is(redis.Nil, cmd.Err()) {
		utils.CaptureApmErrorFromTransaction(cmd.Err(), ctx)
	} else {
		if b, _ := cmd.Bytes(); len(b) > 0 {
			if err := json.Unmarshal(b, &expirationData); err != nil {
				utils.CaptureApmErrorFromTransaction(err, ctx)
			}
		}
	}

	now := time.Now().UTC().Unix()

	for _, ex := range expirationData {
		if now < ex.ExpiresAt {
			idsToIgnore = append(idsToIgnore, ex.SongId)
			newExpirationData = append(newExpirationData, ex)
		}
	}

	if len(idsToIgnore) == 0 {
		idsToIgnore = append(idsToIgnore, 0)
	}

	return newExpirationData, idsToIgnore
}

func (v *DeDuplicator) buildKey(userId int64) string {
	return fmt.Sprintf("%v:lastdata_%v", boilerplate.GetCurrentEnvironment().ToString(), userId)
}
