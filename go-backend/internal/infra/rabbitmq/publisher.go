package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel"
)

// Publisher handles message publishing to RabbitMQ with tracing.
type Publisher struct {
	client   *Client
	exchange string
}

// NewPublisher creates a new message publisher.
func NewPublisher(client *Client, exchange string) *Publisher {
	return &Publisher{client: client, exchange: exchange}
}

// Publish serializes a payload and sends it to the specified exchange.
func (p *Publisher) Publish(ctx context.Context, routingKey string, payload interface{}) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	// Inject trace context into headers
	headers := amqp.Table{}
	otel.GetTextMapPropagator().Inject(ctx, &amqpHeaderCarrier{headers})

	if err := p.client.Channel().PublishWithContext(ctx,
		p.exchange,
		routingKey,
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Body:         body,
			Headers:      headers,
		},
	); err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}
	return nil
}

// amqpHeaderCarrier is a helper for OpenTelemetry context injection.
type amqpHeaderCarrier struct {
	headers amqp.Table
}

func (c *amqpHeaderCarrier) Get(key string) string {
	if v, ok := c.headers[key]; ok {
		s, _ := v.(string) //nolint:forcetypeassert
		return s
	}
	return ""
}

func (c *amqpHeaderCarrier) Set(key string, value string) {
	c.headers[key] = value
}

func (c *amqpHeaderCarrier) Keys() []string {
	keys := make([]string, 0, len(c.headers))
	for k := range c.headers {
		keys = append(keys, k)
	}
	return keys
}
