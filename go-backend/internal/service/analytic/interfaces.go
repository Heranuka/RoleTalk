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

// PracticeSessionRepository defines data access for roleplay sessions.
type PracticeSessionRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*domain.PracticeSession, error)
}

// TopicRepository defines data access for roleplay scenarios.
type TopicRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Topic, error)
}

// MessageService defines the behavior for persisting dialog history.
type MessageService interface {
	GetSessionHistory(ctx context.Context, sessionID uuid.UUID) ([]*domain.Message, error)
}

// Engine defines the prompt rendering logic.
type Engine interface {
	RenderEvaluation(params domain.EvaluationParams) (string, error)
}

// OllamaClient defines the LLM analysis interface.
type OllamaClient interface {
	AnalyzeTranscript(ctx context.Context, prompt string) (map[string]int, error)
}

// Transactor defines the interface for database transaction management.
type Transactor interface {
	WithinTx(ctx context.Context, fn func(ctx context.Context) error) error
}
