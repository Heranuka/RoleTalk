package rabbitmq

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

// SetupTopology declares exchanges, queues, and bindings for the application.
func (c *Client) SetupTopology(exchange, queue, dlx, dlq string) error {
	ch := c.Channel()

	// 1. DLX & DLQ
	if err := ch.ExchangeDeclare(dlx, "direct", true, false, false, false, nil); err != nil {
		return fmt.Errorf("failed to declare exchange: %w", err)
	}
	if _, err := ch.QueueDeclare(dlq, true, false, false, false, nil); err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}
	if err := ch.QueueBind(dlq, dlq, dlx, false, nil); err != nil {
		return fmt.Errorf("failed to bind queue: %w", err)
	}

	// 2. Main Exchange & Queue
	if err := ch.ExchangeDeclare(exchange, "direct", true, false, false, false, nil); err != nil {
		return fmt.Errorf("failed to declare exchange: %w", err)
	}

	args := amqp.Table{
		"x-dead-letter-exchange":    dlx,
		"x-dead-letter-routing-key": dlq,
	}
	if _, err := ch.QueueDeclare(queue, true, false, false, false, args); err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	if err := ch.QueueBind(queue, queue, exchange, false, nil); err != nil {
		return fmt.Errorf("failed to bind queue: %w", err)
	}
	return nil
}
