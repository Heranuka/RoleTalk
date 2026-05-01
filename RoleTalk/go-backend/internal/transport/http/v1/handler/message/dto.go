package message

import (
	"time"

	"go-backend/internal/models/domain"

	"github.com/google/uuid"
)

// messageResponse represents a single dialog message.
type messageResponse struct {
	ID          uuid.UUID `json:"id"`
	SenderRole  string    `json:"role"`      // "user" or "ai"
	TextContent *string   `json:"text"`      // Pointer to allow nulls if needed
	AudioURL    *string   `json:"audio_url"` // Link to the voice file in S3/MinIO
	CreatedAt   time.Time `json:"created_at"`
}

// voiceTurnResponse is the combined response for a single voice interaction.
type voiceTurnResponse struct {
	UserText   string          `json:"user_text"`   // Result of STT (what user said)
	AIResponse messageResponse `json:"ai_response"` // Result of LLM + TTS (what AI replied)
}

// toMessageResponse maps a domain Message entity to the transport DTO.
func toMessageResponse(m *domain.Message) messageResponse {
	return messageResponse{
		ID:          m.ID,
		SenderRole:  string(m.SenderRole),
		TextContent: m.TextContent,
		AudioURL:    m.AudioURL, // Now correctly using AudioURL for voice
		CreatedAt:   m.CreatedAt,
	}
}

// toMessageListResponse converts multiple domain messages to a slice of DTOs.
func toMessageListResponse(msgs []*domain.Message) []messageResponse {
	res := make([]messageResponse, len(msgs))
	for i, m := range msgs {
		res[i] = toMessageResponse(m)
	}
	return res
}
