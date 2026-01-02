package ports

import (
	"context"
)

type EventHandler func(ctx context.Context, event interface{}) error

type EventBus interface {
	Publish(ctx context.Context, topic string, event interface{}) error
	Subscribe(topic string, handler EventHandler)
}
