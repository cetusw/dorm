package eventbus

import (
	"context"
	"errors"
	"log"
	"sync"

	"dorm/pkg/core/ports"
)

type InMemoryEventBus struct {
	handlers map[string][]ports.EventHandler
	mu       sync.RWMutex
	ctx      context.Context
	cancel   context.CancelFunc
	active   sync.WaitGroup
}

func NewInMemoryEventBus() *InMemoryEventBus {
	ctx, cancel := context.WithCancel(context.Background())
	return &InMemoryEventBus{
		ctx:      ctx,
		cancel:   cancel,
		handlers: make(map[string][]ports.EventHandler),
	}
}

func (b *InMemoryEventBus) Subscribe(topic string, handler ports.EventHandler) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.handlers[topic] = append(b.handlers[topic], handler)
}

func (b *InMemoryEventBus) Publish(ctx context.Context, topic string, event interface{}) error {
	b.mu.Lock()
	if b.ctx.Err() != nil {
		b.mu.Unlock()
		return context.Canceled
	}
	handlers := b.handlers[topic]
	b.active.Add(1)
	b.mu.Unlock()
	defer b.active.Done()

	ctx, cancel := context.WithCancel(ctx)
	stopCancellation := context.AfterFunc(b.ctx, cancel)
	defer stopCancellation()
	defer cancel()

	for _, handler := range handlers {
		if err := handler(ctx, event); err != nil {
			if b.ctx.Err() != nil && errors.Is(err, context.Canceled) {
				continue
			}
			log.Printf("ERROR in EventBus handler [topic=%s]: %v", topic, err)
		}
	}

	return nil
}

func (b *InMemoryEventBus) Stop() context.Context {
	b.mu.Lock()
	b.cancel()
	b.mu.Unlock()

	done, cancel := context.WithCancel(context.Background())
	go func() {
		b.active.Wait()
		cancel()
	}()
	return done
}
