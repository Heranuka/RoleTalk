// Package health provides health check configuration for the application infrastructure.
package health

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/hellofresh/health-go/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"go-backend/internal/config"
	"go-backend/internal/infra/minio"
)

// NewHandler initializes system health checks and returns an http.Handler.
func NewHandler(
	ctx context.Context,
	cfg *config.Config,
	log *zap.SugaredLogger,
	pool *pgxpool.Pool,
	storage *minio.Storage,
) (http.Handler, error) {
	h, err := health.New()
	if err != nil {
		return nil, fmt.Errorf("health.New: %w", err)
	}

	// 1. PostgreSQL Check
	if err := h.Register(health.Config{
		Name: "postgres",
		Check: func(ctx context.Context) error {
			return pool.Ping(ctx)
		},
		Timeout: 5 * time.Second,
	}); err != nil {
		log.Errorw("failed to register postgres health check", "error", err)
	}

	// 2. MinIO Check
	if err := h.Register(health.Config{
		Name: "storage_minio",
		Check: func(ctx context.Context) error {
			return storage.HealthCheck(ctx)
		},
		Timeout: 5 * time.Second,
	}); err != nil {
		log.Errorw("failed to register minio health check", "error", err)
	}

	// 3. AI Service Check
	if err := h.Register(health.Config{
		Name: "ai_service",
		Check: func(ctx context.Context) error {
			client := http.Client{Timeout: 2 * time.Second}
			resp, err := client.Get(cfg.API.URL + "/health")
			if err != nil {
				return fmt.Errorf("ai service unreachable: %w", err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				return fmt.Errorf("ai service status: %d", resp.StatusCode)
			}
			return nil
		},
		Timeout: 5 * time.Second,
	}); err != nil {
		log.Errorw("failed to register ai_service health check", "error", err)
	}

	return h.Handler(), nil
}
