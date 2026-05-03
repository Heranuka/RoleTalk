package minio

import (
	"context"
	"strings"

	"go-backend/internal/config"

	"go.uber.org/zap"
)

// FromConfig builds Storage from the MinIO application config.
//
// Credentials and publicEndpoint are required so presigned GET URLs resolve for clients outside the Docker network.
func FromConfig(ctx context.Context, cfg *config.MinIO, log *zap.SugaredLogger) (*Storage, error) {
	publicEndpoint := strings.TrimSpace(cfg.PublicEndpoint)
	if publicEndpoint == "" {
		publicEndpoint = "localhost:9000"
	}

	region := strings.TrimSpace(cfg.Region)
	if region == "" {
		region = "us-east-1"
	}

	return NewStorage(
		ctx,
		cfg.Endpoint,
		publicEndpoint,
		cfg.AccessKey,
		cfg.SecretKey,
		cfg.Bucket,
		cfg.UseSSL,
		region,
		log,
	)
}
