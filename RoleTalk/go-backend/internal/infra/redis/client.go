// Package redis provides Redis client and infrastructure.
package redis

import (
	"context"
	"fmt"
	"time"

	"go-backend/internal/config"

	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// Client wraps the redis.Client and adds logging.
type Client struct {
	*redis.Client
	log *zap.SugaredLogger
}

// New creates a new Redis client and checks the connection.
func New(ctx context.Context, cfg *config.Redis, log *zap.SugaredLogger) (*Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,

		PoolSize:     cfg.Pool.PoolSize,
		MinIdleConns: cfg.Pool.MinIdleConns,
		MaxRetries:   cfg.Pool.MaxRetries,

		DialTimeout:  cfg.DialTimeout,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	})

	// Интеграция с OpenTelemetry — критично для твоего стека!
	// Теперь каждый запрос к редису будет виден в твоих трейсах Tempo.
	if err := redisotel.InstrumentTracing(rdb); err != nil {
		log.Errorf("[REDIS] Failed to instrument tracing: %v", err)
	}

	// Опционально: Метрики для Prometheus
	if err := redisotel.InstrumentMetrics(rdb); err != nil {
		log.Errorf("[REDIS] Failed to instrument metrics: %v", err)
	}

	// Проверка соединения
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping redis: %w", err)
	}

	log.Info("[REDIS] Connected successfully")

	return &Client{
		Client: rdb,
		log:    log,
	}, nil
}

// Close корректно завершает работу
func (c *Client) Close() error {
	c.log.Info("[REDIS] Closing connection...")
	if err := c.Client.Close(); err != nil {
		return fmt.Errorf("failed to close redis connection: %w", err)
	}
	return nil
}
