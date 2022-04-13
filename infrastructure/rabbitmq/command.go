package rabbitmq

import (
	"bytes"
	"context"
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

func (c *CommandBus) Close() error {
	if err := c.Connection.Close(); err != nil {
		log.Println("failed to close rabbitmq conn", err)
		return fmt.Errorf("rabbit.close: %w", err)
	}
	return nil
}

func (c *CommandBus) Dispatch(ctx context.Context, command command.Command) error {
	b, err := encode(command)
	if err != nil {
		return fmt.Errorf("err encoding command, %w", err)
	}
	if err := c.Channel.Publish(exchange, string(command.Type()), false, false, amqp.Publishing{
		AppId:       appID,
		ContentType: encodingType,
		Body:        b.Bytes(),
		Timestamp:   time.Now(),
	}); err != nil {
		return fmt.Errorf("err publishing command to topic, %w", err)
	}
	return nil
}

func (c *CommandBus) Register(t command.Type, handler command.Handler) {
	msgs := initializeChannelConsumption(c.Channel, string(t))
	go handleCommands(msgs, handler)
}

func handleCommands(commands <-chan amqp.Delivery, handler command.Handler) {
	for delivery := range commands {
		var cmd command.Command
		buf := bytes.NewBuffer(delivery.Body)
		if err := decode(buf, &cmd); err != nil {
			log.Printf("err decoding command, %v", err)
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
