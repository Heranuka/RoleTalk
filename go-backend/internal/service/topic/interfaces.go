package topic

import (
	"context"
	"go-backend/internal/models/domain"

	"github.com/google/uuid"
)

// Repository defines the data access layer for this service.
type Repository interface {
	Create(ctx context.Context, t *domain.Topic) (uuid.UUID, error)
	GetOfficial(ctx context.Context) ([]*domain.Topic, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Topic, error)
	GetCommunity(ctx context.Context, limit, offset int) ([]*domain.Topic, error)
	AddLike(ctx context.Context, userID, topicID uuid.UUID) error
	RemoveLike(ctx context.Context, userID, topicID uuid.UUID) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// Transactor defines the transaction management interface.
type Transactor interface {
	WithinTx(ctx context.Context, fn func(ctx context.Context) error) error
}
