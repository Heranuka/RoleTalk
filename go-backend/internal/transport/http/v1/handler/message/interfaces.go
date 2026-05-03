package message

import (
	"context"
	"io"

	"go-backend/internal/models/domain"

	"github.com/google/uuid"
)

// Service defines the business logic for dialog orchestration and history retrieval.
type Service interface {
	// ProcessVoiceTurn handles the full AI loop: STT, LLM interaction, and TTS synthesis.
	// It returns the user's transcription, AI text reply, optional direct MinIO URL, and durable object key for API playback.
	ProcessVoiceTurn(
		ctx context.Context,
		userID, sessionID uuid.UUID,
		practiceLang string,
		audioData io.Reader,
	) (string, string, string, string, error)

	// StreamValidatedSessionAIWave streams a WAV artifact for the AI reply stored under objectKey within the user's session namespace.
	StreamValidatedSessionAIWave(ctx context.Context, userID, sessionID uuid.UUID, objectKey string) (io.ReadCloser, error)

	// GetSessionHistory retrieves all past messages for a specific practice session.
	GetSessionHistory(ctx context.Context, sessionID uuid.UUID) ([]*domain.Message, error)
}
