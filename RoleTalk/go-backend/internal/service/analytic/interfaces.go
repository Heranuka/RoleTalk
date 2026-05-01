package analytic

import (
	"context"

	"go-backend/internal/models/domain"

	"github.com/google/uuid"
)

// Repository defines the expected behavior for persisting analytic data.
type Repository interface {
	Create(ctx context.Context, userID uuid.UUID) error
	GetByUserID(ctx context.Context, userID uuid.UUID) (*domain.UserSkill, error)
	UpdateSkills(ctx context.Context, s *domain.UserSkill) error
}

type PracticeSessionRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*domain.PracticeSession, error)
}

type TopicRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Topic, error)
}

// MessageService defines the behavior for persisting dialog history.
type MessageService interface {
	GetSessionHistory(ctx context.Context, sessionID uuid.UUID) ([]*domain.Message, error)
}

type OllamaClient interface {
	AnalyzeTranscript(ctx context.Context, prompt string) (map[string]int, error)
}

// Transactor defines the interface for database transaction management.
type Transactor interface {
	WithinTx(ctx context.Context, fn func(ctx context.Context) error) error
}
