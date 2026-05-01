// Package ai defines internal interfaces for the AI orchestration layer.
package ai

import (
	"context"
	"io"

	"go-backend/internal/models/domain"

	"github.com/google/uuid"
)

// StorageProvider defines the behavior for persistent binary data storage (S3/MinIO).
type StorageProvider interface {
	// Upload stores an object and returns its public or presigned URL.
	Upload(ctx context.Context, bucket, filename string, src io.Reader) (string, error)
}

// MessageService defines the behavior for persisting and retrieving dialog history.
type MessageService interface {
	// SaveMessage persists a single turn in the database.
	SaveMessage(ctx context.Context, sessionID uuid.UUID, role domain.MessageRole, content, audioURL string) error
	// GetSessionHistory retrieves all messages for a given session, ordered by time.
	GetSessionHistory(ctx context.Context, sessionID uuid.UUID) ([]*domain.Message, error)
}

// TopicRepository provides access to roleplay scenario metadata.
type TopicRepository interface {
	// GetByID retrieves a topic by its unique identifier.
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Topic, error)
}

// PracticeSessionRepository provides access to active or archived user sessions.
type PracticeSessionRepository interface {
	// GetByID retrieves session details, including current status and linked topic.
	GetByID(ctx context.Context, id uuid.UUID) (*domain.PracticeSession, error)
}

// PromptService encapsulates the logic for generating LLM system instructions.
type PromptService interface {
	// RenderRoleplay generates the character system prompt based on topic parameters.
	RenderRoleplay(params domain.RoleplayParams) (string, error)
	// RenderEvaluation generates instructions for the AI to analyze the user's performance.
	RenderEvaluation(params domain.EvaluationParams) (string, error)
}

// UserRepo defines the behavior for retrieving user-specific data.
type UserRepo interface {
	// GetByID returns the domain User model.
	GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
}

// Provider defines the interface for external AI processing (STT/LLM/TTS).
type Provider interface {
	ProcessVoiceTurn(ctx context.Context, audio []byte, lang string, systemPrompt string) (string, string, []byte, error)
}
