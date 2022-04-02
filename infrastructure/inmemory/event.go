package inmemory

import (
	"context"
	"fmt"
	"github.com/hectorgabucio/taterubot-dc/kit/event"
)

// EventBus is an in-memory implementation of the event.Bus.
type EventBus struct {
	handlers map[event.Type][]event.Handler
}

// NewEventBus initializes a new EventBus.
func NewEventBus() *EventBus {
	return &EventBus{
		handlers: make(map[event.Type][]event.Handler),
	}
}

// Publish implements the event.Bus interface.
func (b *EventBus) Publish(ctx context.Context, events []event.Event) error {
	for _, evt := range events {
		handlers, ok := b.handlers[evt.Type()]
		if !ok {
			return nil
		}
		for _, handler := range handlers {
			err := handler.Handle(ctx, evt)
			if err != nil {
				return fmt.Errorf("err invoking handler for event, %w", err)
			}
		}
	}

	return nil
}

// Subscribe implements the event.Bus interface.
func (b *EventBus) Subscribe(evtType event.Type, handler event.Handler) {
	if _, ok := b.handlers[evtType]; !ok {
		b.handlers[evtType] = []event.Handler{}
	}
	b.handlers[evtType] = append(b.handlers[evtType], handler)
}
