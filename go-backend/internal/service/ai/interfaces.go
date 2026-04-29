package ai

import (
	"context"
	"go-backend/internal/models/domain"
	"io"

	"github.com/google/uuid"
)

// StorageProvider defines the behavior for uploading audio files to S3/MinIO.
type StorageProvider interface {
	Upload(ctx context.Context, subdir, filename string, src io.Reader) (string, error)
}

// MessageService defines the behavior for persisting dialog history.
type MessageService interface {
	SaveMessage(ctx context.Context, sessionID uuid.UUID, role, content, audioURL string) error
	GetSessionHistory(ctx context.Context, sessionID uuid.UUID) ([]*domain.Message, error)
}

// UserRepo defines the behavior for retrieving user language preferences.
type UserRepo interface {
	GetByID(ctx context.Context, id uuid.UUID) (interface{}, error) // simplified for example
}
