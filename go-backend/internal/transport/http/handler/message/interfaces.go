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
	// It returns the user's transcribed text, the AI's text response, and the URL to the generated audio.
	ProcessVoiceTurn(
		ctx context.Context,
		userID uuid.UUID,
		sessionID uuid.UUID,
		practiceLang string,
		audioData io.Reader,
	) (string, string, string, error)

	// GetSessionHistory retrieves all past messages for a specific practice session.
	GetSessionHistory(ctx context.Context, sessionID uuid.UUID) ([]*domain.Message, error)
}
