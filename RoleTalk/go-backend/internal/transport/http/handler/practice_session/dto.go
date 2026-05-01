package practice_session

import (
	"time"

	"go-backend/internal/models/domain"

	"github.com/google/uuid"
)

// startSessionRequest defines the payload required to begin a new roleplay.
type startSessionRequest struct {
	TopicID uuid.UUID `json:"topic_id" validate:"required"`
}

// sessionResponse represents the metadata of a practice session.
type sessionResponse struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	TopicID   uuid.UUID `json:"topic_id"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// toSessionResponse maps the domain session entity to a transport DTO.
func toSessionResponse(s *domain.PracticeSession) sessionResponse {
	return sessionResponse{
		ID:     s.ID,
		UserID: s.UserID,

		CreatedAt: s.CreatedAt,
		UpdatedAt: s.UpdatedAt,
	}
}
