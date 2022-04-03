package event

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Bus defines the expected behaviour from an event bus.
type Bus interface {
	// Publish is the method used to publish new events.
	Publish(context.Context, []Event) error
	// Subscribe is the method used to subscribe new event handlers.
	Subscribe(Type, Handler)
	Close()
}

// Handler defines the expected behaviour from an event handler.
type Handler interface {
	Handle(context.Context, Event) error
}

// Type represents a domain event type.
type Type string

// Event represents a domain command.
type Event interface {
	ID() string
	AggregateID() string
	OccurredOn() time.Time
	Type() Type
}

type BaseEvent struct {
	MEventID     string
	MAggregateID string
	MOccurredOn  time.Time
}

func NewBaseEvent(aggregateID string) BaseEvent {
	return BaseEvent{
		MEventID:     uuid.New().String(),
		MAggregateID: aggregateID,
		MOccurredOn:  time.Now(),
	}
}

func (b BaseEvent) ID() string {
	return b.MEventID
}

func (b BaseEvent) OccurredOn() time.Time {
	return b.MOccurredOn
}

func (b BaseEvent) AggregateID() string {
	return b.MAggregateID
}
