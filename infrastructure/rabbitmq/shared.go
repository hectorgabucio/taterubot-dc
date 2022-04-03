package rabbitmq

import (
	"fmt"
	"github.com/streadway/amqp"
)

const EXCHANGE = "taterubot-exchange"

const ENCODING_TYPE = "application/x-encoding-gob"
const APP_ID = "taterubot-rabbit"
const CONSUMER = "consumer-taterubot"

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
		EXCHANGE, // name
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
