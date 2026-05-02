package topic

import (
	"go-backend/internal/models/domain"

	"github.com/google/uuid"
)

type createTopicRequest struct {
	Title           string `json:"title" validate:"required,min=3,max=100"`
	Description     string `json:"description" validate:"required,max=500"`
	Emoji           string `json:"emoji" validate:"required"`
	DifficultyLevel string `json:"difficulty_level" validate:"required,oneof=A1 A2 B1 B2 C1 C2 all"`

	// ДОБАВЬ ЭТИ ПОЛЯ:
	MyRole       string `json:"my_role" validate:"required"`
	PartnerRole  string `json:"partner_role" validate:"required"`
	PartnerEmoji string `json:"partner_emoji" validate:"required"`
	Goal         string `json:"goal" validate:"required"`
}

// Также обнови topicResponse, чтобы Флаттер видел эти данные в ответ
type topicResponse struct {
	ID              uuid.UUID  `json:"id"`
	AuthorID        *uuid.UUID `json:"author_id,omitempty"`
	Title           string     `json:"title"`
	Description     *string    `json:"description"`
	Emoji           *string    `json:"emoji"`
	DifficultyLevel *string    `json:"difficulty_level"`
	IsOfficial      bool       `json:"is_official"`
	LikesCount      int        `json:"likes_count"`

	// НОВЫЕ ПОЛЯ:
	MyRole       string `json:"my_role"`
	PartnerRole  string `json:"partner_role"`
	PartnerEmoji string `json:"partner_emoji"`
	Goal         string `json:"goal"`
}

// toTopicResponse maps a domain.Topic entity to a transport response DTO.
func toTopicResponse(t *domain.Topic) topicResponse {
	return topicResponse{
		ID:              t.ID,
		AuthorID:        t.AuthorID,
		Title:           t.Title,
		Description:     t.Description,
		Emoji:           t.Emoji,
		DifficultyLevel: t.DifficultyLevel,
		IsOfficial:      t.IsOfficial,
		LikesCount:      t.LikesCount,
	}
}

// toTopicListResponse maps a slice of domain topics to a slice of DTOs.
func toTopicListResponse(topics []*domain.Topic) []topicResponse {
	res := make([]topicResponse, len(topics))
	for i, t := range topics {
		res[i] = toTopicResponse(t)
	}
	return res
}
