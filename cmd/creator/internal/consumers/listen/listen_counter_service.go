package listen

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

func (s *listenCounterService) Process(messages map[int64]*listenCount, db *gorm.DB, apmTransaction *apm.Transaction,
	ctx context.Context) []kafka.Message {
	var processed []kafka.Message
	var processedMut sync.Mutex

	pool := workerpool.New(s.maxPoolSize)

	for cId, lData := range messages {
		contentId := cId
		listenData := lData

		pool.Submit(func() {

			var span = apmTransaction.StartSpan("update_listens_count", "task", nil)
			defer span.End()

			innerCtx := apm.ContextWithSpan(boilerplate.CreateCustomContext(ctx, apmTransaction, log.Logger), span)
			apm_helper.AddSpanApmLabel(span, "content_id", fmt.Sprint(contentId))
			if err := db.Exec("update creator_songs set short_listens = ?, full_listens = ?  where id = ?", listenData.ShortListensCount, listenData.ListensCount, contentId).Error; err != nil {
				apm_helper.LogError(err, innerCtx)
				return
			}

			processedMut.Lock()

			processed = append(processed, listenData.Messages...)
			processedMut.Unlock()
		})
	}

	pool.StopWait()

	return processed
}
