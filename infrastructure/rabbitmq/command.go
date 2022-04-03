package rabbitmq

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"github.com/hectorgabucio/taterubot-dc/kit/command"
	"github.com/streadway/amqp"
	"log"
	"time"
)

type CommandBus struct {
	Connection *amqp.Connection
	Channel    *amqp.Channel
}

func NewCommandBus(connectionURL string) (*CommandBus, error) {
	conn, ch, err := establishConnection(connectionURL)
	if err != nil {
		return nil, err
	}
	return &CommandBus{
		Connection: conn,
		Channel:    ch,
	}, nil
}

func (c *CommandBus) Close() {
	err := c.Connection.Close()
	if err != nil {
		log.Println("failed to close rabbitmq conn", err)
	}
}

func (c *CommandBus) Dispatch(ctx context.Context, command command.Command) error {
	var b bytes.Buffer

	gob.Register(command)
	if err := gob.NewEncoder(&b).Encode(&command); err != nil {
		return fmt.Errorf("err encoding command to buffer gob, %w", err)
	}
	err := c.Channel.Publish(EXCHANGE, string(command.Type()), false, false, amqp.Publishing{
		AppId:       APP_ID,
		ContentType: ENCODING_TYPE, // XXX: We will revisit this in future episodes
		Body:        b.Bytes(),
		Timestamp:   time.Now(),
	})
	if err != nil {
		return fmt.Errorf("err publishing command to topic, %w", err)
	}
	return nil
}

func (c *CommandBus) Register(t command.Type, handler command.Handler) {
	q, err := c.Channel.QueueDeclare(
		fmt.Sprintf("QUEUE-%s", t), // name
		false,                      // durable
		false,                      // delete when unused
		true,                       // exclusive
		false,                      // no-wait
		nil,                        // arguments
	)
	if err != nil {
		log.Fatal("err declaring queue", err)
	}
	err = c.Channel.QueueBind(
		q.Name,    // queue name
		string(t), // routing key
		EXCHANGE,  // exchange
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("channel.QueueBind %v", err)
	}
	msgs, err := c.Channel.Consume(
		q.Name,                            // queue
		fmt.Sprintf("%s-%s", CONSUMER, t), // consumer
		false,                             // auto-ack
		false,                             // exclusive
		false,                             // no-local
		false,                             // no-wait
		nil,                               // args
	)
	if err != nil {
		log.Fatalf("channel.Consume %v", err)
	}
	go handleCommands(msgs, handler)
}

func handleCommands(commands <-chan amqp.Delivery, handler command.Handler) {
	for delivery := range commands {
		var cmd command.Command
		buf := bytes.NewBuffer(delivery.Body)
		if err := gob.NewDecoder(buf).Decode(&cmd); err != nil {
			log.Printf("err decoding command gob, %v", err)
			err := delivery.Ack(false)
			if err != nil {
				log.Println("err acking failed decoding", err)
			}
			return
		}

		log.Println("handling command", cmd)
		// TODO had to do this because i use same command type for starting and ending recording, i should separate into different command types
		go func() {
			err := handler.Handle(context.Background(), cmd)
			if err != nil {
				log.Println("err goroutine handling command", err)
			}
		}()

		err := delivery.Ack(false)
		if err != nil {
			log.Println("err ack command", err)
		}
	}
}
