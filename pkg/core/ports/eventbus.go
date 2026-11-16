package ports

import (
	"context"
	"time"
)

type EventHandler func(ctx context.Context, event interface{}) error

type EventBus interface {
	Publish(ctx context.Context, topic string, event interface{}) error
	PublishAsync(topic string, event interface{}, timeout time.Duration) error
	Subscribe(topic string, handler EventHandler)
}
