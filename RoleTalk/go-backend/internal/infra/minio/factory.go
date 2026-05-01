package minio

import (
	"context"

	"go-backend/internal/config"

	"go.uber.org/zap"
)

// FromConfig creates Storage from config.MinIO section.
func FromConfig(ctx context.Context, cfg *config.MinIO, log *zap.SugaredLogger) (*Storage, error) {
	return NewStorage(
		ctx,
		cfg.Endpoint,
		cfg.AccessKey,
		cfg.SecretKey,
		cfg.Bucket,
		cfg.UseSSL,
		log,
	)
}
