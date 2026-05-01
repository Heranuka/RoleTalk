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
	_ context.Context,
	cfg *config.Config,
	log *zap.SugaredLogger,
	pool *pgxpool.Pool,
	storage *minio.Storage,
) (http.Handler, error) {
	h, err := health.New(health.WithComponent(health.Component{
		Name:    cfg.App.Name,
		Version: cfg.App.Version,
	}))
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
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, cfg.API.URL+"/health", nil)
			if err != nil {
				return fmt.Errorf("failed to create request: %w", err)
			}
			resp, err := client.Do(req)
			if err != nil {
				return fmt.Errorf("ai service unreachable: %w", err)
			}
			defer func() { _ = resp.Body.Close() }()
			if resp.StatusCode != http.StatusOK {
				return fmt.Errorf("ai service status: %d", resp.StatusCode)
			}
			return nil
		},
		Timeout: 5 * time.Second,
	}); err != nil {
		log.Errorw("failed to register ai_service health check", "error", err)
	}

	// 4. Ollama Check (LLM Provider)
	if err := h.Register(health.Config{
		Name: "ollama_llm",
		Check: func(ctx context.Context) error {
			client := http.Client{Timeout: 2 * time.Second}
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, cfg.Ollama.URL, nil)
			if err != nil {
				return fmt.Errorf("failed to create request: %w", err)
			}
			resp, err := client.Do(req)
			if err != nil {
				return fmt.Errorf("ollama unreachable: %w", err)
			}
			defer func() { _ = resp.Body.Close() }()
			return nil
		},
		Timeout: 5 * time.Second,
	}); err != nil {
		log.Errorw("failed to register ollama health check", "error", err)
	}

	return h.Handler(), nil
}
