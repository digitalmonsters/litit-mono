package shares

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

type sharesCounterService struct {
	maxPoolSize int
}

func newListenCounterService(maxPoolSize int) *sharesCounterService {
	return &sharesCounterService{
		maxPoolSize: maxPoolSize,
	}
}

func (s *sharesCounterService) Process(messages map[int64]*sharesCount, db *gorm.DB, apmTransaction *apm.Transaction,
	ctx context.Context) []kafka.Message {
	var processed []kafka.Message
	var processedMut sync.Mutex

	pool := workerpool.New(s.maxPoolSize)

	for cId, c := range messages {
		contentId := cId
		contentData := c

		pool.Submit(func() {

			var span = apmTransaction.StartSpan("update_shares_count", "task", nil)
			defer span.End()

			innerCtx := apm.ContextWithSpan(boilerplate.CreateCustomContext(ctx, apmTransaction, log.Logger), span)
			apm_helper.AddSpanApmLabel(span, "content_id", fmt.Sprint(contentId))
			if err := db.Exec("update creator_songs set shares = ?  where id = ?", contentData.SharesCount, contentId).Error; err != nil {
				apm_helper.LogError(err, innerCtx)
				return
			}

			processedMut.Lock()

			processed = append(processed, contentData.Messages...)
			processedMut.Unlock()
		})
	}

	pool.StopWait()

	return processed
}
