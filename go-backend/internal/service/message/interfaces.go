package message

import (
	"context"
	"go-backend/internal/models/domain"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, m *domain.Message) error
	GetBySessionID(ctx context.Context, sessionID uuid.UUID) ([]*domain.Message, error)
}
