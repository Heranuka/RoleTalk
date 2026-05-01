package message

import (
	"context"
	"go-backend/internal/models/domain"
	"time"

	"github.com/google/uuid"
)

// Repository defines the data access layer for this service.
type Repository interface {
	Create(ctx context.Context, m *domain.Message) error
	GetBySessionID(ctx context.Context, sessionID uuid.UUID) ([]*domain.Message, error)
}

// StorageProvider defines the subset of MinIO methods needed by this service.
type StorageProvider interface {
	GetPresignedURL(ctx context.Context, objectKey string, expires time.Duration) (string, error)
}
