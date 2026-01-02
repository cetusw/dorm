package eventbus

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"dorm/pkg/core/domain/events"
)

func TestInMemoryEventBus_PublishSubscribe(t *testing.T) {
	bus := NewInMemoryEventBus()
	ctx := context.Background()

	var receivedEvent events.TaskCompletedEvent
	var receivedCount int

	bus.Subscribe(events.TopicTaskCompleted, func(ctx context.Context, e interface{}) error {
		if event, ok := e.(events.TaskCompletedEvent); ok {
			receivedEvent = event
			receivedCount++
		}
		return nil
	})

	testTaskID := uuid.New()
	eventToSend := events.TaskCompletedEvent{
		TaskID:   testTaskID,
		TaskCost: 5,
	}

	err := bus.Publish(ctx, events.TopicTaskCompleted, eventToSend)
	assert.NoError(t, err)

	assert.Equal(t, 1, receivedCount)
	assert.Equal(t, testTaskID, receivedEvent.TaskID)
	assert.Equal(t, 5, receivedEvent.TaskCost)
}

func TestInMemoryEventBus_NoSubscribers(t *testing.T) {
	bus := NewInMemoryEventBus()
	err := bus.Publish(context.Background(), "unknown.topic", "some data")
	assert.NoError(t, err)
}
