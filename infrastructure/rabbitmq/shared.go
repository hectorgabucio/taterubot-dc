package rabbitmq

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/streadway/amqp"
	"log"
)

const exchange = "taterubot-exchange"

const encodingType = "application/x-encoding-gob"
const appID = "taterubot-rabbit"
const consumer = "consumer-taterubot"

func establishConnection(connectionURL string) (*amqp.Connection, *amqp.Channel, error) {
	conn, err := amqp.Dial(connectionURL)
	if err != nil {
		return nil, nil, fmt.Errorf("amqp.Dial %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, nil, fmt.Errorf("conn.Channel %w", err)
	}

	err = ch.ExchangeDeclare(
		exchange,
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("ch.ExchangeDeclare %w", err)
	}

	return conn, ch, nil
}

func initializeChannelConsumption(channel *amqp.Channel, name string) <-chan amqp.Delivery {
	q, err := channel.QueueDeclare(
		fmt.Sprintf("QUEUE-%s", name),
		false,
		false,
		true,
		false,
		nil,
	)
	if err != nil {
		log.Fatal("err declaring queue", err)
	}
	err = channel.QueueBind(
		q.Name,
		name,
		exchange,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("channel.QueueBind %v", err)
	}

	msgs, err := channel.Consume(
		q.Name,
		fmt.Sprintf("%s-%s", consumer, name),
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("channel.Consume %v", err)
	}
	log.Println("initialized consuming messages on queue", name)
	return msgs
}

func encode(message interface{}) (bytes.Buffer, error) {
	var b bytes.Buffer
	gob.Register(message)
	if err := gob.NewEncoder(&b).Encode(&message); err != nil {
		return bytes.Buffer{}, fmt.Errorf("rabbitmq.encode, %w", err)
	}
	return b, nil
}

func decode(buf *bytes.Buffer, target interface{}) error {
	if err := gob.NewDecoder(buf).Decode(target); err != nil {
		return fmt.Errorf("rabbitmq.decode, %w", err)
	}
	return nil
}
