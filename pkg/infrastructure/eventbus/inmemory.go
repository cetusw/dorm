package eventbus

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"dorm/pkg/core/ports"
)

type InMemoryEventBus struct {
	handlers map[string][]ports.EventHandler
	mu       sync.Mutex
	ctx      context.Context
	cancel   context.CancelFunc
	active   sync.WaitGroup
	stopping bool
	done     chan struct{}
}

func NewInMemoryEventBus() *InMemoryEventBus {
	ctx, cancel := context.WithCancel(context.Background())
	return &InMemoryEventBus{
		ctx:      ctx,
		cancel:   cancel,
		done:     make(chan struct{}),
		handlers: make(map[string][]ports.EventHandler),
	}
}

func (b *InMemoryEventBus) Subscribe(topic string, handler ports.EventHandler) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.handlers[topic] = append(b.handlers[topic], handler)
}

func (b *InMemoryEventBus) register(topic string) ([]ports.EventHandler, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.stopping {
		return nil, fmt.Errorf("publish event %q: event bus is stopping", topic)
	}
	b.active.Add(1)
	return append([]ports.EventHandler(nil), b.handlers[topic]...), nil
}

func (b *InMemoryEventBus) Publish(ctx context.Context, topic string, event interface{}) error {
	handlers, err := b.register(topic)
	if err != nil {
		return err
	}
	defer b.active.Done()

	ctx, cancel := context.WithCancel(ctx)
	stopCancellation := context.AfterFunc(b.ctx, cancel)
	defer stopCancellation()
	defer cancel()

	b.deliver(ctx, topic, event, handlers)
	return nil
}

func (b *InMemoryEventBus) PublishAsync(topic string, event interface{}, timeout time.Duration) error {
	handlers, err := b.register(topic)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(b.ctx, timeout)
	go func() {
		defer b.active.Done()
		defer cancel()
		b.deliver(ctx, topic, event, handlers)
	}()
	return nil
}

func (b *InMemoryEventBus) deliver(ctx context.Context, topic string, event interface{}, handlers []ports.EventHandler) {
	for _, handler := range handlers {
		if err := handler(ctx, event); err != nil {
			if b.ctx.Err() != nil && errors.Is(err, context.Canceled) {
				continue
			}
			log.Printf("ERROR in EventBus handler [topic=%s]: %v", topic, err)
		}
	}
}

func (b *InMemoryEventBus) Stop(ctx context.Context) error {
	b.mu.Lock()
	if !b.stopping {
		b.stopping = true
		go func() {
			b.active.Wait()
			b.cancel()
			close(b.done)
		}()
	}
	b.mu.Unlock()

	select {
	case <-b.done:
		return nil
	case <-ctx.Done():
		b.cancel()
		return fmt.Errorf("drain event handlers: %w", ctx.Err())
	}
}
