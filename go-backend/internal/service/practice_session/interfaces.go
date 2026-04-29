package practice

import (
	"context"
	"go-backend/internal/models/domain"

	"github.com/google/uuid"
)

// Repository defines the data store operations for practice sessions.
type Repository interface {
	Create(ctx context.Context, s *domain.PracticeSession) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.PracticeSession, error)
	GetActiveByUserID(ctx context.Context, userID uuid.UUID) (*domain.PracticeSession, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) error
}

// TopicRepository is required to validate that a scenario exists before starting a session.
type TopicRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Topic, error)
}
