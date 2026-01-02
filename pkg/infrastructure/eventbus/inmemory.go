package eventbus

import (
	"context"
	"log"
	"sync"

	"dorm/pkg/core/ports"
)

type InMemoryEventBus struct {
	handlers map[string][]ports.EventHandler
	mu       sync.RWMutex
}

func NewInMemoryEventBus() *InMemoryEventBus {
	return &InMemoryEventBus{
		handlers: make(map[string][]ports.EventHandler),
	}
}

func (b *InMemoryEventBus) Subscribe(topic string, handler ports.EventHandler) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.handlers[topic] = append(b.handlers[topic], handler)
}

func (b *InMemoryEventBus) Publish(ctx context.Context, topic string, event interface{}) error {
	b.mu.RLock()
	handlers, ok := b.handlers[topic]
	b.mu.RUnlock()

	if !ok {
		return nil
	}

	for _, handler := range handlers {
		if err := handler(ctx, event); err != nil {
			log.Printf("ERROR in EventBus handler [topic=%s]: %v", topic, err)
		}
	}

	return nil
}
