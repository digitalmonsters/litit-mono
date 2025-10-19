package common

import (
	"context"
	"github.com/digitalmonsters/go-common/eventsourcing"
	"github.com/digitalmonsters/go-common/router"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

type MockEventHandler struct {
}

var eventsForHandling []eventsourcing.LikeEvent

func (m *MockEventHandler) Process(messages []eventsourcing.LikeEvent, executionData router.MethodExecutionData) []error {
	eventsForHandling = append(eventsForHandling, messages...)
	return nil
}
func (m *MockEventHandler) Close() []error {
	return nil
}

func TestNewBufferHandler(t *testing.T) {
	var bufferHandler = NewBufferHandler[eventsourcing.LikeEvent]("like", 1*time.Second, &MockEventHandler{}, false, context.TODO())

	bufferHandler.Enqueue(eventsourcing.LikeEvent{
		UserId:    1,
		ContentId: 2,
		Like:      false,
		CreatedAt: time.Now().Unix(),
	})
	bufferHandler.Enqueue([]eventsourcing.LikeEvent{
		{
			UserId:    1,
			ContentId: 2,
			Like:      true,
			CreatedAt: time.Now().Unix(),
		},
		{
			UserId:    2,
			ContentId: 3,
			Like:      true,
			CreatedAt: time.Now().Unix(),
		},
	}...)
	bufferHandler.Enqueue(eventsourcing.LikeEvent{
		UserId:    5,
		ContentId: 12,
		Like:      false,
		CreatedAt: time.Now().Unix(),
	})
	bufferHandler.Enqueue(eventsourcing.LikeEvent{
		UserId:    67,
		ContentId: 50,
		Like:      true,
		CreatedAt: time.Now().Unix(),
	})
	errs := bufferHandler.Flush()
	if len(errs) > 0 {
		t.Fatal(errs[0])
	}
	assert.Equal(t, 4, len(eventsForHandling))

	for _, ev := range eventsForHandling {
		if ev.UserId == 1 {
			assert.Equal(t, int64(1), ev.UserId)
			assert.Equal(t, int64(2), ev.ContentId)
			assert.Equal(t, true, ev.Like)
		} else if ev.UserId == 5 {
			assert.Equal(t, int64(5), ev.UserId)
			assert.Equal(t, int64(12), ev.ContentId)
			assert.Equal(t, false, ev.Like)
		} else if ev.UserId == 2 {
			assert.Equal(t, int64(2), ev.UserId)
			assert.Equal(t, int64(3), ev.ContentId)
			assert.Equal(t, true, ev.Like)
		} else {
			assert.Equal(t, int64(67), ev.UserId)
			assert.Equal(t, int64(50), ev.ContentId)
			assert.Equal(t, true, ev.Like)
		}
	}
}
