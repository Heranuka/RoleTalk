package topic

import (
	"context"
	"go-backend/internal/models/domain"

	"github.com/google/uuid"
)

// Service defines the business logic operations for roleplay scenarios.
type Service interface {
	GetAIRecommended(ctx context.Context) ([]*domain.Topic, error)
	GetCommunityFeed(ctx context.Context, limit, offset int) ([]*domain.Topic, error)
	CreateTopic(ctx context.Context, userID uuid.UUID, title, desc, emoji, level string) (uuid.UUID, error)
	LikeTopic(ctx context.Context, userID, topicID uuid.UUID) error
	RemoveLike(ctx context.Context, userID, topicID uuid.UUID) error
	DeleteTopic(ctx context.Context, userID, topicID uuid.UUID) error
}
