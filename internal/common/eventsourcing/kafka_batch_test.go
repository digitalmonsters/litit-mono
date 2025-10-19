package eventsourcing

import (
	"context"
	"fmt"
	"github.com/digitalmonsters/go-common/boilerplate"
	"github.com/pkg/errors"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestTimeBasedPublishSuccess(t *testing.T) {
	flushEveryMs := int(100 * time.Millisecond)
	writer := &mockWriter{}

	tt := NewKafkaBatchPublisher[UserEvent]("one", boilerplate.KafkaBatchWriterV2Configuration{
		FlushTimeMilliseconds: flushEveryMs,
	}, context.WithValue(context.TODO(), ciRun{}, writer))

	mapped := tt.(*KafkaEventPublisherV2[UserEvent])
	mapped.writer = writer

	called := false

	writer.WriteFn = func(ctx context.Context, msgs ...kafka.Message) error {
		called = true
		assert.Equal(t, 3, len(msgs))

		return nil
	}

	if err := <-mapped.Publish(context.TODO(), UserEvent{UserId: 1}, UserEvent{UserId: 2}, UserEvent{UserId: 3}); err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 3, len(mapped.queue))

	time.Sleep(time.Duration(flushEveryMs) * 2)

	assert.Equal(t, true, called)
	assert.Equal(t, 0, len(mapped.queue))
}

func TestImmediatePublishSuccess(t *testing.T) {
	flushEveryMs := int(100 * time.Millisecond)
	writer := &mockWriter{}

	tt := NewKafkaBatchPublisher[UserEvent]("two", boilerplate.KafkaBatchWriterV2Configuration{
		FlushTimeMilliseconds: flushEveryMs,
	}, context.WithValue(context.TODO(), ciRun{}, writer))

	mapped := tt.(*KafkaEventPublisherV2[UserEvent])
	mapped.writer = writer

	called := false

	writer.WriteFn = func(ctx context.Context, msgs ...kafka.Message) error {
		called = true
		assert.Equal(t, 3, len(msgs))

		return nil
	}

	if err := <-mapped.PublishImmediate(context.TODO(), UserEvent{UserId: 1}, UserEvent{UserId: 2}, UserEvent{UserId: 3}); err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 0, len(mapped.queue))
	assert.Equal(t, true, called)
	assert.Equal(t, 0, len(mapped.queue))
}

func TestPublishImmediateFailed(t *testing.T) {
	flushEveryMs := int(100 * time.Millisecond)
	retryCount := 2

	writer := &mockWriter{}

	tt := NewKafkaBatchPublisher[UserEvent]("three", boilerplate.KafkaBatchWriterV2Configuration{
		FlushTimeMilliseconds: flushEveryMs,
		MaxRetryCount:         retryCount,
	}, context.WithValue(context.TODO(), ciRun{}, writer))

	mapped := tt.(*KafkaEventPublisherV2[UserEvent])
	mapped.writer = writer

	called := false

	expectedMessageCount := 3

	writer.WriteFn = func(ctx context.Context, msgs ...kafka.Message) error {
		called = true
		assert.Equal(t, expectedMessageCount, len(msgs))

		return errors.New("expected error")
	}

	err := <-mapped.PublishImmediate(context.TODO(), UserEvent{UserId: 1}, UserEvent{UserId: 2}, UserEvent{UserId: 3})

	if err == nil {
		t.Fatal(errors.New("error is expected"))
	}

	assert.Equal(t, 3, len(mapped.queue))
	assert.Equal(t, true, called)

	for _, k := range mapped.queue {
		assert.Equal(t, 1, k.currentRetry)
		assert.Equal(t, k.maxRetry, 2)
	}

	time.Sleep(time.Duration(flushEveryMs) * 2)

	for _, k := range mapped.queue {
		assert.Equal(t, 2, k.currentRetry)
		assert.Equal(t, k.maxRetry, 2)
	}

	called = false
	expectedMessageCount = 4

	lastEvent := UserEvent{UserId: 4}
	<-mapped.PublishImmediate(context.TODO(), lastEvent)

	time.Sleep(time.Duration(flushEveryMs) * 2)

	assert.Equal(t, 1, len(mapped.queue)) // previous should be dropped
	assert.Equal(t, 1, mapped.queue[0].currentRetry)
	assert.Equal(t, []byte(lastEvent.GetPublishKey()), mapped.queue[0].message.Key)
	assert.Equal(t, true, called)

	expectedMessageCount = 1
	time.Sleep(2 * time.Second)
	called = false
}

