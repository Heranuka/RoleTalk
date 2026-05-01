//go:generate mockgen -package=user -destination=mocks.go go-backend/internal/transport/http/handler/user Service
package user

import (
	"context"

	"github.com/google/uuid"

	"go-backend/internal/models/domain"
	serviceuser "go-backend/internal/service/user"
)

// Service defines the interface that HTTP handlers use to perform user operations.
type Service interface {
	// GetByID retrieves user details by unique identifier.
	GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)

	// Update modifies an existing user's details.
	Update(ctx context.Context, input serviceuser.UpdateInput) error
}
