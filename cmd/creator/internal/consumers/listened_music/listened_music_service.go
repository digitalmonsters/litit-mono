package listened_music

import (
	"context"
	"fmt"
	"github.com/digitalmonsters/go-common/apm_helper"
	"github.com/digitalmonsters/go-common/boilerplate"
	"github.com/gammazero/workerpool"
	"github.com/rs/zerolog/log"
	"github.com/segmentio/kafka-go"
	"go.elastic.co/apm"
	"gorm.io/gorm"
	"sync"
)

type listenCounterService struct {
	maxPoolSize int
}

func newListenCounterService(maxPoolSize int) *listenCounterService {
	return &listenCounterService{
		maxPoolSize: maxPoolSize,
	}
}

func (s *listenCounterService) Process(messages []*listenEvent, db *gorm.DB, apmTransaction *apm.Transaction,
	ctx context.Context) []kafka.Message {
	var processed []kafka.Message
	var processedMut sync.Mutex

	pool := workerpool.New(s.maxPoolSize)

	for _, lEvent := range messages {
		event := lEvent

		pool.Submit(func() {

			var span = apmTransaction.StartSpan("update_listened_music", "task", nil)
			defer span.End()

			innerCtx := apm.ContextWithSpan(boilerplate.CreateCustomContext(ctx, apmTransaction, log.Logger), span)
			apm_helper.AddSpanApmLabel(span, "content_id", fmt.Sprint(event.ContentId))

			if err := db.Exec("insert into listened_music (user_id, song_id) values (?, ?) on conflict(user_id, song_id) do nothing;",
				event.UserId,
				event.ContentId,
			).Error; err != nil {
				apm_helper.LogError(err, innerCtx)
				return
			}

			processedMut.Lock()

			processed = append(processed, event.Messages...)
			processedMut.Unlock()
		})
	}

	pool.StopWait()

	return processed
}
