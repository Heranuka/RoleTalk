// Package rabbitmq provides RabbitMQ client and infrastructure.
package rabbitmq

import (
	"fmt"
	"go-backend/internal/config"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

// Client manages the RabbitMQ connection and channel with auto-reconnect logic.
type Client struct {
	conn   *amqp.Connection
	cfg    *config.RabbitMQ
	ch     *amqp.Channel
	mu     sync.RWMutex
	url    string
	log    *zap.SugaredLogger
	notify chan *amqp.Error
	done   chan bool
}

// NewClient initializes a new RabbitMQ client and starts the reconnection handler.
func NewClient(cfg *config.RabbitMQ, log *zap.SugaredLogger) (*Client, error) {
	url := fmt.Sprintf("amqp://%s:%s@%s:%d/", cfg.User, cfg.Password, cfg.Host, cfg.Port)
	c := &Client{
		url:  url,
		cfg:  cfg,
		log:  log,
		done: make(chan bool),
	}

	if err := c.connect(); err != nil {
		return nil, err
	}

	go c.handleReconnect()
	return c, nil
}

func (c *Client) connect() error {
	conn, err := amqp.Dial(c.url)
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}
	ch, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to create channel: %w", err)
	}

	c.mu.Lock()
	c.conn = conn
	c.ch = ch
	c.notify = conn.NotifyClose(make(chan *amqp.Error))
	c.mu.Unlock()

	return nil
}

func (c *Client) handleReconnect() {
	for {
		select {
		case <-c.done:
			return
		case err := <-c.notify:
			if err != nil {
				c.log.Errorf("[RABBIT] Connection lost: %v. Retrying...", err)
				for {
					time.Sleep(5 * time.Second)
					if err := c.connect(); err == nil {
						c.log.Info("[RABBIT] Reconnected successfully")
						break
					}
				}
			}
		}
	}
}

// Channel returns the current active RabbitMQ channel.
func (c *Client) Channel() *amqp.Channel {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.ch
}

// Stop gracefully closes the channel and connection.
func (c *Client) Stop() error {
	close(c.done)
	if c.ch != nil {
		_ = c.ch.Close()
	}
	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			return fmt.Errorf("close rabbitmq connection: %w", err)
		}
	}
	return nil
}
