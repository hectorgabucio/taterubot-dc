package rabbitmq

import (
	"bytes"
	"context"
	"encoding/gob"
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

func (b *EventBus) Close() {
	err := b.Connection.Close()
	if err != nil {
		log.Println("failed to close rabbitmq conn", err)
	}
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
		var buf bytes.Buffer
		gob.Register(events[i])
		if err := gob.NewEncoder(&buf).Encode(&events[i]); err != nil {
			return fmt.Errorf("err encoding event to buffer gob, %w", err)
		}
		err := b.Channel.Publish(EXCHANGE, string(events[i].Type()), false, false, amqp.Publishing{
			AppId:       APP_ID,
			ContentType: ENCODING_TYPE, // XXX: We will revisit this in future episodes
			Body:        buf.Bytes(),
			Timestamp:   time.Now(),
		})
		if err != nil {
			return fmt.Errorf("err publishing event to topic, %w", err)
		}
	}
	return nil
}

// Subscribe implements the event.Bus interface.
func (b *EventBus) Subscribe(evtType event.Type, handler event.Handler) {
	q, err := b.Channel.QueueDeclare(
		fmt.Sprintf("QUEUE-%s", evtType), // name
		false,                            // durable
		false,                            // delete when unused
		true,                             // exclusive
		false,                            // no-wait
		nil,                              // arguments
	)
	if err != nil {
		log.Fatal("err declaring queue", err)
	}
	err = b.Channel.QueueBind(
		q.Name,          // queue name
		string(evtType), // routing key
		EXCHANGE,        // exchange
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("channel.QueueBind %v", err)
	}
	msgs, err := b.Channel.Consume(
		q.Name,                                  // queue
		fmt.Sprintf("%s-%s", CONSUMER, evtType), // consumer
		false,                                   // auto-ack
		false,                                   // exclusive
		false,                                   // no-local
		false,                                   // no-wait
		nil,                                     // args
	)
	if err != nil {
		log.Fatalf("channel.Consume %v", err)
	}
	go handleEvents(msgs, handler)
}

func handleEvents(events <-chan amqp.Delivery, handler event.Handler) {
	for delivery := range events {
		var evt event.Event
		buf := bytes.NewBuffer(delivery.Body)
		if err := gob.NewDecoder(buf).Decode(&evt); err != nil {
			log.Printf("err decoding evt gob, %v", err)
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