func TestBatchPublishAfterImmediate(t *testing.T) {
	flushEveryMs := int(100 * time.Millisecond)
	writer := &mockWriter{}

	tt := NewKafkaBatchPublisher[UserEvent]("four", boilerplate.KafkaBatchWriterV2Configuration{
		FlushTimeMilliseconds: flushEveryMs,
	}, context.WithValue(context.TODO(), ciRun{}, writer))

	mapped := tt.(*KafkaEventPublisherV2[UserEvent])
	mapped.writer = writer

	called := false

	writer.WriteFn = func(ctx context.Context, msgs ...kafka.Message) error {
		called = true
		assert.Equal(t, 4, len(msgs))

		assert.Equal(t, fmt.Sprint(6), string(msgs[0].Key))
		assert.Equal(t, fmt.Sprint(8), string(msgs[1].Key))
		assert.Equal(t, fmt.Sprint(9), string(msgs[2].Key))
		assert.Equal(t, fmt.Sprint(6), string(msgs[3].Key))

		return nil
	}

	if err := <-mapped.Publish(context.TODO(), UserEvent{UserId: 6}, UserEvent{UserId: 8}, UserEvent{UserId: 9}); err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 3, len(mapped.queue))

	if err := <-mapped.PublishImmediate(context.TODO(), UserEvent{UserId: 6}); err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 0, len(mapped.queue))
	assert.Equal(t, true, called)
}

func TestClose(t *testing.T) {
	flushEveryMs := int(10000 * time.Millisecond)
	writer := &mockWriter{}

	tt := NewKafkaBatchPublisher[UserEvent]("five", boilerplate.KafkaBatchWriterV2Configuration{
		FlushTimeMilliseconds: flushEveryMs,
	}, context.WithValue(context.TODO(), ciRun{}, writer))

	mapped := tt.(*KafkaEventPublisherV2[UserEvent])
	mapped.writer = writer

	called := false

	writer.WriteFn = func(ctx context.Context, msgs ...kafka.Message) error {
		called = true
		assert.Equal(t, 3, len(msgs))

		return nil
	}

	if err := <-mapped.Publish(context.TODO(), UserEvent{UserId: 1}, UserEvent{UserId: 2}, UserEvent{UserId: 3}); err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 3, len(mapped.queue))

	if err := tt.Close(); err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, true, called)
	assert.Equal(t, 0, len(mapped.queue))
}

func TestBackOff(t *testing.T) {
	flushEveryMs := int(10000 * time.Millisecond)
	writer := &mockWriter{}
	maxRetryCount := 5

	tt := NewKafkaBatchPublisher[UserEvent]("six", boilerplate.KafkaBatchWriterV2Configuration{
		FlushTimeMilliseconds: flushEveryMs,
		MaxRetryCount:         maxRetryCount,
	}, context.WithValue(context.TODO(), ciRun{}, writer))

	mapped := tt.(*KafkaEventPublisherV2[UserEvent])
	mapped.writer = writer

	calledCount := 0

	expectedError := errors.New("expected error")

	writer.WriteFn = func(ctx context.Context, msgs ...kafka.Message) error {
		calledCount += 1
		assert.Equal(t, 3, len(msgs))

		return errors.WithStack(expectedError)
	}

	if err := <-mapped.Publish(context.TODO(), UserEvent{UserId: 1}, UserEvent{UserId: 2}, UserEvent{UserId: 3}); err != nil {
		t.Fatal(err)
	}

	err := tt.Close()

	if err == nil {
		t.Fatal("error is expected")
	}

	assert.True(t, errors.Is(err, expectedError))
	assert.Equal(t, maxRetryCount*2, calledCount) // manually set due to backoff
	assert.Equal(t, 3, len(mapped.queue))
}
