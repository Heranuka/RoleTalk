package analytic

import (
	"context"
	"go-backend/internal/models/domain"

	"github.com/google/uuid"
)

// Service defines the business logic for retrieving and processing user analytics.
type Service interface {
	// GetUserSkills returns the current skill levels for the specified user.
	GetUserSkills(ctx context.Context, userID uuid.UUID) (*domain.UserSkill, error)
}
