package rabbitmq

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

// Consumer handles message consumption from a specific queue.
type Consumer struct {
	client   *Client
	queue    string
	log      *zap.SugaredLogger
	prefetch int
}

// NewConsumer creates a new RabbitMQ consumer instance.
func NewConsumer(client *Client, queue string, prefetch int, log *zap.SugaredLogger) *Consumer {
	return &Consumer{client: client, queue: queue, prefetch: prefetch, log: log}
}

// Start begins consuming messages and passes them to the handler.
func (c *Consumer) Start(ctx context.Context, handler Handler) error {
	ch := c.client.Channel()
	if err := ch.Qos(c.prefetch, 0, false); err != nil {
		return fmt.Errorf("create qos: %w", err)
	}

	msgs, err := ch.Consume(c.queue, "", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("create consumer: %w", err)
	}

	go func() {
		for d := range msgs {
			// Логика обработки
			if err := handler(ctx, d.Body, d.Headers); err != nil {
				c.log.Errorf("[RABBIT] Handler error: %v", err)
				_ = d.Nack(false, false) // В DLQ
			} else {
				_ = d.Ack(false)
			}
		}
	}()

	return nil
}
