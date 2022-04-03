package rabbitmq

import (
	"fmt"
	"github.com/streadway/amqp"
)

const exchange = "taterubot-exchange"

const encodingType = "application/x-encoding-gob"
const appId = "taterubot-rabbit"
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
		exchange, // name
		"topic",  // type
		true,     // durable
		false,    // auto-deleted
		false,    // internal
		false,    // no-wait
		nil,      // arguments
	)
	if err != nil {
		return nil, nil, fmt.Errorf("ch.ExchangeDeclare %w", err)
	}

	return conn, ch, nil
}
