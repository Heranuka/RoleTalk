package practice_session

import (
	"context"

	"go-backend/internal/models/domain"

	"github.com/google/uuid"
)

// Service defines the business logic for managing practice sessions.
type Service interface {
	StartSession(ctx context.Context, userID, topicID uuid.UUID) (uuid.UUID, error)
	CompleteSession(ctx context.Context, sessionID uuid.UUID) error
	GetSession(ctx context.Context, sessionID uuid.UUID) (*domain.PracticeSession, error)
}
