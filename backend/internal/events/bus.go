// Package events is the in-process pub/sub
package events

import (
	"context"
	"errors"
	"sync"

	"github.com/google/uuid"
)

// Type selects which handlers run
// Payload is whatever that event type's contract says it is
// TenantID travels alongside every event because everything in this system is tenant-scoped
type Event struct {
	Type     string
	TenantID uuid.UUID
	Payload  any
}

type Handler func(ctx context.Context, e Event) error

type Bus struct {
	mu       sync.RWMutex
	handlers map[string][]Handler
}

func NewBus() *Bus {
	return &Bus{handlers: make(map[string][]Handler)}
}

func (b *Bus) Subscribe(eventType string, h Handler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers[eventType] = append(b.handlers[eventType], h)
}

func (b *Bus) Publish(ctx context.Context, e Event) error {
	b.mu.RLock()
	handlers := append([]Handler(nil), b.handlers[e.Type]...)
	b.mu.RUnlock()

	var errs []error
	for _, h := range handlers {
		if err := h(ctx, e); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}
