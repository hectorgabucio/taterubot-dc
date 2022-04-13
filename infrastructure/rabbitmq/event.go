package rabbitmq

import (
	"bytes"
	"context"
	"fmt"
	"github.com/hectorgabucio/taterubot-dc/kit/event"
	"github.com/streadway/amqp"
	"log"
	"time"
)

// EventBus is an in-memory implementation of the event.Bus.
type EventBus struct {
	Connection *amqp.Connection
	Channel    *amqp.Channel
}

func (b *EventBus) Close() error {
	if err := b.Connection.Close(); err != nil {
		log.Println("failed to close rabbitmq conn", err)
		return fmt.Errorf("rabbit.close: %w", err)
	}
	return nil
}

// NewEventBus initializes a new EventBus.
func NewEventBus(connectionURL string) (*EventBus, error) {
	conn, ch, err := establishConnection(connectionURL)
	if err != nil {
		return nil, err
	}
	return &EventBus{
		Connection: conn,
		Channel:    ch,
	}, nil
}

// Publish implements the event.Bus interface.
func (b *EventBus) Publish(ctx context.Context, events []event.Event) error {
	// TODO if some event fails, try to publish rest of events instead of returning
	for i := range events {
		buf, err := encode(events[i])
		if err != nil {
			return fmt.Errorf("err encoding event, %w", err)
		}
		if err := b.Channel.Publish(exchange, string(events[i].Type()), false, false, amqp.Publishing{
			AppId:       appID,
			ContentType: encodingType,
			Body:        buf.Bytes(),
			Timestamp:   time.Now(),
		}); err != nil {
			return fmt.Errorf("err publishing event to topic, %w", err)
		}
	}
	return nil
}

// Subscribe implements the event.Bus interface.
func (b *EventBus) Subscribe(evtType event.Type, handler event.Handler) {
	msgs := initializeChannelConsumption(b.Channel, string(evtType))
	go handleEvents(msgs, handler)
}

func handleEvents(events <-chan amqp.Delivery, handler event.Handler) {
	for delivery := range events {
		var evt event.Event
		buf := bytes.NewBuffer(delivery.Body)
		if err := decode(buf, &evt); err != nil {
			log.Printf("err decoding evt, %v", err)
			err := delivery.Ack(false)
			if err != nil {
				log.Println("err acking failed decoding", err)
			}
			return
		}

		log.Println("handling event", evt)
		go func() {
			err := handler.Handle(context.Background(), evt)
			if err != nil {
				log.Println("err goroutine handling event", err)
			}
		}()

		err := delivery.Ack(false)
		if err != nil {
			log.Println("err ack event", err)
		}
	}
}
